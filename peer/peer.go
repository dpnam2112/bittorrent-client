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
	RegisterMsgHandler(name string, handler MsgHandler)

	// Remove message handler by name
	RemoveMsgHandler(string)

	Addr() common.PeerAddr

	Pieces() []common.PieceIndex

	// Instances of this interface are expected to maintain an internal message buffer.
	// These methods enqueue peer messages to the internal message buffer.
	Request(pieceIndex common.PieceIndex, begin, length int)
}

// TODO: Create a new peer
func NewPeer(ctx context.Context, peerAddr common.PeerAddr) (Peer, error) {
	return nil, nil
}
