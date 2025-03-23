package torrent

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSingleFileTorrent(t *testing.T) {
	// Prepare a simple single-file torrent content (using Bencode encoding)
	data := []byte("d8:announce18:http://tracker.com4:infod4:name12:testfile.txt12:piece lengthi524288e6:pieces20:aaaaaaaaaaaaaaaaaaaaee")

	// Use bytes.NewReader to create an io.Reader for the test data.
	reader := bytes.NewReader(data)

	// Parse the torrent data from the reader.
	torrent, err := ParseTorrent(reader)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, "http://tracker.com", torrent.Announce)
	assert.Equal(t, "testfile.txt", torrent.Info.Name)
	assert.Equal(t, int64(524288), torrent.Info.PieceLength)
	assert.Equal(t, 20, len(torrent.Info.Pieces))
}

func TestParseMultiFileTorrent(t *testing.T) {
	// Prepare a multi-file torrent content
	data := []byte("d8:announce18:http://tracker.com4:infod4:name7:myfiles12:piece lengthi524288e6:pieces40:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb5:filesld6:lengthi12345e4:pathl9:file1.txteed6:lengthi67890e4:pathl9:file2.txteeeee")

	// Use bytes.NewReader to create an io.Reader for the test data.
	reader := bytes.NewReader(data)

	// Parse the torrent data from the reader.
	torrent, err := ParseTorrent(reader)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, "http://tracker.com", torrent.Announce)
	assert.Equal(t, "myfiles", torrent.Info.Name)
	assert.Equal(t, int64(524288), torrent.Info.PieceLength)
	assert.Equal(t, 40, len(torrent.Info.Pieces))

	// Verify multi-file entries.
	assert.Equal(t, 2, len(torrent.Info.Files))
	assert.Equal(t, int64(12345), torrent.Info.Files[0].Length)
	assert.Equal(t, []string{"file1.txt"}, torrent.Info.Files[0].Path)
	assert.Equal(t, int64(67890), torrent.Info.Files[1].Length)
	assert.Equal(t, []string{"file2.txt"}, torrent.Info.Files[1].Path)
}

func TestParseInvalidTorrent(t *testing.T) {
	// Prepare invalid torrent data.
	data := []byte("invalid torrent data")

	// Use bytes.NewReader to create an io.Reader for the test data.
	reader := bytes.NewReader(data)

	// Attempt to parse the invalid torrent data.
	_, err := ParseTorrent(reader)
	assert.Error(t, err)
}
