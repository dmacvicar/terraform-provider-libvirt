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

package util

import (
	"reflect"
	"testing"

	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/internal/resource"
)

func TestAssertValid(t *testing.T) {
	type in struct {
		verification types.Verification
		data         []byte
	}
	type out struct {
		err error
	}

	stringDeref := func(s string) *string { return &s }

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{data: []byte("hello")},
			out: out{},
		},
		{
			in: in{
				verification: types.Verification{
					Hash: stringDeref("sha512-9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"),
				},
				data: []byte("hello"),
			},
			out: out{},
		},
		{
			in: in{
				verification: types.Verification{
					Hash: stringDeref("xor-"),
				},
			},
			out: out{err: types.ErrHashUnrecognized},
		},
		{
			in: in{
				verification: types.Verification{
					Hash: stringDeref("sha512-0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"),
				},
				data: []byte("hello"),
			},
			out: out{err: resource.ErrHashMismatch{
				Calculated: "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043",
				Expected:   "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			}},
		},
	}

	for i, test := range tests {
		err := AssertValid(test.in.verification, test.in.data)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("#%d: bad err: want %+v, got %+v", i, test.out.err, err)
		}
	}
}
