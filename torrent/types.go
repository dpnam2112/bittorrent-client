package torrent

// Torrent represents the parsed content of a .torrent file.
type Torrent struct {
    Announce     string   // Primary tracker URL
    AnnounceList [][]string // List of tracker URLs (optional)
    Info         InfoDict // Info dictionary containing file details
}

// InfoDict represents the "info" dictionary in a torrent file.
type InfoDict struct {
    Name        string   // Name of the file or directory
    PieceLength int64    // Length of each piece in bytes
    Pieces      []byte   // Concatenated SHA1 hashes (20 bytes each)
    Length      int64    // Length of the single file (for single-file torrents)
    Files       []FileEntry // List of files (for multi-file torrents)
}

// FileEntry represents a file in a multi-file torrent.
type FileEntry struct {
    Length int64    // Length of the file in bytes
    Path   []string // Path components (for multi-file torrents)
}

