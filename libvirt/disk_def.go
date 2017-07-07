package libvirt

import (
	"math/rand"

	"github.com/libvirt/libvirt-go-xml"
)

const OUI = "05abcd"

func newDefDisk() libvirtxml.DomainDisk {
	return libvirtxml.DomainDisk{
		Type:   "file",
		Device: "disk",
		Target: &libvirtxml.DomainDiskTarget{
			Bus: "virtio",
		},
		Driver: &libvirtxml.DomainDiskDriver{
			Name: "qemu",
			Type: "qcow2",
		},
	}
}

func newCDROM() libvirtxml.DomainDisk {
	return libvirtxml.DomainDisk{
		Type:   "file",
		Device: "cdrom",
		Target: &libvirtxml.DomainDiskTarget{
			Dev: "hda",
			Bus: "ide",
		},
		Driver: &libvirtxml.DomainDiskDriver{
			Name: "qemu",
			Type: "raw",
		},
	}
}

func randomWWN(strlen int) string {
	const chars = "abcdef0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return OUI + string(result)
}
