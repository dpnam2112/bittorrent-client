package trackerclient

// Tracker action is a 32-bit unsigned integer representing actions that the downloader wants to
// make when interacting with a tracker.
const (
	TrackerActionConnect uint32 = 0x01
	TrackerActionAnnounce uint32 = 0x00
)

// A generic payload structure for both UDP request/response.
type UDPPayload interface {
	Serialize() ([]byte, error)
	GetAction() (uint32)
	GetTxnID() (uint32)
}

type udpConnectRequest struct {
	action uint32
	txnID uint32
}

type udpIPV4AnnounceRequest struct {
	connectionID uint32
}
