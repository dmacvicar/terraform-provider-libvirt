package libvirt

import (
	"encoding/xml"
	"fmt"

	libvirt "github.com/digitalocean/go-libvirt"
	"libvirt.org/go/libvirtxml"
)

func newDefVolume() libvirtxml.StorageVolume {
	return libvirtxml.StorageVolume{
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
	}
}

// Creates a volume definition from a XML.
func newDefVolumeFromXML(s string) (libvirtxml.StorageVolume, error) {
	var volumeDef libvirtxml.StorageVolume
	err := xml.Unmarshal([]byte(s), &volumeDef)
	if err != nil {
		return libvirtxml.StorageVolume{}, err
	}
	return volumeDef, nil
}

func newDefVolumeFromLibvirt(virConn *libvirt.Libvirt, volume libvirt.StorageVol) (libvirtxml.StorageVolume, error) {
	volumeDefXML, err := virConn.StorageVolGetXMLDesc(volume, 0)
	if err != nil {
		return libvirtxml.StorageVolume{}, fmt.Errorf("could not get XML description for volume %s: %w", volume.Name, err)
	}
	volumeDef, err := newDefVolumeFromXML(volumeDefXML)
	if err != nil {
		return libvirtxml.StorageVolume{}, fmt.Errorf("could not get a volume definition from XML for %s: %w", volumeDef.Name, err)
	}
	return volumeDef, nil
}

func newDefBackingStoreFromLibvirt(virConn *libvirt.Libvirt, baseVolume libvirt.StorageVol) (libvirtxml.StorageVolumeBackingStore, error) {
	baseVolumeDef, err := newDefVolumeFromLibvirt(virConn, baseVolume)
	if err != nil {
		return libvirtxml.StorageVolumeBackingStore{}, fmt.Errorf("could not get volume: %w", err)
	}
	baseVolPath, err := virConn.StorageVolGetPath(baseVolume)
	if err != nil {
		return libvirtxml.StorageVolumeBackingStore{}, fmt.Errorf("could not get base image path: %w", err)
	}
	var backingStoreVolFormat *libvirtxml.StorageVolumeTargetFormat
	if baseVolumeDef.Target.Format != nil {
		backingStoreVolFormat = &libvirtxml.StorageVolumeTargetFormat{
			Type: baseVolumeDef.Target.Format.Type,
		}
	}
	backingStoreDef := libvirtxml.StorageVolumeBackingStore{
		Path:   baseVolPath,
		Format: backingStoreVolFormat,
	}
	return backingStoreDef, nil
}
