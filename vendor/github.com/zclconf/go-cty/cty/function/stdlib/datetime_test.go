package stdlib

import (
	"fmt"
	"testing"
	"time"

	"github.com/zclconf/go-cty/cty"
)

func TestFormatDate(t *testing.T) {
	tests := []struct {
		Format cty.Value
		Want   cty.Value
		Err    string
	}{
		{
			cty.StringVal(""), // pointless, but valid
			cty.StringVal(""),
			``,
		},
		{
			cty.StringVal("YYYY-MM-DD"),
			cty.StringVal("2006-01-02"),
			``,
		},
		{
			cty.StringVal("EEE, MMM D ''YY"),
			cty.StringVal("Mon, Jan 2 '06"),
			``,
		},
		{
			cty.StringVal("hh:mm:ss"),
			cty.StringVal("15:04:05"),
			``,
		},
		{
			cty.StringVal("H 'o''clock' AA"),
			cty.StringVal("3 o'clock PM"),
			``,
		},
		{
			cty.StringVal("hh:mm:ssZZZZ"),
			cty.StringVal("15:04:05+0000"),
			``,
		},
		{
			cty.StringVal("hh:mm:ssZZZZZ"),
			cty.StringVal("15:04:05+00:00"),
			``,
		},
		{
			cty.StringVal("MMMM"),
			cty.StringVal("January"),
			``,
		},
		{
			cty.StringVal("EEEE"),
			cty.StringVal("Monday"),
			``,
		},
		{
			cty.StringVal("aa"),
			cty.StringVal("pm"),
			``,
		},

		// Some common standard machine-oriented formats
		{
			cty.StringVal("YYYY-MM-DD'T'hh:mm:ssZ"), // RFC3339
			cty.StringVal("2006-01-02T15:04:05Z"),   // (since RFC3339 is the input format too, this is a bit pointless)
			``,
		},
		{
			cty.StringVal("DD MMM YYYY hh:mm ZZZ"), // RFC822
			cty.StringVal("02 Jan 2006 15:04 UTC"),
			``,
		},
		{
			cty.StringVal("EEEE, DD-MMM-YY hh:mm:ss ZZZ"), // RFC850
			cty.StringVal("Monday, 02-Jan-06 15:04:05 UTC"),
			``,
		},
		{
			cty.StringVal("EEE, DD MMM YYYY hh:mm:ss ZZZ"), // RFC1123
			cty.StringVal("Mon, 02 Jan 2006 15:04:05 UTC"),
			``,
		},

		// Invalids
		{
			cty.StringVal("Y"),
			cty.NilVal,
			`invalid date format verb "Y": year must either be "YY" or "YYYY"`,
		},
		{
			cty.StringVal("YYYYY"),
			cty.NilVal,
			`invalid date format verb "YYYYY": year must either be "YY" or "YYYY"`,
		},
		{
			cty.StringVal("A"),
			cty.NilVal,
			`invalid date format verb "A": must be "AA"`,
		},
		{
			cty.StringVal("a"),
			cty.NilVal,
			`invalid date format verb "a": must be "aa"`,
		},
		{
			cty.StringVal("'blah blah"),
			cty.NilVal,
			`unterminated literal '`,
		},
		{
			cty.StringVal("'"),
			cty.NilVal,
			`unterminated literal '`,
		},
	}
	ts := time.Date(2006, time.January, 2, 15, 04, 05, 0, time.UTC)
	timeVal := cty.StringVal(ts.Format(time.RFC3339))
	for _, test := range tests {
		t.Run(test.Format.GoString(), func(t *testing.T) {
			got, err := FormatDate(test.Format, timeVal)

			if test.Err != "" {
				if err == nil {
					t.Fatalf("no error; want error %q", test.Err)
				}

				if got, want := err.Error(), test.Err; got != want {
					t.Fatalf("wrong error\ngot:  %s\nwant: %s", got, want)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}

				if !got.RawEquals(test.Want) {
					t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
				}
			}
		})
	}

	parseErrTests := []struct {
		Timestamp cty.Value
		Err       string
	}{
		{
			cty.StringVal(""),
			`not a valid RFC3339 timestamp: end of string before year`,
		},
		{
			cty.StringVal("2017-01-02"),
			`not a valid RFC3339 timestamp: missing required time introducer 'T'`,
		},
		{
			cty.StringVal(`2017-12-02t00:00:00Z`),
			`not a valid RFC3339 timestamp: missing required time introducer 'T'`,
		},
		{
			cty.StringVal("2017:01:02"),
			`not a valid RFC3339 timestamp: found ":01:02" where "-" is expected`,
		},
		{
			cty.StringVal("2017"),
			`not a valid RFC3339 timestamp: end of string where "-" is expected`,
		},
		{
			cty.StringVal("2017-01-02T"),
			`not a valid RFC3339 timestamp: end of string before hour`,
		},
		{
			cty.StringVal("2017-01-02T00"),
			`not a valid RFC3339 timestamp: end of string where ":" is expected`,
		},
		{
			cty.StringVal("2017-01-02T00:00:00"),
			`not a valid RFC3339 timestamp: end of string before UTC offset`,
		},
		{
			cty.StringVal("2017-01-02T26:00:00Z"),
			// This one generates an odd message due to an apparent quirk in
			// the Go time parser. Ideally it would use "26" as the errant string.
			`not a valid RFC3339 timestamp: cannot use ":00:00Z" as hour`,
		},
		{
			cty.StringVal("2017-13-02T00:00:00Z"),
			// This one generates an odd message due to an apparent quirk in
			// the Go time parser. Ideally it would use "13" as the errant string.
			`not a valid RFC3339 timestamp: cannot use "-02T00:00:00Z" as month`,
		},
		{
			cty.StringVal("2017-02-31T00:00:00Z"),
			`not a valid RFC3339 timestamp: day out of range`,
		},
		{
			cty.StringVal(`"2017-12-02T00:00:00Z"`),
			`not a valid RFC3339 timestamp: cannot use "\"2017-12-02T00:00:00Z\"" as year`,
		},
		{
			cty.StringVal(`2-12-02T00:00:00Z`),
			// Go parser seems to be trying to parse "2-12" as a year here,
			// producing a confusing error message.
			`not a valid RFC3339 timestamp: cannot use "-02T00:00:00Z" as year`,
		},
	}
	for _, test := range parseErrTests {
		t.Run(fmt.Sprintf("%s parse error", test.Timestamp.AsString()), func(t *testing.T) {
			_, err := FormatDate(cty.StringVal(""), test.Timestamp)

			if err == nil {
				t.Fatalf("no error; want error %q", test.Err)
			}

			if got, want := err.Error(), test.Err; got != want {
				t.Fatalf("wrong error\ngot:  %s\nwant: %s", got, want)
			}
		})
	}
}
