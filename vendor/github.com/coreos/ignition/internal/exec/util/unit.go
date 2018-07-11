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
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	configUtil "github.com/coreos/ignition/config/util"
	"github.com/coreos/ignition/internal/config/types"

	"github.com/vincent-petithory/dataurl"
)

const (
	presetPath               string      = "/etc/systemd/system-preset/20-ignition.preset"
	DefaultPresetPermissions os.FileMode = 0644
)

func FileFromSystemdUnit(unit types.Unit) (*FetchOp, error) {
	u, err := url.Parse(dataurl.EncodeBytes([]byte(unit.Contents)))
	if err != nil {
		return nil, err
	}
	return &FetchOp{
		Path: filepath.Join(SystemdUnitsPath(), string(unit.Name)),
		Url:  *u,
		Mode: configUtil.IntToPtr(int(DefaultFilePermissions)),
	}, nil
}

func FileFromNetworkdUnit(unit types.Networkdunit) (*FetchOp, error) {
	u, err := url.Parse(dataurl.EncodeBytes([]byte(unit.Contents)))
	if err != nil {
		return nil, err
	}
	return &FetchOp{
		Path: filepath.Join(NetworkdUnitsPath(), string(unit.Name)),
		Url:  *u,
		Mode: configUtil.IntToPtr(int(DefaultFilePermissions)),
	}, nil
}

func FileFromSystemdUnitDropin(unit types.Unit, dropin types.SystemdDropin) (*FetchOp, error) {
	u, err := url.Parse(dataurl.EncodeBytes([]byte(dropin.Contents)))
	if err != nil {
		return nil, err
	}
	return &FetchOp{
		Path: filepath.Join(SystemdDropinsPath(string(unit.Name)), string(dropin.Name)),
		Url:  *u,
		Mode: configUtil.IntToPtr(int(DefaultFilePermissions)),
	}, nil
}

func FileFromNetworkdUnitDropin(unit types.Networkdunit, dropin types.NetworkdDropin) (*FetchOp, error) {
	u, err := url.Parse(dataurl.EncodeBytes([]byte(dropin.Contents)))
	if err != nil {
		return nil, err
	}
	return &FetchOp{
		Path: filepath.Join(NetworkdDropinsPath(string(unit.Name)), string(dropin.Name)),
		Url:  *u,
		Mode: configUtil.IntToPtr(int(DefaultFilePermissions)),
	}, nil
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
	return u.appendLineToPreset(fmt.Sprintf("enable %s", unit.Name))
}

func (u Util) DisableUnit(unit types.Unit) error {
	return u.appendLineToPreset(fmt.Sprintf("disable %s", unit.Name))
}

func (u Util) appendLineToPreset(data string) error {
	path := u.JoinPath(presetPath)
	if err := MkdirForFile(path); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, DefaultPresetPermissions)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(data + "\n")
	return err
}
