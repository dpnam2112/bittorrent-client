// Scripts to experiment calling Bittorrent tracker via UDP.
// Here is an example of calling this script twice:
//➜  bittorrent-client git:(main) ✗ go run scripts/udp_connect.go
//
// Resolved UDP address: 93.158.213.92:1337
// Raw response: 00000000e7c5323eea06ea83b9e43785
// Parsed Response Fields:
//   [0–3]   Action        : 0
//   [4–7]   Transaction ID: 3888460350
//   [8–15]  Connection ID : 0xea06ea83b9e43785
// The received response is valid.
// ➜  bittorrent-client git:(main) ✗ go run scripts/udp_connect.go
// Resolved UDP address: 93.158.213.92:1337
// Raw response: 00000000f9ad9043ea06ea83b9e43785
// Parsed Response Fields:
//   [0–3]   Action        : 0
//   [4–7]   Transaction ID: 4188901443
//   [8–15]  Connection ID : 0xea06ea83b9e43785
// The received response is valid.
//
// Purpose of each fields:
// - connection ID: Represents a connection between the client and the tracker. In a connection,
// multiple operations (identified by transaction ID field) can be performed.
//
// - transaction ID: Represents a transaction/operation. Due to the unreliability of the UDP, this
// field can be used to validate the tracker's response to the request (tracker has to respond with
// the same transaction ID as the transaction ID of the request made by the client.
//
// As the above demonstration shows, two consecutive runs result in the same connection ID.
// According to Bittorrent spec: "A connection ID can be used for multiple requests. A client can
// use a connection ID until one minute after it has received it. Trackers should accept the
// connection ID until two minutes after it has been send."

package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"time"
)

const (
	TRACKER_UDP_PROTOCOL_ID uint64 = 0x41727101980 // constant to identify the protocol to be used, i.e, UDP in this case)
	TRACKER_ACTION_CONNECT  uint32 = 0x0
	TRACKER_ACTION_ANNOUNCE uint32 = 0x1
)

func genTransactionID() uint32 {
	return rand.Uint32()
}

func main() {
	trackerURL := "udp://tracker.opentrackr.org:1337"

	parsedURL, err := url.Parse(trackerURL)
	if err != nil {
		fmt.Println("Failed to parse the URI:", err)
		return
	}

	udpRAddr, _ := net.ResolveUDPAddr("udp", parsedURL.Host)
	fmt.Println("Resolved UDP address:", udpRAddr)
	conn, err := net.DialUDP("udp", nil, udpRAddr)

	if err != nil {
		fmt.Println("Failed to initialize an UDP connection:", err)
		return
	}

	defer conn.Close()

	txID := genTransactionID()

	// UDP request format (https://www.bittorrent.org/beps/bep_0015.html):
	// Offset  Size            Name            Value
	// 0       64-bit integer  protocol_id     0x41727101980 // magic constant
	// 8       32-bit integer  action          0 // connect
	// 12      32-bit integer  transaction_id
	// 16
	req := make([]byte, 16)
	binary.BigEndian.PutUint64(req[0:8], TRACKER_UDP_PROTOCOL_ID)
	binary.BigEndian.PutUint32(req[8:12], TRACKER_ACTION_CONNECT)
	binary.BigEndian.PutUint32(req[12:16], txID)

	_, err = conn.Write(req)
	if err != nil {
		fmt.Println("Failed to send data over the UDP connection:", err)
		return
	}

	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Read the response
	// Response format
	// Offset  Size            Name            Value
	// 0       32-bit integer  action          0 // connect
	// 4       32-bit integer  transaction_id
	// 8       64-bit integer  connection_id
	// 16

	// Read the response
	resp := make([]byte, 16)
	_, err = conn.Read(resp)
	if err != nil {
		fmt.Println("Failed to read response: %v", err)
		return
	}

	// Log raw response bytes in hex
	fmt.Printf("Raw response: %x\n", resp)

	// Parse fields
	respAction := binary.BigEndian.Uint32(resp[0:4])
	respTxID := binary.BigEndian.Uint32(resp[4:8])
	respConnectionID := binary.BigEndian.Uint64(resp[8:16])

	// Log parsed values with offsets
	fmt.Printf("Parsed Response Fields:\n")
	fmt.Printf("  [0–3]   Action        : %d\n", respAction)
	fmt.Printf("  [4–7]   Transaction ID: %d\n", respTxID)
	fmt.Printf("  [8–15]  Connection ID : 0x%x\n", respConnectionID)

	// Validate
	if respAction != TRACKER_ACTION_CONNECT || respTxID != txID {
		fmt.Println("Invalid response: action=%d txnID=%d (expected %d)", respAction, respTxID, txID)
	} else {
		fmt.Println("The received response is valid.")
	}
}
