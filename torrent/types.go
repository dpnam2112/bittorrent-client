package torrent

import (
	"fmt"
	"strings"
)

// Torrent represents the parsed content of a .torrent file.
type Torrent struct {
	Announce     string     // Primary tracker URL
	AnnounceList [][]string // List of tracker URLs (optional)
	Info         InfoDict   // Info dictionary containing file details
}

// InfoDict represents the "info" dictionary in a torrent file.
type InfoDict struct {
	Name        string      // Name of the file or directory
	PieceLength int64       // Length of each piece in bytes
	Pieces      []byte      // Concatenated SHA1 hashes (20 bytes each)
	Length      int64       // Length of the single file (for single-file torrents)
	Files       []FileEntry // List of files (for multi-file torrents)
}

// FileEntry represents a file in a multi-file torrent.
type FileEntry struct {
	Length int64    // Length of the file in bytes
	Path   []string // Path components (for multi-file torrents)
}

func (t Torrent) String() string {
	var sb strings.Builder

	sb.WriteString("=== Torrent Info ===\n")
	sb.WriteString(fmt.Sprintf("Announce: %s\n", t.Announce))

	if len(t.AnnounceList) > 0 {
		sb.WriteString("Announce List:\n")
		for _, tier := range t.AnnounceList {
			sb.WriteString("  - ")
			sb.WriteString(strings.Join(tier, ", "))
			sb.WriteByte('\n')
		}
	}

	info := t.Info
	sb.WriteString(fmt.Sprintf("Name: %s\n", info.Name))
	sb.WriteString(fmt.Sprintf("Piece Length: %d\n", info.PieceLength))

	// Safely print info about raw pieces
	numPieces := len(info.Pieces) / 20
	sb.WriteString(fmt.Sprintf("Pieces: %d pieces (%d bytes total)\n", numPieces, len(info.Pieces)))

	// File info
	if len(info.Files) > 0 {
		sb.WriteString("Files:\n")
		for _, f := range info.Files {
			sb.WriteString(fmt.Sprintf("  - %s (%d bytes)\n", strings.Join(f.Path, "/"), f.Length))
		}
	} else {
		// Single-file mode
		sb.WriteString(fmt.Sprintf("Single File Length: %d bytes\n", info.Length))
	}

	return sb.String()
}
