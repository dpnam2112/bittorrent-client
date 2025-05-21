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

	"github.com/dpnam2112/bittorrent-client/torrent"
	"github.com/dpnam2112/bittorrent-client/peerwire"
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
	torrent, err := torrent.ParseTorrent(reader)
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
    
	conn, err := peerwire.InitiatePeerWireConnection(addrStr, peerID, "", infohash, *slog.Default())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
}
