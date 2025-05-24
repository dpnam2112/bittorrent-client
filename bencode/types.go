package bencode

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

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

// BString represents a bencoded string (raw bytes).
type BString struct {
	Value []byte
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
	raw  []byte
}

// GetRaw returns raw bencode representation of the dictionary.
func (v *BDict) GetRawBencode() []byte {
	raw := make([]byte, len(v.raw))
	copy(raw, v.raw)
	return raw
}

func (b *BDict) GetType() BencodeType {
	return BencodeDict
}

func BValueToString(v BValue, indent int) string {
	var buf bytes.Buffer
	pad := bytes.Repeat([]byte("  "), indent)

	switch val := v.(type) {
	case *BInt:
		fmt.Fprintf(&buf, "%s[Int] %d\n", pad, val.Value)

	case *BString:
		const maxLen = 20
		display := val.Value
		truncated := false
		if len(val.Value) > maxLen {
			display = val.Value[:maxLen]
			truncated = true
		}

		if isPrintable(display) {
			fmt.Fprintf(&buf, "%s[String] \"%s\"", pad, display)
		} else {
			fmt.Fprintf(&buf, "%s[String] 0x%s", pad, hex.EncodeToString(display))
		}
		if truncated {
			fmt.Fprintf(&buf, "... (truncated)")
		}
		buf.WriteByte('\n')

	case *BList:
		fmt.Fprintf(&buf, "%s[List] (\n", pad)
		for _, item := range val.Values {
			buf.WriteString(BValueToString(item, indent+1))
		}
		fmt.Fprintf(&buf, "%s)\n", pad)

	case *BDict:
		fmt.Fprintf(&buf, "%s[Dict] {\n", pad)
		for key, value := range val.Dict {
			fmt.Fprintf(&buf, "%s  Key: \"%s\"\n", pad, key)
			buf.WriteString(BValueToString(value, indent+1))
		}
		fmt.Fprintf(&buf, "%s}\n", pad)

	default:
		fmt.Fprintf(&buf, "%s[Unknown Type]\n", pad)
	}

	return buf.String()
}

func isPrintable(data []byte) bool {
	for _, b := range data {
		if b < 32 || b > 126 {
			return false
		}
	}
	return true
}
