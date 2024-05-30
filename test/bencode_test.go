package test

import (
	"fmt"
	"testing"

	"github.com/dpnam2112/bittorrent-client/parser"
	"github.com/stretchr/testify/assert"
)

func TestParseBencodeString(t *testing.T) {
  remaining, val, err := parser.ParseBencode("1:a")
  assert.Equal(t, "", remaining)
  assert.Equal(t, "a", val)
  assert.NoError(t, err)

  remaining, val, err = parser.ParseBencode("0:")
  assert.Equal(t, "", remaining)
  assert.Equal(t, "", val)
  assert.NoError(t, err)

  // Syntax error in the input string
  remaining, val, err = parser.ParseBencode("x2:abc")
  assert.Error(t, err)
  assert.Equal(t, "x2:abc", remaining)
  assert.Equal(t, val, nil)

  remaining, val, err = parser.ParseBencode("2!:abc")
  assert.Error(t, err)
  assert.Equal(t, "2!:abc", remaining)
  assert.Equal(t, nil, val)

  remaining, val, err = parser.ParseBencode("10$:a")
  assert.Error(t, err)
  assert.Equal(t, "10$:a", remaining)
  assert.Equal(t, nil, val)

  remaining, val, err = parser.ParseBencode("2::a")
  assert.NoError(t, err)
  assert.Equal(t, remaining, "")
  assert.Equal(t, ":a", val)

  // leading zeros in the length part
  remaining, val, err = parser.ParseBencode("02::a")
  assert.NoError(t, err)
  assert.Equal(t, "", remaining)
  assert.Equal(t, ":a", val)

  // Negative length
  remaining, val, err = parser.ParseBencode("-1:a")
  assert.Error(t, err)
  assert.Equal(t, "-1:a", remaining)
  assert.Equal(t, nil, val)

  remaining, val, err = parser.ParseBencode("-2:ab")
  assert.Error(t, err)
  assert.Equal(t, "-2:ab", remaining)
  assert.Equal(t, nil, val)

  // Long string
  remaining, val, err = parser.ParseBencode("12:123456789012")
  assert.NoError(t, err)
  assert.Equal(t, "", remaining)
  assert.Equal(t, "123456789012", val)

  s := "abcwifieeirwjrwriwruvsfjkadfjieqie83e19jr29rj2rjofjafdmqdiqdhquhdusdks><odjwiereir::sidsifq0eee}}][p"
  remaining, val, err =  parser.ParseBencode(fmt.Sprintf("%d:%s", len(s), s))
  assert.NoError(t, err)
  assert.Equal(t, "", remaining)
  assert.Equal(t, s, val)
}

func TestParseBencodeInt(t *testing.T) {
  remaining, val, err := parser.ParseBencode("i123e")
  assert.Equal(t, "", remaining)
  assert.Equal(t, int64(123), val)
  assert.NoError(t, err)

  remaining, val, err = parser.ParseBencode("i0e")
  assert.Equal(t, "", remaining)
  assert.Equal(t, int64(0), val)
  assert.NoError(t, err)

  // Leading zeros
  remaining, val, err = parser.ParseBencode("i-0e")
  assert.Equal(t, "i-0e", remaining)
  assert.Equal(t, nil, val)
  assert.Error(t, err)

  remaining, val, err = parser.ParseBencode("i00e")
  assert.Equal(t, "i00e", remaining)
  assert.Equal(t, nil, val)
  assert.Error(t, err)

  // Syntax error
  remaining, val, err = parser.ParseBencode("i122d")
  assert.Equal(t, "i122d", remaining)
  assert.Equal(t, nil, val)
  assert.Error(t, err)

  remaining, val, err = parser.ParseBencode("ie")
  assert.Equal(t, "ie", remaining)
  assert.Equal(t, nil, val)
  assert.Error(t, err)

  remaining, val, err = parser.ParseBencode("i+e")
  assert.Equal(t, nil, val)

  remaining, val, err = parser.ParseBencode("i+0e")
  assert.Equal(t, nil, val)

  remaining, val, err = parser.ParseBencode("i0.e")
  assert.Equal(t, nil, val)

  remaining, val, err = parser.ParseBencode("i-1.0e")
  assert.Equal(t, nil, val)

  // valid testcases
  remaining, val, err = parser.ParseBencode("i-12e")
  assert.Equal(t, "", remaining)
  assert.Equal(t, int64(-12), val)
  assert.NoError(t, err)

  remaining, val, err = parser.ParseBencode("i99839e")
  assert.Equal(t, "", remaining)
  assert.Equal(t, int64(99839), val)

  remaining, val, err = parser.ParseBencode("i-99839e")
  assert.Equal(t, int64(-99839), val)
}
