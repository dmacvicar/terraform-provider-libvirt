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
	ErrFilesystemInvalidFormat     = errors.New("invalid filesystem format")
	ErrFilesystemNoMountPath       = errors.New("filesystem is missing mount or path")
	ErrFilesystemMountAndPath      = errors.New("filesystem has both mount and path defined")
	ErrUsedCreateAndMountOpts      = errors.New("cannot use both create object and mount-level options field")
	ErrUsedCreateAndWipeFilesystem = errors.New("cannot use both create object and wipeFilesystem field")
	ErrWarningCreateDeprecated     = errors.New("the create object has been deprecated in favor of mount-level options")
	ErrExt4LabelTooLong            = errors.New("filesystem labels cannot be longer than 16 characters when using ext4")
	ErrBtrfsLabelTooLong           = errors.New("filesystem labels cannot be longer than 256 characters when using btrfs")
	ErrXfsLabelTooLong             = errors.New("filesystem labels cannot be longer than 12 characters when using xfs")
	ErrSwapLabelTooLong            = errors.New("filesystem labels cannot be longer than 15 characters when using swap")
	ErrVfatLabelTooLong            = errors.New("filesystem labels cannot be longer than 11 characters when using vfat")
)

func (f Filesystem) Validate() report.Report {
	r := report.Report{}
	if f.Mount == nil && f.Path == nil {
		r.Add(report.Entry{
			Message: ErrFilesystemNoMountPath.Error(),
			Kind:    report.EntryError,
		})
	}
	if f.Mount != nil {
		if f.Path != nil {
			r.Add(report.Entry{
				Message: ErrFilesystemMountAndPath.Error(),
				Kind:    report.EntryError,
			})
		}
		if f.Mount.Create != nil {
			if f.Mount.WipeFilesystem {
				r.Add(report.Entry{
					Message: ErrUsedCreateAndWipeFilesystem.Error(),
					Kind:    report.EntryError,
				})
			}
			if len(f.Mount.Options) > 0 {
				r.Add(report.Entry{
					Message: ErrUsedCreateAndMountOpts.Error(),
					Kind:    report.EntryError,
				})
			}
			r.Add(report.Entry{
				Message: ErrWarningCreateDeprecated.Error(),
				Kind:    report.EntryWarning,
			})
		}
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
	case "ext4", "btrfs", "xfs", "swap", "vfat":
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

func (m Mount) ValidateLabel() report.Report {
	r := report.Report{}
	if m.Label == nil {
		return r
	}
	switch m.Format {
	case "ext4":
		if len(*m.Label) > 16 {
			// source: man mkfs.ext4
			r.Add(report.Entry{
				Message: ErrExt4LabelTooLong.Error(),
				Kind:    report.EntryError,
			})
		}
	case "btrfs":
		if len(*m.Label) > 256 {
			// source: man mkfs.btrfs
			r.Add(report.Entry{
				Message: ErrBtrfsLabelTooLong.Error(),
				Kind:    report.EntryError,
			})
		}
	case "xfs":
		if len(*m.Label) > 12 {
			// source: man mkfs.xfs
			r.Add(report.Entry{
				Message: ErrXfsLabelTooLong.Error(),
				Kind:    report.EntryError,
			})
		}
	case "swap":
		// mkswap's man page does not state a limit on label size, but through
		// experimentation it appears that mkswap will truncate long labels to
		// 15 characters, so let's enforce that.
		if len(*m.Label) > 15 {
			r.Add(report.Entry{
				Message: ErrSwapLabelTooLong.Error(),
				Kind:    report.EntryError,
			})
		}
	case "vfat":
		if len(*m.Label) > 11 {
			// source: man mkfs.fat
			r.Add(report.Entry{
				Message: ErrVfatLabelTooLong.Error(),
				Kind:    report.EntryError,
			})
		}
	}
	return r
}
