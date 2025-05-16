package trackerclient

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"net"
)

type TrackerAction int32

// Tracker action is a 32-bit unsigned integer representing actions that the downloader wants to
// make when interacting with a tracker.
const (
	TrackerActionConnect  TrackerAction = 0x00
	TrackerActionAnnounce TrackerAction = 0x01
)

type AnnounceEvent uint32

const (
	AnnounceEventNone      AnnounceEvent = 0
	AnnounceEventCompleted AnnounceEvent = 1
	AnnounceEventStarted   AnnounceEvent = 2
	AnnounceEventStopped   AnnounceEvent = 3
)

const (
	UDPAnnounceRequestSize = 98
	UDPConnectRequestSize = 16
	UDPConnectResponseSize = 16
)

type PeerAddr struct {
	IP   net.IP
	Port uint16
}

type TrackerUDPAnnounceRequest struct {
	ConnectionID int64
	TxnID        int32
	InfoHash     [20]byte
	PeerID       [20]byte
	Downloaded   int32
	Uploaded     int32
	Left         int32
	Event        AnnounceEvent
	IPAddr       *net.IP
	Port         uint16
	Key          int32
	NumWant      int32
}

// Schema for the announce response returned by a tracker.
// Leechers is the number of leechers currently downloading the file.
// Seeders is the number of seeders currently seeding the file
// TxnID is the transaction ID.
// PeerAddresses is the list of peer addresses requested by the client
type TrackerUDPAnnounceResponse struct {
	TxnID         int32
	Interval      int32
	Leechers      int32
	Seeders       int32
	PeerAddresses []PeerAddr
}

type TrackerUDPErrorResponse struct {
	TxnID   int32
	message string
}

// Serialize request to bytes
// Format:
// IPv4 announce request:
//
// Offset  Size    Name    Value
// 0       64-bit integer  connection_id
// 8       32-bit integer  action          1 // announce
// 12      32-bit integer  transaction_id
// 16      20-byte string  info_hash
// 36      20-byte string  peer_id
// 56      64-bit integer  downloaded
// 64      64-bit integer  left
// 72      64-bit integer  uploaded
// 80      32-bit integer  event           0 // 0: none; 1: completed; 2: started; 3: stopped
// 84      32-bit integer  IP address      0 // default
// 88      32-bit integer  key
// 92      32-bit integer  num_want        -1 // default
// 96      16-bit integer  port
// 98
func (req TrackerUDPAnnounceRequest) Marshal() []byte {
	serializedReq := make([]byte, UDPAnnounceRequestSize)

	binary.BigEndian.PutUint64(serializedReq[0:8], uint64(req.ConnectionID))
	binary.BigEndian.PutUint32(serializedReq[8:12], uint32(req.GetActionCode()))
	binary.BigEndian.PutUint32(serializedReq[12:16], uint32(req.TxnID))
	copy(serializedReq[16:36], req.InfoHash[:])
	copy(serializedReq[36:56], req.PeerID[:])
	binary.BigEndian.PutUint64(serializedReq[56:64], uint64(req.Downloaded))
	binary.BigEndian.PutUint64(serializedReq[64:72], uint64(req.Left))
	binary.BigEndian.PutUint64(serializedReq[72:80], uint64(req.Uploaded))
	binary.BigEndian.PutUint32(serializedReq[80:84], uint32(req.Event))

	if req.IPAddr == nil {
		binary.BigEndian.PutUint32(serializedReq[84:88], uint32(0))
	} else {
		rawIpv4 := req.IPAddr.To4()
		copy(serializedReq[84:88], rawIpv4)
	}

	binary.BigEndian.PutUint32(serializedReq[88:92], 0)
	binary.BigEndian.PutUint32(serializedReq[92:96], uint32(req.NumWant))
	binary.BigEndian.PutUint16(serializedReq[96:98], req.Port)

	return serializedReq
}

// Response format:
// Offset      Size            Name            Value
// 0           32-bit integer  action          1 // announce
// 4           32-bit integer  transaction_id
// 8           32-bit integer  interval
// 12          32-bit integer  leechers
// 16          32-bit integer  seeders
// 20 + 6 * n  32-bit integer  IP address
// 24 + 6 * n  16-bit integer  TCP port
// 20 + 6 * N
func UnmarshalTrackerUDPAnnounceResponse(rawResponse []byte) (*TrackerUDPAnnounceResponse, error) {
	// Validate the response size
	responseSize := len(rawResponse)
	if (responseSize - 20) % 6 != 0 {
		return nil, errors.New("Announce response's size is invalid. Currently only IPv4 is supported.")
	}

	action := binary.BigEndian.Uint32(rawResponse[:4])
	if TrackerAction(action) != TrackerActionAnnounce {
		return nil, errors.New("Value of the field 'action' in the response is invalid.")
	}

	// Parse addreses of peers
	peerCount := (responseSize - 20) / 6
	peers := []PeerAddr{}
	for i := 0; i < peerCount; i++ {
		ipOffset := 20 + 6 * i
		portOffset := 24 + 6 * i
		newPeer := PeerAddr{
			IP: net.IP(rawResponse[ipOffset:ipOffset + 4]),
			Port: binary.BigEndian.Uint16(rawResponse[portOffset:portOffset + 2]),
		}
		peers = append(peers, newPeer)
	}

	return &TrackerUDPAnnounceResponse{
		TxnID: int32(binary.BigEndian.Uint32(rawResponse[4:8])),
		Interval: int32(binary.BigEndian.Uint32(rawResponse[8:12])),
		Leechers: int32(binary.BigEndian.Uint32(rawResponse[12:16])),
		Seeders: int32(binary.BigEndian.Uint32(rawResponse[16:20])),
		PeerAddresses: peers,
	}, nil
}

func (req TrackerUDPAnnounceRequest) GetActionCode() TrackerAction {
	return TrackerActionAnnounce
}

type TrackerUDPConnectRequest struct {
	TxnID int32
}

// if genTxnID is set to true, the transaction ID will be randomly generated.
// otherwise, this field will be set to 0.
func CreateTrackerUDPConnectRequest(genTxnID bool) *TrackerUDPConnectRequest {
	var txnID int32 = 0
	if genTxnID {
		txnID = rand.Int31()
	}
	return &TrackerUDPConnectRequest{
		TxnID: txnID,
	}
}

// Request format:
// Offset  Size            Name            Value
// 0       64-bit integer  protocol_id     0x41727101980 // magic constant
// 8       32-bit integer  action          0 // connect
// 12      32-bit integer  transaction_id
// 16
func (req TrackerUDPConnectRequest) Marshal() []byte {
	rawRequest := make([]byte, UDPConnectRequestSize)
	binary.BigEndian.PutUint64(rawRequest[0:8], 0x41727101980)
	binary.BigEndian.PutUint32(rawRequest[8:12], uint32(req.GetActionCode()))
	binary.BigEndian.PutUint32(rawRequest[12:], uint32(req.TxnID))
	return rawRequest
}

func (req TrackerUDPConnectRequest) GetActionCode() TrackerAction {
	return TrackerActionConnect
}

type TrackerUDPConnectResponse struct {
	TxnID int32
	ConnectionID int64
}

func (req TrackerUDPConnectResponse) GetActionCode() TrackerAction {
	return TrackerActionConnect
}

// Offset  Size            Name            Value
// 0       32-bit integer  action          0 // connect
// 4       32-bit integer  transaction_id
// 8       64-bit integer  connection_id
// 16
func UnmarshalTrackerUDPConnectResponse(rawResponse []byte) (*TrackerUDPConnectResponse, error) {
	action := binary.BigEndian.Uint32(rawResponse[:4])

	if TrackerAction(action) != TrackerActionConnect {
		return nil, fmt.Errorf("Invalid value of field 'action', expect '%d', but got: '%d'.", TrackerActionConnect, action)
	}

	txnID := int32(binary.BigEndian.Uint32(rawResponse[4:8]))
	connID := int64(binary.BigEndian.Uint64(rawResponse[8:16]))

	return &TrackerUDPConnectResponse{
		TxnID: txnID,
		ConnectionID: connID,
	}, nil
}
