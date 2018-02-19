package msgpack

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestRoundTrip(t *testing.T) {
	bigNumber := &big.Float{}
	bigNumber.Parse("9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 10)

	tests := []struct {
		Value cty.Value
		Type  cty.Type
	}{

		{
			cty.StringVal("hello"),
			cty.String,
		},
		{
			cty.StringVal(""),
			cty.String,
		},
		{
			cty.NullVal(cty.String),
			cty.String,
		},
		{
			cty.UnknownVal(cty.String),
			cty.String,
		},

		{
			cty.True,
			cty.Bool,
		},
		{
			cty.False,
			cty.Bool,
		},
		{
			cty.NullVal(cty.Bool),
			cty.Bool,
		},
		{
			cty.UnknownVal(cty.Bool),
			cty.Bool,
		},

		{
			cty.NumberIntVal(1),
			cty.Number,
		},
		{
			cty.NumberFloatVal(1.5),
			cty.Number,
		},
		{
			cty.NumberVal(bigNumber),
			cty.Number,
		},
		{
			cty.PositiveInfinity,
			cty.Number,
		},
		{
			cty.NegativeInfinity,
			cty.Number,
		},

		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			cty.List(cty.String),
		},
		{
			cty.ListVal([]cty.Value{
				cty.UnknownVal(cty.String),
			}),
			cty.List(cty.String),
		},
		{
			cty.ListVal([]cty.Value{
				cty.NullVal(cty.String),
			}),
			cty.List(cty.String),
		},
		{
			cty.NullVal(cty.List(cty.String)),
			cty.List(cty.String),
		},
		{
			cty.ListValEmpty(cty.String),
			cty.List(cty.String),
		},

		{
			cty.SetVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			cty.Set(cty.String),
		},
		{
			cty.SetVal([]cty.Value{
				cty.UnknownVal(cty.String),
			}),
			cty.Set(cty.String),
		},
		{
			cty.SetVal([]cty.Value{
				cty.NullVal(cty.String),
			}),
			cty.Set(cty.String),
		},
		{
			cty.SetValEmpty(cty.String),
			cty.Set(cty.String),
		},

		{
			cty.MapVal(map[string]cty.Value{
				"greeting": cty.StringVal("hello"),
			}),
			cty.Map(cty.String),
		},
		{
			cty.MapVal(map[string]cty.Value{
				"greeting": cty.UnknownVal(cty.String),
			}),
			cty.Map(cty.String),
		},
		{
			cty.MapVal(map[string]cty.Value{
				"greeting": cty.NullVal(cty.String),
			}),
			cty.Map(cty.String),
		},
		{
			cty.MapValEmpty(cty.String),
			cty.Map(cty.String),
		},

		{
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			cty.Tuple([]cty.Type{cty.String}),
		},
		{
			cty.TupleVal([]cty.Value{
				cty.UnknownVal(cty.String),
			}),
			cty.Tuple([]cty.Type{cty.String}),
		},
		{
			cty.TupleVal([]cty.Value{
				cty.NullVal(cty.String),
			}),
			cty.Tuple([]cty.Type{cty.String}),
		},
		{
			cty.EmptyTupleVal,
			cty.EmptyTuple,
		},

		{
			cty.ObjectVal(map[string]cty.Value{
				"greeting": cty.StringVal("hello"),
			}),
			cty.Object(map[string]cty.Type{
				"greeting": cty.String,
			}),
		},
		{
			cty.ObjectVal(map[string]cty.Value{
				"greeting": cty.UnknownVal(cty.String),
			}),
			cty.Object(map[string]cty.Type{
				"greeting": cty.String,
			}),
		},
		{
			cty.ObjectVal(map[string]cty.Value{
				"greeting": cty.NullVal(cty.String),
			}),
			cty.Object(map[string]cty.Type{
				"greeting": cty.String,
			}),
		},
		{
			cty.EmptyObjectVal,
			cty.EmptyObject,
		},

		{
			cty.NullVal(cty.String),
			cty.DynamicPseudoType,
		},
		{
			cty.DynamicVal,
			cty.DynamicPseudoType,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			cty.List(cty.DynamicPseudoType),
		},
		{
			cty.ListVal([]cty.Value{
				cty.NullVal(cty.String),
			}),
			cty.List(cty.DynamicPseudoType),
		},
		{
			cty.ListVal([]cty.Value{
				cty.DynamicVal,
			}),
			cty.List(cty.DynamicPseudoType),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v as %#v", test.Value, test.Type), func(t *testing.T) {
			b, err := Marshal(test.Value, test.Type)
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("encoded as %x", b)

			got, err := Unmarshal(b, test.Type)
			if err != nil {
				t.Fatal(err)
			}

			if !got.RawEquals(test.Value) {
				t.Errorf(
					"value did not round-trip\ninput:  %#v\nresult: %#v",
					test.Value, got,
				)
			}
		})
	}
}
