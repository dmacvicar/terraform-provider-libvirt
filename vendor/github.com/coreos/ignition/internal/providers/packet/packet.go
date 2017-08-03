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

// The packet provider fetches a remote configuration from the packet.net
// userdata metadata service URL.

package packet

import (
	"net/url"

	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/config/validate/report"
	"github.com/coreos/ignition/internal/providers/util"
	"github.com/coreos/ignition/internal/resource"
)

var (
	userdataUrl = url.URL{
		Scheme: "https",
		Host:   "metadata.packet.net",
		Path:   "userdata",
	}
)

func FetchConfig(f resource.Fetcher) (types.Config, report.Report, error) {
	// Packet's metadata service returns "Not Acceptable" when queried
	// with the default Accept header.
	headers := resource.ConfigHeaders
	headers.Set("Accept", "*/*")
	data, err := f.FetchToBuffer(userdataUrl, resource.FetchOptions{
		Headers: headers,
	})
	if err != nil {
		return types.Config{}, report.Report{}, err
	}

	return util.ParseConfig(f.Logger, data)
}
