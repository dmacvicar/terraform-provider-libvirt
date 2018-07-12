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

package config

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	from "github.com/coreos/ignition/config/v2_3_experimental/types"
	"github.com/coreos/ignition/internal/config/types"
)

func TestTranslate(t *testing.T) {
	type in struct {
		config from.Config
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
			in: in{config: from.Config{
				Ignition: from.Ignition{
					Config: from.IgnitionConfig{
						Append: []from.ConfigReference{
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
								Verification: from.Verification{
									Hash: strToPtr("func2-sum2"),
								},
							},
						},
						Replace: &from.ConfigReference{
							Source: (&url.URL{
								Scheme: "data",
								Opaque: ",file3",
							}).String(),
							Verification: from.Verification{
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
			in: in{config: from.Config{
				Ignition: from.Ignition{
					Timeouts: from.Timeouts{
						HTTPResponseHeaders: nil,
						HTTPTotal:           nil,
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{
					Version: types.MaxVersion.String(),
					Timeouts: types.Timeouts{
						HTTPResponseHeaders: nil,
						HTTPTotal:           nil,
					},
				},
			}},
		},
		{
			in: in{config: from.Config{
				Ignition: from.Ignition{
					Timeouts: from.Timeouts{
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
			in: in{config: from.Config{
				Ignition: from.Ignition{
					Timeouts: from.Timeouts{
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
			in: in{config: from.Config{
				Ignition: from.Ignition{
					Security: from.Security{
						TLS: from.TLS{
							CertificateAuthorities: []from.CaReference{
								{
									Source: "https://example.com/ca.pem",
								},
							},
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{
					Version: types.MaxVersion.String(),
					Security: types.Security{
						TLS: types.TLS{
							CertificateAuthorities: []types.CaReference{
								{
									Source: "https://example.com/ca.pem",
								},
							},
						},
					},
				},
			}},
		},
		{
			in: in{config: from.Config{
				Ignition: from.Ignition{
					Security: from.Security{
						TLS: from.TLS{
							CertificateAuthorities: []from.CaReference{
								{
									Source: "https://example.com/ca.pem",
								},
								{
									Source: "https://example.com/ca2.pem",
									Verification: from.Verification{
										Hash: strToPtr("sha512-adbebebd234245380"),
									},
								},
							},
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{
					Version: types.MaxVersion.String(),
					Security: types.Security{
						TLS: types.TLS{
							CertificateAuthorities: []types.CaReference{
								{
									Source: "https://example.com/ca.pem",
								},
								{
									Source: "https://example.com/ca2.pem",
									Verification: types.Verification{
										Hash: strToPtr("sha512-adbebebd234245380"),
									},
								},
							},
						},
					},
				},
			}},
		},
		{
			in: in{config: from.Config{
				Ignition: from.Ignition{Version: from.MaxVersion.String()},
				Storage: from.Storage{
					Disks: []from.Disk{
						{
							Device:    "/dev/sda",
							WipeTable: true,
							Partitions: []from.Partition{
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
			in: in{config: from.Config{
				Ignition: from.Ignition{Version: from.MaxVersion.String()},
				Storage: from.Storage{
					Raid: []from.Raid{
						{
							Name:  "fast",
							Level: "raid0",
							Devices: []from.Device{
								from.Device("/dev/sdc"),
								from.Device("/dev/sdd"),
							},
							Spares: 2,
						},
						{
							Name:  "durable",
							Level: "raid1",
							Devices: []from.Device{
								from.Device("/dev/sde"),
								from.Device("/dev/sdf"),
							},
							Spares: 3,
						},
						{
							Name:  "fast-and-durable",
							Level: "raid10",
							Devices: []from.Device{
								from.Device("/dev/sdg"),
								from.Device("/dev/sdh"),
								from.Device("/dev/sdi"),
								from.Device("/dev/sdj"),
							},
							Spares: 0,
							Options: []from.RaidOption{
								"--this-is-a-flag",
							},
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
						{
							Name:  "fast-and-durable",
							Level: "raid10",
							Devices: []types.Device{
								types.Device("/dev/sdg"),
								types.Device("/dev/sdh"),
								types.Device("/dev/sdi"),
								types.Device("/dev/sdj"),
							},
							Spares: 0,
							Options: []types.RaidOption{
								"--this-is-a-flag",
							},
						},
					},
				},
			}},
		},
		{
			in: in{config: from.Config{
				Ignition: from.Ignition{Version: from.MaxVersion.String()},
				Storage: from.Storage{
					Filesystems: []from.Filesystem{
						{
							Name: "filesystem-0",
							Mount: &from.Mount{
								Device: "/dev/disk/by-partlabel/ROOT",
								Format: "btrfs",
								Create: &from.Create{
									Force:   true,
									Options: []from.CreateOption{"-L", "ROOT"},
								},
								Label:          strToPtr("ROOT"),
								Options:        []from.MountOption{"--nodiscard"},
								UUID:           strToPtr("8A7A6E26-5E8F-4CCA-A654-46215D4696AC"),
								WipeFilesystem: true,
							},
						},
						{
							Name: "filesystem-1",
							Mount: &from.Mount{
								Device:         "/dev/disk/by-partlabel/DATA",
								Format:         "ext4",
								Label:          strToPtr("DATA"),
								Options:        []from.MountOption{"-b", "1024"},
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
			in: in{config: from.Config{
				Ignition: from.Ignition{Version: from.MaxVersion.String()},
				Storage: from.Storage{
					Files: []from.File{
						{
							Node: from.Node{
								Filesystem: "filesystem-0",
								Path:       "/opt/file1",
								User:       &from.NodeUser{ID: intToPtr(500)},
								Group:      &from.NodeGroup{ID: intToPtr(501)},
								Overwrite:  boolToPtr(false),
							},
							FileEmbedded1: from.FileEmbedded1{
								Mode: intToPtr(0664),
								Contents: from.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file1",
									}).String(),
									Verification: from.Verification{
										Hash: strToPtr("foobar"),
									},
								},
							},
						},
						{
							Node: from.Node{
								Filesystem: "filesystem-0",
								Path:       "/opt/file2",
								User:       &from.NodeUser{ID: intToPtr(502)},
								Group:      &from.NodeGroup{ID: intToPtr(503)},
							},
							FileEmbedded1: from.FileEmbedded1{
								Mode:   intToPtr(0644),
								Append: true,
								Contents: from.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file2",
									}).String(),
									Compression: "gzip",
								},
							},
						},
						{
							Node: from.Node{
								Filesystem: "filesystem-1",
								Path:       "/opt/file3",
								User:       &from.NodeUser{ID: intToPtr(1000)},
								Group:      &from.NodeGroup{ID: intToPtr(1001)},
							},
							FileEmbedded1: from.FileEmbedded1{
								Mode: intToPtr(0400),
								Contents: from.FileContents{
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
								User:       &types.NodeUser{ID: intToPtr(500)},
								Group:      &types.NodeGroup{ID: intToPtr(501)},
								Overwrite:  boolToPtr(false),
							},
							FileEmbedded1: types.FileEmbedded1{
								Mode: intToPtr(0664),
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
								User:       &types.NodeUser{ID: intToPtr(502)},
								Group:      &types.NodeGroup{ID: intToPtr(503)},
							},
							FileEmbedded1: types.FileEmbedded1{
								Mode:   intToPtr(0644),
								Append: true,
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
								User:       &types.NodeUser{ID: intToPtr(1000)},
								Group:      &types.NodeGroup{ID: intToPtr(1001)},
							},
							FileEmbedded1: types.FileEmbedded1{
								Mode: intToPtr(0400),
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
			in: in{config: from.Config{
				Ignition: from.Ignition{Version: from.MaxVersion.String()},
				Storage: from.Storage{
					Directories: []from.Directory{
						{
							Node: from.Node{
								Filesystem: "filesystem-1",
								Path:       "/opt/dir1",
								User:       &from.NodeUser{ID: intToPtr(500)},
								Group:      &from.NodeGroup{ID: intToPtr(501)},
								Overwrite:  boolToPtr(false),
							},
							DirectoryEmbedded1: from.DirectoryEmbedded1{
								Mode: intToPtr(0664),
							},
						},
						{
							Node: from.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/dir2",
								User:       &from.NodeUser{ID: intToPtr(502)},
								Group:      &from.NodeGroup{ID: intToPtr(503)},
							},
							DirectoryEmbedded1: from.DirectoryEmbedded1{
								Mode: intToPtr(0644),
							},
						},
						{
							Node: from.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/dir3",
								User:       &from.NodeUser{ID: intToPtr(1000)},
								Group:      &from.NodeGroup{ID: intToPtr(1001)},
							},
							DirectoryEmbedded1: from.DirectoryEmbedded1{
								Mode: intToPtr(0400),
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
								User:       &types.NodeUser{ID: intToPtr(500)},
								Group:      &types.NodeGroup{ID: intToPtr(501)},
								Overwrite:  boolToPtr(false),
							},
							DirectoryEmbedded1: types.DirectoryEmbedded1{
								Mode: intToPtr(0664),
							},
						},
						{
							Node: types.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/dir2",
								User:       &types.NodeUser{ID: intToPtr(502)},
								Group:      &types.NodeGroup{ID: intToPtr(503)},
							},
							DirectoryEmbedded1: types.DirectoryEmbedded1{
								Mode: intToPtr(0644),
							},
						},
						{
							Node: types.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/dir3",
								User:       &types.NodeUser{ID: intToPtr(1000)},
								Group:      &types.NodeGroup{ID: intToPtr(1001)},
							},
							DirectoryEmbedded1: types.DirectoryEmbedded1{
								Mode: intToPtr(0400),
							},
						},
					},
				},
			}},
		},
		{
			in: in{config: from.Config{
				Ignition: from.Ignition{Version: from.MaxVersion.String()},
				Storage: from.Storage{
					Links: []from.Link{
						{
							Node: from.Node{
								Filesystem: "filesystem-1",
								Path:       "/opt/link1",
								User:       &from.NodeUser{ID: intToPtr(500)},
								Group:      &from.NodeGroup{ID: intToPtr(501)},
								Overwrite:  boolToPtr(true),
							},
							LinkEmbedded1: from.LinkEmbedded1{
								Hard:   false,
								Target: "/opt/file1",
							},
						},
						{
							Node: from.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/link2",
								User:       &from.NodeUser{ID: intToPtr(502)},
								Group:      &from.NodeGroup{ID: intToPtr(503)},
							},
							LinkEmbedded1: from.LinkEmbedded1{
								Hard:   true,
								Target: "/opt/file2",
							},
						},
						{
							Node: from.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/link3",
								User:       &from.NodeUser{ID: intToPtr(1000)},
								Group:      &from.NodeGroup{ID: intToPtr(1001)},
							},
							LinkEmbedded1: from.LinkEmbedded1{
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
								User:       &types.NodeUser{ID: intToPtr(500)},
								Group:      &types.NodeGroup{ID: intToPtr(501)},
								Overwrite:  boolToPtr(true),
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
								User:       &types.NodeUser{ID: intToPtr(502)},
								Group:      &types.NodeGroup{ID: intToPtr(503)},
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
								User:       &types.NodeUser{ID: intToPtr(1000)},
								Group:      &types.NodeGroup{ID: intToPtr(1001)},
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
			in: in{from.Config{
				Systemd: from.Systemd{
					Units: []from.Unit{
						{
							Name:     "test1.service",
							Enable:   true,
							Contents: "test1 contents",
							Dropins: []from.SystemdDropin{
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
							Dropins: []types.SystemdDropin{
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
			in: in{from.Config{
				Networkd: from.Networkd{
					Units: []from.Networkdunit{
						{
							Name:     "test1.network",
							Contents: "test1 contents",
						},
						{
							Name: "test2.network",
							Dropins: []from.NetworkdDropin{
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
							Name: "test2.network",
							Dropins: []types.NetworkdDropin{
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
					},
				},
			}},
		},
		{
			in: in{from.Config{
				Ignition: from.Ignition{Version: from.MaxVersion.String()},
				Passwd: from.Passwd{
					Users: []from.PasswdUser{
						{
							Name:              "user 1",
							PasswordHash:      strToPtr("password 1"),
							SSHAuthorizedKeys: []from.SSHAuthorizedKey{"key1", "key2"},
						},
						{
							Name:              "user 2",
							PasswordHash:      strToPtr("password 2"),
							SSHAuthorizedKeys: []from.SSHAuthorizedKey{"key3", "key4"},
							Create: &from.Usercreate{
								UID:          intToPtr(123),
								Gecos:        "gecos",
								HomeDir:      "/home/user 2",
								NoCreateHome: true,
								PrimaryGroup: "wheel",
								Groups:       []from.UsercreateGroup{"wheel", "plugdev"},
								NoUserGroup:  true,
								System:       true,
								NoLogInit:    true,
								Shell:        "/bin/zsh",
							},
						},
						{
							Name:              "user 3",
							PasswordHash:      strToPtr("password 3"),
							SSHAuthorizedKeys: []from.SSHAuthorizedKey{"key5", "key6"},
							UID:               intToPtr(123),
							Gecos:             "gecos",
							HomeDir:           "/home/user 2",
							NoCreateHome:      true,
							PrimaryGroup:      "wheel",
							Groups:            []from.Group{"wheel", "plugdev"},
							NoUserGroup:       true,
							System:            true,
							NoLogInit:         true,
							Shell:             "/bin/zsh",
						},
						{
							Name:              "user 4",
							PasswordHash:      strToPtr("password 4"),
							SSHAuthorizedKeys: []from.SSHAuthorizedKey{"key7", "key8"},
							Create:            &from.Usercreate{},
						},
					},
					Groups: []from.PasswdGroup{
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
							Groups:            []types.Group{"wheel", "plugdev"},
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
		config := Translate(test.in.config)
		assert.Equal(t, test.out.config, config, "#%d: bad config", i)
	}
}
