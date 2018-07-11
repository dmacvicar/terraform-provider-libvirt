// Copyright 2017 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package blackbox

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/coreos/ignition/internal/distro"
	"github.com/coreos/ignition/tests/types"
)

func run(t *testing.T, ctx context.Context, command string, args ...string) ([]byte, error) {
	out, err := exec.CommandContext(ctx, command, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed: %q: %v\n%s", command, err, out)
	}
	return out, nil
}

// Runs the command even if the context has exired. Should be used for cleanup
// operations
func runWithoutContext(t *testing.T, command string, args ...string) ([]byte, error) {
	out, err := exec.Command(command, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed: %q: %v\n%s", command, err, out)
	}
	return out, nil
}

func prepareRootPartitionForPasswd(t *testing.T, ctx context.Context, partitions []*types.Partition) error {
	mountPath := getRootLocation(partitions)
	if mountPath == "" {
		// Guess there's no root partition
		return nil
	}
	dirs := []string{
		filepath.Join(mountPath, "home"),
		filepath.Join(mountPath, "usr", "bin"),
		filepath.Join(mountPath, "usr", "sbin"),
		filepath.Join(mountPath, "usr", "lib64"),
		filepath.Join(mountPath, "etc"),
	}
	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	symlinks := []string{"lib64", "bin", "sbin"}
	for _, symlink := range symlinks {
		err := os.Symlink(
			filepath.Join(mountPath, "usr", symlink),
			filepath.Join(mountPath, symlink))
		if err != nil {
			return err
		}
	}

	// TODO: use the architecture, not hardcode amd64
	err := exec.CommandContext(ctx, "cp", "bin/amd64/id-stub", filepath.Join(mountPath, distro.IdCmd())).Run()
	if err != nil {
		return err
	}
	// TODO: needed for user_group_lookup.c
	err = exec.CommandContext(ctx, "cp", "/lib64/libnss_files.so.2", filepath.Join(mountPath, "usr", "lib64")).Run()
	return err
}

func getRootLocation(partitions []*types.Partition) string {
	for _, p := range partitions {
		if p.Label == "ROOT" {
			return p.MountPath
		}
	}
	return ""
}

// returns true if no error, false if error
func runIgnition(t *testing.T, ctx context.Context, stage, root, cwd string, appendEnv []string) error {
	args := []string{"-clear-cache", "-oem", "file", "-stage", stage, "-root", root, "-log-to-stdout"}
	cmd := exec.CommandContext(ctx, "ignition", args...)
	t.Log("ignition", args)
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(), appendEnv...)
	out, err := cmd.CombinedOutput()
	t.Logf("PID: %d", cmd.Process.Pid)
	t.Logf("Ignition output:\n%s", string(out))
	if strings.Contains(string(out), "panic") {
		return fmt.Errorf("ignition panicked")
	}
	return err
}

// pickPartition will return the partition device corresponding to a
// partition with a given label on the given loop device
func pickPartition(t *testing.T, device string, partitions []*types.Partition, label string) string {
	for _, p := range partitions {
		if p.Label == label {
			return fmt.Sprintf("%sp%d", device, p.Number)
		}
	}
	return ""
}

// createVolume will create the image file of the specified size, create a
// partition table in it, and generate mount paths for every partition
func createVolume(t *testing.T, ctx context.Context, index int, imageFile string, size int64, cylinders int, heads int, sectorsPerTrack int, partitions []*types.Partition) (err error) {
	// attempt to create the file, will leave already existing files alone.
	// os.Truncate requires the file to already exist
	var out *os.File
	out, err = os.Create(imageFile)
	if err != nil {
		return err
	}
	defer func() {
		// Delete the image file if this function exits with an error
		if err != nil {
			os.Remove(imageFile)
		}
	}()
	out.Close()

	// Truncate the file to the given size
	err = os.Truncate(imageFile, size)
	if err != nil {
		return err
	}

	err = createPartitionTable(t, ctx, imageFile, partitions)
	if err != nil {
		return err
	}

	for counter, partition := range partitions {
		if partition.TypeCode == "blank" || partition.FilesystemType == "" {
			continue
		}

		partition.MountPath = filepath.Join(os.TempDir(), fmt.Sprintf("hd%dp%d", index, counter))
		if err := os.Mkdir(partition.MountPath, 0777); err != nil {
			return err
		}
		mountPath := partition.MountPath
		defer func() {
			// Delete the mount path if this function exits with an error
			if err != nil {
				os.RemoveAll(mountPath)
			}
		}()
	}
	return nil
}

// setDevices will create devices for each of the partitions in the imageFile,
// and then will format each partition according to what's described in the
// partitions argument.
func setDevices(t *testing.T, ctx context.Context, imageFile string, partitions []*types.Partition) (string, error) {
	out, err := run(t, ctx, "losetup", "-Pf", "--show", imageFile)
	if err != nil {
		return "", err
	}
	loopDevice := strings.TrimSpace(string(out))

	for _, partition := range partitions {
		if partition.TypeCode == "blank" || partition.FilesystemType == "" {
			continue
		}
		partition.Device = fmt.Sprintf("%sp%d", loopDevice, partition.Number)
		err := formatPartition(t, ctx, partition)
		if err != nil {
			return "", err
		}
	}
	return loopDevice, nil
}

func destroyDevice(t *testing.T, loopDevice string) error {
	_, err := runWithoutContext(t, "losetup", "-d", loopDevice)
	return err
}

func formatPartition(t *testing.T, ctx context.Context, partition *types.Partition) error {
	var mkfs string
	var opts, label, uuid []string

	switch partition.FilesystemType {
	case "vfat":
		mkfs = "mkfs.vfat"
		label = []string{"-n", partition.FilesystemLabel}
		uuid = []string{"-i", partition.FilesystemUUID}
	case "ext2", "ext4":
		mkfs = "mke2fs"
		opts = []string{
			"-t", partition.FilesystemType, "-b", "4096",
			"-i", "4096", "-I", "128", "-e", "remount-ro",
		}
		label = []string{"-L", partition.FilesystemLabel}
		uuid = []string{"-U", partition.FilesystemUUID}
	case "btrfs":
		mkfs = "mkfs.btrfs"
		label = []string{"--label", partition.FilesystemLabel}
		uuid = []string{"--uuid", partition.FilesystemUUID}
	case "xfs":
		mkfs = "mkfs.xfs"
		label = []string{"-L", partition.FilesystemLabel}
		uuid = []string{"-m", "uuid=" + partition.FilesystemUUID}
	case "swap":
		mkfs = "mkswap"
		label = []string{"-L", partition.FilesystemLabel}
		uuid = []string{"-U", partition.FilesystemUUID}
	default:
		if partition.FilesystemType == "blank" ||
			partition.FilesystemType == "" {
			return nil
		}
		return fmt.Errorf("Unknown partition: %v", partition.FilesystemType)
	}

	if partition.FilesystemLabel != "" {
		opts = append(opts, label...)
	}
	if partition.FilesystemUUID != "" {
		opts = append(opts, uuid...)
	}
	opts = append(opts, partition.Device)

	_, err := run(t, ctx, mkfs, opts...)
	if err != nil {
		return err
	}

	if (partition.FilesystemType == "ext2" || partition.FilesystemType == "ext4") && partition.TypeCode == "coreos-usr" {
		// this is done to mirror the functionality from disk_util
		opts := []string{
			"-U", "clear", "-T", "20091119110000", "-c", "0", "-i", "0",
			"-m", "0", "-r", "0", "-e", "remount-ro", partition.Device,
		}
		_, err = run(t, ctx, "tune2fs", opts...)
		if err != nil {
			return err
		}
	}
	return nil
}

func createPartitionTable(t *testing.T, ctx context.Context, imageFile string, partitions []*types.Partition) error {
	opts := []string{imageFile}
	hybrids := []int{}
	for _, p := range partitions {
		if p.TypeCode == "blank" || p.Length == 0 {
			continue
		}
		opts = append(opts, fmt.Sprintf(
			"--new=%d:%d:+%d", p.Number, p.Offset, p.Length))
		opts = append(opts, fmt.Sprintf(
			"--change-name=%d:%s", p.Number, p.Label))
		if p.TypeGUID != "" {
			opts = append(opts, fmt.Sprintf(
				"--typecode=%d:%s", p.Number, p.TypeGUID))
		}
		if p.GUID != "" {
			opts = append(opts, fmt.Sprintf(
				"--partition-guid=%d:%s", p.Number, p.GUID))
		}
		if p.Hybrid {
			hybrids = append(hybrids, p.Number)
		}
	}
	if len(hybrids) > 0 {
		if len(hybrids) > 3 {
			return fmt.Errorf("Can't have more than three hybrids")
		} else {
			opts = append(opts, fmt.Sprintf("-h=%s", intJoin(hybrids, ":")))
		}
	}
	_, err := run(t, ctx, "sgdisk", opts...)
	return err
}

func mountRootPartition(t *testing.T, ctx context.Context, partitions []*types.Partition) (bool, error) {
	for _, partition := range partitions {
		if partition.Label != "ROOT" {
			continue
		}
		args := []string{partition.Device, partition.MountPath}
		_, err := run(t, ctx, "mount", args...)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func mountPartitions(t *testing.T, ctx context.Context, partitions []*types.Partition) error {
	for _, partition := range partitions {
		if partition.FilesystemType == "" || partition.FilesystemType == "swap" || partition.Label == "ROOT" {
			continue
		}
		args := []string{partition.Device, partition.MountPath}
		_, err := run(t, ctx, "mount", args...)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateTypeGUID(t *testing.T, partition *types.Partition) error {
	partitionTypes := map[string]string{
		"coreos-resize":   "3884DD41-8582-4404-B9A8-E9B84F2DF50E",
		"data":            "0FC63DAF-8483-4772-8E79-3D69D8477DE4",
		"coreos-rootfs":   "5DFBF5F4-2848-4BAC-AA5E-0D9A20B745A6",
		"bios":            "21686148-6449-6E6F-744E-656564454649",
		"efi":             "C12A7328-F81F-11D2-BA4B-00A0C93EC93B",
		"coreos-reserved": "C95DC21A-DF0E-4340-8D7B-26CBFA9A03E0",
	}

	if partition.TypeCode == "" || partition.TypeCode == "blank" {
		return nil
	}

	partition.TypeGUID = partitionTypes[partition.TypeCode]
	if partition.TypeGUID == "" {
		return fmt.Errorf("Unknown TypeCode: %s", partition.TypeCode)
	}
	return nil
}

func intJoin(ints []int, delimiter string) string {
	strArr := []string{}
	for _, i := range ints {
		strArr = append(strArr, strconv.Itoa(i))
	}
	return strings.Join(strArr, delimiter)
}

func removeEmpty(strings []string) []string {
	var r []string
	for _, str := range strings {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func generateUUID(t *testing.T, ctx context.Context) (string, error) {
	out, err := run(t, ctx, "uuidgen")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func createFilesForPartitions(t *testing.T, partitions []*types.Partition) error {
	for _, partition := range partitions {
		err := createDirectoriesFromSlice(t, partition.MountPath, partition.Directories)
		if err != nil {
			return err
		}
		createFilesFromSlice(t, partition.MountPath, partition.Files)
		if err != nil {
			return err
		}
		createLinksFromSlice(t, partition.MountPath, partition.Links)
		if err != nil {
			return err
		}
	}
	return nil
}

func createFilesFromSlice(t *testing.T, basedir string, files []types.File) error {
	for _, file := range files {
		err := os.MkdirAll(filepath.Join(
			basedir, file.Directory), 0755)
		if err != nil {
			return err
		}
		f, err := os.Create(filepath.Join(
			basedir, file.Directory, file.Name))
		if err != nil {
			return err
		}
		defer f.Close()
		if file.Contents != "" {
			writer := bufio.NewWriter(f)
			_, err := writer.WriteString(file.Contents)
			if err != nil {
				return err
			}
			writer.Flush()
		}
	}
	return nil
}

func createDirectoriesFromSlice(t *testing.T, basedir string, dirs []types.Directory) error {
	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(
			basedir, dir.Directory), 0755)
		if err != nil {
			return err
		}
		err = os.Mkdir(filepath.Join(
			basedir, dir.Directory, dir.Name), os.FileMode(dir.Mode))
		if err != nil {
			return err
		}
	}
	return nil
}

func createLinksFromSlice(t *testing.T, basedir string, links []types.Link) error {
	for _, link := range links {
		err := os.MkdirAll(filepath.Join(
			basedir, link.Directory), 0755)
		if err != nil {
			return err
		}
		if link.Hard {
			err = os.Link(link.Target, filepath.Join(basedir, link.Directory, link.Name))
		} else {
			err = os.Symlink(link.Target, filepath.Join(basedir, link.Directory, link.Name))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func unmountRootPartition(t *testing.T, partitions []*types.Partition) error {
	for _, partition := range partitions {
		if partition.Label != "ROOT" {
			continue
		}

		_, err := runWithoutContext(t, "umount", partition.Device)
		if err != nil {
			return err
		}
	}
	return nil
}

func unmountPartitions(t *testing.T, partitions []*types.Partition) error {
	for _, partition := range partitions {
		if partition.FilesystemType == "" || partition.FilesystemType == "swap" || partition.Label == "ROOT" {
			continue
		}

		_, err := runWithoutContext(t, "umount", partition.Device)
		if err != nil {
			return err
		}
	}
	return nil
}

func setExpectedPartitionsDrive(actual []*types.Partition, expected []*types.Partition) {
	for _, a := range actual {
		for _, e := range expected {
			if a.Number == e.Number {
				e.MountPath = a.MountPath
				e.Device = a.Device
				break
			}
		}
	}
}
