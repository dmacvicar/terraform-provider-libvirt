// Copyright 2018 CoreOS, Inc.
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

package files

import (
	"github.com/coreos/ignition/tests/register"
	"github.com/coreos/ignition/tests/types"
)

func init() {
	register.Register(register.NegativeTest, AppendToDirectory())
	register.Register(register.NegativeTest, AppendAndOverwrite())
}

func AppendToDirectory() types.Test {
	name := "Append To Directory"
	in := types.GetBaseDisk()
	out := in
	mntDevices := []types.MntDevice{
		{
			Label:        "EFI-SYSTEM",
			Substitution: "$DEVICE",
		},
	}

	config := `{
	    "ignition": {"version": "2.2.0"},
	    "storage": {
	      "files": [{
	      "filesystem": "root",
	      "path": "/foo/bar",
	      "contents": { "source": "data:,hello%20world%0A" },
	      "append": true
	    }]
	  }
	}`
	in[0].Partitions.AddDirectories("ROOT", []types.Directory{
		{
			Node: types.Node{
				Directory: "foo",
				Name:      "bar",
			},
		},
	})

	return types.Test{
		Name:       name,
		In:         in,
		Out:        out,
		MntDevices: mntDevices,
		Config:     config,
	}
}

func AppendAndOverwrite() types.Test {
	name := "Append and Overwrite"
	in := types.GetBaseDisk()
	out := in
	mntDevices := []types.MntDevice{
		{
			Label:        "EFI-SYSTEM",
			Substitution: "$DEVICE",
		},
	}

	config := `{
		"ignition": {"version": "2.2.0"},
		"storage": {
		  "files": [{
	      "filesystem": "root",
	      "path": "/foo/bar",
	      "contents": { "source": "data:,hello%20world%0A" },
	      "append": true,
		  "overwrite": true
	    }]
	  }
	}`

	return types.Test{
		Name:              name,
		In:                in,
		Out:               out,
		MntDevices:        mntDevices,
		Config:            config,
		ConfigShouldBeBad: true,
	}
}
