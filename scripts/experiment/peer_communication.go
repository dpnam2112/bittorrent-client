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
	"fmt"
	"log/slog"
	"os"

	"github.com/dpnam2112/bittorrent-client/peerwire"
	"github.com/dpnam2112/bittorrent-client/torrentparser"
)

func main() {
	// Configure logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	path := "./sample_torrents/backlog.torrent"
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	reader := bufio.NewReader(f)
	torrent, err := torrentparser.ParseTorrent(reader)
	infohash := torrent.Info().Hash()

	fmt.Println("Parsed the torrent:", path)
	fmt.Println("Torrent's content:")
	fmt.Println(torrent.String())
	fmt.Println("Torrent's infohash:", fmt.Sprintf("% x\n", infohash))

	addrStr := "127.0.0.1:6881"

	// Example peer ID (20 bytes, usually starts with client ID like "-GT0001-")
	peerIDStr := "-GO0001-123456789012"
	peerID := [20]byte{}
	copy(peerID[:], []byte(peerIDStr))

	conn, err := peerwire.CreatePeerWireConnection(addrStr, *slog.Default())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Step 1: Send handshake
	recvHandshake, err := conn.Handshake(peerID, "", infohash)
	if err != nil {
		panic(err)
	}
	fmt.Println("Received handshake from:", recvHandshake.PeerID)

	// Step 2: Expect Bitfield
	msg, err := conn.ReadPeerMessage()
	if err != nil {
		panic(err)
	}
	fmt.Println("Received message of type:", msg.Type())

	if msg.Type() == peerwire.TypeBitfield {
		bitfield := peerwire.BitFieldMessagePayload(msg.Payload())
		fmt.Printf("Bitfield: % x\n", bitfield)
		fmt.Println("Is bit 0 set:", bitfield.IsSet(0))
	}

	// Step 3: Wait for Unchoke
	for {
		msg, err = conn.ReadPeerMessage()
		if err != nil {
			panic(err)
		}
		fmt.Println("Received message of type:", msg.Type())

		if msg.Type() == peerwire.TypeUnchoke {
			fmt.Println("Unchoked by peer, ready to send requests")
			break
		}
	}

	// Optional: Send Interested if you want to download
	// messages := []peerwire.PeerMessage{
	//     peerwire.CreateInterestedMessage(),
	// }
	// conn.SendPeerMessages(messages)

	// Step 4: Request piece
	err = conn.SendPeerMessages([]peerwire.PeerMessage{
		peerwire.CreateInterestedMessage(),
		peerwire.CreateRequestMessage(0, 0, 1741),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Sent Request for piece 0, offset 0, length 16KB")

	// Step 5: Wait for Piece
	for msg.Type() != peerwire.TypePiece {

		msg, err = conn.ReadPeerMessage()
		if err != nil {
			panic(err)
		}
		fmt.Println("Received message of type:", msg.Type())
	}

	fmt.Println("Received a piece message.")
	piecePayload := peerwire.PieceMessagePayload(msg.Payload())
	fmt.Println("Piece length:", len(piecePayload.Piece()))
	fmt.Println("Piece content:", string(piecePayload.Piece()[:100]), "... (concatenated)")
}
