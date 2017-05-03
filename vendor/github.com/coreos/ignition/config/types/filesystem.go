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
	"fmt"

	"github.com/coreos/ignition/config/validate/report"
)

var (
	ErrFilesystemInvalidFormat = errors.New("invalid filesystem format")
	ErrFilesystemNoMountPath   = errors.New("filesystem is missing mount or path")
	ErrFilesystemMountAndPath  = errors.New("filesystem has both mount and path defined")
)

func (f Filesystem) Validate() report.Report {
	r := report.Report{}
	if f.Mount == nil && f.Path == nil {
		return report.ReportFromError(ErrFilesystemNoMountPath, report.EntryError)
	}
	if f.Mount != nil && f.Path != nil {
		return report.ReportFromError(ErrFilesystemMountAndPath, report.EntryError)
	}
	return r
}

func (f Filesystem) ValidatePath() report.Report {
	r := report.Report{}
	if f.Path != nil && validatePath(*f.Path) != nil {
		r.Add(report.Entry{
			Message: fmt.Sprintf("filesystem %q: path not absolute", f.Name),
			Kind:    report.EntryError,
		})
	}
	return r
}

func (m Mount) Validate() report.Report {
	r := report.Report{}
	switch m.Format {
	case "ext4", "btrfs", "xfs", "swap":
	default:
		r.Add(report.Entry{
			Message: ErrFilesystemInvalidFormat.Error(),
			Kind:    report.EntryError,
		})
	}
	return r
}

func (m Mount) ValidateDevice() report.Report {
	r := report.Report{}
	if err := validatePath(m.Device); err != nil {
		r.Add(report.Entry{
			Message: err.Error(),
			Kind:    report.EntryError,
		})
	}
	return r
}
