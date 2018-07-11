package stdlib

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		Format  cty.Value
		Args    []cty.Value
		Want    cty.Value
		WantErr string
	}{
		{
			cty.StringVal(""),
			nil,
			cty.StringVal(""),
			``,
		},
		{
			cty.StringVal("hello"),
			nil,
			cty.StringVal("hello"),
			``,
		},
		{
			cty.StringVal("100%% successful"),
			nil,
			cty.StringVal("100% successful"),
			``,
		},
		{
			cty.StringVal("100%%"),
			nil,
			cty.StringVal("100%"),
			``,
		},

		// Default formats
		{
			cty.StringVal("string %v"),
			[]cty.Value{cty.StringVal("hello")},
			cty.StringVal("string hello"),
			``,
		},
		{
			cty.StringVal("string %[2]v"),
			[]cty.Value{cty.True, cty.StringVal("hello")},
			cty.StringVal("string hello"),
			``,
		},
		{
			cty.StringVal("string %#v"),
			[]cty.Value{cty.StringVal("hello")},
			cty.StringVal(`string "hello"`),
			``,
		},
		{
			cty.StringVal("number %v"),
			[]cty.Value{cty.NumberIntVal(2)},
			cty.StringVal("number 2"),
			``,
		},
		{
			cty.StringVal("number %#v"),
			[]cty.Value{cty.NumberIntVal(2)},
			cty.StringVal("number 2"),
			``,
		},
		{
			cty.StringVal("bool %v"),
			[]cty.Value{cty.True},
			cty.StringVal("bool true"),
			``,
		},
		{
			cty.StringVal("bool %#v"),
			[]cty.Value{cty.True},
			cty.StringVal("bool true"),
			``,
		},
		{
			cty.StringVal("object %v"),
			[]cty.Value{cty.EmptyObjectVal},
			cty.StringVal("object {}"),
			``,
		},
		{
			cty.StringVal("tuple %v"),
			[]cty.Value{cty.EmptyTupleVal},
			cty.StringVal("tuple []"),
			``,
		},
		{
			cty.StringVal("tuple with unknown %v"),
			[]cty.Value{cty.TupleVal([]cty.Value{
				cty.UnknownVal(cty.String),
			})},
			cty.UnknownVal(cty.String),
			``,
		},
		{
			cty.StringVal("%%%v"),
			[]cty.Value{cty.False},
			cty.StringVal("%false"),
			``,
		},
		{
			cty.StringVal("%v"),
			[]cty.Value{cty.NullVal(cty.Bool)},
			cty.StringVal("null"),
			``,
		},

		// Strings
		{
			cty.StringVal("Hello, %s!"),
			[]cty.Value{cty.StringVal("Ermintrude")},
			cty.StringVal("Hello, Ermintrude!"),
			``,
		},
		{
			cty.StringVal("Hello, %[2]s!"),
			[]cty.Value{cty.StringVal("Stephen"), cty.StringVal("Ermintrude")},
			cty.StringVal("Hello, Ermintrude!"),
			``,
		},
		{
			cty.StringVal("Hello, %q... if that _is_ your real name!"),
			[]cty.Value{cty.StringVal("Ermintrude")},
			cty.StringVal(`Hello, "Ermintrude"... if that _is_ your real name!`),
			``,
		},
		{
			cty.StringVal("This statement is %s"),
			[]cty.Value{cty.False},
			cty.StringVal("This statement is false"),
			``,
		},
		{
			cty.StringVal("This statement is %q"),
			[]cty.Value{cty.False},
			cty.StringVal(`This statement is "false"`),
			``,
		},
		{
			cty.StringVal("%s"),
			[]cty.Value{cty.NullVal(cty.String)},
			cty.NilVal,
			`unsupported value for "%s" at 0: null value cannot be formatted`,
		},
		{
			cty.StringVal("%10s"),
			[]cty.Value{cty.StringVal("hello")},
			cty.StringVal(`     hello`),
			``,
		},
		{
			cty.StringVal("%-10s"),
			[]cty.Value{cty.StringVal("hello")},
			cty.StringVal(`hello     `),
			``,
		},
		{
			cty.StringVal("%4s"),
			[]cty.Value{cty.StringVal("💃🏿")},
			cty.StringVal(`   💃🏿`), // three spaces because this emoji sequence is a single grapheme cluster
			``,
		},
		{
			cty.StringVal("%-4s"),
			[]cty.Value{cty.StringVal("💃🏿")},
			cty.StringVal(`💃🏿   `), // three spaces because this emoji sequence is a single grapheme cluster
			``,
		},
		{
			cty.StringVal("%q"),
			[]cty.Value{cty.StringVal("💃🏿")},
			cty.StringVal(`"💃🏿"`),
			``,
		},
		{
			cty.StringVal("%6q"),
			[]cty.Value{cty.StringVal("💃🏿")},
			cty.StringVal(`   "💃🏿"`), // three spaces because this emoji sequence is a single grapheme cluster
			``,
		},
		{
			cty.StringVal("%-6q"),
			[]cty.Value{cty.StringVal("💃🏿")},
			cty.StringVal(`"💃🏿"   `), // three spaces because this emoji sequence is a single grapheme cluster
			``,
		},
		{
			cty.StringVal("%.2s"),
			[]cty.Value{cty.StringVal("hello")},
			cty.StringVal(`he`),
			``,
		},
		{
			cty.StringVal("%.2q"),
			[]cty.Value{cty.StringVal("hello")},
			cty.StringVal(`"he"`),
			``,
		},
		{
			cty.StringVal("%.5s"),
			[]cty.Value{cty.StringVal("日本語日本語")},
			cty.StringVal(`日本語日本`),
			``,
		},
		{
			cty.StringVal("%.1q"),
			[]cty.Value{cty.StringVal("日本語日本語")},
			cty.StringVal(`"日"`),
			``,
		},
		{
			cty.StringVal("%.10s"),
			[]cty.Value{cty.StringVal("hello")},
			cty.StringVal(`hello`),
			``,
		},
		{
			cty.StringVal("%4.2s"),
			[]cty.Value{cty.StringVal("hello")},
			cty.StringVal(`  he`),
			``,
		},
		{
			cty.StringVal("%6.2q"),
			[]cty.Value{cty.StringVal("hello")},
			cty.StringVal(`  "he"`),
			``,
		},
		{
			cty.StringVal("%-4.2s"),
			[]cty.Value{cty.StringVal("hello")},
			cty.StringVal(`he  `),
			``,
		},
		{
			cty.StringVal("%q"),
			[]cty.Value{cty.StringVal("Hello\nWorld")},
			cty.StringVal(`"Hello\nWorld"`),
			``,
		},

		// Booleans
		{
			cty.StringVal("This statement is %t"),
			[]cty.Value{cty.False},
			cty.StringVal("This statement is false"),
			``,
		},
		{
			cty.StringVal("This statement is %[2]t"),
			[]cty.Value{cty.True, cty.False},
			cty.StringVal("This statement is false"),
			``,
		},
		{
			cty.StringVal("This statement is %t"),
			[]cty.Value{cty.True},
			cty.StringVal("This statement is true"),
			``,
		},
		{
			cty.StringVal("This statement is %t"),
			[]cty.Value{cty.StringVal("false")},
			cty.StringVal("This statement is false"),
			``,
		},

		// Integer Numbers
		{
			cty.StringVal("%d green bottles standing on the wall"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.StringVal("10 green bottles standing on the wall"),
			``,
		},
		{
			cty.StringVal("%[2]d things"),
			[]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(10)},
			cty.StringVal("10 things"),
			``,
		},
		{
			cty.StringVal("%+d green bottles standing on the wall"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.StringVal("+10 green bottles standing on the wall"),
			``,
		},
		{
			cty.StringVal("% d green bottles standing on the wall"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.StringVal(" 10 green bottles standing on the wall"),
			``,
		},
		{
			cty.StringVal("%5d green bottles standing on the wall"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.StringVal("   10 green bottles standing on the wall"),
			``,
		},
		{
			cty.StringVal("%-5d green bottles standing on the wall"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.StringVal("10    green bottles standing on the wall"),
			``,
		},
		{
			cty.StringVal("%d green bottles standing on the wall"),
			[]cty.Value{cty.True},
			cty.NilVal,
			`unsupported value for "%d" at 0: number required`,
		},
		{
			cty.StringVal("%b"),
			[]cty.Value{cty.NumberIntVal(5)},
			cty.StringVal("101"),
			``,
		},
		{
			cty.StringVal("%o"),
			[]cty.Value{cty.NumberIntVal(9)},
			cty.StringVal("11"),
			``,
		},
		{
			cty.StringVal("%x"),
			[]cty.Value{cty.NumberIntVal(254)},
			cty.StringVal("fe"),
			``,
		},
		{
			cty.StringVal("%X"),
			[]cty.Value{cty.NumberIntVal(254)},
			cty.StringVal("FE"),
			``,
		},

		// Floating-point numbers
		{
			cty.StringVal("%f things"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.StringVal("10.000000 things"),
			``,
		},
		{
			cty.StringVal("%[2]f things"),
			[]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(10)},
			cty.StringVal("10.000000 things"),
			``,
		},
		{
			cty.StringVal("%+f things"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.StringVal("+10.000000 things"),
			``,
		},
		{
			cty.StringVal("% f things"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.StringVal(" 10.000000 things"),
			``,
		},
		{
			cty.StringVal("%+f things"),
			[]cty.Value{cty.NumberIntVal(-10)},
			cty.StringVal("-10.000000 things"),
			``,
		},
		{
			cty.StringVal("% f things"),
			[]cty.Value{cty.NumberIntVal(-10)},
			cty.StringVal("-10.000000 things"),
			``,
		},
		{
			cty.StringVal("%f things"),
			[]cty.Value{cty.StringVal("100000000000000000000000000000000000001")},
			cty.StringVal("100000000000000000000000000000000000001.000000 things"),
			``,
		},
		{
			cty.StringVal("%f things"),
			[]cty.Value{cty.StringVal("1.00000000000000000000000000000000000001")},
			cty.StringVal("1.000000 things"),
			``,
		},
		{
			cty.StringVal("%.4f things"),
			[]cty.Value{cty.StringVal("1.00000000000000000000000000000000000001")},
			cty.StringVal("1.0000 things"),
			``,
		},
		{
			cty.StringVal("%.1f things"),
			[]cty.Value{cty.StringVal("1.06")},
			cty.StringVal("1.1 things"),
			``,
		},
		{
			cty.StringVal("%e things"),
			[]cty.Value{cty.NumberIntVal(1000)},
			cty.StringVal("1.000000e+03 things"),
			``,
		},
		{
			cty.StringVal("%E things"),
			[]cty.Value{cty.NumberIntVal(1000)},
			cty.StringVal("1.000000E+03 things"),
			``,
		},
		{
			cty.StringVal("%g things"),
			[]cty.Value{cty.NumberIntVal(1000)},
			cty.StringVal("1000 things"),
			``,
		},
		{
			cty.StringVal("%G things"),
			[]cty.Value{cty.NumberIntVal(1000)},
			cty.StringVal("1000 things"),
			``,
		},
		{
			cty.StringVal("%g things"),
			[]cty.Value{cty.StringVal("0.00000000000000000000001")},
			cty.StringVal("1e-23 things"),
			``,
		},
		{
			cty.StringVal("%G things"),
			[]cty.Value{cty.StringVal("0.00000000000000000000001")},
			cty.StringVal("1E-23 things"),
			``,
		},

		// Unknowns
		{
			cty.UnknownVal(cty.String),
			[]cty.Value{cty.True},
			cty.UnknownVal(cty.String),
			``,
		},
		{
			cty.UnknownVal(cty.Bool),
			[]cty.Value{cty.True},
			cty.NilVal,
			`string required, but received bool`,
		},
		{
			cty.StringVal("Hello, %s!"),
			[]cty.Value{cty.UnknownVal(cty.String)},
			cty.UnknownVal(cty.String),
			``,
		},
		{
			cty.StringVal("Hello, %[2]s!"),
			[]cty.Value{cty.UnknownVal(cty.String), cty.StringVal("Ermintrude")},
			cty.UnknownVal(cty.String),
			``,
		},

		// Invalids
		{
			cty.StringVal("%s is not in the args list"),
			nil,
			cty.NilVal,
			`not enough arguments for "%s" at 0: need index 1 but have 0 total`,
		},
		{
			cty.StringVal("%[3]s is not in the args list"),
			[]cty.Value{cty.True, cty.True},
			cty.NilVal,
			`not enough arguments for "%[3]s" at 0: need index 3 but have 2 total`,
		},
		{
			cty.StringVal("%v %v %v"),
			[]cty.Value{cty.True, cty.True},
			cty.NilVal,
			`not enough arguments for "%v" at 6: need index 3 but have 2 total`,
		},
		{
			cty.StringVal("%z is not a valid sequence"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.NilVal,
			`unsupported format verb 'z' in "%z" at offset 0`,
		},
		{
			cty.StringVal("%#z is not a valid sequence"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.NilVal,
			`unsupported format verb 'z' in "%#z" at offset 0`,
		},
		{
			cty.StringVal("%012z is not a valid sequence"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.NilVal,
			`unsupported format verb 'z' in "%012z" at offset 0`,
		},
		{
			cty.StringVal("%☠ is not a valid sequence"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.NilVal,
			`unrecognized format character '☠' at offset 1`,
		},
		{
			cty.StringVal("%💃🏿 is not a valid sequence"),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.NilVal,
			`unrecognized format character '💃' at offset 1`, // since this is a grammar-level error, we don't get the full grapheme cluster
		},
		{
			cty.NullVal(cty.String),
			[]cty.Value{cty.NumberIntVal(10)},
			cty.NilVal,
			`must not be null`,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d-%#v", i, test.Format), func(t *testing.T) {
			got, err := Format(test.Format, test.Args...)

			if test.WantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
			} else {
				if err == nil {
					t.Fatalf("no error; want %q", test.WantErr)
				}
				errStr := err.Error()
				if errStr != test.WantErr {
					t.Fatalf("wrong error\ngot:  %s\nwant: %s", errStr, test.WantErr)
				}
				return
			}

			if test.Want != cty.NilVal {
				if !got.RawEquals(test.Want) {
					t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
				}
			} else {
				t.Errorf("unexpected success %#v; want error", got)
			}
		})
	}
}
func TestFormatList(t *testing.T) {
	tests := []struct {
		Format  cty.Value
		Args    []cty.Value
		Want    cty.Value
		WantErr string
	}{
		{
			cty.StringVal(""),
			nil,
			cty.ListVal([]cty.Value{
				cty.StringVal(""),
			}),
			``,
		},
		{
			cty.StringVal("hello"),
			nil,
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			``,
		},
		{
			cty.StringVal("100%% successful"),
			nil,
			cty.ListVal([]cty.Value{
				cty.StringVal("100% successful"),
			}),
			``,
		},
		{
			cty.StringVal("100%%"),
			nil,
			cty.ListVal([]cty.Value{
				cty.StringVal("100%"),
			}),
			``,
		},

		{
			cty.StringVal("%s"),
			[]cty.Value{cty.StringVal("hello")},
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			``,
		},
		{
			cty.StringVal("%s"),
			[]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("hello"),
				}),
			},
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			``,
		},
		{
			cty.StringVal("%s"),
			[]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("hello"),
					cty.StringVal("world"),
				}),
			},
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
				cty.StringVal("world"),
			}),
			``,
		},
		{
			cty.StringVal("%s %s"),
			[]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("hello"),
					cty.StringVal("goodbye"),
				}),
				cty.ListVal([]cty.Value{
					cty.StringVal("world"),
					cty.StringVal("universe"),
				}),
			},
			cty.ListVal([]cty.Value{
				cty.StringVal("hello world"),
				cty.StringVal("goodbye universe"),
			}),
			``,
		},
		{
			cty.StringVal("%s %s"),
			[]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("hello"),
					cty.StringVal("goodbye"),
				}),
				cty.StringVal("world"),
			},
			cty.ListVal([]cty.Value{
				cty.StringVal("hello world"),
				cty.StringVal("goodbye world"),
			}),
			``,
		},
		{
			cty.StringVal("%s %s"),
			[]cty.Value{
				cty.StringVal("hello"),
				cty.ListVal([]cty.Value{
					cty.StringVal("world"),
					cty.StringVal("universe"),
				}),
			},
			cty.ListVal([]cty.Value{
				cty.StringVal("hello world"),
				cty.StringVal("hello universe"),
			}),
			``,
		},
		{
			cty.StringVal("%s %s"),
			[]cty.Value{
				cty.ListVal([]cty.Value{
					cty.StringVal("hello"),
					cty.StringVal("goodbye"),
				}),
				cty.ListVal([]cty.Value{
					cty.StringVal("world"),
				}),
			},
			cty.ListValEmpty(cty.String),
			`argument 2 has length 1, which is inconsistent with argument 1 of length 2`,
		},
		{
			cty.StringVal("%s"),
			[]cty.Value{cty.EmptyObjectVal},
			cty.ListValEmpty(cty.String),
			`error on format iteration 0: unsupported value for "%s" at 0: string required`,
		},
		{
			cty.StringVal("%v"),
			[]cty.Value{cty.EmptyTupleVal},
			cty.ListValEmpty(cty.String), // no items because our given tuple is empty
			``,
		},
		{
			cty.StringVal("%v"),
			[]cty.Value{cty.NullVal(cty.List(cty.String))},
			cty.ListVal([]cty.Value{
				cty.StringVal("null"), // single item because a null list is interpreted as a single null
			}),
			``,
		},

		{
			cty.UnknownVal(cty.String),
			[]cty.Value{
				cty.True,
			},
			cty.UnknownVal(cty.List(cty.String)),
			``,
		},
		{
			cty.StringVal("%v"),
			[]cty.Value{
				cty.UnknownVal(cty.String),
			},
			cty.ListVal([]cty.Value{
				cty.UnknownVal(cty.String),
			}),
			``,
		},
		{
			cty.StringVal("%v"),
			[]cty.Value{
				cty.ListVal([]cty.Value{
					cty.TupleVal([]cty.Value{cty.StringVal("hello")}),
					cty.TupleVal([]cty.Value{cty.UnknownVal(cty.String)}),
					cty.TupleVal([]cty.Value{cty.StringVal("world")}),
				}),
			},
			cty.ListVal([]cty.Value{
				cty.StringVal(`["hello"]`),
				cty.UnknownVal(cty.String),
				cty.StringVal(`["world"]`),
			}),
			``,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d-%#v", i, test.Format), func(t *testing.T) {
			got, err := FormatList(test.Format, test.Args...)

			if test.WantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
			} else {
				if err == nil {
					t.Fatalf("no error; want %q", test.WantErr)
				}
				errStr := err.Error()
				if errStr != test.WantErr {
					t.Fatalf("wrong error\ngot:  %s\nwant: %s", errStr, test.WantErr)
				}
				return
			}

			if test.Want != cty.NilVal {
				if !got.RawEquals(test.Want) {
					t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
				}
			} else {
				t.Errorf("unexpected success %#v; want error", got)
			}
		})
	}
}
