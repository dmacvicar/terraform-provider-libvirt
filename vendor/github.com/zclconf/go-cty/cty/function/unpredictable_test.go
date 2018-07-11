package function

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestUnpredictable(t *testing.T) {
	f := New(&Spec{
		Params: []Parameter{
			{
				Name: "fixed",
				Type: cty.Bool,
			},
		},
		VarParam: &Parameter{
			Name: "variadic",
			Type: cty.String,
		},
		Type: func(args []cty.Value) (cty.Type, error) {
			if len(args) == 1 {
				return cty.Bool, nil
			} else {
				return cty.String, nil
			}
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			return cty.NullVal(retType), nil
		},
	})

	uf := Unpredictable(f)

	{
		predVal, err := f.Call([]cty.Value{cty.True})
		if err != nil {
			t.Fatal(err)
		}
		if !predVal.RawEquals(cty.NullVal(cty.Bool)) {
			t.Fatal("wrong predictable result")
		}
	}

	t.Run("argument type error", func(t *testing.T) {
		_, err := uf.Call([]cty.Value{cty.StringVal("hello")})
		if err == nil {
			t.Fatal("call successful; want error")
		}
	})

	t.Run("type check 1", func(t *testing.T) {
		ty, err := uf.ReturnTypeForValues([]cty.Value{cty.True})
		if err != nil {
			t.Fatal(err)
		}
		if !ty.Equals(cty.Bool) {
			t.Errorf("wrong type %#v; want %#v", ty, cty.Bool)
		}
	})

	t.Run("type check 2", func(t *testing.T) {
		ty, err := uf.ReturnTypeForValues([]cty.Value{cty.True, cty.StringVal("hello")})
		if err != nil {
			t.Fatal(err)
		}
		if !ty.Equals(cty.String) {
			t.Errorf("wrong type %#v; want %#v", ty, cty.String)
		}
	})

	t.Run("call", func(t *testing.T) {
		v, err := uf.Call([]cty.Value{cty.True})
		if err != nil {
			t.Fatal(err)
		}
		if !v.RawEquals(cty.UnknownVal(cty.Bool)) {
			t.Errorf("wrong result %#v; want %#v", v, cty.UnknownVal(cty.Bool))
		}
	})

}
