package common

import "context"

type (
	// Internal ID of a torrent
	TorrentID  uint16
	PieceIndex uint32
	PeerID     [20]byte
	InfoHash   [20]byte
	// BlockSize and BlockOFfset 's sizes follow the specification of peer message protocol
	BlockSize uint32

	BlockOffset uint32
	PeerAddr    struct {
		Host string
		Port uint16
	}

	// Blocks are identified by the triple of (piece index, begin, length)
	BlockID struct {
		PieceIndex PieceIndex
		Begin      BlockOffset
		Size       BlockSize
	}
)

const (
	DefaultBlockSize BlockSize = 16384 // 16 kB
)

type Closer interface {
	Close() error
}

type LifeCycle interface {
	Start(context context.Context) error
	Closer
}
