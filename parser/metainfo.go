package parser

type MetaInfo struct {
	info     Info
	announce string
}

type Info interface {
	getInfoMode()
}

const (
	INFO_SINGLE_FILE   int = 0
	INFO_MULTIPLE_FILE int = 1
)

type BaseInfo struct {
	pieceLength int64
	pieces      string
	private     int64
}

type SingleFileInfo struct {
	fileName string
	length   int64
	md5sum   string
}

type MultiFileInfo struct {
	dirName string
	files   []struct {
		length string
		md5sum string
		path   []string
	}
}

// TODO: Implement interface function for MultiFileInfo and SingleFileInfo here...
func (info *SingleFileInfo) getInfoMode() int {
	return INFO_SINGLE_FILE
}

func (info *MultiFileInfo) getInfoMode() int {
	return INFO_MULTIPLE_FILE
}

func ConstructInfo(dict map[any]any) Info {
	return nil
}
