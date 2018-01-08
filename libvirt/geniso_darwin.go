// +build darwin

package libvirt

import (
	"os/exec"
	"path/filepath"
)

// GenIsoCmd generate darwin/macosx cloud-init image.
// returns a cmd ready to be executed
func GenIsoCmd(isoDest string, tmpDir string, userdata string, metadata string) *exec.Cmd {
	return exec.Command(
		" hdiutil",
		"makehybrid",
		"-iso",
		"-joliet",
		"-o",
		isoDest,
		filepath.Join(tmpDir, userdata),
		filepath.Join(tmpDir, metadata))
}
