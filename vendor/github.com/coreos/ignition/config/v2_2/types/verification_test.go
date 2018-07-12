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
	"reflect"
	"testing"

	"github.com/coreos/ignition/config/validate/report"
)

func TestHashParts(t *testing.T) {
	type in struct {
		data string
	}
	type out struct {
		err error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in: in{data: `"sha512-0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"`},
		},
		{
			in:  in{data: `"sha512:01234567"`},
			out: out{err: ErrHashMalformed},
		},
	}

	for i, test := range tests {
		fun, sum, err := Verification{Hash: &test.in.data}.HashParts()
		if err != test.out.err {
			t.Fatalf("#%d: bad error: want %+v, got %+v", i, test.out.err, err)
		}
		if err == nil && fun+"-"+sum != test.in.data {
			t.Fatalf("#%d: bad hash: want %+v, got %+v", i, test.in.data, fun+"-"+sum)
		}
	}
}

func TestHashValidate(t *testing.T) {
	type in struct {
		v Verification
	}
	type out struct {
		err error
	}

	h1 := "xor-abcdef"
	h2 := "sha512-123"
	h3 := "sha512-0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{v: Verification{Hash: &h1}},
			out: out{err: ErrHashUnrecognized},
		},
		{
			in:  in{v: Verification{Hash: &h2}},
			out: out{err: ErrHashWrongSize},
		},
		{
			in:  in{v: Verification{Hash: &h3}},
			out: out{},
		},
	}

	for i, test := range tests {
		err := test.in.v.Validate()
		if !reflect.DeepEqual(report.ReportFromError(test.out.err, report.EntryError), err) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.out.err, err)
		}
	}
}
