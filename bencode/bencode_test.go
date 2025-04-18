package bencode

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBencodeString(t *testing.T) {
	// Test a simple string: "1:a"
	remaining, val, err := ParseBencode([]byte("1:a"))
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), remaining)
	bs, ok := val.(*BString)
	assert.True(t, ok, "expected *BString")
	assert.Equal(t, []byte("a"), bs.Value)

	// Test an empty string: "0:"
	remaining, val, err = ParseBencode([]byte("0:"))
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), remaining)
	bs, ok = val.(*BString)
	assert.True(t, ok, "expected *BString")
	assert.Equal(t, []byte(""), bs.Value)

	// Syntax errors in the input string
	remaining, val, err = ParseBencode([]byte("x2:abc"))
	assert.Error(t, err)
	assert.Equal(t, []byte("x2:abc"), remaining)
	assert.Nil(t, val)

	remaining, val, err = ParseBencode([]byte("2!:abc"))
	assert.Error(t, err)
	assert.Equal(t, []byte("2!:abc"), remaining)
	assert.Nil(t, val)

	remaining, val, err = ParseBencode([]byte("10$:a"))
	assert.Error(t, err)
	assert.Equal(t, []byte("10$:a"), remaining)
	assert.Nil(t, val)

	// Test string with a colon prefix: "2::a"
	remaining, val, err = ParseBencode([]byte("2::a"))
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), remaining)
	bs, ok = val.(*BString)
	assert.True(t, ok)
	assert.Equal(t, []byte(":a"), bs.Value)

	// Leading zeros in the length part: "02::a"
	remaining, val, err = ParseBencode([]byte("02::a"))
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), remaining)
	bs, ok = val.(*BString)
	assert.True(t, ok)
	assert.Equal(t, []byte(":a"), bs.Value)

	// Negative length should fail.
	remaining, val, err = ParseBencode([]byte("-1:a"))
	assert.Error(t, err)
	assert.Equal(t, []byte("-1:a"), remaining)
	assert.Nil(t, val)

	remaining, val, err = ParseBencode([]byte("-2:ab"))
	assert.Error(t, err)
	assert.Equal(t, []byte("-2:ab"), remaining)
	assert.Nil(t, val)

	// Test a long string.
	remaining, val, err = ParseBencode([]byte("12:123456789012"))
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), remaining)
	bs, ok = val.(*BString)
	assert.True(t, ok)
	assert.Equal(t, []byte("123456789012"), bs.Value)

	// Test a long random string.
	s := "abcwifieeirwjrwriwruvsfjkadfjieqie83e19jr29rj2rjofjafdmqdiqdhquhdusdks><odjwiereir::sidsifq0eee}}][p"
	encoded := fmt.Sprintf("%d:%s", len(s), s)
	remaining, val, err = ParseBencode([]byte(encoded))
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), remaining)
	bs, ok = val.(*BString)
	assert.True(t, ok)
	assert.Equal(t, []byte(s), bs.Value)
}

func TestParseBencodeInt(t *testing.T) {
	// Test a positive integer.
	remaining, val, err := ParseBencode([]byte("i123e"))
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), remaining)
	bi, ok := val.(*BInt)
	assert.True(t, ok, "expected *BInt")
	assert.Equal(t, int64(123), bi.Value)

	// Test zero.
	remaining, val, err = ParseBencode([]byte("i0e"))
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), remaining)
	bi, ok = val.(*BInt)
	assert.True(t, ok)
	assert.Equal(t, int64(0), bi.Value)

	// Leading zeros or invalid formatting should fail.
	remaining, val, err = ParseBencode([]byte("i-0e"))
	assert.Error(t, err)
	assert.Equal(t, []byte("i-0e"), remaining)
	assert.Nil(t, val)

	remaining, val, err = ParseBencode([]byte("i00e"))
	assert.Error(t, err)
	assert.Equal(t, []byte("i00e"), remaining)
	assert.Nil(t, val)

	// Syntax errors.
	remaining, val, err = ParseBencode([]byte("i122d"))
	assert.Error(t, err)
	assert.Equal(t, []byte("i122d"), remaining)
	assert.Nil(t, val)

	remaining, val, err = ParseBencode([]byte("ie"))
	assert.Error(t, err)
	assert.Equal(t, []byte("ie"), remaining)
	assert.Nil(t, val)

	remaining, val, err = ParseBencode([]byte("i+e"))
	assert.Error(t, err)
	assert.Nil(t, val)

	remaining, val, err = ParseBencode([]byte("i+0e"))
	assert.Error(t, err)
	assert.Nil(t, val)

	remaining, val, err = ParseBencode([]byte("i0.e"))
	assert.Error(t, err)
	assert.Nil(t, val)

	remaining, val, err = ParseBencode([]byte("i-1.0e"))
	assert.Error(t, err)
	assert.Nil(t, val)

	// Valid negative integer.
	remaining, val, err = ParseBencode([]byte("i-12e"))
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), remaining)
	bi, ok = val.(*BInt)
	assert.True(t, ok)
	assert.Equal(t, int64(-12), bi.Value)

	remaining, val, err = ParseBencode([]byte("i99839e"))
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), remaining)
	bi, ok = val.(*BInt)
	assert.True(t, ok)
	assert.Equal(t, int64(99839), bi.Value)

	remaining, val, err = ParseBencode([]byte("i-99839e"))
	bi, ok = val.(*BInt)
	assert.True(t, ok)
	assert.Equal(t, int64(-99839), bi.Value)
}

func TestParseSimpleBencodeList(t *testing.T) {
	// Test a simple list: li1ei2e3:abce
	remaining, val, err := ParseBencode([]byte("li1ei2e3:abce"))
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), remaining)

	bl, ok := val.(*BList)
	assert.True(t, ok, "expected *BList")
	assert.Equal(t, 3, len(bl.Values))

	// Validate first element (integer 1).
	bi, ok := bl.Values[0].(*BInt)
	assert.True(t, ok)
	assert.Equal(t, int64(1), bi.Value)

	// Validate second element (integer 2).
	bi, ok = bl.Values[1].(*BInt)
	assert.True(t, ok)
	assert.Equal(t, int64(2), bi.Value)

	// Validate third element (string "abc").
	bs, ok := bl.Values[2].(*BString)
	assert.True(t, ok)
	assert.Equal(t, []byte("abc"), bs.Value)
}

func TestParseComplexBencodeList(t *testing.T) {
	// Test a complex list: li1ei2e3:abcli3ei4e2:abee
	remaining, val, err := ParseBencode([]byte("li1ei2e3:abcli3ei4e2:abee"))
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), remaining)

	bl, ok := val.(*BList)
	assert.True(t, ok)
	assert.Equal(t, 4, len(bl.Values))

	// The fourth element should be a sublist.
	sublist, ok := bl.Values[3].(*BList)
	assert.True(t, ok)
	assert.Equal(t, 3, len(sublist.Values))

	bi, ok := sublist.Values[0].(*BInt)
	assert.True(t, ok)
	assert.Equal(t, int64(3), bi.Value)

	bi, ok = sublist.Values[1].(*BInt)
	assert.True(t, ok)
	assert.Equal(t, int64(4), bi.Value)

	bs, ok := sublist.Values[2].(*BString)
	assert.True(t, ok)
	assert.Equal(t, []byte("ab"), bs.Value)
}

func TestParseDictionarySuccess(t *testing.T) {
	expectedDicts := map[string]map[string]any{
		"de":                {},
		"d1:a1:be":          {"a": "b"},
		"d2:abi3ee":         {"ab": int64(3)},
		"d2:abli1ei2ei3eee": {"ab": []any{int64(1), int64(2), int64(3)}},
	}

	for bencodeStr, expected := range expectedDicts {
		remaining, val, err := ParseBencode([]byte(bencodeStr))
		assert.NoError(t, err)
		assert.Equal(t, []byte(""), remaining)

		bd, ok := val.(*BDict)
		assert.True(t, ok, "expected *BDict")

		// Convert bd.Dict to a map[string]any by extracting underlying values.
		converted := make(map[string]any)
		for key, bval := range bd.Dict {
			switch v := bval.(type) {
			case *BString:
				converted[key] = string(v.Value)
			case *BInt:
				converted[key] = v.Value
			case *BList:
				var list []any
				for _, elem := range v.Values {
					switch x := elem.(type) {
					case *BString:
						list = append(list, string(x.Value))
					case *BInt:
						list = append(list, x.Value)
					default:
						list = append(list, x)
					}
				}
				converted[key] = list
			default:
				converted[key] = v
			}
		}
		assert.True(t, reflect.DeepEqual(converted, expected))
	}
}

func TestParseDictionaryFail(t *testing.T) {
	invalidDicts := []string{
		"di3ei4e",         // Key is not a string.
		"d1:a1:b",         // Missing 'e' to end the dictionary.
		"d1:a",            // Missing value for the key.
		"d1:a1:b1:c",      // Missing 'e' to end the dictionary.
		"di32e1:b1:ci3ee", // Contains a non-string key.
		"d1:a1:bd",        // Nested dictionary without proper end.
	}

	for _, bencodeStr := range invalidDicts {
		_, val, err := ParseBencode([]byte(bencodeStr))
		assert.Error(t, err)
		assert.Nil(t, val)
	}
}
