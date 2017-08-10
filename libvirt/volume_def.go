package libvirt

import (
	"encoding/xml"
	"fmt"

	libvirt "github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
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

// Creates a volume definition from a XML
func newDefVolumeFromXML(s string) (libvirtxml.StorageVolume, error) {
	var volumeDef libvirtxml.StorageVolume
	err := xml.Unmarshal([]byte(s), &volumeDef)
	if err != nil {
		return libvirtxml.StorageVolume{}, err
	}
	return volumeDef, nil
}

func newDefVolumeFromLibvirt(volume *libvirt.StorageVol) (libvirtxml.StorageVolume, error) {
	name, err := volume.GetName()
	if err != nil {
		return libvirtxml.StorageVolume{}, fmt.Errorf("could not get name for volume: %s.", err)
	}
	volumeDefXml, err := volume.GetXMLDesc(0)
	if err != nil {
		return libvirtxml.StorageVolume{}, fmt.Errorf("could not get XML description for volume %s: %s.", name, err)
	}
	volumeDef, err := newDefVolumeFromXML(volumeDefXml)
	if err != nil {
		return libvirtxml.StorageVolume{}, fmt.Errorf("could not get a volume definition from XML for %s: %s.", volumeDef.Name, err)
	}
	return volumeDef, nil
}

func newDefBackingStoreFromLibvirt(baseVolume *libvirt.StorageVol) (libvirtxml.StorageVolumeBackingStore, error) {
	baseVolumeDef, err := newDefVolumeFromLibvirt(baseVolume)
	if err != nil {
		return libvirtxml.StorageVolumeBackingStore{}, fmt.Errorf("could not get volume: %s.", err)
	}
	baseVolPath, err := baseVolume.GetPath()
	if err != nil {
		return libvirtxml.StorageVolumeBackingStore{}, fmt.Errorf("could not get base image path: %s", err)
	}
	backingStoreDef := libvirtxml.StorageVolumeBackingStore{
		Path: baseVolPath,
		Format: &libvirtxml.StorageVolumeTargetFormat{
			Type: baseVolumeDef.Target.Format.Type,
		},
	}
	return backingStoreDef, nil
}
