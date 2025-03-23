package bencode

// BencodeType is an enum that defines the type of bencoded value.
type BencodeType int

const (
	BencodeInt BencodeType = iota
	BencodeString
	BencodeList
	BencodeDict
)

// BValue is the interface implemented by all bencoded values.
type BValue interface {
    GetType() BencodeType
}

// BInt represents a bencoded integer.
type BInt struct {
    Value int64
}

func (b *BInt) GetType() BencodeType {
    return BencodeInt
}

// BString represents a bencoded string.
type BString struct {
    Value string
}

func (b *BString) GetType() BencodeType {
    return BencodeString
}

// BList represents a bencoded list.
type BList struct {
    Values []BValue
}

func (b *BList) GetType() BencodeType {
    return BencodeList
}

// BDict represents a bencoded dictionary.
type BDict struct {
    Dict map[string]BValue
}

func (b *BDict) GetType() BencodeType {
    return BencodeDict
}


