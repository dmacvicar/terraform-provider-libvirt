// Copyright 2016 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/coreos/ignition/config/validate/report"
)

func TestURLValidate(t *testing.T) {
	type in struct {
		u string
	}
	type out struct {
		err error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{u: ""},
			out: out{},
		},
		{
			in:  in{u: "http://example.com"},
			out: out{},
		},
		{
			in:  in{u: "https://example.com"},
			out: out{},
		},
		{
			in:  in{u: "oem:///foobar"},
			out: out{},
		},
		{
			in:  in{u: "data:,example%20file%0A"},
			out: out{},
		},
		{
			in:  in{u: "bad://"},
			out: out{err: ErrInvalidScheme},
		},
	}

	for i, test := range tests {
		u, err := url.Parse(test.in.u)
		if err != nil {
			t.Errorf("URL failed to parse. This is an error with the test")
		}
		r := Url(*u).Validate()
		if !reflect.DeepEqual(report.ReportFromError(test.out.err, report.EntryError), r) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.out.err, r)
		}
	}
}
