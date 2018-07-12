package msgpack

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestImpliedType(t *testing.T) {
	tests := []struct {
		Input string
		Want  cty.Type
	}{
		{
			"\xc0",
			cty.DynamicPseudoType,
		},
		{
			"\x01", // positive fixnum
			cty.Number,
		},
		{
			"\xff", // negative fixnum
			cty.Number,
		},
		{
			"\xcc\x04", // uint8
			cty.Number,
		},
		{
			"\xcd\x00\x04", // uint16
			cty.Number,
		},
		{
			"\xce\x00\x04\x02\x01", // uint32
			cty.Number,
		},
		{
			"\xcf\x00\x04\x02\x01\x00\x04\x02\x01", // uint64
			cty.Number,
		},
		{
			"\xd0\x04", // int8
			cty.Number,
		},
		{
			"\xd1\x00\x04", // int16
			cty.Number,
		},
		{
			"\xd2\x00\x04\x02\x01", // int32
			cty.Number,
		},
		{
			"\xd3\x00\x04\x02\x01\x00\x04\x02\x01", // int64
			cty.Number,
		},
		{
			"\xca\x01\x01\x01\x01", // float32
			cty.Number,
		},
		{
			"\xcb\x01\x01\x01\x01\x01\x01\x01\x01", // float64
			cty.Number,
		},
		{
			"\xd4\x00\x00", // fixext1 (unknown value)
			cty.DynamicPseudoType,
		},
		{
			"\xd5\x00\x00\x00", // fixext2 (unknown value)
			cty.DynamicPseudoType,
		},
		{
			"\xa0", // fixstr (length zero)
			cty.String,
		},
		{
			"\xa1\xff", // fixstr (length one)
			cty.String,
		},
		{
			"\xd9\x00", // str8 (length zero)
			cty.String,
		},
		{
			"\xd9\x01\xff", // str8 (length one)
			cty.String,
		},
		{
			"\xda\x00\x00", // str16 (length zero)
			cty.String,
		},
		{
			"\xda\x00\x01\xff", // str16 (length one)
			cty.String,
		},
		{
			"\xdb\x00\x00\x00\x00", // str32 (length zero)
			cty.String,
		},
		{
			"\xdb\x00\x00\x00\x01\xff", // str32 (length one)
			cty.String,
		},
		{
			"\xc2", // false
			cty.Bool,
		},
		{
			"\xc3", // true
			cty.Bool,
		},
		{
			"\x90", // fixarray (length zero)
			cty.EmptyTuple,
		},
		{
			"\x91\xa0", // fixarray (length one, element is empty string)
			cty.Tuple([]cty.Type{cty.String}),
		},
		{
			"\xdc\x00\x00", // array16 (length zero)
			cty.EmptyTuple,
		},
		{
			"\xdc\x00\x01\xc2", // array16 (length one, element is bool)
			cty.Tuple([]cty.Type{cty.Bool}),
		},
		{
			"\xdd\x00\x00\x00\x00", // array32 (length zero)
			cty.EmptyTuple,
		},
		{
			"\xdd\x00\x00\x00\x01\xc2", // array32 (length one, element is bool)
			cty.Tuple([]cty.Type{cty.Bool}),
		},
		{
			"\x80", // fixmap (length zero)
			cty.EmptyObject,
		},
		{
			"\x81\xa1a\xc2", // fixmap (length one, "a" => bool)
			cty.Object(map[string]cty.Type{"a": cty.Bool}),
		},
		{
			"\xde\x00\x00", // map16 (length zero)
			cty.EmptyObject,
		},
		{
			"\xde\x00\x01\xa1a\xc2", // map16 (length one, "a" => bool)
			cty.Object(map[string]cty.Type{"a": cty.Bool}),
		},
		{
			"\xdf\x00\x00\x00\x00", // map32 (length zero)
			cty.EmptyObject,
		},
		{
			"\xdf\x00\x00\x00\x01\xa1a\xc2", // map32 (length one, "a" => bool)
			cty.Object(map[string]cty.Type{"a": cty.Bool}),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%x", test.Input), func(t *testing.T) {
			got, err := ImpliedType([]byte(test.Input))

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.Equals(test.Want) {
				t.Errorf(
					"wrong type\ninput: %q\ngot:   %#v\nwant:  %#v",
					test.Input, got, test.Want,
				)
			}
		})
	}
}
