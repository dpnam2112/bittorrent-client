package peer


type MsgHandler interface {
	HandleMessage(peerMessage PeerMessage) error
}
