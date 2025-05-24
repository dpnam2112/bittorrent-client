package peerwire

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

type Message interface {
	Length() uint32
}

type HandshakeMessage interface {
	Message
	Protocol() string
	InfoHash() [20]byte
	PeerID() [20]byte
}

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
	return string(msg[1 : 1+pStrLen])
}

func (msg handshakeMessage) Length() uint32 {
	pstrLen := msg[0]
	return 49 + uint32(pstrLen)
}

// Create peer-wire protocol handshake message
func createHandshakeMessage(protocolName string, infohash, peerID [20]byte) (handshakeMessage, error) {
	if protocolName == "" {
		protocolName = "BitTorrent protocol"
	}

	if len(protocolName) > 255 {
		return nil, fmt.Errorf("Protocol name's length must not exceed 255.")
	}

	size := 1 + len(protocolName) + 8 + 20 + 20
	msg := make([]byte, size)
	msg[0] = uint8(len(protocolName))
	copy(msg[1:1+len(protocolName)], []byte(protocolName))
	copy(msg[1+len(protocolName)+8:1+len(protocolName)+28], infohash[:])
	copy(msg[1+len(protocolName)+28:1+len(protocolName)+48], peerID[:])

	return msg, nil
}

// readHandshakeMessage tries reading raw representation (slice) of the handshake message. if the format
// is invalid, return an error.
// The input can contains a handshake followed with multiple peer messages. In this case, the
// function only returns the representation of the handshake message only.
func readHandshakeMessage(r *bufio.Reader) (handshakeMessage, error) {
	pstrLen, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("Error while reading handshake message from reader: %w", err)
	}

	size := 1 + int(pstrLen) + 48
	msg := make([]byte, size)
	msg[0] = pstrLen
	io.ReadFull(r, msg[1:1+pstrLen+48])

	return handshakeMessage(msg), nil
}

func (msg handshakeMessage) InfoHash() [20]byte {
	infohashOffset := 1 + msg[0] + 8
	infoHash := [20]byte{}
	copy(infoHash[:], msg[infohashOffset:infohashOffset+20])
	return infoHash
}

func (msg handshakeMessage) PeerID() [20]byte {
	peerIDOffset := 1 + msg[0] + 8 + 20
	peerID := [20]byte{}
	copy(peerID[:], msg[peerIDOffset:peerIDOffset+20])
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
	TypeHaveNone      PeerMsgType = 0xf
	// KeepAlive is a special case — it has no ID and length is 0
	TypeKeepAlive PeerMsgType = 255 // Reserved for internal handling
)

func (t PeerMsgType) String() string {
	switch t {
	case TypeChoke:
		return "Choke"
	case TypeUnchoke:
		return "Unchoke"
	case TypeInterested:
		return "Interested"
	case TypeNotInterested:
		return "NotInterested"
	case TypeHave:
		return "Have"
	case TypeBitfield:
		return "Bitfield"
	case TypeRequest:
		return "Request"
	case TypePiece:
		return "Piece"
	case TypeCancel:
		return "Cancel"
	case TypePort:
		return "Port"
	case TypeHaveNone:
		return "HaveNone"
	case TypeKeepAlive:
		return "KeepAlive"
	default:
		return fmt.Sprintf("UnknownPeerMsgType(%d)", uint8(t))
	}
}

// Peer wire protocol message format:
//
// Each message (after the handshake) is length-prefixed and consists of:
//
//	[ length_prefix (4 bytes) ][ message_id (1 byte) ][ payload (variable) ]
//
// - length_prefix: 4-byte big-endian uint32 indicating the length of message_id + payload.
// - message_id: 1 byte indicating the message type (not present if length == 0).
// - payload: varies by message type (e.g., piece index, block data).
//
// A message with length_prefix == 0 is a Keep-Alive message and has no message_id or payload.
type PeerMessage interface {
	Message
	Type() PeerMsgType
	Payload() MessagePayload
	Raw() []byte
}

type peerMessage []byte
type MessagePayload []byte

func readPeerMessage(r *bufio.Reader) (PeerMessage, error) {
	// read message length
	prefLenBytes := make([]byte, 4)
	_, err := io.ReadFull(r, prefLenBytes)

	if err != nil {
		return nil, fmt.Errorf("An error occurred while reading peer message: %w", err)
	}

	bodyLen := binary.BigEndian.Uint32(prefLenBytes)

	// 'unmarshal' data to a message instance
	msg := make([]byte, 4+bodyLen)
	copy(msg[:4], prefLenBytes)
	_, err = io.ReadFull(r, msg[4:])

	if err != nil {
		return nil, fmt.Errorf("An error occurred while reading peer message: %w", err)
	}

	return peerMessage(msg), nil
}

func (msg peerMessage) Raw() []byte {
	return []byte(msg)
}

func (msg peerMessage) Type() PeerMsgType {
	if len(msg) == 4 {
		return TypeKeepAlive
	}
	return PeerMsgType(msg[4])
}

func (msg peerMessage) Length() uint32 {
	return binary.BigEndian.Uint32(msg[:4])
}

func (msg peerMessage) Payload() MessagePayload {
	return MessagePayload(msg[5 : 4+msg.Length()])
}

func AsPeerMessage(raw []byte) (PeerMessage, error) {
	if len(raw) < 5 {
		return nil, fmt.Errorf("Invalid slice length. Expect length >= 5.")
	}

	msgLength := binary.BigEndian.Uint32(raw[:4])
	return peerMessage(raw[:4+msgLength]), nil
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

func (payload PieceMessagePayload) Piece() []byte {
	return payload[8:]
}

func createPeerMessage(msgType PeerMsgType, payload MessagePayload) PeerMessage {
	payloadLen := 0

	if payload != nil {
		payloadLen = len(payload)
		raw := make([]byte, 5+payloadLen)
		lenFieldValue := 1 + len(payload)
		binary.BigEndian.PutUint32(raw[:4], uint32(lenFieldValue))
		raw[4] = byte(msgType)
		copy(raw[5:], payload)
		return peerMessage(raw)
	} else {
		raw := make([]byte, 5)
		binary.BigEndian.PutUint32(raw, 1)
		raw[4] = byte(msgType)
		return peerMessage(raw)
	}
}

type BitFieldMessagePayload MessagePayload

func (payload BitFieldMessagePayload) IsSet(i int) bool {
	return !((payload[i/8] >> (7 - i%8)) == 0x0)
}

func CreateRequestMessage(index int, begin int, length int) PeerMessage {
	msgPayload := make(MessagePayload, 12)
	binary.BigEndian.PutUint32(msgPayload[:4], uint32(index))
	binary.BigEndian.PutUint32(msgPayload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(msgPayload[8:12], uint32(length))
	return createPeerMessage(TypeRequest, msgPayload)
}

func CreateChokeMessage() PeerMessage {
	return createPeerMessage(TypeChoke, nil)
}

func CreateUnchokeMessage() PeerMessage {
	return createPeerMessage(TypeUnchoke, nil)
}

func CreateInterestedMessage() PeerMessage {
	return createPeerMessage(TypeInterested, nil)
}

func CreateNotInterestedMessage() PeerMessage {
	return createPeerMessage(TypeNotInterested, nil)
}

func CreateHaveMessage(index int) PeerMessage {
	raw := make([]byte, 9)
	binary.BigEndian.PutUint32(raw[:4], 5)
	raw[4] = byte(TypeHave)
	binary.BigEndian.PutUint32(raw[5:], uint32(index))
	return peerMessage(raw)
}

func CreateHaveNoneMessage() PeerMessage {
	return createPeerMessage(TypeHaveNone, nil)
}
