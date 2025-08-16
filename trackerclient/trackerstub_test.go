package tracker

import (
	"encoding/binary"
	"fmt"
	"net"
	"testing"
)

type udpTrackerStub struct {
	conn     *net.UDPConn
	addr     *net.UDPAddr
	interval uint32
	peers    []struct{ IP net.IP; Port uint16 }
	// behavior toggles
	forceError    bool
	mismatchTxnID bool
}

func newUDPTrackerStub(t *testing.T, intervalSec uint32, opts ...func(*udpTrackerStub)) *udpTrackerStub {
	t.Helper()

	la, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil { t.Fatalf("resolve: %v", err) }
	c, err := net.ListenUDP("udp", la)
	if err != nil { t.Fatalf("listen: %v", err) }

	s := &udpTrackerStub{
		conn:     c,
		addr:     c.LocalAddr().(*net.UDPAddr),
		interval: intervalSec,
		peers: []struct{ IP net.IP; Port uint16 }{
			{net.IPv4(10,0,0,1), 6881},
			{net.IPv4(10,0,0,2), 6881},
		},
	}
	for _, o := range opts { o(s) }

	go s.serve(t)
	t.Cleanup(s.close)
	return s
}
func withError() func(*udpTrackerStub)       { return func(s *udpTrackerStub) { s.forceError = true } }
func withTxnMismatch() func(*udpTrackerStub) { return func(s *udpTrackerStub) { s.mismatchTxnID = true } }

func (s *udpTrackerStub) URL() string { return fmt.Sprintf("udp://%s", s.addr.String()) }
func (s *udpTrackerStub) close()      { _ = s.conn.Close() }

func (s *udpTrackerStub) serve(t *testing.T) {
	buf := make([]byte, 2048)
	for {
		n, raddr, err := s.conn.ReadFromUDP(buf)
		if err != nil { return } // socket closed
		if n < 16 { continue }

		action := binary.BigEndian.Uint32(buf[8:12])
		switch action {
		case 0: // CONNECT
			resp := make([]byte, 16)
			binary.BigEndian.PutUint32(resp[0:4], 0)
			txn := binary.BigEndian.Uint32(buf[12:16])
			binary.BigEndian.PutUint32(resp[4:8], txn)
			binary.BigEndian.PutUint64(resp[8:16], 0x1122334455667788)
			_, _ = s.conn.WriteToUDP(resp, raddr)

		case 1: // ANNOUNCE
			if s.forceError {
				msg := []byte("forced error")
				resp := make([]byte, 8+len(msg))
				binary.BigEndian.PutUint32(resp[0:4], 3) // error
				binary.BigEndian.PutUint32(resp[4:8], binary.BigEndian.Uint32(buf[12:16]))
				copy(resp[8:], msg)
				_, _ = s.conn.WriteToUDP(resp, raddr)
				continue
			}

			peers := s.peers
			resp := make([]byte, 20+6*len(peers))
			binary.BigEndian.PutUint32(resp[0:4], 1) // announce

			txn := binary.BigEndian.Uint32(buf[12:16])
			if s.mismatchTxnID { txn ^= 0xdeadbeef }
			binary.BigEndian.PutUint32(resp[4:8], txn)

			binary.BigEndian.PutUint32(resp[8:12], s.interval)
			binary.BigEndian.PutUint32(resp[12:16], 5) // leechers
			binary.BigEndian.PutUint32(resp[16:20], uint32(len(peers))) // seeders (arbitrary)

			for i, p := range peers {
				off := 20 + i*6
				copy(resp[off:off+4], p.IP.To4())
				binary.BigEndian.PutUint16(resp[off+4:off+6], p.Port)
			}
			_, _ = s.conn.WriteToUDP(resp, raddr)

		default:
			// no-op
		}
	}
}

