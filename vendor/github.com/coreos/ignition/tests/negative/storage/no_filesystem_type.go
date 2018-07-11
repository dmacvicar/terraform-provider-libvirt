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
	register.Register(register.NegativeTest, NoFilesystemType())
	register.Register(register.NegativeTest, NoFilesystemTypeWithForce())
	register.Register(register.NegativeTest, NoFilesystemTypeWithWipeFilesystem())
}

func NoFilesystemType() types.Test {
	name := "No Filesystem Type"
	in := types.GetBaseDisk()
	out := in
	mntDevices := []types.MntDevice{
		{
			Label:        "EFI-SYSTEM",
			Substitution: "$DEVICE",
		},
	}
	config := `{
		"ignition": {"version": "2.1.0"},
		"storage": {
			"filesystems": [{
				"mount": {
					"device": "$DEVICE"
				},
				"name": "foobar"
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

func NoFilesystemTypeWithForce() types.Test {
	name := "No Filesystem Type w/ Force"
	in := types.GetBaseDisk()
	out := in
	mntDevices := []types.MntDevice{
		{
			Label:        "EFI-SYSTEM",
			Substitution: "$DEVICE",
		},
	}
	config := `{
		"ignition": {"version": "2.0.0"},
		"storage": {
			"filesystems": [{
				"mount": {
					"device": "$DEVICE",
					"create": {
						"force": true
					}
				},
				"name": "foobar"
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

func NoFilesystemTypeWithWipeFilesystem() types.Test {
	name := "No Filesystem Type w/ wipeFilesystem"
	in := types.GetBaseDisk()
	out := in
	mntDevices := []types.MntDevice{
		{
			Label:        "EFI-SYSTEM",
			Substitution: "$DEVICE",
		},
	}
	config := `{
		"ignition": {"version": "2.1.0"},
		"storage": {
			"filesystems": [{
				"mount": {
					"device": "$DEVICE",
					"wipeFilesystem": true
				},
				"name": "foobar"
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
