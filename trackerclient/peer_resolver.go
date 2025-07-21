package trackerclient

import (
	"context"

	"github.com/dpnam2112/bittorrent-client/common"
	"github.com/dpnam2112/bittorrent-client/torrentparser"
)

type AnnoucementData struct {
	Uploaded   int64
	Downloaded int64
	Left       int64
	Event      AnnounceEvent
	numWant	uint16
}

type PeerDiscoveryHandler func(peers []common.PeerAddr) error

// TrackerPeerResolver resolves peers by sending announcement requests to trackers.
// For each announcement request, trackers only returns a subset of peers. It's the
// responsibility of the user (caller) to manage peer connections and to track which peers are
// already connected to, which are not.
type TrackerPeerResolver interface {
	common.LifeCycle

	// Set data for tracker announcement, including metric: uploaded, downloaded, left
	// AnnouncementData would be sent to all trackers periodically, and the interval between each
	// annoucement requests depends on the variable `interval` in the annoucement response returned
	// by each tracker.
	SetAnnoucementData(AnnoucementData)

	// Explicitly send annoucement requests to trackers
	Announce(context.Context, AnnoucementData) error

	// Add handler function that would be called after the announce request is sent to the
	// tracker(s).
	AddPeerDiscoveredHandler(handler PeerDiscoveryHandler)
}

func NewTrackerPeerResolver(metainfo *torrentparser.TorrentMetainfo) TrackerPeerResolver {
	// TODO: Implement construction logic
	return nil
}
