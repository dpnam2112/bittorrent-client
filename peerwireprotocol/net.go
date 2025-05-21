package peerwireprotocol

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
	"reflect"
	"time"
)


type PeerWireConnection interface {
	SendPeerMessages(messages []PeerMessage) error
	ReadPeerMessages(reader io.Reader, n int) ([]PeerMessage, error)
	Close() error
}

type peerWireConnection struct {
	logger slog.Logger
	lpeerID [20]byte
	conn net.Conn
}

func InitiatePeerWireConnection(
	rAddr string,
	peerID [20]byte,
	protocolName string,
	infoHash [20]byte,
	logger slog.Logger,
) (PeerWireConnection, error) {
	if len(peerID) != 20 {
		return nil, fmt.Errorf("peerID's length must be exactly 20.")
	}

	conn, err := net.DialTimeout("tcp", rAddr, 5 * time.Second)

	if err != nil {
		return nil, fmt.Errorf("Error while initiating peer wire connection: %w", err)
	}

	// handshaking
	handshakeMsg, err := createHandshakeMessage(protocolName, infoHash, peerID)
	if err != nil {
		return nil, fmt.Errorf("Error while initiating peer wire connection: %w", err)
	}

	n, err := conn.Write(handshakeMsg)
	if err != nil {
		return nil, fmt.Errorf("Error while initiating peer wire connection: %w", err)
	}
	logger.Debug("Sent handshake message", "remote_addr", rAddr, "raw_msg", fmt.Sprintf("% x", handshakeMsg), "bytes_sent_count", n)


	buf := make([]byte, 1024)
	n, err = conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("Error while initiating peer wire connection: %w", err)
	}
	logger.Debug("Receive messages", "remote_addr", rAddr, "raw_msg", fmt.Sprintf("% x", buf[:n]))

	recvHandshakeMsg, err := readHandshakeMessage(buf[:n])
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
	logger.Debug("Handshake message", "remote_addr", rAddr, "raw_msg", fmt.Sprintf("% x", recvHandshakeMsg))

	c := peerWireConnection{
		logger: logger,
		conn: conn,
		lpeerID: peerID,
	}

	return &c, nil
}

// SendMessages sends a list of messages to the peer over a connection-oriented protocol.
func (c *peerWireConnection) SendPeerMessages(messages []PeerMessage) error {
	buf := bufio.NewWriter(c.conn)

	for i := 0; i < len(messages); i++ {
		_, err := buf.Write(messages[i].Payload())
		if err != nil {
			return fmt.Errorf("An error occurred while sending messages: %w", err)
		}
	}

	err := buf.Flush()
	if err != nil {
		return fmt.Errorf("An error occurred while sending messages: %w", err)
	}

	return err
}

func (c *peerWireConnection) SendHandshakeMessage(infohash [20]byte) error {
	return nil
}

func (c *peerWireConnection) ReadPeerMessages(reader io.Reader, n int) ([]PeerMessage, error) {
	// TODO
	return nil, nil
}

func (c *peerWireConnection) Close() error {
	err := c.conn.Close()
	if err != nil {
		return fmt.Errorf("Error while closing the client: %w", err)
	}
	return err
}
