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

package system

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/coreos/ignition/config/validate/report"
	"github.com/coreos/ignition/internal/config/types"
	"github.com/coreos/ignition/internal/distro"
	"github.com/coreos/ignition/internal/log"
	"github.com/coreos/ignition/internal/providers"
	"github.com/coreos/ignition/internal/providers/util"
	"github.com/coreos/ignition/internal/resource"
)

const (
	baseFilename    = "base.ign"
	defaultFilename = "default.ign"
	userFilename    = "user.ign"
)

func FetchBaseConfig(logger *log.Logger) (types.Config, report.Report, error) {
	return fetchConfig(logger, baseFilename)
}

func FetchDefaultConfig(logger *log.Logger) (types.Config, report.Report, error) {
	return fetchConfig(logger, defaultFilename)
}

func FetchConfig(f resource.Fetcher) (types.Config, report.Report, error) {
	return fetchConfig(f.Logger, userFilename)
}

func fetchConfig(logger *log.Logger, filename string) (types.Config, report.Report, error) {
	path := filepath.Join(distro.SystemConfigDir(), filename)
	logger.Info("reading system config file %q", path)

	rawConfig, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		logger.Info("no config at %q", path)
		return types.Config{}, report.Report{}, providers.ErrNoProvider
	} else if err != nil {
		logger.Err("couldn't read config %q: %v", path, err)
		return types.Config{}, report.Report{}, err
	}
	return util.ParseConfig(logger, rawConfig)
}
