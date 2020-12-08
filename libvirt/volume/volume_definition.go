package volume

import libvirtxml "github.com/libvirt/libvirt-go-xml"

type VolumeDefinition libvirtxml.StorageVolume

func DefaultDefinition() VolumeDefinition {
	return VolumeDefinition(libvirtxml.StorageVolume{
		Target: &libvirtxml.StorageVolumeTarget{
			Format: &libvirtxml.StorageVolumeTargetFormat{
				Type: "qcow2",
			},
			Permissions: &libvirtxml.StorageVolumeTargetPermissions{
				Mode: "644",
			},
		},
		Capacity: &libvirtxml.StorageVolumeSize{
			Unit:  "bytes",
			Value: 1,
		},
	})
}
