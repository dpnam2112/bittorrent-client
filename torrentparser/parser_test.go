package torrentparser

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSingleFileTorrent(t *testing.T) {
	data := []byte("d8:announce18:http://tracker.com4:infod4:name12:testfile.txt12:piece lengthi524288e6:pieces20:aaaaaaaaaaaaaaaaaaaaee")
	reader := bytes.NewReader(data)

	torrent, err := ParseTorrent(reader)
	assert.NoError(t, err)

	assert.Equal(t, "http://tracker.com", torrent.Announce())
	assert.Equal(t, "testfile.txt", torrent.Info().Name())
	assert.Equal(t, int64(524288), torrent.Info().PieceLength())
	assert.Equal(t, 20, len(torrent.Info().Pieces()))

	fmt.Println(string(torrent.Info().rawBencode))

	infoDictBencode := "d4:name12:testfile.txt12:piece lengthi524288e6:pieces20:aaaaaaaaaaaaaaaaaaaae"

	assert.True(t, reflect.DeepEqual(torrent.Info().rawBencode, []byte(infoDictBencode)))
	assert.True(t, reflect.DeepEqual(sha1.Sum([]byte(infoDictBencode)), torrent.Info().Hash()))
}

func TestParseMultiFileTorrent(t *testing.T) {
	data := []byte("d8:announce18:http://tracker.com4:infod4:name7:myfiles12:piece lengthi524288e6:pieces40:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb5:filesld6:lengthi12345e4:pathl9:file1.txteed6:lengthi67890e4:pathl9:file2.txteeeee")
	reader := bytes.NewReader(data)

	torrent, err := ParseTorrent(reader)
	assert.NoError(t, err)

	info := torrent.Info()

	assert.Equal(t, "http://tracker.com", torrent.Announce())
	assert.Equal(t, "myfiles", info.Name())
	assert.Equal(t, int64(524288), info.PieceLength())
	assert.Equal(t, 40, len(info.Pieces()))

	files := info.Files()
	assert.Len(t, files, 2)

	assert.Equal(t, int64(12345), files[0].Length())
	assert.Equal(t, []string{"file1.txt"}, files[0].Path())

	assert.Equal(t, int64(67890), files[1].Length())
	assert.Equal(t, []string{"file2.txt"}, files[1].Path())
}

func TestParseInvalidTorrent(t *testing.T) {
	data := []byte("invalid torrent data")
	reader := bytes.NewReader(data)

	_, err := ParseTorrent(reader)
	assert.Error(t, err)
}
