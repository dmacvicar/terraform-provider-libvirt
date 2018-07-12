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

// The cmdline provider fetches a remote configuration from the URL specified
// in the kernel boot option "coreos.config.url".

package cmdline

import (
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/coreos/ignition/config/validate/report"
	"github.com/coreos/ignition/internal/config/types"
	"github.com/coreos/ignition/internal/distro"
	"github.com/coreos/ignition/internal/log"
	"github.com/coreos/ignition/internal/providers"
	"github.com/coreos/ignition/internal/providers/util"
	"github.com/coreos/ignition/internal/resource"
)

const (
	cmdlineUrlFlag = "coreos.config.url"
)

func FetchConfig(f resource.Fetcher) (types.Config, report.Report, error) {
	url, err := readCmdline(f.Logger)
	if err != nil {
		return types.Config{}, report.Report{}, err
	}

	if url == nil {
		return types.Config{}, report.Report{}, providers.ErrNoProvider
	}

	data, err := f.FetchToBuffer(*url, resource.FetchOptions{
		Headers: resource.ConfigHeaders,
	})
	if err != nil {
		return types.Config{}, report.Report{}, err
	}

	return util.ParseConfig(f.Logger, data)
}

func readCmdline(logger *log.Logger) (*url.URL, error) {
	args, err := ioutil.ReadFile(distro.KernelCmdlinePath())
	if err != nil {
		logger.Err("couldn't read cmdline: %v", err)
		return nil, err
	}

	rawUrl := parseCmdline(args)
	logger.Debug("parsed url from cmdline: %q", rawUrl)
	if rawUrl == "" {
		logger.Info("no config URL provided")
		return nil, nil
	}

	url, err := url.Parse(rawUrl)
	if err != nil {
		logger.Err("failed to parse url: %v", err)
		return nil, err
	}

	return url, err
}

func parseCmdline(cmdline []byte) (url string) {
	for _, arg := range strings.Split(string(cmdline), " ") {
		parts := strings.SplitN(strings.TrimSpace(arg), "=", 2)
		key := parts[0]

		if key != cmdlineUrlFlag {
			continue
		}

		if len(parts) == 2 {
			url = parts[1]
		}
	}

	return
}
