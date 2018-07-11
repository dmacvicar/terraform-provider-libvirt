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

package systemd

import (
	"github.com/coreos/ignition/tests/register"
	"github.com/coreos/ignition/tests/types"
)

func init() {
	register.Register(register.PositiveTest, CreateSystemdService())
}

func CreateSystemdService() types.Test {
	name := "Create a systemd service"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	config := `{
		"ignition": { "version": "2.0.0" },
		"systemd": {
			"units": [{
				"name": "example.service",
				"enable": true,
				"contents": "[Service]\nType=oneshot\nExecStart=/usr/bin/echo Hello World\n\n[Install]\nWantedBy=multi-user.target"
			}]
		}
	}`
	out[0].Partitions.AddFiles("ROOT", []types.File{
		{
			Node: types.Node{
				Name:      "example.service",
				Directory: "etc/systemd/system",
			},
			Contents: "[Service]\nType=oneshot\nExecStart=/usr/bin/echo Hello World\n\n[Install]\nWantedBy=multi-user.target",
		},
		{
			Node: types.Node{
				Name:      "20-ignition.preset",
				Directory: "etc/systemd/system-preset",
			},
			Contents: "enable example.service\n",
		},
	})

	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}
