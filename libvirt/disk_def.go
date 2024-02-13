package libvirt

import (
	"fmt"
	"math/rand"

	"libvirt.org/go/libvirtxml"
)

const oui = "05abcd"

// note, source is not initialized.
func newDefDisk(i int) libvirtxml.DomainDisk {
	return libvirtxml.DomainDisk{
		Device: "disk",
		Target: &libvirtxml.DomainDiskTarget{
			Bus: "virtio",
			Dev: fmt.Sprintf("vd%s", diskLetterForIndex(i)),
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
		//nolint:gosec // math.rand is enough for this
		result[i] = chars[rand.Intn(len(chars))]
	}
	return oui + string(result)
}
