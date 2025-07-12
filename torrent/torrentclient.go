package torrentclient

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dpnam2112/bittorrent-client/common"
	"github.com/dpnam2112/bittorrent-client/peer"
	"github.com/dpnam2112/bittorrent-client/torrentparser"
	"github.com/dpnam2112/bittorrent-client/trackerclient"
)

type TorrentClient interface {
	common.LifeCycle
}


type torrentClientImpl struct {
	metainfo *torrentparser.TorrentMetainfo
	trackerPeerResolver trackerclient.TrackerPeerResolver
	connectedPeers []peer.Peer
	ctx context.Context
	Logger slog.Logger
}


func (c *torrentClientImpl) NewTorrentClient(metainfo *torrentparser.TorrentMetainfo, logger slog.Logger) TorrentClient {
	client := torrentClientImpl{}

	client.metainfo = metainfo
	client.trackerPeerResolver = trackerclient.NewTrackerPeerResolver(client.metainfo, -1)

	// the handling logic is triggerred whenever the resolver discovers new peers.
	client.trackerPeerResolver.RegisterHandler(c.handlePeerDiscovery)
	client.Logger = logger

	return &client
}


func (c *torrentClientImpl) Start(context context.Context) error {
	c.ctx = context
	if err := c.trackerPeerResolver.Start(context); err != nil {
		return fmt.Errorf("Error when starting tracker peer resolver: %w", err)
	}

	return nil
}


// handlePeerDiscovery is used as callback for PeerResolver component.
// every time new peers are discovered, this function is invoked for establishing peer connections.
func (c *torrentClientImpl) handlePeerDiscovery(peers []peer.Peer) error {
	for _, peer := range peers {
		go c.connectPeer(peer)
	}

	return nil
}


func (c *torrentClientImpl) connectPeer(peer peer.Peer) {
	if err := peer.Start(c.ctx); err != nil {
		c.Logger.Error("Error when connecting to peer:", "err", err)
	}

	// TODO: Avoid race condition here
	c.connectedPeers = append(c.connectedPeers, peer)
}


func (c *torrentClientImpl) Close() error {
	if err := c.trackerPeerResolver.Close(); err != nil {
		c.Logger.Error("Error when closing trackerPeerResolver:", "err", err)
	}

	return nil
}
