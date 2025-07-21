package torrent

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
	Download(context.Context) error
	ClientID() common.PeerID
}

type blockDownloadingState struct {
	blockID common.BlockID

	// Peers used to download the block
	// A block request may be sent to multiple peers so that the average rtt for retrieving a block
	// can be reduced (downloading speed varies across peers)
	usedPeers []common.PeerAddr

	downloaded bool
}

type torrentClientImpl struct {
	metainfo            *torrentparser.TorrentMetainfo
	trackerPeerResolver trackerclient.TrackerPeerResolver
	peers               map[common.PeerAddr]peer.Peer
	discoveredPeerAddrs []common.PeerAddr
	torrentStorage      TorrentStorage
	logger              slog.Logger
	cancelFn            context.CancelFunc
	pieceDownloaders    map[common.PieceIndex]PieceDownloader

	clientID common.PeerID
}

func (c *torrentClientImpl) NewTorrentClient(metainfo *torrentparser.TorrentMetainfo, location string, logger slog.Logger) TorrentClient {
	client := torrentClientImpl{}

	client.metainfo = metainfo
	client.trackerPeerResolver = trackerclient.NewTrackerPeerResolver(client.metainfo)
	client.torrentStorage = NewTorrentStorage(metainfo, location)
	client.logger = logger

	client.initPieceDownloader()
	client.initTrackerPeerResolver()

	return &client
}

func (c *torrentClientImpl) initPieceDownloader() error {
	if c.torrentStorage == nil {
		return fmt.Errorf("Error initializing piece downloader: field 'torrentStorage' is not set.")
	}

	if c.metainfo == nil {
		return fmt.Errorf("Error initializing piece downloader: field 'metainfo' is not set.")
	}

	for idx, _ := range c.metainfo.Info().Pieces() {
		pieceIndex := common.PieceIndex(idx)
		pieceDownloader, err := NewPieceDownloader(pieceIndex, *c.metainfo, c.torrentStorage)
		if err != nil {
			return fmt.Errorf("Error initializing piece downloader: %w", err)
		}
		c.pieceDownloaders[pieceIndex] = pieceDownloader
	}

	return nil
}

func (c *torrentClientImpl) ClientID() common.PeerID {
	// TODO: Implement logic to generate client ID
	idStr := "-UT3530-1n2k3j4h5l6m"
	var id common.PeerID
	copy(id[:], idStr[:20])
	return id
}

func (c *torrentClientImpl) Start(ctx context.Context) error {
	if err := c.trackerPeerResolver.Start(ctx); err != nil {
		return fmt.Errorf("Error when starting tracker peer resolver: %w", err)
	}

	_, cancelFn := context.WithCancel(ctx)
	c.cancelFn = cancelFn

	return nil
}

func (c *torrentClientImpl) Download(ctx context.Context) error {
	peerDiscoveryHandler := func(peerAddrs []common.PeerAddr) error {
		for _, peerAddr := range peerAddrs {
			if c.peers[peerAddr] != nil {
				continue
			}

			newPeer, err := peer.NewPeer(ctx, peerAddr)

			if err != nil {
				c.logger.Error("Error creating a new peer", "err", err)
				continue
			}

			c.peers[peerAddr] = newPeer
		}

		return nil
	}

	c.trackerPeerResolver.AddPeerDiscoveredHandler(peerDiscoveryHandler)
	announcementData := trackerclient.AnnoucementData{
		Uploaded:   0,
		Downloaded: 0,
		Left:       0,
		Event:      trackerclient.AnnounceEventStarted,
	}

	c.trackerPeerResolver.SetAnnoucementData(announcementData)
	c.trackerPeerResolver.Announce(ctx, announcementData)

	// Calculate number of peers that own a piece, for each piece
	// TODO: Add a loop to download every piece until all pieces are downloaded
	// TODO: Skip pieces that are already downloaded
	piecePeerCounts := make([]int, len(c.metainfo.Info().Pieces()))
	for _, peer := range c.peers {
		pieceIndices := peer.Pieces()

		for _, index := range pieceIndices {
			piecePeerCounts[index]++
		}
	}

	selectedPiece, avail := 0, piecePeerCounts[0]

	for pieceIndex, peerCount := range piecePeerCounts {
		if peerCount < avail {
			selectedPiece = pieceIndex
		}
	}

	// TODO: Implement piece downloading strategy here
	// e.g: Rarest first, Random first, etc.
	c.downloadPiece(ctx, common.PieceIndex(selectedPiece))
	return nil
}

func (c *torrentClientImpl) downloadPiece(ctx context.Context, pieceIndex common.PieceIndex) error {
	pieceDownloader := c.pieceDownloaders[pieceIndex]

	if err := pieceDownloader.DownloadWithContext(ctx); err != nil {
		return fmt.Errorf("Error downloading the piece: %w", err)
	}

	return nil
}

func (c *torrentClientImpl) Close() error {
	if err := c.trackerPeerResolver.Close(); err != nil {
		c.logger.Error("Error when closing trackerPeerResolver:", "err", err)
	}

	return nil
}

// TODO: Initialize peer resolver and handler functions that will be called when new peers are
// discovered.
func (c *torrentClientImpl) initTrackerPeerResolver() error
