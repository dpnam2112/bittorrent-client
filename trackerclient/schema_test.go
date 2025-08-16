package tracker

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
)

// --- helpers ---
func b20(s string) [20]byte {
	var a [20]byte
	copy(a[:], []byte(s))
	return a
}

// --- Connect response ---

func TestUnmarshalTrackerUDPConnectResponse_Valid(t *testing.T) {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint32(buf[0:4], uint32(TrackerActionConnect))
	binary.BigEndian.PutUint32(buf[4:8], 0xAABBCCDD)
	binary.BigEndian.PutUint64(buf[8:16], 0x1122334455667788)

	resp, err := UnmarshalTrackerUDPConnectResponse(buf)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if uint32(resp.TxnID) != 0xAABBCCDD || resp.ConnectionID != 0x1122334455667788 {
		t.Fatalf("bad resp: %#v", resp)
	}
}

func TestUnmarshalTrackerUDPConnectResponse_InvalidAction(t *testing.T) {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint32(buf[0:4], uint32(TrackerActionAnnounce)) // wrong
	_, err := UnmarshalTrackerUDPConnectResponse(buf)
	if err == nil {
		t.Fatalf("expected error")
	}
}

// --- Error response ---

func TestUnmarshalTrackerUDPErrorResponse_Valid(t *testing.T) {
	msg := []byte("oops")
	buf := make([]byte, 8+len(msg))
	binary.BigEndian.PutUint32(buf[0:4], uint32(TrackerActionError))
	binary.BigEndian.PutUint32(buf[4:8], 0x12345678)
	copy(buf[8:], msg)

	resp, err := UnmarshalTrackerUDPErrorResponse(buf)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp.TxnID != 0x12345678 || resp.Message != "oops" {
		t.Fatalf("bad resp: %#v", resp)
	}
}

func TestUnmarshalTrackerUDPErrorResponse_InvalidAction(t *testing.T) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint32(buf[0:4], uint32(TrackerActionConnect)) // wrong
	_, err := UnmarshalTrackerUDPErrorResponse(buf)
	if err == nil {
		t.Fatalf("expected error")
	}
}

// --- Announce request marshal ---

func TestTrackerUDPAnnounceRequest_Marshal(t *testing.T) {
	req := TrackerUDPAnnounceRequest{
		ConnectionID: 0x1122334455667788,
		TxnID:        123,
		InfoHash:     b20("0123456789abcdefghij"),
		PeerID:       b20("-UT0001-abcdefghij"),
		Downloaded:   1,
		Uploaded:     2,
		Left:         3,
		Event:        AnnounceEventStarted,
		IPAddr:       net.IPv4(1, 2, 3, 4),
		Port:         6881,
		NumWant:      -1,
	}
	raw := req.Marshal()
	if len(raw) != UDPAnnounceRequestSize {
		t.Fatalf("size: got %d want %d", len(raw), UDPAnnounceRequestSize)
	}
	// sanity checks
	if got := binary.BigEndian.Uint32(raw[8:12]); TrackerAction(got) != TrackerActionAnnounce {
		t.Fatalf("action mismatch")
	}
	if !bytes.Equal(raw[16:36], req.InfoHash[:]) || !bytes.Equal(raw[36:56], req.PeerID[:]) {
		t.Fatalf("hash/peerID mismatch")
	}
	if got := binary.BigEndian.Uint16(raw[96:98]); got != 6881 {
		t.Fatalf("port mismatch: %d", got)
	}
}

// --- Announce response unmarshal ---

func TestUnmarshalTrackerUDPAnnounceResponse_Valid(t *testing.T) {
	buf := make([]byte, 20+6*2)
	binary.BigEndian.PutUint32(buf[0:4], uint32(TrackerActionAnnounce))
	binary.BigEndian.PutUint32(buf[4:8], 7)   // txn
	binary.BigEndian.PutUint32(buf[8:12], 10) // interval
	binary.BigEndian.PutUint32(buf[12:16], 3) // leechers
	binary.BigEndian.PutUint32(buf[16:20], 2) // seeders
	copy(buf[20:24], net.IPv4(1, 2, 3, 4).To4())
	binary.BigEndian.PutUint16(buf[24:26], 6881)
	copy(buf[26:30], net.IPv4(5, 6, 7, 8).To4())
	binary.BigEndian.PutUint16(buf[30:32], 51413)

	resp, err := UnmarshalTrackerUDPAnnounceResponse(buf)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp.TxnID != 7 || resp.Interval != 10 || len(resp.PeerAddresses) != 2 {
		t.Fatalf("bad resp: %#v", resp)
	}
	if !bytes.Equal(resp.PeerAddresses[0].IP.To4(), net.IPv4(1, 2, 3, 4).To4()) {
		t.Fatalf("peer[0] ip mismatch")
	}
}

func TestUnmarshalTrackerUDPAnnounceResponse_InvalidSize(t *testing.T) {
	// size not congruent to 20+6*n
	buf := make([]byte, 21)
	binary.BigEndian.PutUint32(buf[0:4], uint32(TrackerActionAnnounce))
	_, err := UnmarshalTrackerUDPAnnounceResponse(buf)
	if err == nil {
		t.Fatalf("expected size error")
	}
}

func TestUnmarshalTrackerUDPAnnounceResponse_InvalidAction(t *testing.T) {
	buf := make([]byte, 20)
	binary.BigEndian.PutUint32(buf[0:4], uint32(TrackerActionConnect)) // wrong
	_, err := UnmarshalTrackerUDPAnnounceResponse(buf)
	if err == nil {
		t.Fatalf("expected action error")
	}
}

// --- Misc ---

func TestGetActionFromRawResp(t *testing.T) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(TrackerActionError))
	if got := getActionFromRawResp(buf); got != TrackerActionError {
		t.Fatalf("got %v", got)
	}
}

