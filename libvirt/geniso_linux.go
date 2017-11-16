// +build linux

package libvirt

import (
	"os/exec"
	"path/filepath"
)

// GenIsoCmd generate cloud-init iso command for linux systems.
func GenIsoCmd(isoDest string, tmpDir string, userdata string, metadata string) *exec.Cmd {
	return exec.Command(
		"genisoimage",
		"-output",
		isoDest,
		"-volid",
		"cidata",
		"-joliet",
		"-rock",
		filepath.Join(tmpDir, userdata),
		filepath.Join(tmpDir, metadata))
}
