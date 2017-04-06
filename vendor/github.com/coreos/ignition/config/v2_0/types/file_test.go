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

func TestFileValidate(t *testing.T) {
	type in struct {
		mode FileMode
	}
	type out struct {
		err error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{mode: FileMode(0)},
			out: out{},
		},
		{
			in:  in{mode: FileMode(0644)},
			out: out{},
		},
		{
			in:  in{mode: FileMode(01755)},
			out: out{},
		},
		{
			in:  in{mode: FileMode(07777)},
			out: out{},
		},
		{
			in:  in{mode: FileMode(010000)},
			out: out{err: ErrFileIllegalMode},
		},
	}

	for i, test := range tests {
		err := test.in.mode.Validate()
		if !reflect.DeepEqual(report.ReportFromError(test.out.err, report.EntryError), err) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.out.err, err)
		}
	}
}
