package torrent

import (
	"errors"
	"fmt"
	"io"

	"github.com/dpnam2112/bittorrent-client/bencode"
)

// ParseTorrent reads and parses a torrent from an io.Reader into a Torrent struct.
func ParseTorrent(reader io.Reader) (*Torrent, error) {
	// Read the entire content from the reader.
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read torrent data: %w", err)
	}

	// Parse the Bencoded data.
	remaining, value, err := bencode.ParseBencode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bencoded data: %w", err)
	}
	if len(remaining) != 0 {
		return nil, errors.New("unexpected extra data after parsing")
	}

	// Check if the value is a dictionary.
	dict, ok := value.(*bencode.BDict)
	if !ok {
		return nil, errors.New("torrent data is not a dictionary")
	}

	// Initialize the Torrent struct.
	torrent := &Torrent{}

	// Parse announce URL.
	if announceVal, ok := dict.Dict["announce"].(*bencode.BString); ok {
		torrent.Announce = string(announceVal.Value)
	}

	// Parse announce list (optional).
	if announceListVal, ok := dict.Dict["announce-list"].(*bencode.BList); ok {
		var announceList [][]string
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
		torrent.AnnounceList = announceList
	}

	// Parse info dictionary.
	if infoVal, ok := dict.Dict["info"].(*bencode.BDict); ok {
		torrent.Info = parseInfoDict(infoVal)
	} else {
		return nil, errors.New("missing info dictionary in torrent data")
	}

	return torrent, nil
}

// parseInfoDict parses the "info" dictionary from a torrent file.
func parseInfoDict(infoDict *bencode.BDict) InfoDict {
	info := InfoDict{}

	// Parse name.
	if nameVal, ok := infoDict.Dict["name"].(*bencode.BString); ok {
		info.Name = string(nameVal.Value)
	}

	// Parse piece length.
	if pieceLengthVal, ok := infoDict.Dict["piece length"].(*bencode.BInt); ok {
		info.PieceLength = pieceLengthVal.Value
	}

	// Parse pieces (concatenated SHA-1 hashes).
	if piecesVal, ok := infoDict.Dict["pieces"].(*bencode.BString); ok {
		info.Pieces = piecesVal.Value
	}

	// Parse length (for single-file torrents).
	if lengthVal, ok := infoDict.Dict["length"].(*bencode.BInt); ok {
		info.Length = lengthVal.Value
	}

	// Parse files (for multi-file torrents).
	if filesVal, ok := infoDict.Dict["files"].(*bencode.BList); ok {
		var files []FileEntry
		for _, fileVal := range filesVal.Values {
			if fileDict, ok := fileVal.(*bencode.BDict); ok {
				var file FileEntry

				// Parse length.
				if lengthVal, ok := fileDict.Dict["length"].(*bencode.BInt); ok {
					file.Length = lengthVal.Value
				}

				// Parse path (as a list of strings).
				if pathVal, ok := fileDict.Dict["path"].(*bencode.BList); ok {
					var path []string
					for _, pathElem := range pathVal.Values {
						if str, ok := pathElem.(*bencode.BString); ok {
							path = append(path, string(str.Value))
						}
					}
					file.Path = path
				}

				files = append(files, file)
			}
		}
		info.Files = files
	}

	return info
}

