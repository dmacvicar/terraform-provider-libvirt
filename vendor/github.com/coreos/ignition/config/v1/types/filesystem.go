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
	"errors"
)

var (
	ErrFilesystemInvalidFormat = errors.New("invalid filesystem format")
)

type Filesystem struct {
	Device Path              `json:"device,omitempty"`
	Format FilesystemFormat  `json:"format,omitempty"`
	Create *FilesystemCreate `json:"create,omitempty"`
	Files  []File            `json:"files,omitempty"`
}

type FilesystemCreate struct {
	Force   bool        `json:"force,omitempty"`
	Options MkfsOptions `json:"options,omitempty"`
}
type filesystem Filesystem

func (f *Filesystem) UnmarshalJSON(data []byte) error {
	tf := filesystem(*f)
	if err := json.Unmarshal(data, &tf); err != nil {
		return err
	}
	*f = Filesystem(tf)
	return f.AssertValid()
}

func (f Filesystem) AssertValid() error {
	if err := f.Device.AssertValid(); err != nil {
		return err
	}
	if err := f.Format.AssertValid(); err != nil {
		return err
	}
	return nil
}

type FilesystemFormat string
type filesystemFormat FilesystemFormat

func (f *FilesystemFormat) UnmarshalJSON(data []byte) error {
	tf := filesystemFormat(*f)
	if err := json.Unmarshal(data, &tf); err != nil {
		return err
	}
	*f = FilesystemFormat(tf)
	return f.AssertValid()
}

func (f FilesystemFormat) AssertValid() error {
	switch f {
	case "ext4", "btrfs", "xfs":
		return nil
	default:
		return ErrFilesystemInvalidFormat
	}
}

type MkfsOptions []string
type mkfsOptions MkfsOptions

func (o *MkfsOptions) UnmarshalJSON(data []byte) error {
	to := mkfsOptions(*o)
	if err := json.Unmarshal(data, &to); err != nil {
		return err
	}
	*o = MkfsOptions(to)
	return o.AssertValid()
}

func (o MkfsOptions) AssertValid() error {
	return nil
}
