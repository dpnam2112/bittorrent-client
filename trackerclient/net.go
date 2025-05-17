package trackerclient

import (
	"fmt"
	"log/slog"
	"net"
	"time"
)

type TrackerUDPClient struct {
	Logger *slog.Logger
}

func (client *TrackerUDPClient) SendConnectRequest(trackerIP net.IP, trackerPort int, readTimeout time.Duration) (*TrackerUDPConnectResponse, error) {
	// remote address
	raddr := net.UDPAddr{
		Port: trackerPort,
		IP: trackerIP,
	}

	conn, err := net.DialUDP("udp", nil, &raddr)
	if err != nil {
		return nil, fmt.Errorf("Failed to create an UDP socket to %s: %w", raddr.IP.String(), err)
	}

	conn.SetReadDeadline(time.Now().Add(readTimeout * time.Second))
	defer conn.Close()

	// Generate a UDP connection request with transaction ID randomly generated.
	connectRequest := CreateTrackerUDPConnectRequest(true)
	rawConnectRequest := connectRequest.Marshal()

	client.Logger.Debug(
		"Send a connect request to a tracker",
		"request_payload", connectRequest,
		"raw_payload", fmt.Sprintf("% x", rawConnectRequest),
	)

	_, err = conn.Write(connectRequest.Marshal())
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

	client.Logger.Debug(
		"Received connect response from tracker",
		"raw_payload", fmt.Sprintf("% x\n", buf[:n]),
	)

	connectResp, err := UnmarshalTrackerUDPConnectResponse(buf[:UDPConnectResponseSize])
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal connect response: %w", err)
	}

	client.Logger.Debug("Received connect response from the tracker", "response_payload", connectResp)
	return connectResp, nil
}

func (client *TrackerUDPClient) SendAnnounceRequest(
	trackerIP net.IP,
	trackerPort int,
	readTimeout time.Duration,
	r *TrackerUDPAnnounceRequest,
) (*TrackerUDPAnnounceResponse, error) {
	raddr := net.UDPAddr{
		IP: trackerIP,
		Port: trackerPort,
	}

	conn, err := net.DialUDP("udp", nil, &raddr)
	if err != nil {
		return nil, fmt.Errorf("Failed to open an UDP socket to %s: %w", raddr.String(), err)
	}

	request := r.Marshal()
	client.Logger.Debug("Send an announce request to the tracker", "raw_payload", fmt.Sprintf("% x\n", request))
	_, err = conn.Write(request)

	if err != nil {
		return nil, fmt.Errorf("Failed to send an UDP packet to %s: %w", raddr.String(), err)
	}

	// Max size of an IP packet is 65535 bytes
	// An UDP packet is just a thin wrapper of an IP packet
	responseBuf := make([]byte, 65535)
	conn.SetReadDeadline(time.Now().Add(readTimeout * time.Second))
	n, _, err := conn.ReadFromUDP(responseBuf)

	if err != nil {
		return nil, fmt.Errorf("Failed to read UDP announce response from %s: %w", raddr.String(), err)
	}

	client.Logger.Debug("Received response", "response_size", n, "raw_payload", fmt.Sprintf("% x", responseBuf[:n]))
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
