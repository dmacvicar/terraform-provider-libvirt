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

package regression

import (
	"github.com/coreos/ignition/tests/register"
	"github.com/coreos/ignition/tests/types"
)

func init() {
	register.Register(register.NegativeTest, VFATIgnoresWipeFilesystem())
}

func VFATIgnoresWipeFilesystem() types.Test {
	// Originally found in https://github.com/coreos/bugs/issues/2055

	name := "Regression: VFAT ignores WipeFilesystem"
	in := types.GetBaseDisk()
	out := in
	mntDevices := []types.MntDevice{
		{
			Label:        "OEM",
			Substitution: "$DEVICE",
		},
	}
	config := `{
                "ignition": {"version": "2.1.0"},
                "storage": {
                        "filesystems": [{
                                "mount": {
                                        "device": "$DEVICE",
                                        "format": "vfat",
                                        "wipeFilesystem": false
                                }}]
                        }
        }`

	return types.Test{
		Name:       name,
		In:         in,
		Out:        out,
		MntDevices: mntDevices,
		Config:     config,
	}
}
