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

func (client *TrackerUDPClient) SendConnectRequest(trackerIP string, trackerPort int, readTimeout time.Duration) (*TrackerUDPConnectResponse, error) {
	// remote address
	raddr := net.UDPAddr{
		Port: trackerPort,
		IP: net.ParseIP(trackerIP),
	}

	conn, err := net.DialUDP("udp", nil, &raddr)
	if err != nil {
		return nil, fmt.Errorf("Failed to open an UDP connection: %w", err)
	}

	conn.SetReadDeadline(time.Now().Add(readTimeout * time.Second))
	defer conn.Close()

	// Generate a UDP connection request with transaction ID randomly generated.
	connectRequest := CreateTrackerUDPConnectRequest(true)
	client.Logger.Debug("Send a connect request to a tracker", "request_payload", connectRequest)

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

	connectResp, err := UnmarshalTrackerUDPConnectResponse(buf[:UDPConnectResponseSize])
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal connect response: %w", err)
	}

	client.Logger.Debug("Received connect response from the tracker", "response_payload", connectResp)
	return connectResp, nil
}
