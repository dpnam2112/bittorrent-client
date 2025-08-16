package tracker

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/dpnam2112/bittorrent-client/common"
)

// Purpose: Connect to a local stub tracker and perform a single announce.
// Asserts we get a non-empty peer list and a positive interval.
func TestTrackerClient_Announce_HappyPath(t *testing.T) {
	stub := newUDPTrackerStub(t, 1)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	tc, err := NewTrackerClientFromURL(context.Background(), stub.URL(), *logger)
	if err != nil { t.Fatalf("NewTrackerClientFromURL: %v", err) }

	var ih common.InfoHash
	var pid common.PeerID
	copy(ih[:], "0123456789abcdefghij")
	copy(pid[:], "-UT0001-abcdefghij")

	resp, err := tc.Announce(context.Background(), AnnounceRequest{
		InfoHash: ih, PeerID: pid, Event: AnnounceEventStarted, NumWant: -1, Port: 6881,
	})
	if err != nil { t.Fatalf("announce error: %v", err) }
	if resp.Interval <= 0 { t.Fatalf("want positive interval, got %d", resp.Interval) }
	if len(resp.PeerAddrs) == 0 { t.Fatalf("want peers, got none") }
}

// Purpose: If tracker returns an "error" action, the client should surface an error.
func TestTrackerClient_Announce_ErrorPacket(t *testing.T) {
	stub := newUDPTrackerStub(t, 1, withError())
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tc, err := NewTrackerClientFromURL(context.Background(), stub.URL(), *logger)
	if err != nil { t.Fatalf("client: %v", err) }

	var ih common.InfoHash
	var pid common.PeerID
	_, err = tc.Announce(context.Background(), AnnounceRequest{InfoHash: ih, PeerID: pid})
	if err == nil { t.Fatalf("expected error from tracker, got nil") }
}

// Purpose: If the response txn ID doesn't match request txn ID, client should fail loudly.
func TestTrackerClient_Announce_TxnMismatch(t *testing.T) {
	stub := newUDPTrackerStub(t, 1, withTxnMismatch())
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tc, err := NewTrackerClientFromURL(context.Background(), stub.URL(), *logger)
	if err != nil { t.Fatalf("client: %v", err) }

	var ih common.InfoHash
	var pid common.PeerID
	_, err = tc.Announce(context.Background(), AnnounceRequest{InfoHash: ih, PeerID: pid})
	if err == nil { t.Fatalf("expected txn mismatch error") }
}

