package trackerclient

import (
	"context"

	"github.com/dpnam2112/bittorrent-client/common"
)

type AnnoucementData struct {
	Uploaded   int64
	Downloaded int64
	Left       int64
	Event      AnnounceEvent
	numWant    uint16
}

type PeerDiscoveryHandler func(peers []common.PeerAddr)

// TrackerPeerResolver resolves peers by sending announcement requests to trackers.
// For each announcement request, trackers only returns a subset of peers. It's the
// responsibility of the user (caller) to manage peer connections and to track which peers are
// already connected to, which are not.
type TrackerPeerResolver interface {
	// Explicitly send annoucement requests to trackers
	Announce(context.Context, AnnoucementData) error

	SetAnnoucementData(AnnoucementData)
	AnnoucementData() AnnoucementData

	// can be used as a goroutine to automatically send annoucement to every tracker after a
	// specific interval. the interval between two annoucements for a tracker is based on the
	// annoucement response from that tracker.
	RunAutoAnnouncer(context.Context) error

	// Add handler function that would be called after the announce request is sent to the
	// tracker(s).
	AddPeerDiscoveredHandler(handler PeerDiscoveryHandler)
}

func NewTrackerPeerResolver(metainfo *common.TorrentMetainfo) (TrackerPeerResolver, error)
