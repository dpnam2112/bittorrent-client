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

func TestParseSimpleBencodeList(t *testing.T) {
  remaining, val, _ := parser.ParseBencode("li1ei2e3:abce")

  list, isList := val.([]any)

  assert.Equal(t, "", remaining)
  assert.True(t, isList)
  assert.Equal(t, 3, len(list))

  ele0, correctType0 := list[0].(int64)
  assert.True(t, correctType0)
  assert.Equal(t, int64(1), ele0)

  ele1, correctType1 := list[1].(int64)
  assert.True(t, correctType1)
  assert.Equal(t, int64(2), ele1)

  ele2, correctType2 := list[2].(string)
  assert.True(t, correctType2)
  assert.Equal(t, "abc", ele2)
}

func TestParseComplexBencodeList(t *testing.T) {
  remaining, val, _ := parser.ParseBencode("li1ei2e3:abcli3ei4e2:abee")

  list, isList := val.([]any)

  assert.Equal(t, "", remaining)
  assert.True(t, isList)
  assert.Equal(t, 4, len(list))

  sublist, isList2 := list[3].([]any)
  
  assert.True(t, isList2)
  assert.Equal(t, 3, len(sublist))

  sublistEle0, isNum0 := sublist[0].(int64)
  assert.Equal(t, int64(3), sublistEle0)
  assert.True(t, isNum0)

  sublistEle1, isNum1 := sublist[1].(int64)
  assert.Equal(t, int64(4), sublistEle1)
  assert.True(t, isNum1)

  sublistEle2, isString2 := sublist[2].(string)
  assert.Equal(t, "ab", sublistEle2)
  assert.True(t, isString2)
}

func TestParseDictionary(t *testing.T) {

}
