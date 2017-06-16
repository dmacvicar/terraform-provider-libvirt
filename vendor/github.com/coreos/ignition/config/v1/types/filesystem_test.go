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
)

func TestFilesystemFormatUnmarshalJSON(t *testing.T) {
	type in struct {
		data string
	}
	type out struct {
		format FilesystemFormat
		err    error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{data: `"ext4"`},
			out: out{format: FilesystemFormat("ext4")},
		},
		{
			in:  in{data: `"bad"`},
			out: out{format: FilesystemFormat("bad"), err: ErrFilesystemInvalidFormat},
		},
	}

	for i, test := range tests {
		var format FilesystemFormat
		err := json.Unmarshal([]byte(test.in.data), &format)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.out.err, err)
		}
		if !reflect.DeepEqual(test.out.format, format) {
			t.Errorf("#%d: bad format: want %#v, got %#v", i, test.out.format, format)
		}
	}
}

func TestFilesystemFormatAssertValid(t *testing.T) {
	type in struct {
		format FilesystemFormat
	}
	type out struct {
		err error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{format: FilesystemFormat("ext4")},
			out: out{},
		},
		{
			in:  in{format: FilesystemFormat("btrfs")},
			out: out{},
		},
		{
			in:  in{format: FilesystemFormat("")},
			out: out{err: ErrFilesystemInvalidFormat},
		},
	}

	for i, test := range tests {
		err := test.in.format.AssertValid()
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.out.err, err)
		}
	}
}

func TestMkfsOptionsUnmarshalJSON(t *testing.T) {
	type in struct {
		data string
	}
	type out struct {
		options MkfsOptions
		err     error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{data: `["--label=ROOT"]`},
			out: out{options: MkfsOptions([]string{"--label=ROOT"})},
		},
	}

	for i, test := range tests {
		var options MkfsOptions
		err := json.Unmarshal([]byte(test.in.data), &options)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.out.err, err)
		}
		if !reflect.DeepEqual(test.out.options, options) {
			t.Errorf("#%d: bad format: want %#v, got %#v", i, test.out.options, options)
		}
	}
}

func TestFilesystemUnmarshalJSON(t *testing.T) {
	type in struct {
		data string
	}
	type out struct {
		filesystem Filesystem
		err        error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{data: `{"device": "/foo", "format": "ext4"}`},
			out: out{filesystem: Filesystem{Device: "/foo", Format: "ext4"}},
		},
		{
			in:  in{data: `{"format": "ext4"}`},
			out: out{filesystem: Filesystem{Format: "ext4"}, err: ErrPathRelative},
		},
	}

	for i, test := range tests {
		var filesystem Filesystem
		err := json.Unmarshal([]byte(test.in.data), &filesystem)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.out.err, err)
		}
		if !reflect.DeepEqual(test.out.filesystem, filesystem) {
			t.Errorf("#%d: bad filesystem: want %#v, got %#v", i, test.out.filesystem, filesystem)
		}
	}
}

func TestFilesystemAssertValid(t *testing.T) {
	type in struct {
		filesystem Filesystem
	}
	type out struct {
		err error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{filesystem: Filesystem{Device: "/foo", Format: "ext4"}},
			out: out{},
		},
		{
			in:  in{filesystem: Filesystem{Device: "/foo"}},
			out: out{err: ErrFilesystemInvalidFormat},
		},
		{
			in:  in{filesystem: Filesystem{Format: "ext4"}},
			out: out{err: ErrPathRelative},
		},
		{
			in:  in{filesystem: Filesystem{}},
			out: out{err: ErrPathRelative},
		},
	}

	for i, test := range tests {
		err := test.in.filesystem.AssertValid()
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.out.err, err)
		}
	}
}
