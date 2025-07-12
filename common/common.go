package common

import "context"

// Internal ID of a torrent
type TorrentID uint16


type PeerID [20]byte

type InfoHash [20]byte

type PeerAddr struct {
    Host string
    Port uint16
}

type LifeCycle interface {
	Start(context context.Context) error
	Close() error
}
