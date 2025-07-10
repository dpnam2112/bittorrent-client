package torrentparser

import (
	"crypto/sha1"
	"fmt"
	"strings"
)

// TorrentMetainfo represents the parsed content of a .torrent file.
type TorrentMetainfo struct {
	announce     string
	announceList [][]string
	info         InfoDict
}

func NewTorrentMetainfo(announce string, announceList [][]string, info InfoDict) TorrentMetainfo {
	return TorrentMetainfo{
		announce:     announce,
		announceList: announceList,
		info:         info,
	}
}

func (t TorrentMetainfo) Announce() string {
	return t.announce
}

func (t TorrentMetainfo) AnnounceList() [][]string {
	return t.announceList
}

func (t TorrentMetainfo) Info() InfoDict {
	return t.info
}

// InfoDict represents the "info" dictionary in a torrent file.
type InfoDict struct {
	name        string
	pieceLength int64
	pieces      []byte
	length      int64
	files       []FileEntry
	rawBencode  []byte // raw bencode representation of the info dictionary. This is used to compute SHA-1 hash of the torrent.
	hash        [20]byte
}

func (i InfoDict) Name() string {
	return i.name
}

func (i InfoDict) PieceLength() int64 {
	return i.pieceLength
}

func (i InfoDict) Pieces() []byte {
	return i.pieces
}

func (i InfoDict) Length() int64 {
	return i.length
}

func (i InfoDict) Files() []FileEntry {
	return i.files
}

func (i InfoDict) Hash() [20]byte {
	// Calculate SHA-1 hash of the info dictionary.
	rawBencode := i.rawBencode
	hash := sha1.Sum(rawBencode)
	return hash
}

// FileEntry represents a file in a multi-file torrent.
type FileEntry struct {
	length int64
	path   []string
}

func NewFileEntry(length int64, path []string) FileEntry {
	return FileEntry{
		length: length,
		path:   path,
	}
}

func (f FileEntry) Length() int64 {
	return f.length
}

func (f FileEntry) Path() []string {
	return f.path
}

func (t TorrentMetainfo) String() string {
	var sb strings.Builder

	sb.WriteString("=== TorrentMetainfo Info ===\n")
	sb.WriteString(fmt.Sprintf("Announce: %s\n", t.Announce()))

	if len(t.AnnounceList()) > 0 {
		sb.WriteString("Announce List:\n")
		for _, tier := range t.AnnounceList() {
			sb.WriteString("  - ")
			sb.WriteString(strings.Join(tier, ", "))
			sb.WriteByte('\n')
		}
	}

	info := t.Info()
	sb.WriteString(fmt.Sprintf("Name: %s\n", info.Name()))
	sb.WriteString(fmt.Sprintf("Piece Length: %d\n", info.PieceLength()))

	numPieces := len(info.Pieces()) / 20
	sb.WriteString(fmt.Sprintf("Pieces: %d pieces (%d bytes total)\n", numPieces, len(info.Pieces())))

	if len(info.Files()) > 0 {
		sb.WriteString("Files:\n")
		for _, f := range info.Files() {
			sb.WriteString(fmt.Sprintf("  - %s (%d bytes)\n", strings.Join(f.Path(), "/"), f.Length()))
		}
	} else {
		sb.WriteString(fmt.Sprintf("Single File Length: %d bytes\n", info.Length()))
	}

	return sb.String()
}
