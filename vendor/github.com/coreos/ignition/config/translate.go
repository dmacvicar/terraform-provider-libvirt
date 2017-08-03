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
	"fmt"
	"net/url"
	"strings"

	"github.com/coreos/ignition/config/types"
	v1 "github.com/coreos/ignition/config/v1/types"
	v2_0 "github.com/coreos/ignition/config/v2_0/types"
	v2_1 "github.com/coreos/ignition/config/v2_1/types"

	"github.com/vincent-petithory/dataurl"
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

func TranslateFromV1(old v1.Config) types.Config {
	config := types.Config{
		Ignition: types.Ignition{
			Version: v2_0.MaxVersion.String(),
		},
	}

	for _, oldDisk := range old.Storage.Disks {
		disk := types.Disk{
			Device:    string(oldDisk.Device),
			WipeTable: oldDisk.WipeTable,
		}

		for _, oldPartition := range oldDisk.Partitions {
			disk.Partitions = append(disk.Partitions, types.Partition{
				Label:    string(oldPartition.Label),
				Number:   oldPartition.Number,
				Size:     int(oldPartition.Size),
				Start:    int(oldPartition.Start),
				TypeGUID: string(oldPartition.TypeGUID),
			})
		}

		config.Storage.Disks = append(config.Storage.Disks, disk)
	}

	for _, oldArray := range old.Storage.Arrays {
		array := types.Raid{
			Name:   oldArray.Name,
			Level:  oldArray.Level,
			Spares: oldArray.Spares,
		}

		for _, oldDevice := range oldArray.Devices {
			array.Devices = append(array.Devices, types.Device(oldDevice))
		}

		config.Storage.Raid = append(config.Storage.Raid, array)
	}

	for i, oldFilesystem := range old.Storage.Filesystems {
		filesystem := types.Filesystem{
			Name: fmt.Sprintf("_translate-filesystem-%d", i),
			Mount: &types.Mount{
				Device: string(oldFilesystem.Device),
				Format: string(oldFilesystem.Format),
			},
		}

		if oldFilesystem.Create != nil {
			filesystem.Mount.Create = &types.Create{
				Force:   oldFilesystem.Create.Force,
				Options: translateV1MkfsOptionsToV2_2OptionSlice(oldFilesystem.Create.Options),
			}
		}

		config.Storage.Filesystems = append(config.Storage.Filesystems, filesystem)

		for _, oldFile := range oldFilesystem.Files {
			file := types.File{
				Node: types.Node{
					Filesystem: filesystem.Name,
					Path:       string(oldFile.Path),
					User:       types.NodeUser{ID: intToPtr(oldFile.Uid)},
					Group:      types.NodeGroup{ID: intToPtr(oldFile.Gid)},
				},
				FileEmbedded1: types.FileEmbedded1{
					Mode: int(oldFile.Mode),
					Contents: types.FileContents{
						Source: (&url.URL{
							Scheme: "data",
							Opaque: "," + dataurl.EscapeString(oldFile.Contents),
						}).String(),
					},
				},
			}

			config.Storage.Files = append(config.Storage.Files, file)
		}
	}

	for _, oldUnit := range old.Systemd.Units {
		unit := types.Unit{
			Name:     string(oldUnit.Name),
			Enable:   oldUnit.Enable,
			Mask:     oldUnit.Mask,
			Contents: oldUnit.Contents,
		}

		for _, oldDropIn := range oldUnit.DropIns {
			unit.Dropins = append(unit.Dropins, types.Dropin{
				Name:     string(oldDropIn.Name),
				Contents: oldDropIn.Contents,
			})
		}

		config.Systemd.Units = append(config.Systemd.Units, unit)
	}

	for _, oldUnit := range old.Networkd.Units {
		config.Networkd.Units = append(config.Networkd.Units, types.Networkdunit{
			Name:     string(oldUnit.Name),
			Contents: oldUnit.Contents,
		})
	}

	for _, oldUser := range old.Passwd.Users {
		user := types.PasswdUser{
			Name:              oldUser.Name,
			PasswordHash:      strToPtr(oldUser.PasswordHash),
			SSHAuthorizedKeys: translateStringSliceToV2_2SSHAuthorizedKeySlice(oldUser.SSHAuthorizedKeys),
		}

		if oldUser.Create != nil {
			var uid *int
			if oldUser.Create.Uid != nil {
				tmp := int(*oldUser.Create.Uid)
				uid = &tmp
			}

			user.Create = &types.Usercreate{
				UID:          uid,
				Gecos:        oldUser.Create.GECOS,
				HomeDir:      oldUser.Create.Homedir,
				NoCreateHome: oldUser.Create.NoCreateHome,
				PrimaryGroup: oldUser.Create.PrimaryGroup,
				Groups:       translateStringSliceToV2_2UsercreateGroupSlice(oldUser.Create.Groups),
				NoUserGroup:  oldUser.Create.NoUserGroup,
				System:       oldUser.Create.System,
				NoLogInit:    oldUser.Create.NoLogInit,
				Shell:        oldUser.Create.Shell,
			}
		}

		config.Passwd.Users = append(config.Passwd.Users, user)
	}

	for _, oldGroup := range old.Passwd.Groups {
		var gid *int
		if oldGroup.Gid != nil {
			tmp := int(*oldGroup.Gid)
			gid = &tmp
		}
		config.Passwd.Groups = append(config.Passwd.Groups, types.PasswdGroup{
			Name:         oldGroup.Name,
			Gid:          gid,
			PasswordHash: oldGroup.PasswordHash,
			System:       oldGroup.System,
		})
	}

	return config
}

// golang--
func translateV1MkfsOptionsToV2_2OptionSlice(opts v1.MkfsOptions) []types.CreateOption {
	newOpts := make([]types.CreateOption, len(opts))
	for i, o := range opts {
		newOpts[i] = types.CreateOption(o)
	}
	return newOpts
}

// golang--
func translateStringSliceToV2_2SSHAuthorizedKeySlice(keys []string) []types.SSHAuthorizedKey {
	newKeys := make([]types.SSHAuthorizedKey, len(keys))
	for i, k := range keys {
		newKeys[i] = types.SSHAuthorizedKey(k)
	}
	return newKeys
}

// golang--
func translateStringSliceToV2_2UsercreateGroupSlice(groups []string) []types.UsercreateGroup {
	var newGroups []types.UsercreateGroup
	for _, g := range groups {
		newGroups = append(newGroups, types.UsercreateGroup(g))
	}
	return newGroups
}

func TranslateFromV2_0(old v2_0.Config) types.Config {
	translateVerification := func(old v2_0.Verification) types.Verification {
		var ver types.Verification
		if old.Hash != nil {
			// .String() here is a wrapper around MarshalJSON, which will put the hash in quotes
			h := strings.Trim(old.Hash.String(), "\"")
			ver.Hash = &h
		}
		return ver
	}
	translateConfigReference := func(old v2_0.ConfigReference) types.ConfigReference {
		return types.ConfigReference{
			Source:       old.Source.String(),
			Verification: translateVerification(old.Verification),
		}
	}

	config := types.Config{
		Ignition: types.Ignition{
			Version: types.MaxVersion.String(),
		},
	}

	if old.Ignition.Config.Replace != nil {
		ref := translateConfigReference(*old.Ignition.Config.Replace)
		config.Ignition.Config.Replace = &ref
	}

	for _, oldAppend := range old.Ignition.Config.Append {
		config.Ignition.Config.Append =
			append(config.Ignition.Config.Append, translateConfigReference(oldAppend))
	}

	for _, oldDisk := range old.Storage.Disks {
		disk := types.Disk{
			Device:    string(oldDisk.Device),
			WipeTable: oldDisk.WipeTable,
		}

		for _, oldPartition := range oldDisk.Partitions {
			disk.Partitions = append(disk.Partitions, types.Partition{
				Label:    string(oldPartition.Label),
				Number:   oldPartition.Number,
				Size:     int(oldPartition.Size),
				Start:    int(oldPartition.Start),
				TypeGUID: string(oldPartition.TypeGUID),
			})
		}

		config.Storage.Disks = append(config.Storage.Disks, disk)
	}

	for _, oldArray := range old.Storage.Arrays {
		array := types.Raid{
			Name:   oldArray.Name,
			Level:  oldArray.Level,
			Spares: oldArray.Spares,
		}

		for _, oldDevice := range oldArray.Devices {
			array.Devices = append(array.Devices, types.Device(oldDevice))
		}

		config.Storage.Raid = append(config.Storage.Raid, array)
	}

	for _, oldFilesystem := range old.Storage.Filesystems {
		filesystem := types.Filesystem{
			Name: oldFilesystem.Name,
		}

		if oldFilesystem.Mount != nil {
			filesystem.Mount = &types.Mount{
				Device: string(oldFilesystem.Mount.Device),
				Format: string(oldFilesystem.Mount.Format),
			}

			if oldFilesystem.Mount.Create != nil {
				filesystem.Mount.Create = &types.Create{
					Force:   oldFilesystem.Mount.Create.Force,
					Options: translateV2_0MkfsOptionsToV2_2OptionSlice(oldFilesystem.Mount.Create.Options),
				}
			}
		}

		if oldFilesystem.Path != nil {
			p := string(*oldFilesystem.Path)
			filesystem.Path = &p
		}

		config.Storage.Filesystems = append(config.Storage.Filesystems, filesystem)
	}

	for _, oldFile := range old.Storage.Files {
		file := types.File{
			Node: types.Node{
				Filesystem: oldFile.Filesystem,
				Path:       string(oldFile.Path),
				User:       types.NodeUser{ID: intToPtr(oldFile.User.Id)},
				Group:      types.NodeGroup{ID: intToPtr(oldFile.Group.Id)},
			},
			FileEmbedded1: types.FileEmbedded1{
				Mode: int(oldFile.Mode),
				Contents: types.FileContents{
					Compression:  string(oldFile.Contents.Compression),
					Source:       oldFile.Contents.Source.String(),
					Verification: translateVerification(oldFile.Contents.Verification),
				},
			},
		}

		config.Storage.Files = append(config.Storage.Files, file)
	}

	for _, oldUnit := range old.Systemd.Units {
		unit := types.Unit{
			Name:     string(oldUnit.Name),
			Enable:   oldUnit.Enable,
			Mask:     oldUnit.Mask,
			Contents: oldUnit.Contents,
		}

		for _, oldDropIn := range oldUnit.DropIns {
			unit.Dropins = append(unit.Dropins, types.Dropin{
				Name:     string(oldDropIn.Name),
				Contents: oldDropIn.Contents,
			})
		}

		config.Systemd.Units = append(config.Systemd.Units, unit)
	}

	for _, oldUnit := range old.Networkd.Units {
		config.Networkd.Units = append(config.Networkd.Units, types.Networkdunit{
			Name:     string(oldUnit.Name),
			Contents: oldUnit.Contents,
		})
	}

	for _, oldUser := range old.Passwd.Users {
		user := types.PasswdUser{
			Name:              oldUser.Name,
			PasswordHash:      strToPtr(oldUser.PasswordHash),
			SSHAuthorizedKeys: translateStringSliceToV2_2SSHAuthorizedKeySlice(oldUser.SSHAuthorizedKeys),
		}

		if oldUser.Create != nil {
			var u *int
			if oldUser.Create.Uid != nil {
				tmp := int(*oldUser.Create.Uid)
				u = &tmp
			}
			user.Create = &types.Usercreate{
				UID:          u,
				Gecos:        oldUser.Create.GECOS,
				HomeDir:      oldUser.Create.Homedir,
				NoCreateHome: oldUser.Create.NoCreateHome,
				PrimaryGroup: oldUser.Create.PrimaryGroup,
				Groups:       translateStringSliceToV2_2UsercreateGroupSlice(oldUser.Create.Groups),
				NoUserGroup:  oldUser.Create.NoUserGroup,
				System:       oldUser.Create.System,
				NoLogInit:    oldUser.Create.NoLogInit,
				Shell:        oldUser.Create.Shell,
			}
		}

		config.Passwd.Users = append(config.Passwd.Users, user)
	}

	for _, oldGroup := range old.Passwd.Groups {
		var g *int
		if oldGroup.Gid != nil {
			tmp := int(*oldGroup.Gid)
			g = &tmp
		}
		config.Passwd.Groups = append(config.Passwd.Groups, types.PasswdGroup{
			Name:         oldGroup.Name,
			Gid:          g,
			PasswordHash: oldGroup.PasswordHash,
			System:       oldGroup.System,
		})
	}

	return config
}

// golang--
func translateV2_0MkfsOptionsToV2_2OptionSlice(opts v2_0.MkfsOptions) []types.CreateOption {
	newOpts := make([]types.CreateOption, len(opts))
	for i, o := range opts {
		newOpts[i] = types.CreateOption(o)
	}
	return newOpts
}

func TranslateFromV2_1(old v2_1.Config) types.Config {
	translateConfigReference := func(old *v2_1.ConfigReference) *types.ConfigReference {
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
	translateConfigReferenceSlice := func(old []v2_1.ConfigReference) []types.ConfigReference {
		var res []types.ConfigReference
		for _, c := range old {
			res = append(res, *translateConfigReference(&c))
		}
		return res
	}
	translateNetworkdUnitSlice := func(old []v2_1.Networkdunit) []types.Networkdunit {
		var res []types.Networkdunit
		for _, u := range old {
			res = append(res, types.Networkdunit{
				Contents: u.Contents,
				Name:     u.Name,
			})
		}
		return res
	}
	translatePasswdGroupSlice := func(old []v2_1.PasswdGroup) []types.PasswdGroup {
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
	translatePasswdUsercreateGroupSlice := func(old []v2_1.UsercreateGroup) []types.UsercreateGroup {
		var res []types.UsercreateGroup
		for _, g := range old {
			res = append(res, types.UsercreateGroup(g))
		}
		return res
	}
	translatePasswdUsercreate := func(old *v2_1.Usercreate) *types.Usercreate {
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
	translatePasswdUserGroupSlice := func(old []v2_1.PasswdUserGroup) []types.PasswdUserGroup {
		var res []types.PasswdUserGroup
		for _, g := range old {
			res = append(res, types.PasswdUserGroup(g))
		}
		return res
	}
	translatePasswdSSHAuthorizedKeySlice := func(old []v2_1.SSHAuthorizedKey) []types.SSHAuthorizedKey {
		res := make([]types.SSHAuthorizedKey, len(old))
		for i, k := range old {
			res[i] = types.SSHAuthorizedKey(k)
		}
		return res
	}
	translatePasswdUserSlice := func(old []v2_1.PasswdUser) []types.PasswdUser {
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
	translateNodeGroup := func(old v2_1.NodeGroup) types.NodeGroup {
		return types.NodeGroup{
			ID:   old.ID,
			Name: old.Name,
		}
	}
	translateNodeUser := func(old v2_1.NodeUser) types.NodeUser {
		return types.NodeUser{
			ID:   old.ID,
			Name: old.Name,
		}
	}
	translateNode := func(old v2_1.Node) types.Node {
		return types.Node{
			Filesystem: old.Filesystem,
			Group:      translateNodeGroup(old.Group),
			Path:       old.Path,
			User:       translateNodeUser(old.User),
		}
	}
	translateDirectorySlice := func(old []v2_1.Directory) []types.Directory {
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
	translatePartitionSlice := func(old []v2_1.Partition) []types.Partition {
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
	translateDiskSlice := func(old []v2_1.Disk) []types.Disk {
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
	translateFileSlice := func(old []v2_1.File) []types.File {
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
					Mode: x.Mode,
				},
			})
		}
		return res
	}
	translateMountCreateOptionSlice := func(old []v2_1.CreateOption) []types.CreateOption {
		var res []types.CreateOption
		for _, x := range old {
			res = append(res, types.CreateOption(x))
		}
		return res
	}
	translateMountCreate := func(old *v2_1.Create) *types.Create {
		if old == nil {
			return nil
		}
		return &types.Create{
			Force:   old.Force,
			Options: translateMountCreateOptionSlice(old.Options),
		}
	}
	translateMountOptionSlice := func(old []v2_1.MountOption) []types.MountOption {
		var res []types.MountOption
		for _, x := range old {
			res = append(res, types.MountOption(x))
		}
		return res
	}
	translateMount := func(old *v2_1.Mount) *types.Mount {
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
	translateFilesystemSlice := func(old []v2_1.Filesystem) []types.Filesystem {
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
	translateLinkSlice := func(old []v2_1.Link) []types.Link {
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
	translateDeviceSlice := func(old []v2_1.Device) []types.Device {
		var res []types.Device
		for _, x := range old {
			res = append(res, types.Device(x))
		}
		return res
	}
	translateRaidSlice := func(old []v2_1.Raid) []types.Raid {
		var res []types.Raid
		for _, x := range old {
			res = append(res, types.Raid{
				Devices: translateDeviceSlice(x.Devices),
				Level:   x.Level,
				Name:    x.Name,
				Spares:  x.Spares,
			})
		}
		return res
	}
	translateSystemdDropinSlice := func(old []v2_1.Dropin) []types.Dropin {
		var res []types.Dropin
		for _, x := range old {
			res = append(res, types.Dropin{
				Contents: x.Contents,
				Name:     x.Name,
			})
		}
		return res
	}
	translateSystemdUnitSlice := func(old []v2_1.Unit) []types.Unit {
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
			Version: types.MaxVersion.String(),
			Timeouts: types.Timeouts{
				HTTPResponseHeaders: old.Ignition.Timeouts.HTTPResponseHeaders,
				HTTPTotal:           old.Ignition.Timeouts.HTTPTotal,
			},
			Config: types.IgnitionConfig{
				Replace: translateConfigReference(old.Ignition.Config.Replace),
				Append:  translateConfigReferenceSlice(old.Ignition.Config.Append),
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
