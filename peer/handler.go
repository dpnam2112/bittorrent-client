package peer

import "github.com/dpnam2112/bittorrent-client/common"

type MsgHandler struct {
	onPieceMessage func (peerAddr common.PeerAddr, msg PieceMessagePayload) error
	onBitFieldMessage func (peerAddr common.PeerAddr, msg BitFieldMessagePayload) error
}
