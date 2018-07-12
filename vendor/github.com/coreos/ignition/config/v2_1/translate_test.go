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

package v2_1

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/coreos/ignition/config/util"
	v2_0 "github.com/coreos/ignition/config/v2_0/types"
	"github.com/coreos/ignition/config/v2_1/types"
)

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
									Hash: util.StrToPtr("func2-sum2"),
								},
							},
						},
						Replace: &types.ConfigReference{
							Source: (&url.URL{
								Scheme: "data",
								Opaque: ",file3",
							}).String(),
							Verification: types.Verification{
								Hash: util.StrToPtr("func3-sum3"),
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
								User:       types.NodeUser{ID: util.IntToPtr(500)},
								Group:      types.NodeGroup{ID: util.IntToPtr(501)},
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
								User:       types.NodeUser{ID: util.IntToPtr(502)},
								Group:      types.NodeGroup{ID: util.IntToPtr(503)},
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
								User:       types.NodeUser{ID: util.IntToPtr(1000)},
								Group:      types.NodeGroup{ID: util.IntToPtr(1001)},
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
							PasswordHash:      util.StrToPtr("password 1"),
							SSHAuthorizedKeys: []types.SSHAuthorizedKey{"key1", "key2"},
						},
						{
							Name:              "user 2",
							PasswordHash:      util.StrToPtr("password 2"),
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
							PasswordHash:      util.StrToPtr("password 3"),
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
