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

package file

import (
	"io/ioutil"
	"os"

	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/config/validate/report"
	"github.com/coreos/ignition/internal/log"
	"github.com/coreos/ignition/internal/providers/util"
	"github.com/coreos/ignition/internal/resource"
)

const (
	cfgFilenameEnvVar = "IGNITION_CONFIG_FILE"
	defaultFilename   = "config.ign"
)

func FetchConfig(logger *log.Logger, _ *resource.HttpClient) (types.Config, report.Report, error) {
	filename := os.Getenv(cfgFilenameEnvVar)
	if filename == "" {
		filename = defaultFilename
		logger.Info("using default filename")
	}
	logger.Info("using config file at %q", filename)

	rawConfig, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Err("couldn't read config %q: %v", filename, err)
		return types.Config{}, report.Report{}, err
	}
	return util.ParseConfig(logger, rawConfig)
}
