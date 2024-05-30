package parser

import (
  "errors"
  "fmt"
  "strconv"
)

// Take a bencode string as input, return:
// first value: the remaining string that is not parsed yet. If there is an error that occurs, the
// first value is the input value.
// second value: parsed value, the type is one of the following: 64-bit integer, list, dictionary,
// string.
// third value: error, this error value is used for debugging.
func ParseBencode(data string) (string, any, error) {
  if data[0] == 'i' {
    // parse integer.
    // The bencoded integer has the following format: i<int>e.

    // assume that 'i' is consumed, the scanner starts with the second character.
    i := 1

    // the starting position of the integer
    start := i

    if data[i] == '-' {
      i++
      if i == len(data) || data[i] == '0' {
      return data, nil, errors.New(fmt.Sprintf("Bad payload: Failed to parse integer, at %s...", data[start:i]))
      }
    }

    // Check for leading zeros.
    if data[i] == '0' && (i == len(data) - 1 || data[i + 1] != 'e'){
      return data, nil, errors.New(fmt.Sprintf("Bad payload: Failed to parse integer, at %s...", data[start:i]))
    }

    // Consume digits.
    for i < len(data) && data[i] >= '0' && data[i] <= '9' {
      i++;
    }

    if i == len(data) || data[i] != 'e' {
      return data, nil, errors.New(fmt.Sprintf("Bad payload: Failed to parse integer, at %s...", data[start:i]))
    }

    val, parseErr := strconv.ParseInt(data[start:i], 10, 64)
    if (parseErr != nil) {
      return data, nil, errors.New(fmt.Sprintf("Bad payload: Failed to parse integer, at %s...", data[start:i]))
    }

    // Consume 'e'
    i++
    return data[i:], val, nil
  } else if data[0] >= '0' && data[0] <= '9' {
    // Parse string.
    // The bencoded string has the following format: <length>:<str>
    i := 0
    start := i
    for data[i] != ':' {
      i++
    }

    strLen, parseErr := strconv.ParseInt(data[start:i], 10, 32)

    if parseErr != nil {
      return data, nil, errors.New("Bad payload: expect string's length in the bencoded string.")
    }

    // consume ':'
    i++
    stringStart := i

    // consume string's characters.
    for j := 0; int64(j) < strLen; j++ {
      i++
    }

    return data[i:], data[stringStart:i], nil
  } else if (data[0] == 'l') {
    // Parse list.
    // The bencoded list has the following format: l<bencoded values>e

    var list []any

    // 'l' is ignored
    remaining := data[1:]

    for remaining[0] != 'e' {
      var (
	val any
	err error
      )

      remaining, val, err = ParseBencode(remaining)
      if err != nil {
	return data, nil, errors.New("Bad payload: Cannot parse the bencoded list's value.")
      }
      list = append(list, val)
    }

    if (remaining[0] != 'e') {
      return data, nil, errors.New("Bad payload: bencoded list must be terminated with an 'e' character.")
    }

    return remaining[1:], list, nil
  } else if (data[0] == 'd') {
    dict := make(map[any]any)

    // delimiter is ignored
    remaining := data[1:]

    for remaining[0] != 'e' {
      var (
	key, value any
	err error
      )

      remaining, key, err = ParseBencode(remaining)
      if err != nil {
	return data, nil, err
      }
      remaining, value, err = ParseBencode(remaining)
      if err != nil {
	return data, nil, err
      }

      dict[key] = value
    }

    return remaining, dict, nil
  } else {
    return data, nil, errors.New(fmt.Sprintf("Bad payload: Unexpected character '%s'", string(data[0])))
  }
}
