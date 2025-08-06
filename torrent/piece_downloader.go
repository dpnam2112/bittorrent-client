package torrent

import (
	"context"

	"github.com/dpnam2112/bittorrent-client/common"
	"github.com/dpnam2112/bittorrent-client/peer"
	"github.com/dpnam2112/bittorrent-client/torrentparser"
)


type AllBlocksReceivedHandler func (common.PieceIndex)
type BlockReceivedHandler func (common.BlockID, []byte)

type DownloadEventHandler struct {
	onAllBlocksReceived AllBlocksReceivedHandler
	onBlockReceived BlockReceivedHandler
}


// PieceDownloader downloads a specific piece from peers in the torrent swarm.
// The implementation should organize a data structure to manage piece blocks and track the
// downloading status for each block.
type PieceDownloader interface {
	// Add a list of peers to download the piece
	// Implementations should be thread-safe (i.e, safe to be concurrently accessed)
	AddPeers([]peer.Peer)

	// Index of the piece the downloader is downloading.
	PieceIndex() common.PieceIndex

	// Download the piece.
	// If there is an error during the downloading process, return that error.
	// Otherwise, this method runs until the downloading finishes and return nil.
	//
	// DownloadFinishedHandler would be called when a piece is downloaded successfully.
	// A piece download is considered 'finished' when all blocks of the piece is already downloaded.
	// The downloader is not responsible for verification step.
	//
	// BlockReceivedHandler is called when a block is received.
	// e.g, the user of this interface may want to store the received blocks in a persistent storage
	// like a storage implementation (disk, in-memory, etc.)
	Download(DownloadEventHandler) error

	// Same as Download(), but has a context as paramter.
	DownloadWithContext(context.Context, DownloadEventHandler) error
}

func NewPieceDownloader(
	pieceIndex common.PieceIndex,
	metainfo torrentparser.TorrentMetainfo,
) (PieceDownloader, error) {
	return &pieceDownloaderImpl{}, nil
}

type pieceDownloaderImpl struct {
	pieceIndex     common.PieceIndex
	metainfo       torrentparser.TorrentMetainfo
	peers          []peer.Peer
}

func (downloader *pieceDownloaderImpl) DownloadWithContext(ctx context.Context, BlockReceivedHandler, DownloadFinishedHandler) error
func (downloader *pieceDownloaderImpl) Download(BlockReceivedHandler, DownloadFinishedHandler) error
func (downloader *pieceDownloaderImpl) AddPeers(peers []peer.Peer)
