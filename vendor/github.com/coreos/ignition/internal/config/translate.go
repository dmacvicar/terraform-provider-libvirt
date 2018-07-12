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
	from "github.com/coreos/ignition/config/v2_3_experimental/types"
	"github.com/coreos/ignition/internal/config/types"
)

func intToPtr(x int) *int {
	return &x
}

func strToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func boolToPtr(b bool) *bool {
	return &b
}

func Translate(old from.Config) types.Config {
	translateConfigReference := func(old *from.ConfigReference) *types.ConfigReference {
		if old == nil {
			return nil
		}
		return &types.ConfigReference{
			Source: old.Source,
			Verification: types.Verification{
				Hash: old.Verification.Hash,
			},
		}
	}
	translateConfigReferenceSlice := func(old []from.ConfigReference) []types.ConfigReference {
		var res []types.ConfigReference
		for _, c := range old {
			res = append(res, *translateConfigReference(&c))
		}
		return res
	}
	translateCertificateAuthoritySlice := func(old []from.CaReference) []types.CaReference {
		var res []types.CaReference
		for _, x := range old {
			res = append(res, types.CaReference{
				Source: x.Source,
				Verification: types.Verification{
					Hash: x.Verification.Hash,
				},
			})
		}
		return res
	}
	translateNetworkdDropinSlice := func(old []from.NetworkdDropin) []types.NetworkdDropin {
		var res []types.NetworkdDropin
		for _, x := range old {
			res = append(res, types.NetworkdDropin{
				Contents: x.Contents,
				Name:     x.Name,
			})
		}
		return res
	}
	translateNetworkdUnitSlice := func(old []from.Networkdunit) []types.Networkdunit {
		var res []types.Networkdunit
		for _, u := range old {
			res = append(res, types.Networkdunit{
				Contents: u.Contents,
				Name:     u.Name,
				Dropins:  translateNetworkdDropinSlice(u.Dropins),
			})
		}
		return res
	}
	translatePasswdGroupSlice := func(old []from.PasswdGroup) []types.PasswdGroup {
		var res []types.PasswdGroup
		for _, g := range old {
			res = append(res, types.PasswdGroup{
				Gid:          g.Gid,
				Name:         g.Name,
				PasswordHash: g.PasswordHash,
				System:       g.System,
			})
		}
		return res
	}
	translatePasswdUsercreateGroupSlice := func(old []from.UsercreateGroup) []types.UsercreateGroup {
		var res []types.UsercreateGroup
		for _, g := range old {
			res = append(res, types.UsercreateGroup(g))
		}
		return res
	}
	translatePasswdUsercreate := func(old *from.Usercreate) *types.Usercreate {
		if old == nil {
			return nil
		}
		return &types.Usercreate{
			Gecos:        old.Gecos,
			Groups:       translatePasswdUsercreateGroupSlice(old.Groups),
			HomeDir:      old.HomeDir,
			NoCreateHome: old.NoCreateHome,
			NoLogInit:    old.NoLogInit,
			NoUserGroup:  old.NoUserGroup,
			PrimaryGroup: old.PrimaryGroup,
			Shell:        old.Shell,
			System:       old.System,
			UID:          old.UID,
		}
	}
	translatePasswdUserGroupSlice := func(old []from.Group) []types.Group {
		var res []types.Group
		for _, g := range old {
			res = append(res, types.Group(g))
		}
		return res
	}
	translatePasswdSSHAuthorizedKeySlice := func(old []from.SSHAuthorizedKey) []types.SSHAuthorizedKey {
		res := make([]types.SSHAuthorizedKey, len(old))
		for i, k := range old {
			res[i] = types.SSHAuthorizedKey(k)
		}
		return res
	}
	translatePasswdUserSlice := func(old []from.PasswdUser) []types.PasswdUser {
		var res []types.PasswdUser
		for _, u := range old {
			res = append(res, types.PasswdUser{
				Create:            translatePasswdUsercreate(u.Create),
				Gecos:             u.Gecos,
				Groups:            translatePasswdUserGroupSlice(u.Groups),
				HomeDir:           u.HomeDir,
				Name:              u.Name,
				NoCreateHome:      u.NoCreateHome,
				NoLogInit:         u.NoLogInit,
				NoUserGroup:       u.NoUserGroup,
				PasswordHash:      u.PasswordHash,
				PrimaryGroup:      u.PrimaryGroup,
				SSHAuthorizedKeys: translatePasswdSSHAuthorizedKeySlice(u.SSHAuthorizedKeys),
				Shell:             u.Shell,
				System:            u.System,
				UID:               u.UID,
			})
		}
		return res
	}
	translateNodeGroup := func(old *from.NodeGroup) *types.NodeGroup {
		if old == nil {
			return nil
		}
		return &types.NodeGroup{
			ID:   old.ID,
			Name: old.Name,
		}
	}
	translateNodeUser := func(old *from.NodeUser) *types.NodeUser {
		if old == nil {
			return nil
		}
		return &types.NodeUser{
			ID:   old.ID,
			Name: old.Name,
		}
	}
	translateNode := func(old from.Node) types.Node {
		return types.Node{
			Filesystem: old.Filesystem,
			Group:      translateNodeGroup(old.Group),
			Path:       old.Path,
			User:       translateNodeUser(old.User),
			Overwrite:  old.Overwrite,
		}
	}
	translateDirectorySlice := func(old []from.Directory) []types.Directory {
		var res []types.Directory
		for _, x := range old {
			res = append(res, types.Directory{
				Node: translateNode(x.Node),
				DirectoryEmbedded1: types.DirectoryEmbedded1{
					Mode: x.DirectoryEmbedded1.Mode,
				},
			})
		}
		return res
	}
	translatePartitionSlice := func(old []from.Partition) []types.Partition {
		var res []types.Partition
		for _, x := range old {
			res = append(res, types.Partition{
				GUID:     x.GUID,
				Label:    x.Label,
				Number:   x.Number,
				Size:     x.Size,
				Start:    x.Start,
				TypeGUID: x.TypeGUID,
			})
		}
		return res
	}
	translateDiskSlice := func(old []from.Disk) []types.Disk {
		var res []types.Disk
		for _, x := range old {
			res = append(res, types.Disk{
				Device:     x.Device,
				Partitions: translatePartitionSlice(x.Partitions),
				WipeTable:  x.WipeTable,
			})
		}
		return res
	}
	translateFileSlice := func(old []from.File) []types.File {
		var res []types.File
		for _, x := range old {
			res = append(res, types.File{
				Node: translateNode(x.Node),
				FileEmbedded1: types.FileEmbedded1{
					Contents: types.FileContents{
						Compression: x.Contents.Compression,
						Source:      x.Contents.Source,
						Verification: types.Verification{
							Hash: x.Contents.Verification.Hash,
						},
					},
					Mode:   x.Mode,
					Append: x.Append,
				},
			})
		}
		return res
	}
	translateMountCreateOptionSlice := func(old []from.CreateOption) []types.CreateOption {
		var res []types.CreateOption
		for _, x := range old {
			res = append(res, types.CreateOption(x))
		}
		return res
	}
	translateMountCreate := func(old *from.Create) *types.Create {
		if old == nil {
			return nil
		}
		return &types.Create{
			Force:   old.Force,
			Options: translateMountCreateOptionSlice(old.Options),
		}
	}
	translateMountOptionSlice := func(old []from.MountOption) []types.MountOption {
		var res []types.MountOption
		for _, x := range old {
			res = append(res, types.MountOption(x))
		}
		return res
	}
	translateMount := func(old *from.Mount) *types.Mount {
		if old == nil {
			return nil
		}
		return &types.Mount{
			Create:         translateMountCreate(old.Create),
			Device:         old.Device,
			Format:         old.Format,
			Label:          old.Label,
			Options:        translateMountOptionSlice(old.Options),
			UUID:           old.UUID,
			WipeFilesystem: old.WipeFilesystem,
		}
	}
	translateFilesystemSlice := func(old []from.Filesystem) []types.Filesystem {
		var res []types.Filesystem
		for _, x := range old {
			res = append(res, types.Filesystem{
				Mount: translateMount(x.Mount),
				Name:  x.Name,
				Path:  x.Path,
			})
		}
		return res
	}
	translateLinkSlice := func(old []from.Link) []types.Link {
		var res []types.Link
		for _, x := range old {
			res = append(res, types.Link{
				Node: translateNode(x.Node),
				LinkEmbedded1: types.LinkEmbedded1{
					Hard:   x.Hard,
					Target: x.Target,
				},
			})
		}
		return res
	}
	translateDeviceSlice := func(old []from.Device) []types.Device {
		var res []types.Device
		for _, x := range old {
			res = append(res, types.Device(x))
		}
		return res
	}
	translateRaidOptionSlice := func(old []from.RaidOption) []types.RaidOption {
		var res []types.RaidOption
		for _, x := range old {
			res = append(res, types.RaidOption(x))
		}
		return res
	}
	translateRaidSlice := func(old []from.Raid) []types.Raid {
		var res []types.Raid
		for _, x := range old {
			res = append(res, types.Raid{
				Devices: translateDeviceSlice(x.Devices),
				Level:   x.Level,
				Name:    x.Name,
				Spares:  x.Spares,
				Options: translateRaidOptionSlice(x.Options),
			})
		}
		return res
	}
	translateSystemdDropinSlice := func(old []from.SystemdDropin) []types.SystemdDropin {
		var res []types.SystemdDropin
		for _, x := range old {
			res = append(res, types.SystemdDropin{
				Contents: x.Contents,
				Name:     x.Name,
			})
		}
		return res
	}
	translateSystemdUnitSlice := func(old []from.Unit) []types.Unit {
		var res []types.Unit
		for _, x := range old {
			res = append(res, types.Unit{
				Contents: x.Contents,
				Dropins:  translateSystemdDropinSlice(x.Dropins),
				Enable:   x.Enable,
				Enabled:  x.Enabled,
				Mask:     x.Mask,
				Name:     x.Name,
			})
		}
		return res
	}
	config := types.Config{
		Ignition: types.Ignition{
			Version: from.MaxVersion.String(),
			Timeouts: types.Timeouts{
				HTTPResponseHeaders: old.Ignition.Timeouts.HTTPResponseHeaders,
				HTTPTotal:           old.Ignition.Timeouts.HTTPTotal,
			},
			Config: types.IgnitionConfig{
				Replace: translateConfigReference(old.Ignition.Config.Replace),
				Append:  translateConfigReferenceSlice(old.Ignition.Config.Append),
			},
			Security: types.Security{
				TLS: types.TLS{
					CertificateAuthorities: translateCertificateAuthoritySlice(old.Ignition.Security.TLS.CertificateAuthorities),
				},
			},
		},
		Networkd: types.Networkd{
			Units: translateNetworkdUnitSlice(old.Networkd.Units),
		},
		Passwd: types.Passwd{
			Groups: translatePasswdGroupSlice(old.Passwd.Groups),
			Users:  translatePasswdUserSlice(old.Passwd.Users),
		},
		Storage: types.Storage{
			Directories: translateDirectorySlice(old.Storage.Directories),
			Disks:       translateDiskSlice(old.Storage.Disks),
			Files:       translateFileSlice(old.Storage.Files),
			Filesystems: translateFilesystemSlice(old.Storage.Filesystems),
			Links:       translateLinkSlice(old.Storage.Links),
			Raid:        translateRaidSlice(old.Storage.Raid),
		},
		Systemd: types.Systemd{
			Units: translateSystemdUnitSlice(old.Systemd.Units),
		},
	}
	return config
}
