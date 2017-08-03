// Copyright 2017 CoreOS, Inc.
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

// The virtualbox provider fetches the configuration from raw data on a partition
// with the GUID 99570a8a-f826-4eb0-ba4e-9dd72d55ea13

package virtualbox

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/coreos/ignition/config"
	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/config/validate/report"
	"github.com/coreos/ignition/internal/providers/util"
	"github.com/coreos/ignition/internal/resource"
)

const (
	configPath = "/dev/disk/by-partuuid/99570a8a-f826-4eb0-ba4e-9dd72d55ea13"
)

func FetchConfig(f resource.Fetcher) (types.Config, report.Report, error) {
	f.Logger.Debug("Attempting to read config drive")
	rawConfig, err := ioutil.ReadFile(configPath)
	if os.IsNotExist(err) {
		f.Logger.Info("Path to ignition config does not exist, assuming no config")
		return types.Config{}, report.Report{}, config.ErrEmpty
	} else if err != nil {
		f.Logger.Err("Error reading ignition config: %v", err)
		return types.Config{}, report.Report{}, err
	}
	trimmedConfig := bytes.TrimRight(rawConfig, "\x00")
	return util.ParseConfig(f.Logger, trimmedConfig)
}
