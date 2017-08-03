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

package v2_1

import (
	"reflect"
	"testing"

	"github.com/coreos/go-semver/semver"

	"github.com/coreos/ignition/config/types"
)

func TestAppend(t *testing.T) {
	type in struct {
		oldConfig types.Config
		newConfig types.Config
	}
	type out struct {
		config types.Config
	}

	tests := []struct {
		in  in
		out out
	}{
		// empty
		{
			in: in{
				oldConfig: types.Config{},
				newConfig: types.Config{},
			},
			out: out{config: types.Config{}},
		},

		// merge tags
		{
			in: in{
				oldConfig: types.Config{},
				newConfig: types.Config{
					Ignition: types.Ignition{
						Version: semver.Version{Major: 2}.String(),
					},
				},
			},
			out: out{config: types.Config{}},
		},
		{
			in: in{
				oldConfig: types.Config{
					Ignition: types.Ignition{
						Version: semver.Version{Major: 2}.String(),
					},
				},
				newConfig: types.Config{},
			},
			out: out{config: types.Config{
				Ignition: types.Ignition{
					Version: semver.Version{Major: 2}.String(),
				},
			}},
		},
		{
			in: in{
				oldConfig: types.Config{},
				newConfig: types.Config{
					Ignition: types.Ignition{
						Config: types.IgnitionConfig{
							Replace: &types.ConfigReference{},
						},
					},
				},
			},
			out: out{config: types.Config{
				Ignition: types.Ignition{
					Config: types.IgnitionConfig{
						Replace: &types.ConfigReference{},
					},
				},
			}},
		},
		{
			in: in{
				oldConfig: types.Config{
					Ignition: types.Ignition{
						Config: types.IgnitionConfig{
							Replace: &types.ConfigReference{},
						},
					},
				},
				newConfig: types.Config{},
			},
			out: out{config: types.Config{}},
		},

		// old
		{
			in: in{
				oldConfig: types.Config{
					Storage: types.Storage{
						Disks: []types.Disk{
							{
								WipeTable: true,
								Partitions: []types.Partition{
									{Number: 1},
									{Number: 2},
								},
							},
						},
					},
				},
				newConfig: types.Config{},
			},
			out: out{config: types.Config{
				Storage: types.Storage{
					Disks: []types.Disk{
						{
							WipeTable: true,
							Partitions: []types.Partition{
								{Number: 1},
								{Number: 2},
							},
						},
					},
				},
			}},
		},

		// new
		{
			in: in{
				oldConfig: types.Config{},
				newConfig: types.Config{
					Systemd: types.Systemd{
						Units: []types.Unit{
							{Name: "test1.service"},
							{Name: "test2.service"},
						},
					},
				},
			},
			out: out{config: types.Config{
				Systemd: types.Systemd{
					Units: []types.Unit{
						{Name: "test1.service"},
						{Name: "test2.service"},
					},
				},
			}},
		},

		// both
		{
			in: in{
				oldConfig: types.Config{
					Passwd: types.Passwd{
						Users: []types.PasswdUser{
							{Name: "oldUser"},
						},
					},
				},
				newConfig: types.Config{
					Passwd: types.Passwd{
						Users: []types.PasswdUser{
							{Name: "newUser"},
						},
					},
				},
			},
			out: out{config: types.Config{
				Passwd: types.Passwd{
					Users: []types.PasswdUser{
						{Name: "oldUser"},
						{Name: "newUser"},
					},
				},
			}},
		},
	}

	for i, test := range tests {
		config := Append(test.in.oldConfig, test.in.newConfig)
		if !reflect.DeepEqual(test.out.config, config) {
			t.Errorf("#%d: bad config: want %+v, got %+v", i, test.out.config, config)
		}
	}
}
