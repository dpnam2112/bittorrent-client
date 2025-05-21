package peerwire

import (
	"encoding/binary"
	"fmt"
)

// Handshake message format (in order):
//
// Field        | Size (bytes) | Description
// ------------ | ------------ | ----------------------------------------------
// pstrlen      | 1            | Length of the protocol identifier (`pstr`)
// pstr         | variable     | Protocol identifier (default: "BitTorrent protocol")
// reserved     | 8            | Reserved for extension flags (DHT, Fast, etc.)
// info_hash    | 20           | SHA-1 hash of the torrent metadata
// peer_id      | 20           | Unique identifier for the peer
//
// Total size: 49 + len(pstr) bytes (usually 68 bytes if pstr is 19)
type handshakeMessage []byte

func (msg handshakeMessage) Protocol() string {
	pStrLen := msg[0]
	return string(msg[1:1 + pStrLen])
}

// Create peer-wire protocol handshake message
func createHandshakeMessage(protocolName string, infohash, peerID [20]byte,) (handshakeMessage, error) {
	if protocolName == "" {
		protocolName = "BitTorrent protocol"
	}

	if len(protocolName) > 255 {
		return nil, fmt.Errorf("Protocol name's length must not exceed 255.")
	}

	size := 1 + len(protocolName) + 8 + 20 + 20
	msg := make([]byte, size)
	msg[0] = uint8(len(protocolName))
	copy(msg[1:1 + len(protocolName)], []byte(protocolName))
	copy(msg[1 + len(protocolName) + 8:1 + len(protocolName) + 28], infohash[:])
	copy(msg[1 + len(protocolName) + 28:1 + len(protocolName) + 48], peerID[:])

	return msg, nil
}

// readHandshakeMessage tries reading raw representation (slice) of the handshake message. if the format
// is invalid, return an error.
// The input can contains a handshake followed with multiple peer messages. In this case, the
// function only returns the representation of the handshake message only.
func readHandshakeMessage(raw []byte) (handshakeMessage, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("Message's length is 0.")
	}

	pStrLen := uint8(raw[0])
	handshakeMsgSize := int(pStrLen) + 49
	if len(raw) < handshakeMsgSize  {
		return nil, fmt.Errorf("Invalid message length.")
	}

	return handshakeMessage(raw[:handshakeMsgSize]), nil
}

func (msg handshakeMessage) InfoHash() [20]byte {
	infohashOffset := 1 + msg[0] + 8
	infoHash := [20]byte{}
	copy(infoHash[:], msg[infohashOffset:infohashOffset + 20])
	return infoHash
}

func (msg handshakeMessage) PeerID() [20]byte {
	peerIDOffset := 1 + msg[0] + 8 + 20
	peerID := [20]byte{}
	copy(peerID[:], msg[peerIDOffset:peerIDOffset + 20])
	return peerID
}


type PeerMsgType uint8


const (
	TypeChoke         PeerMsgType = 0
	TypeUnchoke       PeerMsgType = 1
	TypeInterested    PeerMsgType = 2
	TypeNotInterested PeerMsgType = 3
	TypeHave          PeerMsgType = 4
	TypeBitfield      PeerMsgType = 5
	TypeRequest       PeerMsgType = 6
	TypePiece         PeerMsgType = 7
	TypeCancel        PeerMsgType = 8
	TypePort          PeerMsgType = 9
	// KeepAlive is a special case — it has no ID and length is 0
	TypeKeepAlive PeerMsgType = 255 // Reserved for internal handling
)

type PeerMessage interface {
	Type() PeerMsgType
	Length() uint32
	Payload() MessagePayload
}

type peerMessage []byte
type MessagePayload []byte

func (msg peerMessage) Type() PeerMsgType {
	return PeerMsgType(msg[0])
}

func (msg peerMessage) Length() uint32 {
	return binary.BigEndian.Uint32(msg[:4])
}

func (msg peerMessage) Payload() MessagePayload {
	return MessagePayload(msg[5:1 + msg.Length()])
}

func AsPeerMessage(raw []byte) (PeerMessage, error) {
	if len(raw) < 5 {
		return nil, fmt.Errorf("Invalid slice length. Expect length >= 5.")
	}

	msgLength := binary.BigEndian.Uint32(raw[:4])
	return peerMessage(raw[:4 + msgLength]), nil
}

type HaveMessagePayload MessagePayload

func (payload HaveMessagePayload) Index() uint32 {
	return binary.BigEndian.Uint32(payload[:4])
}

type PieceMessagePayload MessagePayload

func (payload PieceMessagePayload) Index() uint32 {
	return binary.BigEndian.Uint32(payload[:4])
}

func (payload PieceMessagePayload) Begin() uint32 {
	return binary.BigEndian.Uint32(payload[4:8])
}

func (payload PieceMessagePayload) Piece() uint32 {
	return binary.BigEndian.Uint32(payload[8:])
}

func createPeerMessage(msgType PeerMsgType, payload MessagePayload) PeerMessage {
	raw := make([]byte, 5 + len(payload))
	length := 1 + len(payload)
	binary.BigEndian.PutUint32(raw, uint32(length))
	raw[4] = byte(msgType)
	return peerMessage(raw)
}
