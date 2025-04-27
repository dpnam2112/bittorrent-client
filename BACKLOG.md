# 📝 BitTorrent Client Backlog

This backlog outlines the planned features and improvements for building a modern BitTorrent client in Go.

---

## ✅ Completed
- [x] Parse `.torrent` files (single-file and multi-file)
- [x] Implement Bencode decoder (string, int, list, dict)
- [x] CLI command to display parsed `.torrent` content

---

## 🧩 Core Protocol Features (Planned)
- [ ] **Peer Discovery**
  - [ ] Connect to tracker via HTTP/UDP
  - [ ] Parse and decode tracker response
  - [ ] Support `announce-list`
- [ ] **Peer Wire Protocol**
  - [ ] Handshake implementation
  - [ ] Message parsing (choke, unchoke, interested, bitfield, request, piece, etc.)
  - [ ] Keep-alive and connection timeout handling
- [ ] **Piece Management**
  - [ ] Piece download and verification using SHA1
  - [ ] Prioritize rarest pieces
- [ ] **Data Storage**
  - [ ] Write pieces to disk in correct file layout
  - [ ] Handle multi-file torrents correctly
- [ ] **Download Engine**
  - [ ] Parallel piece downloads
  - [ ] Basic choking/unchoking logic
  - [ ] Resume partial downloads

---

## 🌐 Modern Protocol Support
- [ ] **DHT (Distributed Hash Table)** for peer discovery without trackers
- [ ] **PEX (Peer Exchange)** to share known peers with others
- [ ] **uTP (Micro Transport Protocol)** support
- [ ] **Magnet URI Support** for starting downloads without `.torrent` files

---

## 🧪 Testing & Debugging
- [ ] Add integration tests for peer communication
- [ ] Simulate and test edge cases (missing pieces, bad hashes, etc.)
- [ ] Verbose logging/debug CLI flag

---

## 🛠 Developer Utilities
- [ ] CLI for downloading a torrent
- [ ] Progress bar for active downloads
- [ ] Torrent verification tool (check pieces from disk)
