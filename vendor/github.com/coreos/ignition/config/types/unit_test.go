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
	"errors"
	"reflect"
	"testing"

	"github.com/coreos/ignition/config/validate/report"
)

func TestSystemdUnitValidateContents(t *testing.T) {
	type in struct {
		unit Unit
	}
	type out struct {
		err error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{unit: Unit{Name: "test.service", Contents: "[Foo]\nQux=Bar"}},
			out: out{err: nil},
		},
		{
			in:  in{unit: Unit{Name: "test.service", Contents: "[Foo"}},
			out: out{err: errors.New("invalid unit content: unable to find end of section")},
		},
		{
			in:  in{unit: Unit{Name: "test.service", Contents: "", Dropins: []Dropin{{}}}},
			out: out{err: nil},
		},
	}

	for i, test := range tests {
		err := test.in.unit.ValidateContents()
		if !reflect.DeepEqual(report.ReportFromError(test.out.err, report.EntryError), err) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.out.err, err)
		}
	}
}

func TestSystemdUnitValidateName(t *testing.T) {
	type in struct {
		unit string
	}
	type out struct {
		err error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{unit: "test.service"},
			out: out{err: nil},
		},
		{
			in:  in{unit: "test.socket"},
			out: out{err: nil},
		},
		{
			in:  in{unit: "test.blah"},
			out: out{err: ErrInvalidSystemdExt},
		},
	}

	for i, test := range tests {
		err := Unit{Name: test.in.unit, Contents: "[Foo]\nQux=Bar"}.ValidateName()
		if !reflect.DeepEqual(report.ReportFromError(test.out.err, report.EntryError), err) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.out.err, err)
		}
	}
}

func TestSystemdUnitDropInValidate(t *testing.T) {
	type in struct {
		unit Dropin
	}
	type out struct {
		err error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{unit: Dropin{Name: "test.conf", Contents: "[Foo]\nQux=Bar"}},
			out: out{err: nil},
		},
		{
			in:  in{unit: Dropin{Name: "test.conf", Contents: "[Foo"}},
			out: out{err: errors.New("invalid unit content: unable to find end of section")},
		},
	}

	for i, test := range tests {
		err := test.in.unit.Validate()
		if !reflect.DeepEqual(report.ReportFromError(test.out.err, report.EntryError), err) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.out.err, err)
		}
	}
}

func TestNetworkdUnitNameValidate(t *testing.T) {
	type in struct {
		unit string
	}
	type out struct {
		err error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{unit: "test.network"},
			out: out{err: nil},
		},
		{
			in:  in{unit: "test.link"},
			out: out{err: nil},
		},
		{
			in:  in{unit: "test.netdev"},
			out: out{err: nil},
		},
		{
			in:  in{unit: "test.blah"},
			out: out{err: ErrInvalidNetworkdExt},
		},
	}

	for i, test := range tests {
		err := Networkdunit{Name: test.in.unit, Contents: "[Foo]\nQux=Bar"}.Validate()
		if !reflect.DeepEqual(report.ReportFromError(test.out.err, report.EntryError), err) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.out.err, err)
		}
	}
}

func TestNetworkdUnitValidate(t *testing.T) {
	type in struct {
		unit Networkdunit
	}
	type out struct {
		err error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{unit: Networkdunit{Name: "test.network", Contents: "[Foo]\nQux=Bar"}},
			out: out{err: nil},
		},
		{
			in:  in{unit: Networkdunit{Name: "test.network", Contents: "[Foo"}},
			out: out{err: errors.New("invalid unit content: unable to find end of section")},
		},
	}

	for i, test := range tests {
		err := test.in.unit.Validate()
		if !reflect.DeepEqual(report.ReportFromError(test.out.err, report.EntryError), err) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.out.err, err)
		}
	}
}
