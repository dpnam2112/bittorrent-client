package peer

import "github.com/dpnam2112/bittorrent-client/common"



type Peer interface {
	common.LifeCycle
	Addr() common.PeerAddr
}
