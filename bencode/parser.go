package bencode

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

// ParseBencode is the entry point for parsing a bencoded byte slice.
// It returns the remaining bytes, the parsed bencoded value (as a BValue), and an error if one occurs.
func ParseBencode(data []byte) ([]byte, BValue, error) {
	if len(data) == 0 {
		return data, nil, errors.New("empty input")
	}

	switch data[0] {
	case 'i':
		return parseInt(data)
	case 'l':
		return parseList(data)
	case 'd':
		return parseDict(data)
	default:
		if data[0] >= '0' && data[0] <= '9' {
			return parseString(data)
		}
		return data, nil, fmt.Errorf("bad payload: unexpected character '%c'", data[0])
	}
}

// parseInt parses a bencoded integer of the form i<int>e.
func parseInt(data []byte) ([]byte, *BInt, error) {
	if len(data) < 3 || data[0] != 'i' {
		return data, nil, errors.New("bad payload: invalid integer encoding")
	}

	i := 1 // skip 'i'
	start := i

	// Handle negative integers.
	if data[i] == '-' {
		i++
		if i >= len(data) || data[i] == '0' {
			return data, nil, fmt.Errorf("bad payload: expected valid digit after '-', at '%s...'", data[start:i])
		}
	}

	// Check for leading zeros.
	if data[i] == '0' && i+1 < len(data) && data[i+1] != 'e' {
		return data, nil, fmt.Errorf("bad payload: leading zeros in integer at '%s...'", data[start:i+2])
	}

	// Consume digits.
	for i < len(data) && data[i] >= '0' && data[i] <= '9' {
		i++
	}

	// 'e' must appear at the end of the integer bencode substring.
	if i >= len(data) || data[i] != 'e' {
		return data, nil, fmt.Errorf("bad payload: expected 'e' at end of integer, at '%s...'", data[start:i])
	}

	val, err := strconv.ParseInt(string(data[start:i]), 10, 64)
	if err != nil {
		return data, nil, fmt.Errorf("bad payload: failed to parse integer at '%s...'", data[start:i])
	}

	return data[i+1:], &BInt{Value: val}, nil
}

// parseString parses a bencoded string of the form <length>:<string>.
// Note: The string part is returned as raw bytes.
func parseString(data []byte) ([]byte, *BString, error) {
	colonIdx := bytes.IndexByte(data, ':')
	if colonIdx < 0 {
		return data, nil, errors.New("bad payload: missing ':' in string encoding")
	}

	length, err := strconv.ParseInt(string(data[:colonIdx]), 10, 32)
	if err != nil {
		return data, nil, errors.New("bad payload: failed to parse string length")
	}

	start := colonIdx + 1
	if start+int(length) > len(data) {
		return data, nil, errors.New("bad payload: string length exceeds available data")
	}

	str := data[start : start+int(length)]
	return data[start+int(length):], &BString{Value: str}, nil
}

// parseList parses a bencoded list of the form l<bencoded values>e.
func parseList(data []byte) ([]byte, *BList, error) {
	if len(data) < 2 || data[0] != 'l' {
		return data, nil, errors.New("bad payload: invalid list encoding")
	}

	var values []BValue
	remaining := data[1:]
	var (
		val BValue
		err error
	)

	for len(remaining) > 0 && remaining[0] != 'e' {
		remaining, val, err = ParseBencode(remaining)
		if err != nil {
			return data, nil, errors.New("bad payload: cannot parse list element")
		}
		values = append(values, val)
	}

	if len(remaining) == 0 || remaining[0] != 'e' {
		return data, nil, errors.New("bad payload: list not terminated with 'e'")
	}

	return remaining[1:], &BList{Values: values}, nil
}

// parseDict parses a bencoded dictionary of the form d<bencoded pairs>e.
func parseDict(data []byte) ([]byte, *BDict, error) {
	if len(data) < 2 || data[0] != 'd' {
		return data, nil, errors.New("bad payload: invalid dictionary encoding")
	}

	dict := make(map[string]BValue)
	remaining := data[1:]

	for len(remaining) != 0 && remaining[0] != 'e' {
		var (
			keyVal *BString
			val    BValue
			err    error
		)

		remaining, keyVal, err = parseString(remaining)
		if err != nil {
			return data, nil, err
		}

		// Use the string value of the key bytes.
		keyStr := string(keyVal.Value)

		remaining, val, err = ParseBencode(remaining)
		if err != nil {
			return data, nil, err
		}

		dict[keyStr] = val
	}

	if len(remaining) == 0 {
		return data, nil, errors.New("bad payload: dictionary not terminated with 'e'")
	}

	return remaining[1:], &BDict{Dict: dict}, nil
}
