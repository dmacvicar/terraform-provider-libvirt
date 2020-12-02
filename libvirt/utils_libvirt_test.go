package libvirt

import (
	"testing"
)

func TestParseUUID(t *testing.T) {
	const uuidStr string = "19fdc2f2-fa64-46f3-bacf-42a8aafca6dd"
	uuid := parseUUID(uuidStr)
	wantUUID := [16]byte{
		0x19, 0xfd, 0xc2, 0xf2, 0xfa, 0x64, 0x46, 0xf3,
		0xba, 0xcf, 0x42, 0xa8, 0xaa, 0xfc, 0xa6, 0xdd,
	}
	if uuid != wantUUID {
		t.Errorf("expected UUID %q, got %q", wantUUID, uuid)
	}

	uuidToStr := uuidString(uuid)
	if uuidToStr != uuidStr {
		t.Errorf("expected UUID %q, got %q", uuidStr, uuidToStr)
	}
}
