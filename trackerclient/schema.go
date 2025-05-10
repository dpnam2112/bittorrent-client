package trackerclient

import (
	"encoding/binary"
	"errors"
	"net"
)

type TrackerAction int32

// Tracker action is a 32-bit unsigned integer representing actions that the downloader wants to
// make when interacting with a tracker.
const (
	TrackerActionConnect  TrackerAction = 0x01
	TrackerActionAnnounce TrackerAction = 0x00
)

type AnnounceEvent uint32

const (
	AnnounceEventNone      AnnounceEvent = 0
	AnnounceEventCompleted AnnounceEvent = 1
	AnnounceEventStarted   AnnounceEvent = 2
	AnnounceEventStopped   AnnounceEvent = 3
)

const (
	UDPAnnounceRequestSize = 98
)

type PeerAddr struct {
	IP   net.IP
	Port uint16
}

type TrackerUDPAnnounceRequest struct {
	ConnectionID int64
	TxnID        int32
	InfoHash     []byte
	PeerID       int32
	Downloaded   int32
	Uploaded     int32
	Left         int32
	Event        AnnounceEvent
	IPAddr       *net.IP
	Port         uint16
	Key          int32
	NumWant      int32
}

type TrackerUDPAnnounceResponse struct {
	TxnID         int32
	Interval      int32
	Leechers      int32
	Seeders       int32
	PeerAddresses []PeerAddr
}

type TrackerUDPErrorResponse struct {
	TxnID   int32
	message string
}

// Serialize request to bytes
// Format:
// IPv4 announce request:
//
// Offset  Size    Name    Value
// 0       64-bit integer  connection_id
// 8       32-bit integer  action          1 // announce
// 12      32-bit integer  transaction_id
// 16      20-byte string  info_hash
// 36      20-byte string  peer_id
// 56      64-bit integer  downloaded
// 64      64-bit integer  left
// 72      64-bit integer  uploaded
// 80      32-bit integer  event           0 // 0: none; 1: completed; 2: started; 3: stopped
// 84      32-bit integer  IP address      0 // default
// 88      32-bit integer  key
// 92      32-bit integer  num_want        -1 // default
// 96      16-bit integer  port
// 98
func (req TrackerUDPAnnounceRequest) Marshal() []byte {
	serializedReq := make([]byte, UDPAnnounceRequestSize)

	binary.BigEndian.PutUint64(serializedReq[0:8], uint64(req.ConnectionID))
	binary.BigEndian.PutUint32(serializedReq[8:12], uint32(req.GetActionCode()))
	binary.BigEndian.PutUint32(serializedReq[12:16], uint32(req.TxnID))
	copy(serializedReq[16:36], req.InfoHash)
	binary.BigEndian.PutUint32(serializedReq[36:56], uint32(req.PeerID))
	binary.BigEndian.PutUint64(serializedReq[56:64], uint64(req.Downloaded))
	binary.BigEndian.PutUint64(serializedReq[64:72], uint64(req.Left))
	binary.BigEndian.PutUint64(serializedReq[72:80], uint64(req.Uploaded))
	binary.BigEndian.PutUint32(serializedReq[80:84], uint32(req.Event))

	if req.IPAddr == nil {
		binary.BigEndian.PutUint32(serializedReq[84:88], uint32(0))
	} else {
		rawIpv4 := req.IPAddr.To4()
		copy(serializedReq[84:88], rawIpv4)
	}

	binary.BigEndian.PutUint32(serializedReq[88:92], 0)
	binary.BigEndian.PutUint32(serializedReq[92:96], uint32(req.NumWant))
	binary.BigEndian.PutUint16(serializedReq[96:98], req.Port)

	return serializedReq
}

// Response format:
// Offset      Size            Name            Value
// 0           32-bit integer  action          1 // announce
// 4           32-bit integer  transaction_id
// 8           32-bit integer  interval
// 12          32-bit integer  leechers
// 16          32-bit integer  seeders
// 20 + 6 * n  32-bit integer  IP address
// 24 + 6 * n  16-bit integer  TCP port
// 20 + 6 * N
func UnmarshalTrackerUDPAnnounceResponse(rawResponse []byte) (*TrackerUDPAnnounceResponse, error) {
	// Validate the response size
	responseSize := len(rawResponse)
	if (responseSize - 20) % 6 != 0 {
		return nil, errors.New("Announce response's size is invalid. Currently only IPv4 is supported.")
	}

	action := binary.BigEndian.Uint32(rawResponse[:4])
	if TrackerAction(action) != TrackerActionAnnounce {
		return nil, errors.New("Value of the field 'action' in the response is invalid.")
	}

	// Parse addreses of peers
	peerCount := (responseSize - 20) / 6
	peers := []PeerAddr{}
	for i := 0; i < peerCount; i++ {
		ipOffset := 20 + 6 * i
		portOffset := 24 + 6 * i
		newPeer := PeerAddr{
			IP: net.IP(rawResponse[ipOffset:ipOffset + 4]),
			Port: binary.BigEndian.Uint16(rawResponse[portOffset:portOffset + 2]),
		}
		peers = append(peers, newPeer)
	}

	return &TrackerUDPAnnounceResponse{
		TxnID: int32(binary.BigEndian.Uint32(rawResponse[4:8])),
		Interval: int32(binary.BigEndian.Uint32(rawResponse[8:12])),
		Leechers: int32(binary.BigEndian.Uint32(rawResponse[12:16])),
		Seeders: int32(binary.BigEndian.Uint32(rawResponse[16:20])),
	}, nil
}

func (req TrackerUDPAnnounceRequest) GetActionCode() TrackerAction {
	return 0x1
}
