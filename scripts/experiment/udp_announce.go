// Sample output:
// 
// ➜  bittorrent-client git:(tracker-client) ✗ go run scripts/experiment/udp_announce.go
// Parsed the torrent
// Torrent's content:
// === Torrent Info ===
// Announce: udp://tracker.opentrackr.org:1337/announce
// Name: linuxmint-22.1-cinnamon-64bit.iso
// Piece Length: 2097152
// Pieces: 1422 pieces (28440 bytes total)
// Single File Length: 2980511744 bytes
// 
// Torrent's infohash: 5d 4d 25 e0 e6 66 47 c7 e2 89 20 2d 78 80 75 d6 88 4e c0 02
// 
// time=2025-05-17T15:48:16.669+07:00 level=DEBUG msg="Send a connect request to a tracker" request_payload=&{TxnID:182969796} raw_payload="00 00 04 17 27 10 19 80 00 00 00 00 0a e7 e5 c4"
// time=2025-05-17T15:48:17.000+07:00 level=DEBUG msg="Received connect response from tracker" raw_payload="00 00 00 00 0a e7 e5 c4 3f 71 e3 1f a8 d9 98 47\n"
// time=2025-05-17T15:48:17.000+07:00 level=DEBUG msg="Received connect response from the tracker" response_payload="&{TxnID:182969796 ConnectionID:4571684821874088007}"
// time=2025-05-17T15:48:17.001+07:00 level=DEBUG msg="Send an announce request to the tracker" raw_payload="3f 71 e3 1f a8 d9 98 47 00 00 00 01 0f 61 01 3e 5d 4d 25 e0 e6 66 47 c7 e2 89 20 2d 78 80 75 d6 88 4e c0 02 64 70 6e 61 6d 32 31 31 32 62 69 74 6f 72 72 65 6e 74 31 32 00 00 00 00 00 00 00 0a 00 00 00 00 00 00 1b 39 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 0a 1b 39\n"
// time=2025-05-17T15:48:17.410+07:00 level=DEBUG msg="Received response" response_size=80 raw_payload="00 00 00 01 0f 61 01 3e 00 00 06 f9 00 00 00 1f 00 00 09 cc 5c 60 7c 29 90 f4 5f d3 13 5f fb 85 56 7f d0 d6 3d 2c cc 0b a3 9c e0 2c 68 a6 eb 3c c8 d5 4f 70 1f 0e 22 c3 ae 70 e2 7b 74 c1 55 c1 01 8b c1 4c 51 11 10 46 e3 33 ac dc 76 0e 05 3c"
// %+v
//  &{258015550 1785 31 2508 [{92.96.124.41 37108} {95.211.19.95 64389} {86.127.208.214 15660} {204.11.163.156 57388} {104.166.235.60 51413} {79.112.31.14 8899} {174.112.226.123 29889} {85.193.1.139 49484} {81.17.16.70 58163} {172.220.118.14 1340}]}


package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/dpnam2112/bittorrent-client/torrent"
	"github.com/dpnam2112/bittorrent-client/trackerclient"
)

func main() {
	rand.Seed(time.Now().UnixNano()) // Seed the RNG

	f, err := os.Open("./sample_torrents/linuxmint-22.1-cinnamon-64bit.iso.torrent")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	reader := bufio.NewReader(f)


	torrent, err := torrent.ParseTorrent(reader)
	fmt.Println("Parsed the torrent")
	fmt.Println("Torrent's content:")
	fmt.Println(torrent.String())
	fmt.Println("Torrent's infohash:", fmt.Sprintf("% x\n", torrent.Info().Hash()))

	parsedURL, err := url.Parse(torrent.Announce())
	if err != nil {
		panic(err)
	}

	IPs, err := net.LookupIP(parsedURL.Hostname())

	if err != nil {
		panic(err)
	}

	var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	client := trackerclient.TrackerUDPClient{
		Logger: logger,
	}

	IP := IPs[0]
	port, err := strconv.Atoi(parsedURL.Port())
	if err != nil {
		panic(err)
	}

	connResp, err := client.SendConnectRequest(IP, port, 30)
	if err != nil {
		panic(err)
	}

	var peerID [20]byte
	copy(peerID[:], "dpnam2112bitorrent12")

	announceRequest := trackerclient.TrackerUDPAnnounceRequest{
		ConnectionID: connResp.ConnectionID,
		TxnID: rand.Int31(),
		Downloaded: 10,
		Uploaded: 0,
		Left: 6969,
		Event: trackerclient.AnnounceEventNone,
		IPAddr: nil,
		Port: 6969,
		Key: 0,
		NumWant: 10,
		InfoHash: torrent.Info().Hash(),
		PeerID: peerID,
	}

	resp, err := client.SendAnnounceRequest(IP, port, 30, &announceRequest)
	if err != nil {
		panic(err)
	}
	fmt.Println("%+v\n", resp)
}
