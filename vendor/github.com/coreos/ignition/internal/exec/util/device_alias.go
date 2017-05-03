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

package util

import (
	"os"
	"path/filepath"
)

const deviceAliasDir = "/dev_aliases"

// DeviceAlias returns the aliased form of the supplied path.
// Note device paths in ignition are always absolute.
func DeviceAlias(path string) string {
	return filepath.Join(deviceAliasDir, filepath.Clean(path))
}

// CreateDeviceAlias creates a device alias for the supplied path.
// On success the canonicalized path used as the alias target is returned.
func CreateDeviceAlias(path string) (string, error) {
	target, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}

	alias := DeviceAlias(path)

	if err := os.Remove(alias); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}

		if err = os.MkdirAll(filepath.Dir(alias), 0750); err != nil {
			return "", err
		}
	}

	if err = os.Symlink(target, alias); err != nil {
		return "", err
	}

	return target, nil
}
