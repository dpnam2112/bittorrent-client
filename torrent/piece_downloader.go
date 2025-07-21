package torrent

import (
	"context"

	"github.com/dpnam2112/bittorrent-client/common"
	"github.com/dpnam2112/bittorrent-client/peer"
	"github.com/dpnam2112/bittorrent-client/torrentparser"
)


type PieceDownloader interface {
	// Add a list of peers to download the piece
	AddPeers([]peer.Peer)

	// Download the piece.
	// If there is an error during the downloading process, return that error.
	// Otherwise, this method runs until the downloading finishes and return nil.
	Download() error

	// Same as Download(), but has a context as paramter.
	DownloadWithContext(context.Context) error
}

func NewPieceDownloader(
	pieceIndex common.PieceIndex,
	metainfo torrentparser.TorrentMetainfo,
	torrentStorage TorrentStorage,
) (PieceDownloader, error) {
	return &pieceDownloaderImpl{}, nil
}


type pieceDownloaderImpl struct {
	pieceIndex common.PieceIndex
	metainfo torrentparser.TorrentMetainfo
	torrentStorage TorrentStorage
	peers []peer.Peer
}

func (downloader *pieceDownloaderImpl) DownloadWithContext(ctx context.Context) error

func (downloader *pieceDownloaderImpl) Download() error

func (downloader *pieceDownloaderImpl) AddPeers(peers []peer.Peer)
