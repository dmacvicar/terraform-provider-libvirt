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

func TestNodeValidatePath(t *testing.T) {
	node := Node{Path: "not/absolute"}
	rep := report.ReportFromError(ErrPathRelative, report.EntryError)
	if receivedRep := node.ValidatePath(); !reflect.DeepEqual(rep, receivedRep) {
		t.Errorf("bad error: want %v, got %v", rep, receivedRep)
	}
}

func TestNodeValidateFilesystem(t *testing.T) {
	tests := []struct {
		node Node
		r    report.Report
	}{
		{
			node: Node{Filesystem: "foo", Path: "/"},
			r:    report.Report{},
		},
		{
			node: Node{Path: "/"},
			r:    report.ReportFromError(ErrNoFilesystem, report.EntryError),
		},
	}
	for i, test := range tests {
		if receivedRep := test.node.ValidateFilesystem(); !reflect.DeepEqual(test.r, receivedRep) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.r, receivedRep)
		}
	}
}

func intToPtr(x int) *int {
	return &x
}

func TestNodeValidateUser(t *testing.T) {
	tests := []struct {
		in  NodeUser
		out report.Report
	}{
		{
			in:  NodeUser{intToPtr(0), ""},
			out: report.Report{},
		},
		{
			in:  NodeUser{intToPtr(1000), ""},
			out: report.Report{},
		},
		{
			in:  NodeUser{nil, "core"},
			out: report.Report{},
		},
		{
			in:  NodeUser{intToPtr(1000), "core"},
			out: report.ReportFromError(ErrBothIDAndNameSet, report.EntryError),
		},
	}

	for i, test := range tests {
		report := test.in.Validate()
		if !reflect.DeepEqual(test.out, report) {
			t.Errorf("#%d: bad report: want %v got %v", i, test.out, report)
		}
	}
}

func TestNodeValidateGroup(t *testing.T) {
	tests := []struct {
		in  NodeGroup
		out report.Report
	}{
		{
			in:  NodeGroup{intToPtr(0), ""},
			out: report.Report{},
		},
		{
			in:  NodeGroup{intToPtr(1000), ""},
			out: report.Report{},
		},
		{
			in:  NodeGroup{nil, "core"},
			out: report.Report{},
		},
		{
			in:  NodeGroup{intToPtr(1000), "core"},
			out: report.ReportFromError(ErrBothIDAndNameSet, report.EntryError),
		},
	}

	for i, test := range tests {
		report := test.in.Validate()
		if !reflect.DeepEqual(test.out, report) {
			t.Errorf("#%d: bad report: want %v got %v", i, test.out, report)
		}
	}
}
