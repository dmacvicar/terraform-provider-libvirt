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
	register.Register(register.PositiveTest, ReformatToBTRFS_2_0_0())
	register.Register(register.PositiveTest, ReformatToXFS_2_0_0())
	register.Register(register.PositiveTest, ReformatToEXT4_2_0_0())
	register.Register(register.PositiveTest, ReformatToBTRFS_2_1_0())
	register.Register(register.PositiveTest, ReformatToXFS_2_1_0())
	register.Register(register.PositiveTest, ReformatToVFAT_2_1_0())
	register.Register(register.PositiveTest, ReformatToEXT4_2_1_0())
	register.Register(register.PositiveTest, ReformatToSWAP_2_1_0())
}

func ReformatToBTRFS_2_0_0() types.Test {
	name := "Reformat a Filesystem to Btrfs"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	mntDevices := []types.MntDevice{
		{
			Label:        "OEM",
			Substitution: "$DEVICE",
		},
	}
	config := `{
	  "ignition": { "version": "2.0.0" },
	  "storage": {
	    "filesystems": [{
	      "mount": {
	        "device": "$DEVICE",
	        "format": "btrfs",
	        "create": {
	          "force": true,
	          "options": [ "--label=OEM", "--uuid=CA7D7CCB-63ED-4C53-861C-1742536059CC" ]
	        }
	      }
	    }]
	  }
	}`
	out[0].Partitions.GetPartition("OEM").FilesystemType = "btrfs"

	return types.Test{
		Name:       name,
		In:         in,
		Out:        out,
		MntDevices: mntDevices,
		Config:     config,
	}
}

func ReformatToXFS_2_0_0() types.Test {
	name := "Reformat a Filesystem to XFS"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	mntDevices := []types.MntDevice{
		{
			Label:        "OEM",
			Substitution: "$DEVICE",
		},
	}
	config := `{
	  "ignition": { "version": "2.0.0" },
	  "storage": {
	    "filesystems": [{
	      "mount": {
	        "device": "$DEVICE",
	        "format": "xfs",
	        "create": {
	          "force": true,
	          "options": [ "-L", "OEM", "-m", "uuid=CA7D7CCB-63ED-4C53-861C-1742536059CC" ]
	        }
	      }
	    }]
	  }
	}`
	out[0].Partitions.GetPartition("OEM").FilesystemType = "xfs"

	return types.Test{
		Name:       name,
		In:         in,
		Out:        out,
		MntDevices: mntDevices,
		Config:     config,
	}
}

func ReformatToVFAT_2_0_0() types.Test {
	name := "Reformat a Filesystem to VFAT"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	mntDevices := []types.MntDevice{
		{
			Label:        "OEM",
			Substitution: "$DEVICE",
		},
	}
	config := `{
	  "ignition": { "version": "2.0.0" },
	  "storage": {
	    "filesystems": [{
	      "mount": {
	        "device": "$DEVICE",
	        "format": "vfat",
	        "create": {
	          "force": true,
	          "options": [ "-n", "OEM", "-i", "CA7D7CCB-63ED-4C53-861C-1742536059CC" ]
	        }
	      }
	    }]
	  }
	}`
	out[0].Partitions.GetPartition("OEM").FilesystemType = "vfat"
	out[0].Partitions.GetPartition("OEM").FilesystemUUID = "CA7D7CCB-63ED-4C53-861C-1742536059CC"

	return types.Test{
		Name:       name,
		In:         in,
		Out:        out,
		MntDevices: mntDevices,
		Config:     config,
	}
}

func ReformatToEXT4_2_0_0() types.Test {
	name := "Reformat a Filesystem to EXT4"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	mntDevices := []types.MntDevice{
		{
			Label:        "OEM",
			Substitution: "$DEVICE",
		},
	}
	config := `{
	  "ignition": { "version": "2.0.0" },
	  "storage": {
	    "filesystems": [{
	      "mount": {
	        "device": "$DEVICE",
	        "format": "ext4",
	        "create": {
	          "force": true,
	          "options": [ "-L", "OEM", "-U", "CA7D7CCB-63ED-4C53-861C-1742536059CC" ]
	        }
	      }
	    }]
	  }
	}`
	in[0].Partitions.GetPartition("OEM").FilesystemType = "ext2"
	out[0].Partitions.GetPartition("OEM").FilesystemType = "ext4"
	out[0].Partitions.GetPartition("OEM").FilesystemUUID = "CA7D7CCB-63ED-4C53-861C-1742536059CC"

	return types.Test{
		Name:       name,
		In:         in,
		Out:        out,
		MntDevices: mntDevices,
		Config:     config,
	}
}

func ReformatToBTRFS_2_1_0() types.Test {
	name := "Reformat a Filesystem to Btrfs"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	mntDevices := []types.MntDevice{
		{
			Label:        "OEM",
			Substitution: "$DEVICE",
		},
	}
	config := `{
	  "ignition": { "version": "2.1.0" },
	  "storage": {
	    "filesystems": [{
	      "mount": {
	        "device": "$DEVICE",
	        "format": "btrfs",
	        "label": "OEM",
		"uuid": "CA7D7CCB-63ED-4C53-861C-1742536059CC",
		"wipeFilesystem": true
	      }
	    }]
	  }
	}`
	out[0].Partitions.GetPartition("OEM").FilesystemType = "btrfs"
	out[0].Partitions.GetPartition("OEM").FilesystemUUID = "CA7D7CCB-63ED-4C53-861C-1742536059CC"

	return types.Test{
		Name:       name,
		In:         in,
		Out:        out,
		MntDevices: mntDevices,
		Config:     config,
	}
}

func ReformatToXFS_2_1_0() types.Test {
	name := "Reformat a Filesystem to XFS"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	mntDevices := []types.MntDevice{
		{
			Label:        "OEM",
			Substitution: "$DEVICE",
		},
	}
	config := `{
	  "ignition": { "version": "2.1.0" },
	  "storage": {
	    "filesystems": [{
	      "mount": {
	        "device": "$DEVICE",
	        "format": "xfs",
	        "label": "OEM",
		"uuid": "CA7D7CCB-63ED-4C53-861C-1742536059CC",
		"wipeFilesystem": true
	      }
	    }]
	  }
	}`
	out[0].Partitions.GetPartition("OEM").FilesystemType = "xfs"
	out[0].Partitions.GetPartition("OEM").FilesystemUUID = "CA7D7CCB-63ED-4C53-861C-1742536059CC"

	return types.Test{
		Name:       name,
		In:         in,
		Out:        out,
		MntDevices: mntDevices,
		Config:     config,
	}
}

func ReformatToVFAT_2_1_0() types.Test {
	name := "Reformat a Filesystem to VFAT"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	mntDevices := []types.MntDevice{
		{
			Label:        "OEM",
			Substitution: "$DEVICE",
		},
	}
	config := `{
	  "ignition": { "version": "2.1.0" },
	  "storage": {
	    "filesystems": [{
	      "mount": {
	        "device": "$DEVICE",
	        "format": "vfat",
	        "label": "OEM",
		"uuid": "2e24ec82",
		"wipeFilesystem": true
	      }
	    }]
	  }
	}`
	out[0].Partitions.GetPartition("OEM").FilesystemType = "vfat"
	out[0].Partitions.GetPartition("OEM").FilesystemUUID = "2e24ec82"

	return types.Test{
		Name:       name,
		In:         in,
		Out:        out,
		MntDevices: mntDevices,
		Config:     config,
	}
}

func ReformatToEXT4_2_1_0() types.Test {
	name := "Reformat a Filesystem to EXT4"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	mntDevices := []types.MntDevice{
		{
			Label:        "OEM",
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
	        "label": "OEM",
		"uuid": "CA7D7CCB-63ED-4C53-861C-1742536059CC",
		"wipeFilesystem": true
	      }
	    }]
	  }
	}`
	in[0].Partitions.GetPartition("OEM").FilesystemType = "ext2"
	out[0].Partitions.GetPartition("OEM").FilesystemType = "ext4"
	out[0].Partitions.GetPartition("OEM").FilesystemUUID = "CA7D7CCB-63ED-4C53-861C-1742536059CC"

	return types.Test{
		Name:       name,
		In:         in,
		Out:        out,
		MntDevices: mntDevices,
		Config:     config,
	}
}

func ReformatToSWAP_2_1_0() types.Test {
	name := "Reformat a Filesystem to SWAP"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	mntDevices := []types.MntDevice{
		{
			Label:        "OEM",
			Substitution: "$DEVICE",
		},
	}
	config := `{
	  "ignition": { "version": "2.1.0" },
	  "storage": {
	    "filesystems": [{
	      "mount": {
	        "device": "$DEVICE",
	        "format": "swap",
	        "label": "OEM",
	        "uuid": "CA7D7CCB-63ED-4C53-861C-1742536059CC",
		"wipeFilesystem": true
	      }
	    }]
	  }
	}`
	in[0].Partitions.GetPartition("OEM").FilesystemType = "ext2"
	out[0].Partitions.GetPartition("OEM").FilesystemType = "swap"
	out[0].Partitions.GetPartition("OEM").FilesystemUUID = "CA7D7CCB-63ED-4C53-861C-1742536059CC"

	return types.Test{
		Name:       name,
		In:         in,
		Out:        out,
		MntDevices: mntDevices,
		Config:     config,
	}
}
