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
	register.Register(register.PositiveTest, ModifySystemdService())
	register.Register(register.PositiveTest, MaskSystemdServices())
}

func ModifySystemdService() types.Test {
	name := "Modify Services"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	config := `{
	  "ignition": { "version": "2.0.0" },
	  "systemd": {
	    "units": [{
	      "name": "systemd-networkd.service",
	      "dropins": [{
	        "name": "debug.conf",
	        "contents": "[Service]\nEnvironment=SYSTEMD_LOG_LEVEL=debug"
	      }]
	    }]
	  }
	}`
	out[0].Partitions.AddFiles("ROOT", []types.File{
		{
			Node: types.Node{
				Name:      "debug.conf",
				Directory: "etc/systemd/system/systemd-networkd.service.d",
			},
			Contents: "[Service]\nEnvironment=SYSTEMD_LOG_LEVEL=debug",
		},
	})

	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}

func MaskSystemdServices() types.Test {
	name := "Mask Services"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	config := `{
	  "ignition": { "version": "2.0.0" },
	  "systemd": {
	    "units": [{
	      "name": "systemd-networkd.service",
		  "mask": true
	    }]
	  }
	}`
	out[0].Partitions.AddLinks("ROOT", []types.Link{
		{
			Node: types.Node{
				Name:      "systemd-networkd.service",
				Directory: "etc/systemd/system",
			},
			Target: "/dev/null",
			Hard:   false,
		},
	})

	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}
