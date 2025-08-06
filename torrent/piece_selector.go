package torrent

import (
	"github.com/dpnam2112/bittorrent-client/common"
)

type PieceDownloadingStat struct {
	Downloaded bool
	Availability int
}

// Select the next n pieces to be downloaded.
// The implementations should only return pieces that are not downloaded.
// If all pieces are downloaded, the result should be an empty array.
type PieceSelector interface {
	Select(n int) []common.PieceIndex

	// If all pieces are already downloaded, this method should return nil.
	SelectOne() (common.PieceIndex, error)
}
