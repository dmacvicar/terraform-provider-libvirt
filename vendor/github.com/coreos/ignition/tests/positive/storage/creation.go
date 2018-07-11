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
	register.Register(register.PositiveTest, ForceNewFilesystemOfSameType())
	register.Register(register.PositiveTest, WipeFilesystemWithSameType())
	register.Register(register.PositiveTest, CreateNewPartitions())
	register.Register(register.PositiveTest, AppendPartition())
	register.Register(register.PositiveTest, PartitionSizeStart0())
}

func ForceNewFilesystemOfSameType() types.Test {
	name := "Force new Filesystem Creation of same type"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
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
					"format": "ext4",
					"create": {
						"force": true
					}}
				 }]
			}
	}`

	in[0].Partitions.GetPartition("EFI-SYSTEM").FilesystemType = "ext4"
	out[0].Partitions.GetPartition("EFI-SYSTEM").FilesystemType = "ext4"
	out[0].Partitions.GetPartition("EFI-SYSTEM").Files = []types.File{}
	out[0].Partitions.AddRemovedNodes("EFI-SYSTEM", []types.Node{
		{
			Name:      "multiLine",
			Directory: "path/example",
		}, {
			Name:      "singleLine",
			Directory: "another/path/example",
		}, {
			Name:      "emptyFile",
			Directory: "empty",
		}, {
			Name:      "noPath",
			Directory: "",
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

func WipeFilesystemWithSameType() types.Test {
	name := "Wipe Filesystem with Filesystem of same type"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	mntDevices := []types.MntDevice{
		{
			Label:        "EFI-SYSTEM",
			Substitution: "$DEVICE",
		},
	}
	config := `{
		"ignition": { "version": "2.1.0" },
		"storage": {
			"filesystems": [{
				"mount": {
					"device": "$DEVICE",
					"format": "ext4",
					"wipeFilesystem": true
				}}]
			}
	}`

	in[0].Partitions.GetPartition("EFI-SYSTEM").FilesystemType = "ext4"
	out[0].Partitions.GetPartition("EFI-SYSTEM").FilesystemType = "ext4"
	out[0].Partitions.GetPartition("EFI-SYSTEM").Files = []types.File{}
	out[0].Partitions.AddRemovedNodes("EFI-SYSTEM", []types.Node{
		{
			Name:      "multiLine",
			Directory: "path/example",
		}, {
			Name:      "singleLine",
			Directory: "another/path/example",
		}, {
			Name:      "emptyFile",
			Directory: "empty",
		}, {
			Name:      "noPath",
			Directory: "",
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

func CreateNewPartitions() types.Test {
	name := "Create new partitions"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	config := `{
		"ignition": {"version": "2.1.0"},
		"storage": {
		    "disks": [
			    {
					"device": "$disk1",
					"wipeTable": true,
					"partitions": [
						{
							"label": "important-data",
							"number": 1,
							"size": 65536,
							"typeGuid": "B921B045-1DF0-41C3-AF44-4C6F280D3FAE",
							"guid": "8A7A6E26-5E8F-4CCA-A654-46215D4696AC"
						},
						{
							"label": "ephemeral-data",
							"number": 2,
							"size": 131072,
							"typeGuid": "CA7D7CCB-63ED-4C53-861C-1742536059CC",
							"guid": "A51034E6-26B3-48DF-BEED-220562AC7AD1"
						}
					]
				}
			]
		}
	}`
	// Create dummy partitions. The UUIDs in the input partitions
	// are intentionally different so if Ignition doesn't do the right thing the
	// validation will fail.
	in = append(in, types.Disk{
		Alignment: types.IgnitionAlignment,
		Partitions: types.Partitions{
			{
				Label:    "important-data",
				Number:   1,
				Length:   65536,
				TypeGUID: "B921B045-1DF0-41C3-AF44-4C6F280D3FAE",
				GUID:     "B921B045-1DF0-41C3-AF44-4C6F280D3FAE",
			},
			{
				Label:    "ephemeral-data",
				Number:   2,
				Length:   131072,
				TypeGUID: "CA7D7CCB-63ED-4C53-861C-1742536059CC",
				GUID:     "B921B045-1DF0-41C3-AF44-4C6F280D3FAE",
			},
		},
	})
	out = append(out, types.Disk{
		Alignment: types.IgnitionAlignment,
		Partitions: types.Partitions{
			{
				Label:    "important-data",
				Number:   1,
				Length:   65536,
				TypeGUID: "B921B045-1DF0-41C3-AF44-4C6F280D3FAE",
				GUID:     "8A7A6E26-5E8F-4CCA-A654-46215D4696AC",
			},
			{
				Label:    "ephemeral-data",
				Number:   2,
				Length:   131072,
				TypeGUID: "CA7D7CCB-63ED-4C53-861C-1742536059CC",
				GUID:     "A51034E6-26B3-48DF-BEED-220562AC7AD1",
			},
		},
	})

	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}

func AppendPartition() types.Test {
	name := "Append partition to an existing partition table"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	config := `{
		"ignition": {
			"version": "2.1.0"
		},
		"storage": {
			"disks": [{
				"device": "$disk1",
				"wipeTable": false,
				"partitions": [{
					"label": "additional-partition",
					"number": 3,
					"size": 65536,
					"typeGuid": "F39C522B-9966-4429-A8F8-417CD5D83E5E",
					"guid": "3ED3993F-0016-422B-B134-09FCBA6F66EF"
				}]
			}]
		}
	}`

	in = append(in, types.Disk{
		Alignment: types.IgnitionAlignment,
		Partitions: types.Partitions{
			{
				Label:    "important-data",
				Number:   1,
				Length:   65536,
				TypeGUID: "B921B045-1DF0-41C3-AF44-4C6F280D3FAE",
				GUID:     "8A7A6E26-5E8F-4CCA-A654-46215D4696AC",
			},
			{
				Label:    "ephemeral-data",
				Number:   2,
				Length:   131072,
				TypeGUID: "CA7D7CCB-63ED-4C53-861C-1742536059CC",
				GUID:     "B921B045-1DF0-41C3-AF44-4C6F280D3FAE",
			},
		},
	})
	out = append(out, types.Disk{
		Alignment: types.IgnitionAlignment,
		Partitions: types.Partitions{
			{
				Label:    "important-data",
				Number:   1,
				Length:   65536,
				TypeGUID: "B921B045-1DF0-41C3-AF44-4C6F280D3FAE",
				GUID:     "8A7A6E26-5E8F-4CCA-A654-46215D4696AC",
			},
			{
				Label:    "ephemeral-data",
				Number:   2,
				Length:   131072,
				TypeGUID: "CA7D7CCB-63ED-4C53-861C-1742536059CC",
				GUID:     "B921B045-1DF0-41C3-AF44-4C6F280D3FAE",
			},
			{
				Label:    "additional-partition",
				Number:   3,
				Length:   65536,
				TypeGUID: "F39C522B-9966-4429-A8F8-417CD5D83E5E",
				GUID:     "3ED3993F-0016-422B-B134-09FCBA6F66EF",
			},
		},
	})

	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}

func PartitionSizeStart0() types.Test {
	name := "Create a partition with size and start 0"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	config := `{
		"ignition": {
			"version": "2.1.0"
		},
		"storage": {
			"disks": [{
				"device": "$disk1",
				"wipeTable": false,
				"partitions": [{
					"label": "fills-disk",
					"number": 1,
					"start": 0,
					"size": 0,
					"typeGuid": "F39C522B-9966-4429-A8F8-417CD5D83E5E",
					"guid": "3ED3993F-0016-422B-B134-09FCBA6F66EF"
				}]
			}]
		}
	}`

	in = append(in, types.Disk{
		Alignment: types.IgnitionAlignment,
	})
	out = append(out, types.Disk{
		Alignment: types.IgnitionAlignment,
		Partitions: types.Partitions{
			{
				Label:    "fills-disk",
				Number:   1,
				Length:   65536,
				TypeGUID: "F39C522B-9966-4429-A8F8-417CD5D83E5E",
				GUID:     "3ED3993F-0016-422B-B134-09FCBA6F66EF",
			},
		},
	})

	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}
