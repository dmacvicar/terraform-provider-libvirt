package dataurl

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

type dataURLTest struct {
	InputRawDataURL string
	ExpectedItems   []item
	ExpectedDataURL DataURL
}

func genTestTable() []dataURLTest {
	return []dataURLTest{
		dataURLTest{
			`data:;base64,aGV5YQ==`,
			[]item{
				item{itemDataPrefix, dataPrefix},
				item{itemParamSemicolon, ";"},
				item{itemBase64Enc, "base64"},
				item{itemDataComma, ","},
				item{itemData, "aGV5YQ=="},
				item{itemEOF, ""},
			},
			DataURL{
				defaultMediaType(),
				EncodingBase64,
				[]byte("heya"),
			},
		},
		dataURLTest{
			`data:text/plain;base64,aGV5YQ==`,
			[]item{
				item{itemDataPrefix, dataPrefix},
				item{itemMediaType, "text"},
				item{itemMediaSep, "/"},
				item{itemMediaSubType, "plain"},
				item{itemParamSemicolon, ";"},
				item{itemBase64Enc, "base64"},
				item{itemDataComma, ","},
				item{itemData, "aGV5YQ=="},
				item{itemEOF, ""},
			},
			DataURL{
				MediaType{
					"text",
					"plain",
					map[string]string{},
				},
				EncodingBase64,
				[]byte("heya"),
			},
		},
		dataURLTest{
			`data:text/plain;charset=utf-8;base64,aGV5YQ==`,
			[]item{
				item{itemDataPrefix, dataPrefix},
				item{itemMediaType, "text"},
				item{itemMediaSep, "/"},
				item{itemMediaSubType, "plain"},
				item{itemParamSemicolon, ";"},
				item{itemParamAttr, "charset"},
				item{itemParamEqual, "="},
				item{itemParamVal, "utf-8"},
				item{itemParamSemicolon, ";"},
				item{itemBase64Enc, "base64"},
				item{itemDataComma, ","},
				item{itemData, "aGV5YQ=="},
				item{itemEOF, ""},
			},
			DataURL{
				MediaType{
					"text",
					"plain",
					map[string]string{
						"charset": "utf-8",
					},
				},
				EncodingBase64,
				[]byte("heya"),
			},
		},
		dataURLTest{
			`data:text/plain;charset=utf-8;foo=bar;base64,aGV5YQ==`,
			[]item{
				item{itemDataPrefix, dataPrefix},
				item{itemMediaType, "text"},
				item{itemMediaSep, "/"},
				item{itemMediaSubType, "plain"},
				item{itemParamSemicolon, ";"},
				item{itemParamAttr, "charset"},
				item{itemParamEqual, "="},
				item{itemParamVal, "utf-8"},
				item{itemParamSemicolon, ";"},
				item{itemParamAttr, "foo"},
				item{itemParamEqual, "="},
				item{itemParamVal, "bar"},
				item{itemParamSemicolon, ";"},
				item{itemBase64Enc, "base64"},
				item{itemDataComma, ","},
				item{itemData, "aGV5YQ=="},
				item{itemEOF, ""},
			},
			DataURL{
				MediaType{
					"text",
					"plain",
					map[string]string{
						"charset": "utf-8",
						"foo":     "bar",
					},
				},
				EncodingBase64,
				[]byte("heya"),
			},
		},
		dataURLTest{
			`data:application/json;charset=utf-8;foo="b\"<@>\"r";style=unformatted%20json;base64,eyJtc2ciOiAiaGV5YSJ9`,
			[]item{
				item{itemDataPrefix, dataPrefix},
				item{itemMediaType, "application"},
				item{itemMediaSep, "/"},
				item{itemMediaSubType, "json"},
				item{itemParamSemicolon, ";"},
				item{itemParamAttr, "charset"},
				item{itemParamEqual, "="},
				item{itemParamVal, "utf-8"},
				item{itemParamSemicolon, ";"},
				item{itemParamAttr, "foo"},
				item{itemParamEqual, "="},
				item{itemLeftStringQuote, "\""},
				item{itemParamVal, `b\"<@>\"r`},
				item{itemRightStringQuote, "\""},
				item{itemParamSemicolon, ";"},
				item{itemParamAttr, "style"},
				item{itemParamEqual, "="},
				item{itemParamVal, "unformatted%20json"},
				item{itemParamSemicolon, ";"},
				item{itemBase64Enc, "base64"},
				item{itemDataComma, ","},
				item{itemData, "eyJtc2ciOiAiaGV5YSJ9"},
				item{itemEOF, ""},
			},
			DataURL{
				MediaType{
					"application",
					"json",
					map[string]string{
						"charset": "utf-8",
						"foo":     `b"<@>"r`,
						"style":   "unformatted json",
					},
				},
				EncodingBase64,
				[]byte(`{"msg": "heya"}`),
			},
		},
		dataURLTest{
			`data:xxx;base64,aGV5YQ==`,
			[]item{
				item{itemDataPrefix, dataPrefix},
				item{itemError, "invalid character for media type"},
			},
			DataURL{},
		},
		dataURLTest{
			`data:,`,
			[]item{
				item{itemDataPrefix, dataPrefix},
				item{itemDataComma, ","},
				item{itemEOF, ""},
			},
			DataURL{
				defaultMediaType(),
				EncodingASCII,
				[]byte(""),
			},
		},
		dataURLTest{
			`data:,A%20brief%20note`,
			[]item{
				item{itemDataPrefix, dataPrefix},
				item{itemDataComma, ","},
				item{itemData, "A%20brief%20note"},
				item{itemEOF, ""},
			},
			DataURL{
				defaultMediaType(),
				EncodingASCII,
				[]byte("A brief note"),
			},
		},
		dataURLTest{
			`data:image/svg+xml-im.a.fake;base64,cGllLXN0b2NrX1RoaXJ0eQ==`,
			[]item{
				item{itemDataPrefix, dataPrefix},
				item{itemMediaType, "image"},
				item{itemMediaSep, "/"},
				item{itemMediaSubType, "svg+xml-im.a.fake"},
				item{itemParamSemicolon, ";"},
				item{itemBase64Enc, "base64"},
				item{itemDataComma, ","},
				item{itemData, "cGllLXN0b2NrX1RoaXJ0eQ=="},
				item{itemEOF, ""},
			},
			DataURL{
				MediaType{
					"image",
					"svg+xml-im.a.fake",
					map[string]string{},
				},
				EncodingBase64,
				[]byte("pie-stock_Thirty"),
			},
		},
	}
}

func expectItems(expected, actual []item) bool {
	if len(expected) != len(actual) {
		return false
	}
	for i := range expected {
		if expected[i].t != actual[i].t {
			return false
		}
		if expected[i].val != actual[i].val {
			return false
		}
	}
	return true
}

func equal(du1, du2 *DataURL) (bool, error) {
	if !reflect.DeepEqual(du1.MediaType, du2.MediaType) {
		return false, nil
	}
	if du1.Encoding != du2.Encoding {
		return false, nil
	}

	if du1.Data == nil || du2.Data == nil {
		return false, fmt.Errorf("nil Data")
	}

	if !bytes.Equal(du1.Data, du2.Data) {
		return false, nil
	}
	return true, nil
}

func TestLexDataURLs(t *testing.T) {
	for _, test := range genTestTable() {
		l := lex(test.InputRawDataURL)
		var items []item
		for item := range l.items {
			items = append(items, item)
		}
		if !expectItems(test.ExpectedItems, items) {
			t.Errorf("Expected %v, got %v", test.ExpectedItems, items)
		}
	}
}

func testDataURLs(t *testing.T, factory func(string) (*DataURL, error)) {
	for _, test := range genTestTable() {
		var expectedItemError string
		for _, item := range test.ExpectedItems {
			if item.t == itemError {
				expectedItemError = item.String()
				break
			}
		}
		dataURL, err := factory(test.InputRawDataURL)
		if expectedItemError == "" && err != nil {
			t.Error(err)
			continue
		} else if expectedItemError != "" && err == nil {
			t.Errorf("Expected error \"%s\", got nil", expectedItemError)
			continue
		} else if expectedItemError != "" && err != nil {
			if err.Error() != expectedItemError {
				t.Errorf("Expected error \"%s\", got \"%s\"", expectedItemError, err.Error())
			}
			continue
		}

		if ok, err := equal(dataURL, &test.ExpectedDataURL); err != nil {
			t.Error(err)
		} else if !ok {
			t.Errorf("Expected %v, got %v", test.ExpectedDataURL, *dataURL)
		}
	}
}

func TestDataURLsWithDecode(t *testing.T) {
	testDataURLs(t, func(s string) (*DataURL, error) {
		return Decode(strings.NewReader(s))
	})
}

func TestDataURLsWithDecodeString(t *testing.T) {
	testDataURLs(t, func(s string) (*DataURL, error) {
		return DecodeString(s)
	})
}

func TestDataURLsWithUnmarshalText(t *testing.T) {
	testDataURLs(t, func(s string) (*DataURL, error) {
		d := &DataURL{}
		err := d.UnmarshalText([]byte(s))
		return d, err
	})
}

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		s           string
		roundTripOk bool
	}{
		{`data:text/plain;charset=utf-8;foo=bar;base64,aGV5YQ==`, true},
		{`data:;charset=utf-8;foo=bar;base64,aGV5YQ==`, false},
		{`data:text/plain;charset=utf-8;foo="bar";base64,aGV5YQ==`, false},
		{`data:text/plain;charset=utf-8;foo="bar",A%20brief%20note`, false},
		{`data:text/plain;charset=utf-8;foo=bar,A%20brief%20note`, true},
	}
	for _, test := range tests {
		dataURL, err := DecodeString(test.s)
		if err != nil {
			t.Error(err)
			continue
		}
		dus := dataURL.String()
		if test.roundTripOk && dus != test.s {
			t.Errorf("Expected %s, got %s", test.s, dus)
		} else if !test.roundTripOk && dus == test.s {
			t.Errorf("Found %s, expected something else", test.s)
		}

		txt, err := dataURL.MarshalText()
		if err != nil {
			t.Error(err)
			continue
		}
		if test.roundTripOk && string(txt) != test.s {
			t.Errorf("MarshalText roundtrip: got '%s', want '%s'", txt, test.s)
		} else if !test.roundTripOk && string(txt) == test.s {
			t.Errorf("MarshalText roundtrip: got '%s', want something else", txt)
		}
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		Data            []byte
		MediaType       string
		ParamPairs      []string
		WillPanic       bool
		ExpectedDataURL *DataURL
	}{
		{
			[]byte(`{"msg": "heya"}`),
			"application/json",
			[]string{},
			false,
			&DataURL{
				MediaType{
					"application",
					"json",
					map[string]string{},
				},
				EncodingBase64,
				[]byte(`{"msg": "heya"}`),
			},
		},
		{
			[]byte(``),
			"application//json",
			[]string{},
			true,
			nil,
		},
		{
			[]byte(``),
			"",
			[]string{},
			true,
			nil,
		},
		{
			[]byte(`{"msg": "heya"}`),
			"text/plain",
			[]string{"charset", "utf-8"},
			false,
			&DataURL{
				MediaType{
					"text",
					"plain",
					map[string]string{
						"charset": "utf-8",
					},
				},
				EncodingBase64,
				[]byte(`{"msg": "heya"}`),
			},
		},
		{
			[]byte(`{"msg": "heya"}`),
			"text/plain",
			[]string{"charset", "utf-8", "name"},
			true,
			nil,
		},
	}
	for _, test := range tests {
		var dataURL *DataURL
		func() {
			defer func() {
				if test.WillPanic {
					if e := recover(); e == nil {
						t.Error("Expected panic didn't happen")
					}
				} else {
					if e := recover(); e != nil {
						t.Errorf("Unexpected panic: %v", e)
					}
				}
			}()
			dataURL = New(test.Data, test.MediaType, test.ParamPairs...)
		}()
		if test.WillPanic {
			if dataURL != nil {
				t.Error("Expected nil DataURL")
			}
		} else {
			if ok, err := equal(dataURL, test.ExpectedDataURL); err != nil {
				t.Error(err)
			} else if !ok {
				t.Errorf("Expected %v, got %v", test.ExpectedDataURL, *dataURL)
			}
		}
	}
}

var golangFavicon = strings.Replace(`AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAD///8AVE44//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb/
/uF2/1ROOP////8A////AFROOP/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+
4Xb//uF2//7hdv9UTjj/////AP///wBUTjj//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7h
dv/+4Xb//uF2//7hdv/+4Xb/VE44/////wD///8AVE44//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2
//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2/1ROOP////8A////AFROOP/+4Xb//uF2//7hdv/+4Xb/
/uF2//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv9UTjj/////AP///wBUTjj//uF2//7hdv/+
4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb/VE44/////wD///8AVE44//7h
dv/+4Xb//uF2//7hdv/+4Xb/z7t5/8Kyev/+4Xb//993///dd///3Xf//uF2/1ROOP////8A////
AFROOP/+4Xb//uF2//7hdv//4Hn/dIzD//v8///7/P//dIzD//7hdv//3Xf//913//7hdv9UTjj/
////AP///wBUTjj//uF2///fd//+4Xb//uF2/6ajif90jMP/dIzD/46Zpv/+4Xb//+F1///feP/+
4Xb/VE44/////wD///8AVE44//7hdv/z1XT////////////Is3L/HyAj/x8gI//Is3L/////////
///z1XT//uF2/1ROOP////8A19nd/1ROOP/+4Xb/5+HS//v+//8RExf/Liwn//7hdv/+4Xb/5+HS
//v8//8RExf/Liwn//7hdv9UTjj/19nd/1ROOP94aDT/yKdO/+fh0v//////ERMX/y4sJ//+4Xb/
/uF2/+fh0v//////ERMX/y4sJ//Ip07/dWU3/1ROOP9UTjj/yKdO/6qSSP/Is3L/9fb7//f6///I
s3L//uF2//7hdv/Is3L////////////Is3L/qpJI/8inTv9UTjj/19nd/1ROOP97c07/qpJI/8in
Tv/Ip07//uF2//7hdv/+4Xb//uF2/8zBlv/Kv4//pZJU/3tzTv9UTjj/19nd/////wD///8A4eLl
/6CcjP97c07/e3NO/1dOMf9BOiX/TkUn/2VXLf97c07/e3NO/6CcjP/h4uX/////AP///wD///8A
////AP///wD///8A////AP///wDq6/H/3N/j/9fZ3f/q6/H/////AP///wD///8A////AP///wD/
//8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAA==`, "\n", "", -1)

func TestEncodeBytes(t *testing.T) {
	mustDecode := func(s string) []byte {
		data, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			panic(err)
		}
		return data
	}
	tests := []struct {
		Data           []byte
		ExpectedString string
	}{
		{
			[]byte(`A brief note`),
			"data:text/plain;charset=utf-8;base64,QSBicmllZiBub3Rl",
		},
		{
			[]byte{0xA, 0xFF, 0x99, 0x34, 0x56, 0x34, 0x00},
			`data:application/octet-stream;base64,Cv+ZNFY0AA==`,
		},
		{
			mustDecode(golangFavicon),
			`data:image/vnd.microsoft.icon;base64,` + golangFavicon,
		},
	}
	for _, test := range tests {
		str := EncodeBytes(test.Data)
		if str != test.ExpectedString {
			t.Errorf("Expected %s, got %s", test.ExpectedString, str)
		}
	}
}

func BenchmarkLex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, test := range genTestTable() {
			l := lex(test.InputRawDataURL)
			for _ = range l.items {
			}
		}
	}
}

const rep = `^data:(?P<mediatype>\w+/[\w\+\-\.]+)?(?P<parameter>(?:;[\w\-]+="?[\w\-\\<>@,";:%]*"?)+)?(?P<base64>;base64)?,(?P<data>.*)$`

func TestRegexp(t *testing.T) {
	re, err := regexp.Compile(rep)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range genTestTable() {
		shouldMatch := true
		for _, item := range test.ExpectedItems {
			if item.t == itemError {
				shouldMatch = false
				break
			}
		}
		// just test it matches, do not parse
		if re.MatchString(test.InputRawDataURL) && !shouldMatch {
			t.Error("doesn't match", test.InputRawDataURL)
		} else if !re.MatchString(test.InputRawDataURL) && shouldMatch {
			t.Error("match", test.InputRawDataURL)
		}
	}
}

func BenchmarkRegexp(b *testing.B) {
	re, err := regexp.Compile(rep)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		for _, test := range genTestTable() {
			_ = re.FindStringSubmatch(test.InputRawDataURL)
		}
	}
}

func ExampleDecodeString() {
	dataURL, err := DecodeString(`data:text/plain;charset=utf-8;base64,aGV5YQ==`)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s, %s", dataURL.MediaType.ContentType(), string(dataURL.Data))
	// Output: text/plain, heya
}

func ExampleDecode() {
	r, err := http.NewRequest(
		"POST", "/",
		strings.NewReader(`data:image/vnd.microsoft.icon;name=golang%20favicon;base64,`+golangFavicon),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	var dataURL *DataURL
	h := func(w http.ResponseWriter, r *http.Request) {
		var err error
		dataURL, err = Decode(r.Body)
		defer r.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}
	w := httptest.NewRecorder()
	h(w, r)
	fmt.Printf("%s: %s", dataURL.Params["name"], dataURL.ContentType())
	// Output: golang favicon: image/vnd.microsoft.icon
}
