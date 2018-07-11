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

// The ec2 provider fetches a remote configuration from the ec2 user-data
// metadata service URL.

package oracleoci

import (
	"encoding/base64"
	"net/url"

	"github.com/coreos/ignition/config/validate/report"
	"github.com/coreos/ignition/internal/config/types"
	"github.com/coreos/ignition/internal/providers/util"
	"github.com/coreos/ignition/internal/resource"
)

var (
	userdataUrl = url.URL{
		Scheme: "http",
		Host:   "169.254.169.254",
		Path:   "opc/v1/instance/metadata/user_data",
	}
)

func FetchConfig(f resource.Fetcher) (types.Config, report.Report, error) {
	data, err := f.FetchToBuffer(userdataUrl, resource.FetchOptions{
		Headers: resource.ConfigHeaders,
	})
	if err == resource.ErrNotFound {
		// The user didn't provide an Ignition config
		return types.Config{}, report.Report{}, nil
	}
	if err != nil {
		return types.Config{}, report.Report{}, err
	}

	var decodedData []byte
	decodedData, err = base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return types.Config{}, report.Report{}, err
	}

	return util.ParseConfig(f.Logger, decodedData)
}
