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

package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/coreos/ignition/config/types"
)

const (
	presetPath               string      = "/etc/systemd/system-preset/20-ignition.preset"
	DefaultPresetPermissions os.FileMode = 0644
)

func FileFromSystemdUnit(unit types.Unit) *File {
	return &File{
		Path:       filepath.Join(SystemdUnitsPath(), string(unit.Name)),
		ReadCloser: ioutil.NopCloser(bytes.NewReader([]byte(unit.Contents))),
		Mode:       DefaultFilePermissions,
	}
}

func FileFromNetworkdUnit(unit types.Networkdunit) *File {
	return &File{
		Path:       filepath.Join(NetworkdUnitsPath(), string(unit.Name)),
		ReadCloser: ioutil.NopCloser(bytes.NewReader([]byte(unit.Contents))),
		Mode:       DefaultFilePermissions,
	}
}

func FileFromUnitDropin(unit types.Unit, dropin types.Dropin) *File {
	return &File{
		Path:       filepath.Join(SystemdDropinsPath(string(unit.Name)), string(dropin.Name)),
		ReadCloser: ioutil.NopCloser(bytes.NewReader([]byte(dropin.Contents))),
		Mode:       DefaultFilePermissions,
	}
}

func (u Util) MaskUnit(unit types.Unit) error {
	path := u.JoinPath(SystemdUnitsPath(), string(unit.Name))
	if err := MkdirForFile(path); err != nil {
		return err
	}
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return os.Symlink("/dev/null", path)
}

func (u Util) EnableUnit(unit types.Unit) error {
	path := u.JoinPath(presetPath)
	if err := MkdirForFile(path); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, DefaultPresetPermissions)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("enable %s\n", unit.Name))
	return err
}
