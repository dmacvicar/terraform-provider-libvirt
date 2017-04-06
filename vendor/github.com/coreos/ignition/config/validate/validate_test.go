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

package validate

import (
	"errors"
	"reflect"
	"testing"

	"github.com/coreos/go-semver/semver"

	// Import into the same namespace to keep config definitions clean
	. "github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/config/validate/report"
)

func TestValidate(t *testing.T) {
	type in struct {
		cfg Config
	}
	type out struct {
		err error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{cfg: Config{Ignition: Ignition{Version: semver.Version{Major: 2}.String()}}},
			out: out{},
		},
		{
			in:  in{cfg: Config{}},
			out: out{err: ErrInvalidVersion},
		},
		{
			in:  in{cfg: Config{Ignition: Ignition{Version: "invalid.version"}}},
			out: out{err: ErrInvalidVersion},
		},
		{
			in:  in{cfg: Config{Ignition: Ignition{Version: "2.2.0"}}},
			out: out{err: ErrNewVersion},
		},
		{
			in:  in{cfg: Config{Ignition: Ignition{Version: "3.0.0"}}},
			out: out{err: ErrNewVersion},
		},
		{
			in:  in{cfg: Config{Ignition: Ignition{Version: "1.0.0"}}},
			out: out{err: ErrOldVersion},
		},
		{
			in: in{cfg: Config{
				Ignition: Ignition{
					Version: semver.Version{Major: 2}.String(),
					Config: IgnitionConfig{
						Replace: &ConfigReference{
							Verification: Verification{
								Hash: func(s string) *string { return &s }("foobar-"),
							},
						},
					},
				},
			}},
			out: out{errors.New("unrecognized hash function")},
		},
		{
			in: in{cfg: Config{
				Ignition: Ignition{Version: semver.Version{Major: 2}.String()},
				Storage: Storage{
					Filesystems: []Filesystem{
						{
							Name: "filesystem1",
							Mount: &Mount{
								Device: "/dev/disk/by-partlabel/ROOT",
								Format: "btrfs",
							},
						},
					},
				},
			}},
			out: out{},
		},
		{
			in: in{cfg: Config{
				Ignition: Ignition{Version: semver.Version{Major: 2}.String()},
				Storage: Storage{
					Filesystems: []Filesystem{
						{
							Name: "filesystem1",
							Path: func(p string) *string { return &p }("/sysroot"),
						},
					},
				},
			}},
			out: out{},
		},
		{
			in: in{cfg: Config{
				Ignition: Ignition{Version: semver.Version{Major: 2}.String()},
				Systemd:  Systemd{Units: []Unit{{Name: "foo.bar", Contents: "[Foo]\nfoo=qux"}}},
			}},
			out: out{err: errors.New("invalid systemd unit extension")},
		},
	}

	for i, test := range tests {
		r := ValidateWithoutSource(reflect.ValueOf(test.in.cfg))
		expectedReport := report.ReportFromError(test.out.err, report.EntryError)
		if !reflect.DeepEqual(expectedReport, r) {
			t.Errorf("#%d: bad error: want %v, got %v", i, expectedReport, r)
		}
	}
}

func TestGetFields(t *testing.T) {
	// basic case
	type Test1 struct {
		A int
		B string
	}
	test1 := Test1{
		A: 1,
		B: "one",
	}

	// test embedded structs
	type Test2 struct {
		C int
		Test1
	}
	test2 := Test2{
		C:     5,
		Test1: test1,
	}

	// test doublely embedded structs
	type Test3 struct {
		D int
		Test2
	}
	test3 := Test3{
		D:     3,
		Test2: test2,
	}
	// test structs embedded via an alias to interface{}
	type Anything interface{}

	test4 := struct {
		E int
		Anything
	}{
		E:        7,
		Anything: test3,
	}

	// test normally contained structs don't cause problems
	test5 := struct {
		E int
		F Test3
	}{
		E: 2,
		F: test3,
	}

	// test non-structs embedded via an alias to interface{} don't cause panics
	test6 := struct {
		E int
		Anything
	}{
		E:        5,
		Anything: 65,
	}

	// test embedded nils
	test7 := struct {
		E int
		Anything
	}{
		E: 5,
	}

	type in struct {
		strct reflect.Value
	}
	type out struct {
		fields []field
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in: in{strct: reflect.ValueOf(test1)},
			out: out{fields: []field{
				{
					Type:  reflect.TypeOf(test1).Field(0),
					Value: reflect.ValueOf(test1.A),
				},
				{
					Type:  reflect.TypeOf(test1).Field(1),
					Value: reflect.ValueOf(test1.B),
				},
			}},
		},
		{
			in: in{strct: reflect.ValueOf(test2)},
			out: out{fields: []field{
				{
					Type:  reflect.TypeOf(test2).Field(0),
					Value: reflect.ValueOf(test2.C),
				},
				{
					Type:  reflect.TypeOf(test1).Field(0),
					Value: reflect.ValueOf(test1.A),
				},
				{
					Type:  reflect.TypeOf(test1).Field(1),
					Value: reflect.ValueOf(test1.B),
				},
			}},
		},
		{
			in: in{strct: reflect.ValueOf(test3)},
			out: out{fields: []field{
				{
					Type:  reflect.TypeOf(test3).Field(0),
					Value: reflect.ValueOf(test3.D),
				},
				{
					Type:  reflect.TypeOf(test2).Field(0),
					Value: reflect.ValueOf(test2.C),
				},
				{
					Type:  reflect.TypeOf(test1).Field(0),
					Value: reflect.ValueOf(test1.A),
				},
				{
					Type:  reflect.TypeOf(test1).Field(1),
					Value: reflect.ValueOf(test1.B),
				},
			}},
		},
		{
			in: in{strct: reflect.ValueOf(test4)},
			out: out{fields: []field{
				{
					Type:  reflect.TypeOf(test4).Field(0),
					Value: reflect.ValueOf(test4.E),
				},
				{
					Type:  reflect.TypeOf(test3).Field(0),
					Value: reflect.ValueOf(test3.D),
				},
				{
					Type:  reflect.TypeOf(test2).Field(0),
					Value: reflect.ValueOf(test2.C),
				},
				{
					Type:  reflect.TypeOf(test1).Field(0),
					Value: reflect.ValueOf(test1.A),
				},
				{
					Type:  reflect.TypeOf(test1).Field(1),
					Value: reflect.ValueOf(test1.B),
				},
			}},
		},
		{
			in: in{strct: reflect.ValueOf(test5)},
			out: out{fields: []field{
				{
					Type:  reflect.TypeOf(test5).Field(0),
					Value: reflect.ValueOf(test5.E),
				},
				{
					Type:  reflect.TypeOf(test5).Field(1),
					Value: reflect.ValueOf(test5.F),
				},
			}},
		},
		{
			in: in{strct: reflect.ValueOf(test6)},
			out: out{fields: []field{
				{
					Type:  reflect.TypeOf(test6).Field(0),
					Value: reflect.ValueOf(test6.E),
				},
				{
					Type:  reflect.TypeOf(test6).Field(1),
					Value: reflect.ValueOf(65),
				},
			}},
		},
		{
			in: in{strct: reflect.ValueOf(test7)},
			out: out{fields: []field{
				{
					Type:  reflect.TypeOf(test7).Field(0),
					Value: reflect.ValueOf(test7.E),
				},
				{
					Type:  reflect.TypeOf(test7).Field(1),
					Value: reflect.ValueOf(nil),
				},
			}},
		},
	}

	for i, test := range tests {
		fields := getFields(test.in.strct)
		// We cannot use reflect.DeepEqual because reflect.DeepEqual(reflect.ValueOf(someinstance),reflect.ValueOf(someinstance))
		// will always return false. We must manually loop over it and convert reflect.Value's to interface{}'s as we go
		for idx, f := range fields {
			if !reflect.DeepEqual(f.Type, test.out.fields[idx].Type) {
				t.Errorf("#%d: bad error with type: want \n%+v, got \n%+v", i, fields, test.out.fields)
			}
			if !reflect.DeepEqual(f.Value.Interface(), test.out.fields[idx].Value.Interface()) {
				t.Errorf("#%d: bad error: want \n%+v, got \n%+v", i, fields, test.out.fields)
			}
		}
	}
}
