package disk

import (
	"testing"
)

func TestVolumeUnmarshal(t *testing.T) {
	diskDef := DefaultDefinition()
	if diskDef.Target.Bus != "virtio" {
		t.Errorf("Expected virtio bus for default disk definition")
	}
}

