package disk

import (
	"fmt"
	"github.com/libvirt/libvirt-go-xml"
)

type DiskDefinition libvirtxml.DomainDisk

func DefaultDefinition() DiskDefinition {
	return DiskDefinition(libvirtxml.DomainDisk{
		Device: "disk",
		Target: &libvirtxml.DomainDiskTarget{
			Bus: "virtio",
			Dev: fmt.Sprintf("vd%s", diskLetterForIndex(0)),
		},
		Driver: &libvirtxml.DomainDiskDriver{
			Name: "qemu",
			Type: "raw",
		},
	})
}
