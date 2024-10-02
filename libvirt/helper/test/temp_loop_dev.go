package test

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type TempBlockDevice struct {
	TempFile   string
	LoopDevice string
}

// returns the temporary file, the device path and the error.
func CreateTempFormattedLoopDevice(t *testing.T, name string) (*TempBlockDevice, error) {
	blockDev, err := CreateTempLoopDevice(t, name)
	if err != nil {
		return nil, err
	}

	//nolint:gosec
	cmd := exec.Command("/sbin/mkfs.ext4", "-F", "-q", blockDev.LoopDevice)
	log.Printf("[DEBUG] executing command: %s", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		if err := cleanupLoop(blockDev.LoopDevice); err != nil {
			return nil, err
		}

		if err := cleanupFile(blockDev.TempFile); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("error formatting file system: %w", err)
	}

	return blockDev, nil
}

// returns the temporary file, the device and the error.
func CreateTempLVMGroupDevice(t *testing.T, name string) (*TempBlockDevice, error) {
	blockDev, err := CreateTempLoopDevice(t, name)
	if err != nil {
		return nil, err
	}

	//nolint:gosec
	cmd := exec.Command("sudo", "pvcreate", blockDev.LoopDevice)
	log.Printf("[DEBUG] executing command: %s", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		if err := cleanupLoop(blockDev.LoopDevice); err != nil {
			return nil, err
		}

		if err := cleanupFile(blockDev.TempFile); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("error creating LVM partition on %s: %w", blockDev.LoopDevice, err)
	}

	//nolint:gosec
	cmd = exec.Command("sudo", "vgcreate", name, blockDev.LoopDevice)
	log.Printf("[DEBUG] executing command: %s", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		if err := cleanupLoop(blockDev.LoopDevice); err != nil {
			return nil, err
		}

		if err := cleanupFile(blockDev.TempFile); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("error creating LVM partition on %s: %w", blockDev.LoopDevice, err)
	}

	return blockDev, nil
}

// returns the temporary file, the device and the error.
func CreateTempLoopDevice(t *testing.T, name string) (*TempBlockDevice, error) {
	log.Print("[DEBUG] creating a temporary file for loop device")

	// Create a 1MB temp file
	filename := filepath.Join(t.TempDir(), name)

	//nolint:gosec
	cmd := exec.Command("dd", "if=/dev/zero", fmt.Sprintf("of=%s", filename), "bs=1024", "count=2048")
	log.Printf("[DEBUG] executing command: %s\n", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		if err := cleanupFile(filename); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("Error creating file %s: %w", filename, err)
	}

	// Find an available loop device.
	cmd = exec.Command("sudo", "/sbin/losetup", "--find")
	loopdevStr, err := cmd.Output()
	log.Printf("[DEBUG] executing command: %s", strings.Join(cmd.Args, " "))
	if err != nil {
		if err := cleanupFile(filename); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("Error searching for available loop device: %w", err)
	}
	loopdev := filepath.Clean(strings.TrimRight(string(loopdevStr), "\n"))

	// give the same permissions to the loop device as the backing file.
	cmd = exec.Command("sudo", "chown", "--reference", filename, loopdev)
	log.Printf("[DEBUG] executing command: %s", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		if err := cleanupFile(filename); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("Error copying permissions from %s: %w", filename, err)
	}

	// attach the file to a loop device.
	cmd = exec.Command("sudo", "/sbin/losetup", loopdev, filename)
	log.Printf("[DEBUG] executing command: %s", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		if err := cleanupFile(filename); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("Error setting up loop device: %w", err)
	}

	log.Printf("[DEBUG] temporary file %s attached to loop device %s", filename, loopdev)

	return &TempBlockDevice{TempFile: filename, LoopDevice: loopdev}, nil
}

func (b *TempBlockDevice) Cleanup() error {
	err := cleanupLoop(b.LoopDevice)
	if err != nil {
		return err
	}

	err = cleanupFile(b.TempFile)
	if err != nil {
		return err
	}

	return nil
}

func cleanupLoop(loopDevice string) error {
	cmd := exec.Command("sudo", "losetup", "-d", loopDevice)
	if err := cmd.Run(); err != nil {
		log.Printf("[DEBUG] error detaching loop device %s: %s", loopDevice, err)
		return err
	}
	log.Printf("[DEBUG] detaching loop device %s", loopDevice)

	return nil
}

func cleanupFile(tempFile string) error {
	if err := os.Remove(tempFile); err != nil {
		log.Printf("[DEBUG] error removing temporary file %s: %s", tempFile, err)
		return err
	}
	log.Printf("[DEBUG] removing temporary file %s", tempFile)

	return nil
}

func (b *TempBlockDevice) String() string {
	return fmt.Sprintf("TempFile: %s, LoopDevice: %s", b.TempFile, b.LoopDevice)
}
