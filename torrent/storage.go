package torrent

import (
	"github.com/dpnam2112/bittorrent-client/common"
	"github.com/dpnam2112/bittorrent-client/torrentparser"
)

type TorrentStorage interface {
	StoreBlock(common.BlockID, []byte)

	GetBlock(common.BlockID) []byte

	ValidatePiece(pieceIndex int)

	Pieces() []common.PieceIndex

	ExistingPieces() []common.PieceIndex
}

func NewTorrentStorage(metainfo *torrentparser.TorrentMetainfo, location string) TorrentStorage
