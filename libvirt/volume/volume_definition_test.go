package volume

import "testing"

func TestVolumeUnmarshal(t *testing.T) {
	volumeDef := DefaultDefinition()
	if volumeDef.Target.Format.Type != "qcow2" {
		t.Errorf("Expected qcow2 volume format for default volume definition")
	}
}
