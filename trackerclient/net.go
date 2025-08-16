package tracker

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/dpnam2112/bittorrent-client/common"
)

type TrackerClient interface {
	Addr() net.Addr
    Announce(context.Context, AnnounceRequest) (AnnounceResponse, error)
	common.Closer
}

type trackerUDPClient struct {
	logger slog.Logger
	ip net.IP
	port uint16
	udpConn *net.UDPConn
	connectionID int64
}


func NewTrackerUDPClient(ip net.IP, port uint16, logger slog.Logger) (TrackerClient, error) {
	// remote address
	raddr := net.UDPAddr{
		Port: int(port),
		IP:   ip,
	}

	conn, err := net.DialUDP("udp", nil, &raddr)

	if err != nil {
		return nil, fmt.Errorf("Failed to create an UDP socket to %s: %w", raddr.IP.String(), err)
	}

	client := trackerUDPClient{ ip: ip, port: port, logger: logger, udpConn: conn, }
	resp, err := client.sendConnectRequest(10 * time.Second)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to UDP tracker %s: %w", raddr.IP.String(), err)
	}

	client.connectionID = resp.ConnectionID
	return &client, nil
}


func (client *trackerUDPClient) Addr() net.Addr {
	return &net.UDPAddr{IP: client.ip, Port: int(client.port)}
}


func (client *trackerUDPClient) Close() error {
	return client.udpConn.Close()
}


func (client *trackerUDPClient) Announce(ctx context.Context, req AnnounceRequest) (AnnounceResponse, error) {
	txnID, err := generateTxnID()
	if err != nil {
		return AnnounceResponse{}, fmt.Errorf("Error generating transaction ID: %w", err)
	}

	udpReq := TrackerUDPAnnounceRequest{
		ConnectionID: client.connectionID,
		TxnID: txnID,
		InfoHash: req.InfoHash,
		PeerID: req.PeerID,
		Uploaded: req.Uploaded,
		Downloaded: req.Downloaded,
		Left: req.Left,
		Event: req.Event,
		IPAddr: req.IP,
	}

	resp, err := client.sendAnnounceRequest(ctx, &udpReq)
	if err != nil {
		return AnnounceResponse{}, fmt.Errorf("Error sending announce request: %w", err)
	}

	if resp.TxnID != udpReq.TxnID {
		return AnnounceResponse{}, fmt.Errorf("Error sending announce request: transaction ID doesn't match.")
	}

	if resp.Action() != TrackerActionAnnounce {
		return AnnounceResponse{}, fmt.Errorf("Error sending announce request: Expect action field to be 'announce' in the response from tracker.")
	}

	return AnnounceResponse{
		Interval: resp.Interval,
		Leechers: resp.Leechers,
		Seeders: resp.Seeders,
		PeerAddrs: resp.PeerAddresses,
	}, nil
}


func (client *trackerUDPClient) sendAnnounceRequest(
	ctx context.Context,
	r *TrackerUDPAnnounceRequest,
) (*TrackerUDPAnnounceResponse, error) {
	conn := client.udpConn
	raddr := conn.RemoteAddr()

	request := r.Marshal()
	client.logger.Debug("Send an announce request to the tracker", "raw_payload", fmt.Sprintf("% x\n", request))
	_, err := conn.Write(request)

	if err != nil {
		return nil, fmt.Errorf("Failed to send an UDP packet to %s: %w", raddr.String(), err)
	}

	// Max size of an IP packet is 65535 bytes
	// An UDP packet is just a thin wrapper of an IP packet
	responseBuf := make([]byte, 65535)

	deadline, ok := ctx.Deadline()
	if !ok {
		// no deadline is set in the context
		deadline = time.Now().Add(20 * time.Second)
	}

	conn.SetReadDeadline(deadline)
	n, _, err := conn.ReadFromUDP(responseBuf)

	if err != nil {
		return nil, fmt.Errorf("Failed to read UDP announce response from %s: %w", raddr.String(), err)
	}

	client.logger.Debug("Received response", "response_size", n, "raw_payload", fmt.Sprintf("% x", responseBuf[:n]))
	action := getActionFromRawResp(responseBuf[:n])

	if action == TrackerActionError {
		errResp, err := UnmarshalTrackerUDPErrorResponse(responseBuf[:n])
		if err != nil {
			return nil, fmt.Errorf("Failed to read UDP error response from %s: %w", raddr.String(), err)
		}

		return nil, fmt.Errorf("Error response from %s: %s", raddr.String(), errResp.Message)
	}

	announceResp, err := UnmarshalTrackerUDPAnnounceResponse(responseBuf[:n])
	if err != nil {
		return nil, fmt.Errorf("Failed to read UDP announce response from %s: %w", raddr.String(), err)
	}

	return announceResp, nil
}

func generateTxnID() (int32, error) {
    var b [4]byte
    _, err := rand.Read(b[:])
    if err != nil {
        return 0, err
    }
    return int32(binary.BigEndian.Uint32(b[:])), nil
}

func (client *trackerUDPClient) sendConnectRequest(readTimeout time.Duration) (*TrackerUDPConnectResponse, error) {
	conn := client.udpConn

	conn.SetReadDeadline(time.Now().Add(readTimeout * time.Second))

	// Generate a UDP connection request with transaction ID randomly generated.
	connectRequest := CreateTrackerUDPConnectRequest(true)
	rawConnectRequest := connectRequest.Marshal()

	client.logger.Debug(
		"Send a connect request to a tracker",
		"request_payload", connectRequest,
		"raw_payload", fmt.Sprintf("% x", rawConnectRequest),
	)

	_, err := conn.Write(connectRequest.Marshal())
	if err != nil {
		return nil, fmt.Errorf("Failed to send a Tracker connect request: %w", err)
	}

	// buffer to store the response
	buf := make([]byte, 512)
	n, _, err := conn.ReadFromUDP(buf)

	if err != nil {
		return nil, fmt.Errorf("Failed to send a Tracker connect request: %w", err)
	}

	if n != UDPConnectResponseSize {
		return nil, fmt.Errorf(
			"Failed to send a Tracker connect request: the response size is invalid. Expect %d, but got %d.",
			UDPConnectResponseSize,
			n,
		)
	}

	client.logger.Debug(
		"Received connect response from tracker",
		"raw_payload", fmt.Sprintf("% x\n", buf[:n]),
	)

	connectResp, err := UnmarshalTrackerUDPConnectResponse(buf[:UDPConnectResponseSize])
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal connect response: %w", err)
	}

	client.logger.Debug("Received connect response from the tracker", "response_payload", connectResp)
	return connectResp, nil
}


func NewTrackerClientFromURL(url string, logger slog.Logger) (TrackerClient, error) {
	host, port, scheme, err := parseTrackerURL(url)
	if err != nil {
		return nil, fmt.Errorf("In peer_resolver.createTrackerClientFromURL, error when parsing URL: %w", err)
	}

	if scheme != "udp" {
		return nil, fmt.Errorf("In peer_resolver.createTrackerClientFromURL, error when parsing URL: scheme is not supported: %s", scheme)
	}

	// resolver
	ips, err := resolveHostToIPs(host)
	if err != nil {
		return nil, fmt.Errorf("In peer_resolver.createTrackerClientFromURL, error when resolving host to IPs: %w", err)
	}

	tracker, err := NewTrackerUDPClient(ips[0], uint16(port), logger)
	if err != nil {
		return nil, fmt.Errorf("In peer_resolver.createTrackerClientFromURL, error when creating a tracker: %w", err)
	}
	return tracker, nil
}

