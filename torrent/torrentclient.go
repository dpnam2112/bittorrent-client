package torrent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dpnam2112/bittorrent-client/common"
	"github.com/dpnam2112/bittorrent-client/peer"
	"github.com/dpnam2112/bittorrent-client/piece"
	"github.com/dpnam2112/bittorrent-client/trackerclient"
)

type TorrentClient interface {
	common.Closer
	Download(context.Context) error
	ClientID() common.PeerID
}

type PeerFactory func (*common.TorrentMetainfo, common.PeerAddr) (peer.Peer, error)
type PieceDownloaderFactory func (*common.TorrentMetainfo, common.PieceIndex) (PieceDownloader, error)

type factoryConfig struct {
	createNewPeer  PeerFactory
	createNewPieceDownloader PieceDownloaderFactory
}

type torrentStat struct {
	downloaded int64 
	uploaded int64
	left int64
}

type torrentClientImpl struct {
	metainfo            *common.TorrentMetainfo
	trackerPeerResolver trackerclient.TrackerPeerResolver
	peers               map[common.PeerAddr]peer.Peer
	pieceRepo			piece.PieceRepository
	pieceSelector		PieceSelector
	logger              slog.Logger
	cancelFn            context.CancelFunc
	pieceDownloaders    map[common.PieceIndex]PieceDownloader
	factoryCfg			factoryConfig
	stat				torrentStat 
	clientID			common.PeerID
}

func (c *torrentClientImpl) NewTorrentClient(metainfo *common.TorrentMetainfo, location string, logger slog.Logger) (TorrentClient, error) {
	var err error

	client := torrentClientImpl{}
	client.metainfo = metainfo

	client.trackerPeerResolver, err = trackerclient.NewTrackerPeerResolver(client.metainfo)
	if err != nil {
		return nil, fmt.Errorf("Error initializing client.trackerPeerResolver: %w", err)
	}

	// initialize stat information (downloaded, uploaded, left)
	c.initTorrentStat()

	client.logger = logger

	return &client, nil
}

func (c *torrentClientImpl) ClientID() common.PeerID {
	// TODO: Implement logic to generate client ID
	idStr := "-UT3530-1n2k3j4h5l6m"
	var id common.PeerID
	copy(id[:], idStr[:20])
	return id
}

func (c *torrentClientImpl) Download(ctx context.Context) error {
	pieceDownloaders, err := c.createPieceDownloaderMap()
	if err != nil {
		return fmt.Errorf("Error when calling Download: %w", err)
	}

	peerDiscoveryHandler := func(peerAddrs []common.PeerAddr) {
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
	}

	c.trackerPeerResolver.AddPeerDiscoveredHandler(peerDiscoveryHandler)
	announcementData := trackerclient.AnnoucementData{
		Uploaded:   c.stat.uploaded,
		Downloaded: c.stat.downloaded,
		Left:       c.stat.left,
		Event:      trackerclient.AnnounceEventStarted,
	}

	c.trackerPeerResolver.SetAnnoucementData(announcementData)

	// Start the background job to send announce requests to trackers.
	go c.trackerPeerResolver.RunAutoAnnouncer(ctx)

	// PieceSelector selects the next piece to be downloaded. If all pieces are downloaded already,
	// it returns nil.
	for pieceIndex, err := c.pieceSelector.SelectOne(); err == nil; {
		pieceDownloader := pieceDownloaders[pieceIndex]
		err := pieceDownloader.DownloadWithContext(ctx, DownloadEventHandler{
			onBlockReceived: c.handleBlockReceived,
			onAllBlocksReceived: c.handleAllBlocksReceived,
		})

		if err != nil {
			return fmt.Errorf("Error when downloading piece %d: %w", pieceIndex, err)
		}
	}

	return nil
}

func (c *torrentClientImpl) Close() error {
	return nil
}

// TODO: Initialize peer resolver and handler functions that will be called when new peers are
// discovered.
func (c *torrentClientImpl) initTrackerPeerResolver() error {
	var err error
	c.trackerPeerResolver, err = trackerclient.NewTrackerPeerResolver(c.metainfo)
	if err != nil {
		return fmt.Errorf("In initTrackerPeerResolver: %w", err)
	}
	return nil
}

func (c *torrentClientImpl) handleBlockReceived(blockID common.BlockID, data []byte) {
	piece := c.pieceRepo.GetPieceByIndex(blockID.PieceIndex)
	if err := piece.StoreBlock(blockID.Begin, blockID.Size); err != nil {
		// TODO: This should be handled gracefully
		c.logger.Error("Error storing block", "PieceIndex", blockID.PieceIndex, "Begin", blockID.Begin, "Size", blockID.Size)
		return
	}
}

// TODO: Handle the event all piece's blocks are downloaded
// If a piece is successfully downloaded, update the torrent stat
func (c *torrentClientImpl) handleAllBlocksReceived(pieceIdx common.PieceIndex) {
	piece := c.pieceRepo.GetPieceByIndex(pieceIdx)
	verified := piece.Verify()

	if verified {
		c.logger.Info("Piece downloaded successfully", "pieceIndex", pieceIdx)
	} else {
		c.logger.Info("Piece verification failed", "pieceIndex", pieceIdx)
	}
}


// Handler implementation for peer messages
// These handler functions are then registered for each peer.
// the appropriate logic is called to handle each type of message.

// Handle bitfield messages from other peers
// Add the peer to the peer list of every piece downloader associated to pieces whose the peer owns.
func (c *torrentClientImpl) onBitfieldMessage(peerAddr common.PeerAddr, payload peer.BitFieldMessagePayload) error {
	var pieceIdx common.PieceIndex = 0
	for ; pieceIdx < common.PieceIndex(len(c.metainfo.Info().Pieces())); pieceIdx++ {
		if payload.IsSet(pieceIdx) {
			c.pieceDownloaders[pieceIdx].AddPeers([]peer.Peer{c.peers[peerAddr]})
		}
	}
	return nil
}

// Initialize piece downloaders for pieces that do not exist in the storage.
func (c *torrentClientImpl) createPieceDownloaderMap() (map[common.PieceIndex]PieceDownloader, error) {
	if c.pieceRepo == nil {
		return nil, fmt.Errorf("Error in initPieceDownloaders: field 'pieceRepo' is not initialized.")
	}

	if c.metainfo == nil {
		return nil, fmt.Errorf("Error initPieceDownloaders: field 'metainfo' is not set.")
	}

	pieceDownloaders := make(map[common.PieceIndex]PieceDownloader)

	pieces := c.pieceRepo.GetAll()
	for _, piece := range pieces {
		if piece.Verify() {
			continue
		}

		newPieceDownloader, err := c.factoryCfg.createNewPieceDownloader(c.metainfo, piece.Index())
		if err != nil {
			return nil, err
		}
		pieceDownloaders[piece.Index()] = newPieceDownloader
	}

	return pieceDownloaders, nil
}


func (c *torrentClientImpl) initTorrentStat() error {
	return nil
}

