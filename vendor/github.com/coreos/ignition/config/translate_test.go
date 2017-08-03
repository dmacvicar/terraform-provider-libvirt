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

package config

import (
	"net/url"
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/assert"

	"github.com/coreos/ignition/config/types"
	v1 "github.com/coreos/ignition/config/v1/types"
	v2_0 "github.com/coreos/ignition/config/v2_0/types"
	v2_1 "github.com/coreos/ignition/config/v2_1/types"
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
			out: out{config: types.Config{Ignition: types.Ignition{Version: v2_0.MaxVersion.String()}}},
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
				Ignition: types.Ignition{Version: semver.Version{Major: 2}.String()},
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
					Raid: []types.Raid{
						{
							Name:    "fast",
							Level:   "raid0",
							Devices: []types.Device{types.Device("/dev/sdc"), types.Device("/dev/sdd")},
							Spares:  2,
						},
						{
							Name:    "durable",
							Level:   "raid1",
							Devices: []types.Device{types.Device("/dev/sde"), types.Device("/dev/sdf")},
							Spares:  3,
						},
					},
					Filesystems: []types.Filesystem{
						{
							Name: "_translate-filesystem-0",
							Mount: &types.Mount{
								Device: "/dev/disk/by-partlabel/ROOT",
								Format: "btrfs",
								Create: &types.Create{
									Force:   true,
									Options: []types.CreateOption{"-L", "ROOT"},
								},
							},
						},
						{
							Name: "_translate-filesystem-1",
							Mount: &types.Mount{
								Device: "/dev/disk/by-partlabel/DATA",
								Format: "ext4",
							},
						},
					},
					Files: []types.File{
						{
							Node: types.Node{
								Filesystem: "_translate-filesystem-0",
								Path:       "/opt/file1",
								User:       types.NodeUser{ID: intToPtr(500)},
								Group:      types.NodeGroup{ID: intToPtr(501)},
							},
							FileEmbedded1: types.FileEmbedded1{
								Mode: 0664,
								Contents: types.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file1",
									}).String(),
								},
							},
						},
						{
							Node: types.Node{
								Filesystem: "_translate-filesystem-0",
								Path:       "/opt/file2",
								User:       types.NodeUser{ID: intToPtr(502)},
								Group:      types.NodeGroup{ID: intToPtr(503)},
							},
							FileEmbedded1: types.FileEmbedded1{
								Mode: 0644,
								Contents: types.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file2",
									}).String(),
								},
							},
						},
						{
							Node: types.Node{
								Filesystem: "_translate-filesystem-1",
								Path:       "/opt/file3",
								User:       types.NodeUser{ID: intToPtr(1000)},
								Group:      types.NodeGroup{ID: intToPtr(1001)},
							},
							FileEmbedded1: types.FileEmbedded1{
								Mode: 0400,
								Contents: types.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file3",
									}).String(),
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
				Ignition: types.Ignition{Version: semver.Version{Major: 2}.String()},
				Systemd: types.Systemd{
					Units: []types.Unit{
						{
							Name:     "test1.service",
							Enable:   true,
							Contents: "test1 contents",
							Dropins: []types.Dropin{
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
				Ignition: types.Ignition{Version: semver.Version{Major: 2}.String()},
				Networkd: types.Networkd{
					Units: []types.Networkdunit{
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
				Ignition: types.Ignition{Version: semver.Version{Major: 2}.String()},
				Passwd: types.Passwd{
					Users: []types.PasswdUser{
						{
							Name:              "user 1",
							PasswordHash:      strToPtr("password 1"),
							SSHAuthorizedKeys: []types.SSHAuthorizedKey{"key1", "key2"},
						},
						{
							Name:              "user 2",
							PasswordHash:      strToPtr("password 2"),
							SSHAuthorizedKeys: []types.SSHAuthorizedKey{"key3", "key4"},
							Create: &types.Usercreate{
								UID:          func(i int) *int { return &i }(123),
								Gecos:        "gecos",
								HomeDir:      "/home/user 2",
								NoCreateHome: true,
								PrimaryGroup: "wheel",
								Groups:       []types.UsercreateGroup{"wheel", "plugdev"},
								NoUserGroup:  true,
								System:       true,
								NoLogInit:    true,
								Shell:        "/bin/zsh",
							},
						},
						{
							Name:              "user 3",
							PasswordHash:      strToPtr("password 3"),
							SSHAuthorizedKeys: []types.SSHAuthorizedKey{"key5", "key6"},
							Create:            &types.Usercreate{},
						},
					},
					Groups: []types.PasswdGroup{
						{
							Name:         "group 1",
							Gid:          func(i int) *int { return &i }(1000),
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

func TestTranslateFromV2_0(t *testing.T) {
	type in struct {
		config v2_0.Config
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
			out: out{config: types.Config{Ignition: types.Ignition{Version: types.MaxVersion.String()}}},
		},
		{
			in: in{config: v2_0.Config{
				Ignition: v2_0.Ignition{
					Config: v2_0.IgnitionConfig{
						Append: []v2_0.ConfigReference{
							{
								Source: v2_0.Url{
									Scheme: "data",
									Opaque: ",file1",
								},
							},
							{
								Source: v2_0.Url{
									Scheme: "data",
									Opaque: ",file2",
								},
								Verification: v2_0.Verification{
									Hash: &v2_0.Hash{
										Function: "func2",
										Sum:      "sum2",
									},
								},
							},
						},
						Replace: &v2_0.ConfigReference{
							Source: v2_0.Url{
								Scheme: "data",
								Opaque: ",file3",
							},
							Verification: v2_0.Verification{
								Hash: &v2_0.Hash{
									Function: "func3",
									Sum:      "sum3",
								},
							},
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{
					Version: types.MaxVersion.String(),
					Config: types.IgnitionConfig{
						Append: []types.ConfigReference{
							{
								Source: (&url.URL{
									Scheme: "data",
									Opaque: ",file1",
								}).String(),
							},
							{
								Source: (&url.URL{
									Scheme: "data",
									Opaque: ",file2",
								}).String(),
								Verification: types.Verification{
									Hash: strToPtr("func2-sum2"),
								},
							},
						},
						Replace: &types.ConfigReference{
							Source: (&url.URL{
								Scheme: "data",
								Opaque: ",file3",
							}).String(),
							Verification: types.Verification{
								Hash: strToPtr("func3-sum3"),
							},
						},
					},
				},
			}},
		},
		{
			in: in{config: v2_0.Config{
				Storage: v2_0.Storage{
					Disks: []v2_0.Disk{
						{
							Device:    v2_0.Path("/dev/sda"),
							WipeTable: true,
							Partitions: []v2_0.Partition{
								{
									Label:    v2_0.PartitionLabel("ROOT"),
									Number:   7,
									Size:     v2_0.PartitionDimension(100),
									Start:    v2_0.PartitionDimension(50),
									TypeGUID: "HI",
								},
								{
									Label:    v2_0.PartitionLabel("DATA"),
									Number:   12,
									Size:     v2_0.PartitionDimension(1000),
									Start:    v2_0.PartitionDimension(300),
									TypeGUID: "LO",
								},
							},
						},
						{
							Device:    v2_0.Path("/dev/sdb"),
							WipeTable: true,
						},
					},
					Arrays: []v2_0.Raid{
						{
							Name:    "fast",
							Level:   "raid0",
							Devices: []v2_0.Path{v2_0.Path("/dev/sdc"), v2_0.Path("/dev/sdd")},
							Spares:  2,
						},
						{
							Name:    "durable",
							Level:   "raid1",
							Devices: []v2_0.Path{v2_0.Path("/dev/sde"), v2_0.Path("/dev/sdf")},
							Spares:  3,
						},
					},
					Filesystems: []v2_0.Filesystem{
						{
							Name: "filesystem-0",
							Mount: &v2_0.FilesystemMount{
								Device: v2_0.Path("/dev/disk/by-partlabel/ROOT"),
								Format: v2_0.FilesystemFormat("btrfs"),
								Create: &v2_0.FilesystemCreate{
									Force:   true,
									Options: v2_0.MkfsOptions([]string{"-L", "ROOT"}),
								},
							},
						},
						{
							Name: "filesystem-1",
							Mount: &v2_0.FilesystemMount{
								Device: v2_0.Path("/dev/disk/by-partlabel/DATA"),
								Format: v2_0.FilesystemFormat("ext4"),
							},
						},
						{
							Name: "filesystem-2",
							Path: func(p v2_0.Path) *v2_0.Path { return &p }("/foo"),
						},
					},
					Files: []v2_0.File{
						{
							Filesystem: "filesystem-0",
							Path:       v2_0.Path("/opt/file1"),
							Mode:       v2_0.FileMode(0664),
							User:       v2_0.FileUser{Id: 500},
							Group:      v2_0.FileGroup{Id: 501},
							Contents: v2_0.FileContents{
								Source: v2_0.Url{
									Scheme: "data",
									Opaque: ",file1",
								},
							},
						},
						{
							Filesystem: "filesystem-0",
							Path:       v2_0.Path("/opt/file2"),
							Mode:       v2_0.FileMode(0644),
							User:       v2_0.FileUser{Id: 502},
							Group:      v2_0.FileGroup{Id: 503},
							Contents: v2_0.FileContents{
								Source: v2_0.Url{
									Scheme: "data",
									Opaque: ",file2",
								},
							},
						},
						{
							Filesystem: "filesystem-1",
							Path:       v2_0.Path("/opt/file3"),
							Mode:       v2_0.FileMode(0400),
							User:       v2_0.FileUser{Id: 1000},
							Group:      v2_0.FileGroup{Id: 1001},
							Contents: v2_0.FileContents{
								Source: v2_0.Url{
									Scheme: "data",
									Opaque: ",file3",
								},
							},
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
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
					Raid: []types.Raid{
						{
							Name:    "fast",
							Level:   "raid0",
							Devices: []types.Device{types.Device("/dev/sdc"), types.Device("/dev/sdd")},
							Spares:  2,
						},
						{
							Name:    "durable",
							Level:   "raid1",
							Devices: []types.Device{types.Device("/dev/sde"), types.Device("/dev/sdf")},
							Spares:  3,
						},
					},
					Filesystems: []types.Filesystem{
						{
							Name: "filesystem-0",
							Mount: &types.Mount{
								Device: "/dev/disk/by-partlabel/ROOT",
								Format: "btrfs",
								Create: &types.Create{
									Force:   true,
									Options: []types.CreateOption{"-L", "ROOT"},
								},
							},
						},
						{
							Name: "filesystem-1",
							Mount: &types.Mount{
								Device: "/dev/disk/by-partlabel/DATA",
								Format: "ext4",
							},
						},
						{
							Name: "filesystem-2",
							Path: func(p string) *string { return &p }("/foo"),
						},
					},
					Files: []types.File{
						{
							Node: types.Node{
								Filesystem: "filesystem-0",
								Path:       "/opt/file1",
								User:       types.NodeUser{ID: intToPtr(500)},
								Group:      types.NodeGroup{ID: intToPtr(501)},
							},
							FileEmbedded1: types.FileEmbedded1{
								Mode: 0664,
								Contents: types.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file1",
									}).String(),
								},
							},
						},
						{
							Node: types.Node{
								Filesystem: "filesystem-0",
								Path:       "/opt/file2",
								User:       types.NodeUser{ID: intToPtr(502)},
								Group:      types.NodeGroup{ID: intToPtr(503)},
							},
							FileEmbedded1: types.FileEmbedded1{
								Mode: 0644,
								Contents: types.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file2",
									}).String(),
								},
							},
						},
						{
							Node: types.Node{
								Filesystem: "filesystem-1",
								Path:       "/opt/file3",
								User:       types.NodeUser{ID: intToPtr(1000)},
								Group:      types.NodeGroup{ID: intToPtr(1001)},
							},
							FileEmbedded1: types.FileEmbedded1{
								Mode: 0400,
								Contents: types.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file3",
									}).String(),
								},
							},
						},
					},
				},
			}},
		},
		{
			in: in{v2_0.Config{
				Systemd: v2_0.Systemd{
					Units: []v2_0.SystemdUnit{
						{
							Name:     "test1.service",
							Enable:   true,
							Contents: "test1 contents",
							DropIns: []v2_0.SystemdUnitDropIn{
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
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Systemd: types.Systemd{
					Units: []types.Unit{
						{
							Name:     "test1.service",
							Enable:   true,
							Contents: "test1 contents",
							Dropins: []types.Dropin{
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
			in: in{v2_0.Config{
				Networkd: v2_0.Networkd{
					Units: []v2_0.NetworkdUnit{
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
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Networkd: types.Networkd{
					Units: []types.Networkdunit{
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
			in: in{v2_0.Config{
				Passwd: v2_0.Passwd{
					Users: []v2_0.User{
						{
							Name:              "user 1",
							PasswordHash:      "password 1",
							SSHAuthorizedKeys: []string{"key1", "key2"},
						},
						{
							Name:              "user 2",
							PasswordHash:      "password 2",
							SSHAuthorizedKeys: []string{"key3", "key4"},
							Create: &v2_0.UserCreate{
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
							Create:            &v2_0.UserCreate{},
						},
					},
					Groups: []v2_0.Group{
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
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Passwd: types.Passwd{
					Users: []types.PasswdUser{
						{
							Name:              "user 1",
							PasswordHash:      strToPtr("password 1"),
							SSHAuthorizedKeys: []types.SSHAuthorizedKey{"key1", "key2"},
						},
						{
							Name:              "user 2",
							PasswordHash:      strToPtr("password 2"),
							SSHAuthorizedKeys: []types.SSHAuthorizedKey{"key3", "key4"},
							Create: &types.Usercreate{
								UID:          func(i int) *int { return &i }(123),
								Gecos:        "gecos",
								HomeDir:      "/home/user 2",
								NoCreateHome: true,
								PrimaryGroup: "wheel",
								Groups:       []types.UsercreateGroup{"wheel", "plugdev"},
								NoUserGroup:  true,
								System:       true,
								NoLogInit:    true,
								Shell:        "/bin/zsh",
							},
						},
						{
							Name:              "user 3",
							PasswordHash:      strToPtr("password 3"),
							SSHAuthorizedKeys: []types.SSHAuthorizedKey{"key5", "key6"},
							Create:            &types.Usercreate{},
						},
					},
					Groups: []types.PasswdGroup{
						{
							Name:         "group 1",
							Gid:          func(i int) *int { return &i }(1000),
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
		config := TranslateFromV2_0(test.in.config)
		assert.Equal(t, config, test.out.config, "#%d: bad config", i)
	}
}

func TestTranslateFromV2_1(t *testing.T) {
	type in struct {
		config v2_1.Config
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
			out: out{config: types.Config{Ignition: types.Ignition{Version: types.MaxVersion.String()}}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{
					Config: v2_1.IgnitionConfig{
						Append: []v2_1.ConfigReference{
							{
								Source: (&url.URL{
									Scheme: "data",
									Opaque: ",file1",
								}).String(),
							},
							{
								Source: (&url.URL{
									Scheme: "data",
									Opaque: ",file2",
								}).String(),
								Verification: v2_1.Verification{
									Hash: strToPtr("func2-sum2"),
								},
							},
						},
						Replace: &v2_1.ConfigReference{
							Source: (&url.URL{
								Scheme: "data",
								Opaque: ",file3",
							}).String(),
							Verification: v2_1.Verification{
								Hash: strToPtr("func3-sum3"),
							},
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{
					Version: types.MaxVersion.String(),
					Config: types.IgnitionConfig{
						Append: []types.ConfigReference{
							{
								Source: (&url.URL{
									Scheme: "data",
									Opaque: ",file1",
								}).String(),
							},
							{
								Source: (&url.URL{
									Scheme: "data",
									Opaque: ",file2",
								}).String(),
								Verification: types.Verification{
									Hash: strToPtr("func2-sum2"),
								},
							},
						},
						Replace: &types.ConfigReference{
							Source: (&url.URL{
								Scheme: "data",
								Opaque: ",file3",
							}).String(),
							Verification: types.Verification{
								Hash: strToPtr("func3-sum3"),
							},
						},
					},
				},
			}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{
					Timeouts: v2_1.Timeouts{
						HTTPResponseHeaders: intToPtr(0),
						HTTPTotal:           intToPtr(0),
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{
					Version: types.MaxVersion.String(),
					Timeouts: types.Timeouts{
						HTTPResponseHeaders: intToPtr(0),
						HTTPTotal:           intToPtr(0),
					},
				},
			}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{
					Timeouts: v2_1.Timeouts{
						HTTPResponseHeaders: intToPtr(50),
						HTTPTotal:           intToPtr(100),
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{
					Version: types.MaxVersion.String(),
					Timeouts: types.Timeouts{
						HTTPResponseHeaders: intToPtr(50),
						HTTPTotal:           intToPtr(100),
					},
				},
			}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{Version: v2_1.MaxVersion.String()},
				Storage: v2_1.Storage{
					Disks: []v2_1.Disk{
						{
							Device:    "/dev/sda",
							WipeTable: true,
							Partitions: []v2_1.Partition{
								{
									Label:    "ROOT",
									Number:   7,
									Size:     100,
									Start:    50,
									TypeGUID: "HI",
									GUID:     "4F68BCE3-E8CD-4DB1-96E7-FBCAF984B709",
								},
								{
									Label:    "DATA",
									Number:   12,
									Size:     1000,
									Start:    300,
									TypeGUID: "LO",
									GUID:     "3B8F8425-20E0-4F3B-907F-1A25A76F98E8",
								},
							},
						},
						{
							Device:    "/dev/sdb",
							WipeTable: true,
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
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
									GUID:     "4F68BCE3-E8CD-4DB1-96E7-FBCAF984B709",
								},
								{
									Label:    "DATA",
									Number:   12,
									Size:     1000,
									Start:    300,
									TypeGUID: "LO",
									GUID:     "3B8F8425-20E0-4F3B-907F-1A25A76F98E8",
								},
							},
						},
						{
							Device:    "/dev/sdb",
							WipeTable: true,
						},
					},
				},
			}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{Version: v2_1.MaxVersion.String()},
				Storage: v2_1.Storage{
					Raid: []v2_1.Raid{
						{
							Name:  "fast",
							Level: "raid0",
							Devices: []v2_1.Device{
								v2_1.Device("/dev/sdc"),
								v2_1.Device("/dev/sdd"),
							},
							Spares: 2,
						},
						{
							Name:  "durable",
							Level: "raid1",
							Devices: []v2_1.Device{
								v2_1.Device("/dev/sde"),
								v2_1.Device("/dev/sdf"),
							},
							Spares: 3,
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Storage: types.Storage{
					Raid: []types.Raid{
						{
							Name:    "fast",
							Level:   "raid0",
							Devices: []types.Device{types.Device("/dev/sdc"), types.Device("/dev/sdd")},
							Spares:  2,
						},
						{
							Name:    "durable",
							Level:   "raid1",
							Devices: []types.Device{types.Device("/dev/sde"), types.Device("/dev/sdf")},
							Spares:  3,
						},
					},
				},
			}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{Version: v2_1.MaxVersion.String()},
				Storage: v2_1.Storage{
					Filesystems: []v2_1.Filesystem{
						{
							Name: "filesystem-0",
							Mount: &v2_1.Mount{
								Device: "/dev/disk/by-partlabel/ROOT",
								Format: "btrfs",
								Create: &v2_1.Create{
									Force:   true,
									Options: []v2_1.CreateOption{"-L", "ROOT"},
								},
								Label:          strToPtr("ROOT"),
								Options:        []v2_1.MountOption{"--nodiscard"},
								UUID:           strToPtr("8A7A6E26-5E8F-4CCA-A654-46215D4696AC"),
								WipeFilesystem: true,
							},
						},
						{
							Name: "filesystem-1",
							Mount: &v2_1.Mount{
								Device:         "/dev/disk/by-partlabel/DATA",
								Format:         "ext4",
								Label:          strToPtr("DATA"),
								Options:        []v2_1.MountOption{"-b", "1024"},
								UUID:           strToPtr("8A7A6E26-5E8F-4CCA-A654-DEADBEEF0101"),
								WipeFilesystem: false,
							},
						},
						{
							Name: "filesystem-2",
							Path: func(p string) *string { return &p }("/foo"),
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Storage: types.Storage{
					Filesystems: []types.Filesystem{
						{
							Name: "filesystem-0",
							Mount: &types.Mount{
								Device: "/dev/disk/by-partlabel/ROOT",
								Format: "btrfs",
								Create: &types.Create{
									Force:   true,
									Options: []types.CreateOption{"-L", "ROOT"},
								},
								Label:          strToPtr("ROOT"),
								Options:        []types.MountOption{"--nodiscard"},
								UUID:           strToPtr("8A7A6E26-5E8F-4CCA-A654-46215D4696AC"),
								WipeFilesystem: true,
							},
						},
						{
							Name: "filesystem-1",
							Mount: &types.Mount{
								Device:         "/dev/disk/by-partlabel/DATA",
								Format:         "ext4",
								Label:          strToPtr("DATA"),
								Options:        []types.MountOption{"-b", "1024"},
								UUID:           strToPtr("8A7A6E26-5E8F-4CCA-A654-DEADBEEF0101"),
								WipeFilesystem: false,
							},
						},
						{
							Name: "filesystem-2",
							Path: strToPtr("/foo"),
						},
					},
				},
			}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{Version: v2_1.MaxVersion.String()},
				Storage: v2_1.Storage{
					Files: []v2_1.File{
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-0",
								Path:       "/opt/file1",
								User:       v2_1.NodeUser{ID: intToPtr(500)},
								Group:      v2_1.NodeGroup{ID: intToPtr(501)},
							},
							FileEmbedded1: v2_1.FileEmbedded1{
								Mode: 0664,
								Contents: v2_1.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file1",
									}).String(),
									Verification: v2_1.Verification{
										Hash: strToPtr("foobar"),
									},
								},
							},
						},
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-0",
								Path:       "/opt/file2",
								User:       v2_1.NodeUser{ID: intToPtr(502)},
								Group:      v2_1.NodeGroup{ID: intToPtr(503)},
							},
							FileEmbedded1: v2_1.FileEmbedded1{
								Mode: 0644,
								Contents: v2_1.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file2",
									}).String(),
									Compression: "gzip",
								},
							},
						},
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-1",
								Path:       "/opt/file3",
								User:       v2_1.NodeUser{ID: intToPtr(1000)},
								Group:      v2_1.NodeGroup{ID: intToPtr(1001)},
							},
							FileEmbedded1: v2_1.FileEmbedded1{
								Mode: 0400,
								Contents: v2_1.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file3",
									}).String(),
								},
							},
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Storage: types.Storage{
					Files: []types.File{
						{
							Node: types.Node{
								Filesystem: "filesystem-0",
								Path:       "/opt/file1",
								User:       types.NodeUser{ID: intToPtr(500)},
								Group:      types.NodeGroup{ID: intToPtr(501)},
							},
							FileEmbedded1: types.FileEmbedded1{
								Mode: 0664,
								Contents: types.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file1",
									}).String(),
									Verification: types.Verification{
										Hash: strToPtr("foobar"),
									},
								},
							},
						},
						{
							Node: types.Node{
								Filesystem: "filesystem-0",
								Path:       "/opt/file2",
								User:       types.NodeUser{ID: intToPtr(502)},
								Group:      types.NodeGroup{ID: intToPtr(503)},
							},
							FileEmbedded1: types.FileEmbedded1{
								Mode: 0644,
								Contents: types.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file2",
									}).String(),
									Compression: "gzip",
								},
							},
						},
						{
							Node: types.Node{
								Filesystem: "filesystem-1",
								Path:       "/opt/file3",
								User:       types.NodeUser{ID: intToPtr(1000)},
								Group:      types.NodeGroup{ID: intToPtr(1001)},
							},
							FileEmbedded1: types.FileEmbedded1{
								Mode: 0400,
								Contents: types.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file3",
									}).String(),
								},
							},
						},
					},
				},
			}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{Version: v2_1.MaxVersion.String()},
				Storage: v2_1.Storage{
					Directories: []v2_1.Directory{
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-1",
								Path:       "/opt/dir1",
								User:       v2_1.NodeUser{ID: intToPtr(500)},
								Group:      v2_1.NodeGroup{ID: intToPtr(501)},
							},
							DirectoryEmbedded1: v2_1.DirectoryEmbedded1{
								Mode: 0664,
							},
						},
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/dir2",
								User:       v2_1.NodeUser{ID: intToPtr(502)},
								Group:      v2_1.NodeGroup{ID: intToPtr(503)},
							},
							DirectoryEmbedded1: v2_1.DirectoryEmbedded1{
								Mode: 0644,
							},
						},
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/dir3",
								User:       v2_1.NodeUser{ID: intToPtr(1000)},
								Group:      v2_1.NodeGroup{ID: intToPtr(1001)},
							},
							DirectoryEmbedded1: v2_1.DirectoryEmbedded1{
								Mode: 0400,
							},
						},
					},
				}},
			},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Storage: types.Storage{
					Directories: []types.Directory{
						{
							Node: types.Node{
								Filesystem: "filesystem-1",
								Path:       "/opt/dir1",
								User:       types.NodeUser{ID: intToPtr(500)},
								Group:      types.NodeGroup{ID: intToPtr(501)},
							},
							DirectoryEmbedded1: types.DirectoryEmbedded1{
								Mode: 0664,
							},
						},
						{
							Node: types.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/dir2",
								User:       types.NodeUser{ID: intToPtr(502)},
								Group:      types.NodeGroup{ID: intToPtr(503)},
							},
							DirectoryEmbedded1: types.DirectoryEmbedded1{
								Mode: 0644,
							},
						},
						{
							Node: types.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/dir3",
								User:       types.NodeUser{ID: intToPtr(1000)},
								Group:      types.NodeGroup{ID: intToPtr(1001)},
							},
							DirectoryEmbedded1: types.DirectoryEmbedded1{
								Mode: 0400,
							},
						},
					},
				},
			}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{Version: v2_1.MaxVersion.String()},
				Storage: v2_1.Storage{
					Links: []v2_1.Link{
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-1",
								Path:       "/opt/link1",
								User:       v2_1.NodeUser{ID: intToPtr(500)},
								Group:      v2_1.NodeGroup{ID: intToPtr(501)},
							},
							LinkEmbedded1: v2_1.LinkEmbedded1{
								Hard:   false,
								Target: "/opt/file1",
							},
						},
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/link2",
								User:       v2_1.NodeUser{ID: intToPtr(502)},
								Group:      v2_1.NodeGroup{ID: intToPtr(503)},
							},
							LinkEmbedded1: v2_1.LinkEmbedded1{
								Hard:   true,
								Target: "/opt/file2",
							},
						},
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/link3",
								User:       v2_1.NodeUser{ID: intToPtr(1000)},
								Group:      v2_1.NodeGroup{ID: intToPtr(1001)},
							},
							LinkEmbedded1: v2_1.LinkEmbedded1{
								Hard:   true,
								Target: "/opt/file3",
							},
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Storage: types.Storage{
					Links: []types.Link{
						{
							Node: types.Node{
								Filesystem: "filesystem-1",
								Path:       "/opt/link1",
								User:       types.NodeUser{ID: intToPtr(500)},
								Group:      types.NodeGroup{ID: intToPtr(501)},
							},
							LinkEmbedded1: types.LinkEmbedded1{
								Hard:   false,
								Target: "/opt/file1",
							},
						},
						{
							Node: types.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/link2",
								User:       types.NodeUser{ID: intToPtr(502)},
								Group:      types.NodeGroup{ID: intToPtr(503)},
							},
							LinkEmbedded1: types.LinkEmbedded1{
								Hard:   true,
								Target: "/opt/file2",
							},
						},
						{
							Node: types.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/link3",
								User:       types.NodeUser{ID: intToPtr(1000)},
								Group:      types.NodeGroup{ID: intToPtr(1001)},
							},
							LinkEmbedded1: types.LinkEmbedded1{
								Hard:   true,
								Target: "/opt/file3",
							},
						},
					},
				},
			}},
		},
		{
			in: in{v2_1.Config{
				Systemd: v2_1.Systemd{
					Units: []v2_1.Unit{
						{
							Name:     "test1.service",
							Enable:   true,
							Contents: "test1 contents",
							Dropins: []v2_1.Dropin{
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
						{
							Name:    "test3.service",
							Enabled: boolToPtr(false),
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Systemd: types.Systemd{
					Units: []types.Unit{
						{
							Name:     "test1.service",
							Enable:   true,
							Contents: "test1 contents",
							Dropins: []types.Dropin{
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
						{
							Name:    "test3.service",
							Enabled: boolToPtr(false),
						},
					},
				},
			}},
		},
		{
			in: in{v2_1.Config{
				Networkd: v2_1.Networkd{
					Units: []v2_1.Networkdunit{
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
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Networkd: types.Networkd{
					Units: []types.Networkdunit{
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
			in: in{v2_1.Config{
				Ignition: v2_1.Ignition{Version: v2_1.MaxVersion.String()},
				Passwd: v2_1.Passwd{
					Users: []v2_1.PasswdUser{
						{
							Name:              "user 1",
							PasswordHash:      strToPtr("password 1"),
							SSHAuthorizedKeys: []v2_1.SSHAuthorizedKey{"key1", "key2"},
						},
						{
							Name:              "user 2",
							PasswordHash:      strToPtr("password 2"),
							SSHAuthorizedKeys: []v2_1.SSHAuthorizedKey{"key3", "key4"},
							Create: &v2_1.Usercreate{
								UID:          intToPtr(123),
								Gecos:        "gecos",
								HomeDir:      "/home/user 2",
								NoCreateHome: true,
								PrimaryGroup: "wheel",
								Groups:       []v2_1.UsercreateGroup{"wheel", "plugdev"},
								NoUserGroup:  true,
								System:       true,
								NoLogInit:    true,
								Shell:        "/bin/zsh",
							},
						},
						{
							Name:              "user 3",
							PasswordHash:      strToPtr("password 3"),
							SSHAuthorizedKeys: []v2_1.SSHAuthorizedKey{"key5", "key6"},
							UID:               intToPtr(123),
							Gecos:             "gecos",
							HomeDir:           "/home/user 2",
							NoCreateHome:      true,
							PrimaryGroup:      "wheel",
							Groups:            []v2_1.PasswdUserGroup{"wheel", "plugdev"},
							NoUserGroup:       true,
							System:            true,
							NoLogInit:         true,
							Shell:             "/bin/zsh",
						},
						{
							Name:              "user 4",
							PasswordHash:      strToPtr("password 4"),
							SSHAuthorizedKeys: []v2_1.SSHAuthorizedKey{"key7", "key8"},
							Create:            &v2_1.Usercreate{},
						},
					},
					Groups: []v2_1.PasswdGroup{
						{
							Name:         "group 1",
							Gid:          func(i int) *int { return &i }(1000),
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
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Passwd: types.Passwd{
					Users: []types.PasswdUser{
						{
							Name:              "user 1",
							PasswordHash:      strToPtr("password 1"),
							SSHAuthorizedKeys: []types.SSHAuthorizedKey{"key1", "key2"},
						},
						{
							Name:              "user 2",
							PasswordHash:      strToPtr("password 2"),
							SSHAuthorizedKeys: []types.SSHAuthorizedKey{"key3", "key4"},
							Create: &types.Usercreate{
								UID:          func(i int) *int { return &i }(123),
								Gecos:        "gecos",
								HomeDir:      "/home/user 2",
								NoCreateHome: true,
								PrimaryGroup: "wheel",
								Groups:       []types.UsercreateGroup{"wheel", "plugdev"},
								NoUserGroup:  true,
								System:       true,
								NoLogInit:    true,
								Shell:        "/bin/zsh",
							},
						},
						{
							Name:              "user 3",
							PasswordHash:      strToPtr("password 3"),
							SSHAuthorizedKeys: []types.SSHAuthorizedKey{"key5", "key6"},
							UID:               intToPtr(123),
							Gecos:             "gecos",
							HomeDir:           "/home/user 2",
							NoCreateHome:      true,
							PrimaryGroup:      "wheel",
							Groups:            []types.PasswdUserGroup{"wheel", "plugdev"},
							NoUserGroup:       true,
							System:            true,
							NoLogInit:         true,
							Shell:             "/bin/zsh",
						},
						{
							Name:              "user 4",
							PasswordHash:      strToPtr("password 4"),
							SSHAuthorizedKeys: []types.SSHAuthorizedKey{"key7", "key8"},
							Create:            &types.Usercreate{},
						},
					},
					Groups: []types.PasswdGroup{
						{
							Name:         "group 1",
							Gid:          func(i int) *int { return &i }(1000),
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
		config := TranslateFromV2_1(test.in.config)
		assert.Equal(t, test.out.config, config, "#%d: bad config", i)
	}
}
