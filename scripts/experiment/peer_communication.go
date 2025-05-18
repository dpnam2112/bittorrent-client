// Setup:
// Create a torrent for ./BACKLOG.md
// Seed the torrent locally with Deluge, at port 6881
// Assume that the downloader (this process) has no pieces.
//
// Flow:
// Go client -> Deluge client: handshake message
// Deluge client -> Go client: handshake message
// Deluge client -> Go client: bitfields message
// Deluge client -> Go client: Unchoke message

package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/dpnam2112/bittorrent-client/torrent"
)

func main() {
	// Configure logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	path := "./sample_torrents/backlog.torrent"
	f, err := os.Open(path)
	defer f.Close()
	reader := bufio.NewReader(f)
	torrent, err := torrent.ParseTorrent(reader)
	infohash := torrent.Info().Hash()

	fmt.Println("Parsed the torrent:", path)
	fmt.Println("Torrent's content:")
	fmt.Println(torrent.String())
	fmt.Println("Torrent's infohash:", fmt.Sprintf("% x\n", infohash))

    addrStr := "127.0.0.1:6881"
    
	conn, err := net.DialTimeout("tcp", addrStr, 5 * time.Second)
	defer conn.Close()

	if err != nil {
		panic(err)
	}

	slog.Debug("Open a TCP connection to a peer", "peerAddr", addrStr)

	msgBuf := bytes.Buffer{}
	const protocol = "BitTorrent protocol"
	pstrLen := byte(len(protocol))

	// Example peer ID (20 bytes, usually starts with client ID like "-GT0001-")
	peerID := []byte("-GO0001-123456789012")

	if len(peerID) != 20 {
		panic("peer ID must be exactly 20 bytes")
	}

	msgBuf.WriteByte(pstrLen)               // 1 byte
	msgBuf.WriteString(protocol)            // 19 bytes
	msgBuf.Write(make([]byte, 8))           // 8 reserved bytes (zeroed)
	msgBuf.Write(infohash[:])               // 20 bytes
	msgBuf.Write(peerID)                    // 20 bytes

	_, err = conn.Write(msgBuf.Bytes())
	if err != nil {
		panic(err)
	}

	slog.Debug("Send handshake message to a peer", "peerAddr", addrStr, "rawMsg", fmt.Sprintf("% x", msgBuf))
	respBuf := make([]byte, 65535)
	n, err := conn.Read(respBuf)
	slog.Debug("Received messages from a peer", "peerAddr", addrStr, "n", n)

	handshakeResp := respBuf[:68]

	if err != nil {
		slog.Error("Error while reading response from a peer", "peerAddr", addrStr, "rawMsg", fmt.Sprintf("% x", handshakeResp))
	}

	slog.Debug("Receive handshake message to a peer", "peerAddr", addrStr, "rawMsg", fmt.Sprintf("% x", handshakeResp))

	// If the downloader is a newcomer, it will receive a bitfield message from the other peer by
	// default.
	bitfieldResp := respBuf[68:n]
	bitfieldRespLen := binary.BigEndian.Uint32(bitfieldResp[:4])
	msgID := bitfieldResp[4]

	if msgID != 0x05 {
		// bitfield message has ID of 5
		slog.Error("Expect bitfield type (5), but got the type:", msgID)
		return
	}
	bitfields := bitfieldResp[5:5 + bitfieldRespLen - 1]
	slog.Debug("Receive a bitfield message", "payload_size", bitfieldRespLen - 1, "payload", fmt.Sprintf("% x", bitfields))
 
	slog.Debug("Remaining", "payload", fmt.Sprintf("% x", respBuf[68 + 4 + bitfieldRespLen:n]))
}
