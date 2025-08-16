package tracker

import (
	"encoding/binary"
	"fmt"
	"net"
	"testing"
)

// udpTrackerStub: minimal in-process UDP tracker that speaks connect + announce.
// Options let you flip error/txn-mismatch without changing tests.
type udpTrackerStub struct {
	conn          *net.UDPConn
	addr          *net.UDPAddr
	interval      uint32
	forceError    bool
	mismatchTxnID bool
}

func newUDPTrackerStub(t *testing.T, intervalSec uint32, opts ...func(*udpTrackerStub)) *udpTrackerStub {
	t.Helper()

	la, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil { t.Fatalf("resolve: %v", err) }
	conn, err := net.ListenUDP("udp", la)
	if err != nil { t.Fatalf("listen: %v", err) }

	s := &udpTrackerStub{conn: conn, addr: conn.LocalAddr().(*net.UDPAddr), interval: intervalSec}
	for _, o := range opts { o(s) }

	go s.serve(t)
	t.Cleanup(func() { _ = s.conn.Close() })
	return s
}

func withError() func(*udpTrackerStub)       { return func(s *udpTrackerStub) { s.forceError = true } }
func withTxnMismatch() func(*udpTrackerStub) { return func(s *udpTrackerStub) { s.mismatchTxnID = true } }
func (s *udpTrackerStub) URL() string        { return fmt.Sprintf("udp://%s", s.addr.String()) }

func (s *udpTrackerStub) serve(t *testing.T) {
	buf := make([]byte, 2048)
	for {
		n, raddr, err := s.conn.ReadFromUDP(buf)
		if err != nil { return } // socket closed
		if n < 16 { continue }

		action := binary.BigEndian.Uint32(buf[8:12]) // request action at offset 8
		switch action {
		case 0: // CONNECT
			resp := make([]byte, 16)
			binary.BigEndian.PutUint32(resp[0:4], 0) // action=connect
			binary.BigEndian.PutUint32(resp[4:8], binary.BigEndian.Uint32(buf[12:16])) // echo txn
			binary.BigEndian.PutUint64(resp[8:16], 0x1122334455667788)                  // connection id
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
			txn := binary.BigEndian.Uint32(buf[12:16])
			if s.mismatchTxnID { txn ^= 0xdeadbeef }

			// action(4) + txn(4) + interval(4) + leechers(4) + seeders(4) + 2 peers (6 each)
			resp := make([]byte, 20+6*2)
			binary.BigEndian.PutUint32(resp[0:4], 1)        // announce
			binary.BigEndian.PutUint32(resp[4:8], txn)      // txn
			binary.BigEndian.PutUint32(resp[8:12], s.interval)
			binary.BigEndian.PutUint32(resp[12:16], 3)      // leechers (arbitrary)
			binary.BigEndian.PutUint32(resp[16:20], 2)      // seeders = count

			// peer 1: 10.0.0.1:6881
			copy(resp[20:24], []byte{10, 0, 0, 1})
			binary.BigEndian.PutUint16(resp[24:26], 6881)
			// peer 2: 10.0.0.2:6881
			copy(resp[26:30], []byte{10, 0, 0, 2})
			binary.BigEndian.PutUint16(resp[30:32], 6881)

			_, _ = s.conn.WriteToUDP(resp, raddr)
		}
	}
}

