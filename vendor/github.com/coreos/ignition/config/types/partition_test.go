// Copyright 2017 CoreOS, Inc.
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

func TestValidateLabel(t *testing.T) {
	type in struct {
		label string
	}
	type out struct {
		report report.Report
	}
	tests := []struct {
		in  in
		out out
	}{
		{
			in{"root"},
			out{report.Report{}},
		},
		{
			in{""},
			out{report.Report{}},
		},
		{
			in{"111111111111111111111111111111111111"},
			out{report.Report{}},
		},
		{
			in{"1111111111111111111111111111111111111"},
			out{report.ReportFromError(ErrLabelTooLong, report.EntryError)},
		},
	}
	for i, test := range tests {
		r := Partition{Label: test.in.label}.ValidateLabel()
		if !reflect.DeepEqual(r, test.out.report) {
			t.Errorf("#%d: wanted %v, got %v", i, test.out.report, r)
		}
	}
}

func TestValidateTypeGUID(t *testing.T) {
	type in struct {
		typeguid string
	}
	type out struct {
		report report.Report
	}
	tests := []struct {
		in  in
		out out
	}{
		{
			in{"5DFBF5F4-2848-4BAC-AA5E-0D9A20B745A6"},
			out{report.Report{}},
		},
		{
			in{""},
			out{report.Report{}},
		},
		{
			in{"not-a-valid-typeguid"},
			out{report.ReportFromError(ErrDoesntMatchGUIDRegex, report.EntryError)},
		},
	}
	for i, test := range tests {
		r := Partition{TypeGUID: test.in.typeguid}.ValidateTypeGUID()
		if !reflect.DeepEqual(r, test.out.report) {
			t.Errorf("#%d: wanted %v, got %v", i, test.out.report, r)
		}
	}
}

func TestValidateGUID(t *testing.T) {
	type in struct {
		guid string
	}
	type out struct {
		report report.Report
	}
	tests := []struct {
		in  in
		out out
	}{
		{
			in{"5DFBF5F4-2848-4BAC-AA5E-0D9A20B745A6"},
			out{report.Report{}},
		},
		{
			in{""},
			out{report.Report{}},
		},
		{
			in{"not-a-valid-typeguid"},
			out{report.ReportFromError(ErrDoesntMatchGUIDRegex, report.EntryError)},
		},
	}
	for i, test := range tests {
		r := Partition{GUID: test.in.guid}.ValidateGUID()
		if !reflect.DeepEqual(r, test.out.report) {
			t.Errorf("#%d: wanted %v, got %v", i, test.out.report, r)
		}
	}
}
