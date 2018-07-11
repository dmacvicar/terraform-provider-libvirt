package stdlib

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestSetUnion(t *testing.T) {
	tests := []struct {
		Input []cty.Value
		Want  cty.Value
	}{
		{
			[]cty.Value{
				cty.SetValEmpty(cty.String),
			},
			cty.SetValEmpty(cty.String),
		},
		{
			[]cty.Value{
				cty.SetValEmpty(cty.String),
				cty.SetValEmpty(cty.String),
			},
			cty.SetValEmpty(cty.String),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.True}),
				cty.SetValEmpty(cty.String),
			},
			cty.SetVal([]cty.Value{cty.StringVal("true")}),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.True}),
				cty.SetVal([]cty.Value{cty.True}),
				cty.SetVal([]cty.Value{cty.False}),
			},
			cty.SetVal([]cty.Value{
				cty.True,
				cty.False,
			}),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.StringVal("a")}),
				cty.SetVal([]cty.Value{cty.StringVal("b")}),
				cty.SetVal([]cty.Value{cty.StringVal("b"), cty.StringVal("c")}),
			},
			cty.SetVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
				cty.StringVal("c"),
			}),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("SetUnion(%#v...)", test.Input), func(t *testing.T) {
			got, err := SetUnion(test.Input...)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestSetIntersection(t *testing.T) {
	tests := []struct {
		Input []cty.Value
		Want  cty.Value
	}{
		{
			[]cty.Value{
				cty.SetValEmpty(cty.String),
			},
			cty.SetValEmpty(cty.String),
		},
		{
			[]cty.Value{
				cty.SetValEmpty(cty.String),
				cty.SetValEmpty(cty.String),
			},
			cty.SetValEmpty(cty.String),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.True}),
				cty.SetValEmpty(cty.String),
			},
			cty.SetValEmpty(cty.String),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.True}),
				cty.SetVal([]cty.Value{cty.True, cty.False}),
				cty.SetVal([]cty.Value{cty.True, cty.False}),
			},
			cty.SetVal([]cty.Value{
				cty.True,
			}),
		},
		{
			[]cty.Value{
				cty.SetVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b")}),
				cty.SetVal([]cty.Value{cty.StringVal("b")}),
				cty.SetVal([]cty.Value{cty.StringVal("b"), cty.StringVal("c")}),
			},
			cty.SetVal([]cty.Value{
				cty.StringVal("b"),
			}),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("SetIntersection(%#v...)", test.Input), func(t *testing.T) {
			got, err := SetIntersection(test.Input...)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
