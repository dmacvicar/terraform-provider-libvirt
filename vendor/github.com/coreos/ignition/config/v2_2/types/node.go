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
	"path/filepath"

	"github.com/coreos/ignition/config/validate/report"
)

var (
	ErrNoFilesystem     = errors.New("no filesystem specified")
	ErrBothIDAndNameSet = errors.New("cannot set both id and name")
)

func (n Node) ValidateFilesystem() report.Report {
	r := report.Report{}
	if n.Filesystem == "" {
		r.Add(report.Entry{
			Message: ErrNoFilesystem.Error(),
			Kind:    report.EntryError,
		})
	}
	return r
}

func (n Node) ValidatePath() report.Report {
	r := report.Report{}
	if err := validatePath(n.Path); err != nil {
		r.Add(report.Entry{
			Message: err.Error(),
			Kind:    report.EntryError,
		})
	}
	return r
}

func (n Node) Depth() int {
	count := 0
	for p := filepath.Clean(string(n.Path)); p != "/"; count++ {
		p = filepath.Dir(p)
	}
	return count
}

func (nu NodeUser) Validate() report.Report {
	r := report.Report{}
	if nu.ID != nil && nu.Name != "" {
		r.Add(report.Entry{
			Message: ErrBothIDAndNameSet.Error(),
			Kind:    report.EntryError,
		})
	}
	return r
}
func (ng NodeGroup) Validate() report.Report {
	r := report.Report{}
	if ng.ID != nil && ng.Name != "" {
		r.Add(report.Entry{
			Message: ErrBothIDAndNameSet.Error(),
			Kind:    report.EntryError,
		})
	}
	return r
}
