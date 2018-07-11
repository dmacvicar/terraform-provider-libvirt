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
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	deviceAliasDir      = "/dev_aliases"
	retrySymlinkDelay   = 10 * time.Millisecond
	retrySymlinkTimeout = 30 * time.Second
	retrySymlinkCount   = int(retrySymlinkTimeout / retrySymlinkDelay)
)

// DeviceAlias returns the aliased form of the supplied path.
// Note device paths in ignition are always absolute.
func DeviceAlias(path string) string {
	return filepath.Join(deviceAliasDir, filepath.Clean(path))
}

// evalSymlinks wraps filepath.EvalSymlinks, retrying if it fails
func evalSymlinks(path string) (res string, err error) {
	for i := 0; i < retrySymlinkCount; i++ {
		res, err = filepath.EvalSymlinks(path)
		if err == nil {
			return res, nil
		} else if os.IsNotExist(err) {
			time.Sleep(retrySymlinkDelay)
		} else {
			return "", err
		}
	}

	return "", fmt.Errorf("Failed to evaluate symlink after %v: %v", retrySymlinkTimeout, err)
}

// CreateDeviceAlias creates a device alias for the supplied path.
// On success the canonicalized path used as the alias target is returned.
func CreateDeviceAlias(path string) (string, error) {
	target, err := evalSymlinks(path)
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
