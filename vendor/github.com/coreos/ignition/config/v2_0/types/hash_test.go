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
	"encoding/json"
	"reflect"
	"testing"

	"github.com/coreos/ignition/config/validate/report"
)

func TestHashUnmarshalJSON(t *testing.T) {
	type in struct {
		data string
	}
	type out struct {
		hash Hash
		err  error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{data: `"sha512-0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"`},
			out: out{hash: Hash{Function: "sha512", Sum: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}},
		},
		{
			in:  in{data: `"xor01234567"`},
			out: out{err: ErrHashMalformed},
		},
	}

	for i, test := range tests {
		var hash Hash
		err := json.Unmarshal([]byte(test.in.data), &hash)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.out.err, err)
		}
		if !reflect.DeepEqual(test.out.hash, hash) {
			t.Errorf("#%d: bad hash: want %+v, got %+v", i, test.out.hash, hash)
		}
	}
}

func TestHashValidate(t *testing.T) {
	type in struct {
		hash Hash
	}
	type out struct {
		err error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{hash: Hash{}},
			out: out{err: ErrHashUnrecognized},
		},
		{
			in:  in{hash: Hash{Function: "xor"}},
			out: out{err: ErrHashUnrecognized},
		},
		{
			in:  in{hash: Hash{Function: "sha512", Sum: "123"}},
			out: out{err: ErrHashWrongSize},
		},
		{
			in:  in{hash: Hash{Function: "sha512", Sum: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}},
			out: out{},
		},
	}

	for i, test := range tests {
		err := test.in.hash.Validate()
		if !reflect.DeepEqual(report.ReportFromError(test.out.err, report.EntryError), err) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.out.err, err)
		}
	}
}
