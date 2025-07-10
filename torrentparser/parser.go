package torrentparser

import (
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/dpnam2112/bittorrent-client/bencode"
)

// ParseTorrent reads and parses a torrent from an io.Reader into a Torrent struct.
func ParseTorrent(reader io.Reader) (*TorrentMetainfo, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read torrent data: %w", err)
	}

	remaining, value, err := bencode.ParseBencode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bencoded data: %w", err)
	}
	if len(remaining) > 0 {
		slog.Warn("unexpected extra data after parsing")
	}

	dict, ok := value.(*bencode.BDict)
	if !ok {
		slog.Error("torrent data is not a dictionary")
		return nil, errors.New("torrent data is not a dictionary")
	}

	var announce string
	var announceList [][]string
	var info InfoDict

	// Parse announce URL.
	if announceVal, ok := dict.Dict["announce"].(*bencode.BString); ok {
		announce = string(announceVal.Value)
	}

	// Parse announce list (optional).
	if announceListVal, ok := dict.Dict["announce-list"].(*bencode.BList); ok {
		for _, listVal := range announceListVal.Values {
			if sublist, ok := listVal.(*bencode.BList); ok {
				var sublistStrings []string
				for _, elem := range sublist.Values {
					if str, ok := elem.(*bencode.BString); ok {
						sublistStrings = append(sublistStrings, string(str.Value))
					}
				}
				announceList = append(announceList, sublistStrings)
			}
		}
	}

	// Parse info dictionary.
	infoVal, ok := dict.Dict["info"].(*bencode.BDict)
	if !ok {
		return nil, errors.New("missing info dictionary in torrent data")
	}
	info = parseInfoDict(infoVal)

	torrent := NewTorrentMetainfo(announce, announceList, info)
	return &torrent, nil
}

// parseInfoDict parses the "info" dictionary from a torrent file.
func parseInfoDict(infoDict *bencode.BDict) InfoDict {
	var (
		name        string
		pieceLength int64
		pieces      []byte
		length      int64
		files       []FileEntry
	)

	// Parse name.
	if nameVal, ok := infoDict.Dict["name"].(*bencode.BString); ok {
		name = string(nameVal.Value)
	}

	// Parse piece length.
	if pieceLengthVal, ok := infoDict.Dict["piece length"].(*bencode.BInt); ok {
		pieceLength = pieceLengthVal.Value
	}

	// Parse pieces (concatenated SHA-1 hashes).
	if piecesVal, ok := infoDict.Dict["pieces"].(*bencode.BString); ok {
		pieces = piecesVal.Value
	}

	// Parse length (for single-file torrents).
	if lengthVal, ok := infoDict.Dict["length"].(*bencode.BInt); ok {
		length = lengthVal.Value
	}

	// Parse files (for multi-file torrents).
	if filesVal, ok := infoDict.Dict["files"].(*bencode.BList); ok {
		for _, fileVal := range filesVal.Values {
			if fileDict, ok := fileVal.(*bencode.BDict); ok {
				var fileLength int64
				var path []string

				if lengthVal, ok := fileDict.Dict["length"].(*bencode.BInt); ok {
					fileLength = lengthVal.Value
				}

				if pathVal, ok := fileDict.Dict["path"].(*bencode.BList); ok {
					for _, pathElem := range pathVal.Values {
						if str, ok := pathElem.(*bencode.BString); ok {
							path = append(path, string(str.Value))
						}
					}
				}

				files = append(files, NewFileEntry(fileLength, path))
			}
		}
	}

	return InfoDict{
		name:        name,
		pieceLength: pieceLength,
		pieces:      pieces,
		length:      length,
		files:       files,
		rawBencode:  infoDict.GetRawBencode(),
	}
}
