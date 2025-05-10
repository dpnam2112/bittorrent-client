package trackerclient

import (
	"encoding/binary"
	"net"
)

// Tracker action is a 32-bit unsigned integer representing actions that the downloader wants to
// make when interacting with a tracker.
const (
	TrackerActionConnect  uint32 = 0x01
	TrackerActionAnnounce uint32 = 0x00
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
func (req TrackerUDPAnnounceRequest) Serialize() []byte {
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

func (req TrackerUDPAnnounceRequest) GetActionCode() int32 {
	return 0x1
}
