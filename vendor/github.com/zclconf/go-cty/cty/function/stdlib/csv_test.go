package stdlib

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestCSVDecode(t *testing.T) {
	tests := []struct {
		Input   cty.Value
		Want    cty.Value
		WantErr string
	}{
		{
			cty.StringVal(csvTest),
			cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"name": cty.StringVal("foo"),
					"size": cty.StringVal("100"),
					"type": cty.StringVal("tiny"),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"name": cty.StringVal("bar"),
					"size": cty.StringVal(""),
					"type": cty.StringVal("huge"),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"name": cty.StringVal("baz"),
					"size": cty.StringVal("50"),
					"type": cty.StringVal("weedy"),
				}),
			}),
			``,
		},
		{
			cty.StringVal(`"just","header","line"`),
			cty.ListValEmpty(cty.Object(map[string]cty.Type{
				"just":   cty.String,
				"header": cty.String,
				"line":   cty.String,
			})),
			``,
		},
		{
			cty.StringVal(``),
			cty.DynamicVal,
			`missing header line`,
		},
		{
			cty.StringVal(`not csv at all`),
			cty.ListValEmpty(cty.Object(map[string]cty.Type{
				"not csv at all": cty.String,
			})),
			``,
		},
		{
			cty.StringVal(`invalid"thing"`),
			cty.DynamicVal,
			`parse error on line 1, column 7: bare " in non-quoted-field`,
		},
		{
			cty.UnknownVal(cty.String),
			cty.DynamicVal, // need to know the value to determine the type
			``,
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			``,
		},
		{
			cty.True,
			cty.DynamicVal,
			`string required, but received bool`,
		},
		{
			cty.NullVal(cty.String),
			cty.DynamicVal,
			`must not be null`,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("CSVDecode(%#v)", test.Input), func(t *testing.T) {
			got, err := CSVDecode(test.Input)
			var errStr string
			if err != nil {
				errStr = err.Error()
			}
			if errStr != test.WantErr {
				t.Fatalf("wrong error\ngot:  %s\nwant: %s", errStr, test.WantErr)
			}
			if err != nil {
				return
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

const csvTest = `"name","size","type"
"foo","100","tiny"
"bar","","huge"
"baz","50","weedy"
`
