// Package libvirt contains wrappers for converting schema
// ids to libvirt RFC4122 UUID type
package libvirt

import (
	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/google/uuid"
)

func parseUUID(uuidStr string) libvirt.UUID {
	return libvirt.UUID(uuid.MustParse(uuidStr))
}

func uuidString(lvUUID libvirt.UUID) string {
	return uuid.UUID(lvUUID).String()
}

func int2bool(b int) bool {
	return b == 1
}

func bool2int(b bool) int32 {
	var i int32
	if b {
		i = 1
	}
	return i
}
