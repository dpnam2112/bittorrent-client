package peer

import (
	"context"

	"github.com/dpnam2112/bittorrent-client/common"
)


// Bittorrent supports two protocols for peer communication: TCP and uTP. But TCP suffers from NAT
// traversal, hence the strategy should be:
// - Try connecting using uTP first
// - If uTP does not work, use TCP as fallback.
type Peer interface {
	common.Closer

	// Register message handler (name + handler instance)
	// Registered handler will be called when a peer receive a peer message
	// If there is already a message handler registered with name `name`, the method will return an
	// error.
	RegisterMsgHandler(name string, handler MsgHandler) error

	// Remove message handler by name
	RemoveMsgHandler(string)

	Addr() common.PeerAddr

	// Return a list of pieces the peer owns.
	Pieces() []common.PieceIndex
}

// TODO: Create a new peer
func NewPeer(ctx context.Context, peerAddr common.PeerAddr) (Peer, error) {
	return nil, nil
}
