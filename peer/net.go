package peer

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"
	"reflect"
	"time"
)

const (
	defaultReadTimeout  int = 5000
	defaultWriteTimeout int = 5000
)

type PeerWireConnection interface {
	SendPeerMessages(messages []PeerMessage) error
	ReadPeerMessage() (PeerMessage, error)
	Handshake(peerID [20]byte, protocolName string, infoHash [20]byte) (HandshakeMessage, error)
	io.Closer
}

type peerWireConnection struct {
	logger       slog.Logger
	lpeerID      [20]byte
	conn         net.Conn
	connReader   *bufio.Reader
	connWriter   *bufio.Writer
	readTimeout  int
	writeTimeout int
}

func CreatePeerWireConnection(
	rAddr string,
	logger slog.Logger,
) (PeerWireConnection, error) {
	conn, err := net.DialTimeout("tcp", rAddr, 5*time.Second)

	if err != nil {
		return nil, fmt.Errorf("Error while initiating peer wire connection: %w", err)
	}

	logger = *logger.With(
		"local_addr", conn.LocalAddr().String(),
		"remote_addr", conn.RemoteAddr().String(),
	)

	c := peerWireConnection{
		logger:     logger,
		conn:       conn,
		connReader: bufio.NewReader(conn),
		connWriter: bufio.NewWriter(conn),
	}

	return &c, nil
}

func (c *peerWireConnection) SetReadTimeout(readTimeout int) {
	if readTimeout < 0 {
		c.readTimeout = defaultReadTimeout
	} else {
		c.readTimeout = readTimeout
	}
}

func (c *peerWireConnection) SetWriteTimeout(writeTimeout int) {
	if writeTimeout < 0 {
		c.writeTimeout = defaultWriteTimeout
	} else {
		c.writeTimeout = writeTimeout
	}
}

func (c *peerWireConnection) Handshake(
	peerID [20]byte,
	protocolName string,
	infoHash [20]byte,
) (HandshakeMessage, error) {
	if len(peerID) != 20 {
		return nil, fmt.Errorf("peerID's length must be exactly 20.")
	}

	rAddr := c.conn.RemoteAddr().String()

	// handshaking
	handshakeMsg, err := createHandshakeMessage(protocolName, infoHash, peerID)
	if err != nil {
		return nil, fmt.Errorf("Error while initiating peer wire connection: %w", err)
	}

	n, err := c.connWriter.Write(handshakeMsg)
	err = c.connWriter.Flush()
	if err != nil {
		return nil, fmt.Errorf("Error while initiating peer wire connection: %w", err)
	}
	c.logger.Debug("Sent handshake message", "raw_msg", fmt.Sprintf("% x", handshakeMsg), "bytes_sent_count", n)

	recvHandshakeMsg, err := c.readHandshakeMessage()
	if err != nil {
		return nil, fmt.Errorf("Error while initiating peer wire connection: %w", err)
	}

	// check if the received handshake message is valid
	if !reflect.DeepEqual(recvHandshakeMsg.Protocol(), handshakeMsg.Protocol()) {
		return nil, fmt.Errorf(
			"The received handshake's protocol field don't match with that of the sent one. Expect: '%s', but got: '%s'.",
			string(handshakeMsg.Protocol()),
			string(recvHandshakeMsg.Protocol()),
		)
	}

	// check if the received handshake message is valid
	if !reflect.DeepEqual(recvHandshakeMsg.Protocol(), handshakeMsg.Protocol()) {
		return nil, fmt.Errorf("Info hash field of received handshake and the sent handshake don't match.")
	}
	c.logger.Debug("Handshake message", "remote_addr", rAddr, "raw_msg", fmt.Sprintf("% x", recvHandshakeMsg))
	return recvHandshakeMsg, nil
}

// readHandshakeMessage tries reading raw representation (slice) of the handshake message. if the format
// is invalid, return an error.
// The input can contains a handshake followed with multiple peer messages. In this case, the
// function only returns the representation of the handshake message only.
func (c *peerWireConnection) readHandshakeMessage() (handshakeMessage, error) {
	r := c.connReader

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

// SendMessages sends a list of messages to the peer over a connection-oriented protocol.
func (c *peerWireConnection) SendPeerMessages(messages []PeerMessage) error {
	writer := c.connWriter

	c.logger.Debug("Start sending peer messages")
	for i, msg := range messages {
		c.logger.Debug("Peer message", "i", i, "type", msg.Type().String())
		_, err := writer.Write(messages[i].Raw())
		if err != nil {
			return fmt.Errorf("An error occurred while sending messages: %w", err)
		}
	}

	err := writer.Flush()
	if err != nil {
		return fmt.Errorf("An error occurred while sending messages: %w", err)
	}

	c.logger.Debug("Complete sending messages", "msg_count", len(messages))
	return err
}

// Read peer message from the connection.
func (c *peerWireConnection) ReadPeerMessage() (PeerMessage, error) {
	peerMsg, err := c.readPeerMessage()
	if err != nil {
		return nil, err
	}
	c.logger.Debug("Receive a peer message", "type", peerMsg.Type())
	return peerMsg, nil
}

func (c *peerWireConnection) Close() error {
	c.connReader = nil
	c.connWriter = nil

	err := c.conn.Close()
	if err != nil {
		return fmt.Errorf("Error while closing the client: %w", err)
	}

	c.logger.Debug("Connection closed", "peer_addr", c.conn.RemoteAddr().String())
	return err
}

func (c *peerWireConnection) readPeerMessage() (PeerMessage, error) {
	r := c.connReader

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
