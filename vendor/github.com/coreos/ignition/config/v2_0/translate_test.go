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

package v2_0

import (
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/assert"

	v1 "github.com/coreos/ignition/config/v1/types"
	"github.com/coreos/ignition/config/v2_0/types"
)

func TestTranslateFromV1(t *testing.T) {
	type in struct {
		config v1.Config
	}
	type out struct {
		config types.Config
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{},
			out: out{config: types.Config{Ignition: types.Ignition{Version: types.IgnitionVersion(types.MaxVersion)}}},
		},
		{
			in: in{config: v1.Config{
				Storage: v1.Storage{
					Disks: []v1.Disk{
						{
							Device:    v1.Path("/dev/sda"),
							WipeTable: true,
							Partitions: []v1.Partition{
								{
									Label:    v1.PartitionLabel("ROOT"),
									Number:   7,
									Size:     v1.PartitionDimension(100),
									Start:    v1.PartitionDimension(50),
									TypeGUID: "HI",
								},
								{
									Label:    v1.PartitionLabel("DATA"),
									Number:   12,
									Size:     v1.PartitionDimension(1000),
									Start:    v1.PartitionDimension(300),
									TypeGUID: "LO",
								},
							},
						},
						{
							Device:    v1.Path("/dev/sdb"),
							WipeTable: true,
						},
					},
					Arrays: []v1.Raid{
						{
							Name:    "fast",
							Level:   "raid0",
							Devices: []v1.Path{v1.Path("/dev/sdc"), v1.Path("/dev/sdd")},
							Spares:  2,
						},
						{
							Name:    "durable",
							Level:   "raid1",
							Devices: []v1.Path{v1.Path("/dev/sde"), v1.Path("/dev/sdf")},
							Spares:  3,
						},
					},
					Filesystems: []v1.Filesystem{
						{
							Device: v1.Path("/dev/disk/by-partlabel/ROOT"),
							Format: v1.FilesystemFormat("btrfs"),
							Create: &v1.FilesystemCreate{
								Force:   true,
								Options: v1.MkfsOptions([]string{"-L", "ROOT"}),
							},
							Files: []v1.File{
								{
									Path:     v1.Path("/opt/file1"),
									Contents: "file1",
									Mode:     v1.FileMode(0664),
									Uid:      500,
									Gid:      501,
								},
								{
									Path:     v1.Path("/opt/file2"),
									Contents: "file2",
									Mode:     v1.FileMode(0644),
									Uid:      502,
									Gid:      503,
								},
							},
						},
						{
							Device: v1.Path("/dev/disk/by-partlabel/DATA"),
							Format: v1.FilesystemFormat("ext4"),
							Files: []v1.File{
								{
									Path:     v1.Path("/opt/file3"),
									Contents: "file3",
									Mode:     v1.FileMode(0400),
									Uid:      1000,
									Gid:      1001,
								},
							},
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.IgnitionVersion(semver.Version{Major: 2})},
				Storage: types.Storage{
					Disks: []types.Disk{
						{
							Device:    "/dev/sda",
							WipeTable: true,
							Partitions: []types.Partition{
								{
									Label:    "ROOT",
									Number:   7,
									Size:     100,
									Start:    50,
									TypeGUID: "HI",
								},
								{
									Label:    "DATA",
									Number:   12,
									Size:     1000,
									Start:    300,
									TypeGUID: "LO",
								},
							},
						},
						{
							Device:    "/dev/sdb",
							WipeTable: true,
						},
					},
					Arrays: []types.Raid{
						{
							Name:    "fast",
							Level:   "raid0",
							Devices: []types.Path{types.Path("/dev/sdc"), types.Path("/dev/sdd")},
							Spares:  2,
						},
						{
							Name:    "durable",
							Level:   "raid1",
							Devices: []types.Path{types.Path("/dev/sde"), types.Path("/dev/sdf")},
							Spares:  3,
						},
					},
					Filesystems: []types.Filesystem{
						{
							Name: "_translate-filesystem-0",
							Mount: &types.FilesystemMount{
								Device: "/dev/disk/by-partlabel/ROOT",
								Format: "btrfs",
								Create: &types.FilesystemCreate{
									Force:   true,
									Options: types.MkfsOptions{"-L", "ROOT"},
								},
							},
						},
						{
							Name: "_translate-filesystem-1",
							Mount: &types.FilesystemMount{
								Device: "/dev/disk/by-partlabel/DATA",
								Format: "ext4",
							},
						},
					},
					Files: []types.File{
						{
							Filesystem: "_translate-filesystem-0",
							Path:       "/opt/file1",
							User:       types.FileUser{Id: 500},
							Group:      types.FileGroup{Id: 501},
							Mode:       0664,
							Contents: types.FileContents{
								Source: types.Url{
									Scheme: "data",
									Opaque: ",file1",
								},
							},
						},
						{
							Filesystem: "_translate-filesystem-0",
							Path:       "/opt/file2",
							User:       types.FileUser{Id: 502},
							Group:      types.FileGroup{Id: 503},
							Mode:       0644,
							Contents: types.FileContents{
								Source: types.Url{
									Scheme: "data",
									Opaque: ",file2",
								},
							},
						},
						{
							Filesystem: "_translate-filesystem-1",
							Path:       "/opt/file3",
							User:       types.FileUser{Id: 1000},
							Group:      types.FileGroup{Id: 1001},
							Mode:       0400,
							Contents: types.FileContents{
								Source: types.Url{
									Scheme: "data",
									Opaque: ",file3",
								},
							},
						},
					},
				},
			}},
		},
		{
			in: in{v1.Config{
				Systemd: v1.Systemd{
					Units: []v1.SystemdUnit{
						{
							Name:     "test1.service",
							Enable:   true,
							Contents: "test1 contents",
							DropIns: []v1.SystemdUnitDropIn{
								{
									Name:     "conf1.conf",
									Contents: "conf1 contents",
								},
								{
									Name:     "conf2.conf",
									Contents: "conf2 contents",
								},
							},
						},
						{
							Name:     "test2.service",
							Mask:     true,
							Contents: "test2 contents",
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.IgnitionVersion(semver.Version{Major: 2})},
				Systemd: types.Systemd{
					Units: []types.SystemdUnit{
						{
							Name:     "test1.service",
							Enable:   true,
							Contents: "test1 contents",
							DropIns: []types.SystemdUnitDropIn{
								{
									Name:     "conf1.conf",
									Contents: "conf1 contents",
								},
								{
									Name:     "conf2.conf",
									Contents: "conf2 contents",
								},
							},
						},
						{
							Name:     "test2.service",
							Mask:     true,
							Contents: "test2 contents",
						},
					},
				},
			}},
		},
		{
			in: in{v1.Config{
				Networkd: v1.Networkd{
					Units: []v1.NetworkdUnit{
						{
							Name:     "test1.network",
							Contents: "test1 contents",
						},
						{
							Name:     "test2.network",
							Contents: "test2 contents",
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.IgnitionVersion(semver.Version{Major: 2})},
				Networkd: types.Networkd{
					Units: []types.NetworkdUnit{
						{
							Name:     "test1.network",
							Contents: "test1 contents",
						},
						{
							Name:     "test2.network",
							Contents: "test2 contents",
						},
					},
				},
			}},
		},
		{
			in: in{v1.Config{
				Passwd: v1.Passwd{
					Users: []v1.User{
						{
							Name:              "user 1",
							PasswordHash:      "password 1",
							SSHAuthorizedKeys: []string{"key1", "key2"},
						},
						{
							Name:              "user 2",
							PasswordHash:      "password 2",
							SSHAuthorizedKeys: []string{"key3", "key4"},
							Create: &v1.UserCreate{
								Uid:          func(i uint) *uint { return &i }(123),
								GECOS:        "gecos",
								Homedir:      "/home/user 2",
								NoCreateHome: true,
								PrimaryGroup: "wheel",
								Groups:       []string{"wheel", "plugdev"},
								NoUserGroup:  true,
								System:       true,
								NoLogInit:    true,
								Shell:        "/bin/zsh",
							},
						},
						{
							Name:              "user 3",
							PasswordHash:      "password 3",
							SSHAuthorizedKeys: []string{"key5", "key6"},
							Create:            &v1.UserCreate{},
						},
					},
					Groups: []v1.Group{
						{
							Name:         "group 1",
							Gid:          func(i uint) *uint { return &i }(1000),
							PasswordHash: "password 1",
							System:       true,
						},
						{
							Name:         "group 2",
							PasswordHash: "password 2",
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.IgnitionVersion(semver.Version{Major: 2})},
				Passwd: types.Passwd{
					Users: []types.User{
						{
							Name:              "user 1",
							PasswordHash:      "password 1",
							SSHAuthorizedKeys: []string{"key1", "key2"},
						},
						{
							Name:              "user 2",
							PasswordHash:      "password 2",
							SSHAuthorizedKeys: []string{"key3", "key4"},
							Create: &types.UserCreate{
								Uid:          func(i uint) *uint { return &i }(123),
								GECOS:        "gecos",
								Homedir:      "/home/user 2",
								NoCreateHome: true,
								PrimaryGroup: "wheel",
								Groups:       []string{"wheel", "plugdev"},
								NoUserGroup:  true,
								System:       true,
								NoLogInit:    true,
								Shell:        "/bin/zsh",
							},
						},
						{
							Name:              "user 3",
							PasswordHash:      "password 3",
							SSHAuthorizedKeys: []string{"key5", "key6"},
							Create:            &types.UserCreate{},
						},
					},
					Groups: []types.Group{
						{
							Name:         "group 1",
							Gid:          func(i uint) *uint { return &i }(1000),
							PasswordHash: "password 1",
							System:       true,
						},
						{
							Name:         "group 2",
							PasswordHash: "password 2",
						},
					},
				},
			}},
		},
	}

	for i, test := range tests {
		config := TranslateFromV1(test.in.config)
		assert.Equal(t, test.out.config, config, "#%d: bad config", i)
	}
}
