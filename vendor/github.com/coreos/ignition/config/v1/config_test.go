// Copyright 2015 CoreOS, Inc.
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

package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/coreos/ignition/config/errors"
	"github.com/coreos/ignition/config/v1/types"
	"github.com/coreos/ignition/config/validate/report"
)

func TestParse(t *testing.T) {
	type in struct {
		config []byte
	}
	type out struct {
		config types.Config
		report report.Report
		err    error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{config: []byte(`{"ignitionVersion": 1}`)},
			out: out{config: types.Config{Version: 1}},
		},
		{
			in: in{config: []byte(`{"ignitionVersion": 1, "storage": {"filesystems": [{"device": "this/is/a/relative/path", "format":"ext4"}]}}`)},
			out: out{
				report: report.Report{
					Entries: []report.Entry{
						{
							Message:   types.ErrPathRelative.Error(),
							Kind:      report.EntryError,
							Line:      1,
							Column:    87,
							Highlight: "    1: {\"ignitionVersion\": 1, \"storage\": {\"filesystems\": [{\"device\": \"this/is/a/relative/path\"\n                                                                                            ^\n",
						},
					},
				},
				err: errors.ErrInvalid,
			},
		},
		{
			in:  in{config: []byte(`{}`)},
			out: out{err: errors.ErrInvalid},
		},
		{
			in:  in{config: []byte{}},
			out: out{err: errors.ErrEmpty},
		},
		{
			in:  in{config: []byte("#cloud-config")},
			out: out{err: errors.ErrCloudConfig},
		},
		{
			in:  in{config: []byte("#cloud-config ")},
			out: out{err: errors.ErrCloudConfig},
		},
		{
			in:  in{config: []byte("#cloud-config\n\r")},
			out: out{err: errors.ErrCloudConfig},
		},
		{
			in: in{config: []byte{0x1f, 0x8b, 0x08, 0x00, 0x03, 0xd6, 0x79, 0x56,
				0x00, 0x03, 0x53, 0x4e, 0xce, 0xc9, 0x2f, 0x4d, 0xd1, 0x4d, 0xce,
				0xcf, 0x4b, 0xcb, 0x4c, 0xe7, 0x02, 0x00, 0x05, 0x56, 0xb3, 0xb8,
				0x0e, 0x00, 0x00, 0x00}},
			out: out{err: errors.ErrCloudConfig},
		},
		{
			in:  in{config: []byte("#!/bin/sh")},
			out: out{err: errors.ErrScript},
		},
		{
			in: in{config: []byte{0x1f, 0x8b, 0x08, 0x00, 0x48, 0xda, 0x79, 0x56,
				0x00, 0x03, 0x53, 0x56, 0xd4, 0x4f, 0xca, 0xcc, 0xd3, 0x2f, 0xce,
				0xe0, 0x02, 0x00, 0x1d, 0x9d, 0xfb, 0x04, 0x0a, 0x00, 0x00, 0x00}},
			out: out{err: errors.ErrScript},
		},
	}

	for i, test := range tests {
		config, rpt, err := Parse(test.in.config)
		assert.Equal(t, test.out.config, config, "#%d: bad config", i)
		assert.Equal(t, test.out.report, rpt, "#%d: bad report", i)
		if len(rpt.Entries) > 0 {
			t.Logf("highlight: %q\n", rpt.Entries[0].Highlight)
		}
		assert.Equal(t, test.out.err, err, "#%d: bad error", i)
	}
}
