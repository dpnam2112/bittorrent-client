package trackerclient

import (
	"github.com/dpnam2112/bittorrent-client/common"
	"github.com/dpnam2112/bittorrent-client/peer"
	"github.com/dpnam2112/bittorrent-client/torrentparser"
)

type AnnounceData struct {
	Uploaded   int
	Downloaded int
	Left       int
	Event      AnnounceEvent
}

type PeerDiscoveryHandler func(peers []peer.Peer) error

// TrackerPeerResolver resolves peers by sending announcement requests to trackers.
// For each announcement request, trackers only returns a subset of peers. It's the
// responsibility of the user (caller) to manage peer connections and to track which peers are
// already connected to, which are not.
type TrackerPeerResolver interface {
	common.LifeCycle

	// Set data for tracker announcement, including metric: uploaded, downloaded, left
	Announce(AnnounceData) error

	// Register handler function that would be called after the announce request is sent to the
	// tracker(s).
	RegisterHandler(handler PeerDiscoveryHandler)
}

func NewTrackerPeerResolver(metainfo *torrentparser.TorrentMetainfo, maxPeerCount int) TrackerPeerResolver {
	// TODO: Implement construction logic
	// maxPeerCount is the maximum number of peers the resolver is able to resolve
	// maxPeerCount = -1 is equivalent to no upper threshold.
	return nil
}
