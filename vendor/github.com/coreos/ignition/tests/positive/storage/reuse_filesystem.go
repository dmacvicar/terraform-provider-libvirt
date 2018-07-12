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

package storage

import (
	"github.com/coreos/ignition/tests/register"
	"github.com/coreos/ignition/tests/types"
)

func init() {
	register.Register(register.PositiveTest, ReuseExistingFilesystem())
}

func ReuseExistingFilesystem() types.Test {
	name := "Reuse Existing Filesystem"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	mntDevices := []types.MntDevice{
		{
			Label:        "important-data",
			Substitution: "$DEVICE",
		},
	}
	config := `{
		"ignition": {"version": "2.1.0"},
		"storage": {
		    "filesystems": [
			    {
					"mount": {
						"device": "$DEVICE",
						"wipeFilesystem": false,
						"format": "btrfs",
						"label": "data",
						"uuid": "8A7A6E26-5E8F-4CCA-A654-46215D4696AC"
					}
				}
			]
		}
	}`
	in = append(in, types.Disk{
		Alignment: types.IgnitionAlignment,
		Partitions: types.Partitions{
			{
				Label:           "important-data",
				Number:          1,
				Length:          2621440,
				FilesystemType:  "btrfs",
				FilesystemLabel: "data",
				FilesystemUUID:  "8A7A6E26-5E8F-4CCA-A654-46215D4696AC",
				Files: []types.File{
					{
						Node: types.Node{
							Name:      "bar",
							Directory: "foo",
						},
						Contents: "example file\n",
					},
				},
			},
		},
	})
	out = append(out, types.Disk{
		Alignment: types.IgnitionAlignment,
		Partitions: types.Partitions{
			{
				Label:           "important-data",
				Number:          1,
				Length:          2621440,
				FilesystemType:  "btrfs",
				FilesystemLabel: "data",
				FilesystemUUID:  "8A7A6E26-5E8F-4CCA-A654-46215D4696AC",
				Files: []types.File{
					{
						Node: types.Node{
							Name:      "bar",
							Directory: "foo",
						},
						Contents: "example file\n",
					},
				},
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
