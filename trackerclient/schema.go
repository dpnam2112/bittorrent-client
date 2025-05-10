package trackerclient

// Tracker action is a 32-bit unsigned integer representing actions that the downloader wants to
// make when interacting with a tracker.
const (
	TrackerActionConnect uint32 = 0x01
	TrackerActionAnnounce uint32 = 0x00
)


type TrackerUDPConnectRequest struct {
	TxnID uint32
}


type AnnounceEvent uint32


const (
	AnnounceEventNone AnnounceEvent = 0
	AnnounceEventCompleted AnnounceEvent = 1
	AnnounceEventStarted AnnounceEvent = 2
	AnnounceEventStopped AnnounceEvent = 3
)


type TrackerUDPIPV4AnnounceRequest struct {
	TxnID int32
	InfoHash int32
	PeerID int32
	Downloaded int32
	Uploaded int32 
	Left int32
	Event AnnounceEvent
	IPAddress int32
	Key int32
	NumWant int32
}


type TrackerUDPAnnounceResponse struct {
	TxnID int32
	Interval int32
	Leechers int32
	Seeders int32
	PeerAddresses []struct {
		Host string
		Port uint16
	}
}


type TrackerUDPErrorResponse struct {
	TxnID int32
	message string
}
