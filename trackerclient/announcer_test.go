package tracker

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/dpnam2112/bittorrent-client/common"
)

// Test purpose: Happy path – start RunAutoAnnouncer, receive at least one success event, then cancel and ensure it exits.
func TestAnnouncer_EmitsEvent_And_Cancels(t *testing.T) {
	stub := newUDPTrackerStub(t, 1) // tracker returns interval=1s
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Build a tracker client from the stub URL.
	tc, err := NewTrackerClientFromURL(context.Background(), stub.URL(), *logger)
	if err != nil {
		t.Fatalf("client: %v", err)
	}

	// Minimal announcer.
	ann := &trackerAnnouncerImpl{
		trackerClient:   tc,
		logger:          *logger,
		announcerEvents: make(chan AnnouncerEvent, 8),
	}

	// Minimal announce request.
	var ih common.InfoHash
	var pid common.PeerID
	copy(ih[:], "0123456789abcdefghij")
	copy(pid[:], "-UT0001-abcdefghij")
	ann.SetAnnounceRequest(AnnounceRequest{
		InfoHash: ih, PeerID: pid,
		Event: AnnounceEventStarted, NumWant: -1, Port: 6881,
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { _ = ann.RunAutoAnnouncer(ctx); close(done) }()

	// Expect one success event.
	select {
	case ev := <-ann.Events():
		if ev.Err != nil {
			t.Fatalf("unexpected error: %v", ev.Err)
		}
		if ev.AnnounceResp == nil {
			t.Fatalf("nil response")
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting for event")
	}

	// Cancel; announcer should exit promptly.
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("announcer did not stop on cancel")
	}
}

