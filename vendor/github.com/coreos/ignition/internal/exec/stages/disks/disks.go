// Copyright 2015 CoreOS, Inc.
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

// The storage stage is responsible for partitioning disks, creating RAID
// arrays, formatting partitions, writing files, writing systemd units, and
// writing network units.

package disks

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/internal/exec/stages"
	"github.com/coreos/ignition/internal/exec/util"
	"github.com/coreos/ignition/internal/log"
	"github.com/coreos/ignition/internal/resource"
	"github.com/coreos/ignition/internal/sgdisk"
	"github.com/coreos/ignition/internal/systemd"
)

const (
	name = "disks"
)

var (
	ErrBadFilesystem = errors.New("filesystem is not of the correct type")
)

func init() {
	stages.Register(creator{})
}

type creator struct{}

func (creator) Create(logger *log.Logger, root string, f resource.Fetcher) stages.Stage {
	return &stage{
		Util: util.Util{
			DestDir: root,
			Logger:  logger,
			Fetcher: f,
		},
	}
}

func (creator) Name() string {
	return name
}

type stage struct {
	util.Util

	client *resource.HttpClient
}

func (stage) Name() string {
	return name
}

func (s stage) Run(config types.Config) bool {
	if err := s.createPartitions(config); err != nil {
		s.Logger.Crit("create partitions failed: %v", err)
		return false
	}

	if err := s.createRaids(config); err != nil {
		s.Logger.Crit("failed to create raids: %v", err)
		return false
	}

	if err := s.createFilesystems(config); err != nil {
		s.Logger.Crit("failed to create filesystems: %v", err)
		return false
	}

	return true
}

// waitOnDevices waits for the devices enumerated in devs as a logged operation
// using ctxt for the logging and systemd unit identity.
func (s stage) waitOnDevices(devs []string, ctxt string) error {
	if err := s.LogOp(
		func() error { return systemd.WaitOnDevices(devs, ctxt) },
		"waiting for devices %v", devs,
	); err != nil {
		return fmt.Errorf("failed to wait on %s devs: %v", ctxt, err)
	}

	return nil
}

// createDeviceAliases creates device aliases for every device in devs.
func (s stage) createDeviceAliases(devs []string) error {
	for _, dev := range devs {
		target, err := util.CreateDeviceAlias(dev)
		if err != nil {
			return fmt.Errorf("failed to create device alias for %q: %v", dev, err)
		}
		s.Logger.Info("created device alias for %q: %q -> %q", dev, util.DeviceAlias(dev), target)
	}

	return nil
}

// waitOnDevicesAndCreateAliases simply wraps waitOnDevices and createDeviceAliases.
func (s stage) waitOnDevicesAndCreateAliases(devs []string, ctxt string) error {
	if err := s.waitOnDevices(devs, ctxt); err != nil {
		return err
	}

	if err := s.createDeviceAliases(devs); err != nil {
		return err
	}

	return nil
}

// createPartitions creates the partitions described in config.Storage.Disks.
func (s stage) createPartitions(config types.Config) error {
	if len(config.Storage.Disks) == 0 {
		return nil
	}
	s.Logger.PushPrefix("createPartitions")
	defer s.Logger.PopPrefix()

	devs := []string{}
	for _, disk := range config.Storage.Disks {
		devs = append(devs, string(disk.Device))
	}

	if err := s.waitOnDevicesAndCreateAliases(devs, "disks"); err != nil {
		return err
	}

	for _, dev := range config.Storage.Disks {
		devAlias := util.DeviceAlias(string(dev.Device))

		err := s.Logger.LogOp(func() error {
			op := sgdisk.Begin(s.Logger, devAlias)
			if dev.WipeTable {
				s.Logger.Info("wiping partition table requested on %q", devAlias)
				op.WipeTable(true)
			}

			for _, part := range dev.Partitions {
				op.CreatePartition(sgdisk.Partition{
					Number:   part.Number,
					Length:   uint64(part.Size),
					Offset:   uint64(part.Start),
					Label:    string(part.Label),
					TypeGUID: string(part.TypeGUID),
					GUID:     string(part.GUID),
				})
			}

			if err := op.Commit(); err != nil {
				return fmt.Errorf("commit failure: %v", err)
			}
			return nil
		}, "partitioning %q", devAlias)
		if err != nil {
			return err
		}
	}

	return nil
}

// createRaids creates the raid arrays described in config.Storage.Raid.
func (s stage) createRaids(config types.Config) error {
	if len(config.Storage.Raid) == 0 {
		return nil
	}
	s.Logger.PushPrefix("createRaids")
	defer s.Logger.PopPrefix()

	devs := []string{}
	for _, array := range config.Storage.Raid {
		for _, dev := range array.Devices {
			devs = append(devs, string(dev))
		}
	}

	if err := s.waitOnDevicesAndCreateAliases(devs, "raids"); err != nil {
		return err
	}

	for _, md := range config.Storage.Raid {
		// FIXME(vc): this is utterly flummoxed by a preexisting md.Name, the magic of device-resident md metadata really interferes with us.
		// It's as if what ignition really needs is to turn off automagic md probing/running before getting started.
		args := []string{
			"--create", md.Name,
			"--force",
			"--run",
			"--level", md.Level,
			"--raid-devices", fmt.Sprintf("%d", len(md.Devices)-md.Spares),
		}

		if md.Spares > 0 {
			args = append(args, "--spare-devices", fmt.Sprintf("%d", md.Spares))
		}

		for _, dev := range md.Devices {
			args = append(args, util.DeviceAlias(string(dev)))
		}

		if _, err := s.Logger.LogCmd(
			exec.Command("/sbin/mdadm", args...),
			"creating %q", md.Name,
		); err != nil {
			return fmt.Errorf("mdadm failed: %v", err)
		}
	}

	return nil
}

// createFilesystems creates the filesystems described in config.Storage.Filesystems.
func (s stage) createFilesystems(config types.Config) error {
	fss := make([]types.Mount, 0, len(config.Storage.Filesystems))
	for _, fs := range config.Storage.Filesystems {
		if fs.Mount != nil {
			fss = append(fss, *fs.Mount)
		}
	}

	if len(fss) == 0 {
		return nil
	}
	s.Logger.PushPrefix("createFilesystems")
	defer s.Logger.PopPrefix()

	devs := []string{}
	for _, fs := range fss {
		devs = append(devs, string(fs.Device))
	}

	if err := s.waitOnDevicesAndCreateAliases(devs, "filesystems"); err != nil {
		return err
	}

	for _, fs := range fss {
		if err := s.createFilesystem(fs); err != nil {
			return err
		}
	}

	// udevd registers an IN_CLOSE_WRITE inotify watch on block device
	// nodes, and synthesizes udev "change" events when the watch fires.
	// mkfs.btrfs triggers multiple such events, the first of which
	// occurs while there is no recognizable filesystem on the
	// partition. Thus, if an existing partition is reformatted as
	// btrfs while keeping the same filesystem label, there will be a
	// synthesized uevent that deletes the /dev/disk/by-label symlink
	// and a second one that restores it. If we didn't account for this,
	// a systemd unit that depended on the by-label symlink (e.g.
	// systemd-fsck-root.service) could have the symlink deleted out
	// from under it.
	//
	// There's no way to fix this completely. We can't wait for the
	// restoring uevent to propagate, since we can't determine which
	// specific uevents were triggered by the mkfs. We can wait for
	// udev to settle, though it's conceivable that the deleting uevent
	// has already been processed and the restoring uevent is still
	// sitting in the inotify queue. In practice the uevent queue will
	// be the slow one, so this should be good enough.
	//
	// Test case: boot failure in coreos.ignition.*.btrfsroot kola test.
	if _, err := s.Logger.LogCmd(
		exec.Command("/bin/udevadm", "settle"),
		"waiting for udev to settle",
	); err != nil {
		return fmt.Errorf("udevadm settle failed: %v", err)
	}

	return nil
}

func (s stage) createFilesystem(fs types.Mount) error {
	info, err := s.readFilesystemInfo(fs)
	if err != nil {
		return err
	}

	if fs.Create != nil {
		// If we are using 2.0.0 semantics...

		if !fs.Create.Force && info.format != "" {
			s.Logger.Err("filesystem detected at %q (found %s) and force was not requested", fs.Device, info.format)
			return ErrBadFilesystem
		}
	} else if !fs.WipeFilesystem {
		// If the filesystem isn't forcefully being created, then we need
		// to check if it is of the correct type or that no filesystem exists.

		if info.format == fs.Format &&
			(fs.Label == nil || info.label == *fs.Label) &&
			(fs.UUID == nil || canonicalizeFilesystemUUID(info.format, info.uuid) == canonicalizeFilesystemUUID(fs.Format, *fs.UUID)) {
			s.Logger.Info("filesystem at %q is already correctly formatted. Skipping mkfs...", fs.Device)
			return nil
		} else if info.format != "" {
			s.Logger.Err("filesystem at %q is not of the correct type, label, or UUID (found %s, %q, %s) and a filesystem wipe was not requested", fs.Device, info.format, info.label, info.uuid)
			return ErrBadFilesystem
		}
	}

	mkfs := ""
	var args []string
	if fs.Create == nil {
		args = translateMountOptionSliceToStringSlice(fs.Options)
	} else {
		args = translateCreateOptionSliceToStringSlice(fs.Create.Options)
	}
	switch fs.Format {
	case "btrfs":
		mkfs = "/sbin/mkfs.btrfs"
		args = append(args, "--force")
		if fs.UUID != nil {
			args = append(args, []string{"-U", canonicalizeFilesystemUUID(fs.Format, *fs.UUID)}...)
		}
		if fs.Label != nil {
			args = append(args, []string{"-L", *fs.Label}...)
		}
	case "ext4":
		mkfs = "/sbin/mkfs.ext4"
		args = append(args, "-F")
		if fs.UUID != nil {
			args = append(args, []string{"-U", canonicalizeFilesystemUUID(fs.Format, *fs.UUID)}...)
		}
		if fs.Label != nil {
			args = append(args, []string{"-L", *fs.Label}...)
		}
	case "xfs":
		mkfs = "/sbin/mkfs.xfs"
		args = append(args, "-f")
		if fs.UUID != nil {
			args = append(args, []string{"-m", "uuid=" + canonicalizeFilesystemUUID(fs.Format, *fs.UUID)}...)
		}
		if fs.Label != nil {
			args = append(args, []string{"-L", *fs.Label}...)
		}
	case "swap":
		mkfs = "/sbin/mkswap"
		args = append(args, "-f")
		if fs.UUID != nil {
			args = append(args, []string{"-U", canonicalizeFilesystemUUID(fs.Format, *fs.UUID)}...)
		}
		if fs.Label != nil {
			args = append(args, []string{"-L", *fs.Label}...)
		}
	case "vfat":
		mkfs = "/sbin/mkfs.vfat"
		// There is no force flag for mkfs.vfat, it always destroys any data on
		// the device at which it is pointed.
		if fs.UUID != nil {
			args = append(args, []string{"-i", canonicalizeFilesystemUUID(fs.Format, *fs.UUID)}...)
		}
		if fs.Label != nil {
			args = append(args, []string{"-n", *fs.Label}...)
		}
	default:
		return fmt.Errorf("unsupported filesystem format: %q", fs.Format)
	}

	devAlias := util.DeviceAlias(string(fs.Device))
	args = append(args, devAlias)
	if _, err := s.Logger.LogCmd(
		exec.Command(mkfs, args...),
		"creating %q filesystem on %q",
		fs.Format, devAlias,
	); err != nil {
		return fmt.Errorf("mkfs failed: %v", err)
	}

	return nil
}

// golang--
func translateMountOptionSliceToStringSlice(opts []types.MountOption) []string {
	newOpts := make([]string, len(opts))
	for i, o := range opts {
		newOpts[i] = string(o)
	}
	return newOpts
}

// golang--
func translateCreateOptionSliceToStringSlice(opts []types.CreateOption) []string {
	newOpts := make([]string, len(opts))
	for i, o := range opts {
		newOpts[i] = string(o)
	}
	return newOpts
}

type filesystemInfo struct {
	format string
	uuid   string
	label  string
}

func (s stage) readFilesystemInfo(fs types.Mount) (filesystemInfo, error) {
	res := filesystemInfo{}
	err := s.Logger.LogOp(
		func() error {
			var err error
			res.format, err = util.FilesystemType(fs.Device)
			if err != nil {
				return err
			}
			res.uuid, err = util.FilesystemUUID(fs.Device)
			if err != nil {
				return err
			}
			res.label, err = util.FilesystemLabel(fs.Device)
			if err != nil {
				return err
			}
			s.Logger.Info("found %s filesystem at %q with uuid %q and label %q", res.format, fs.Device, res.uuid, res.label)
			return nil
		},
		"determining filesystem type of %q", fs.Device,
	)

	return res, err
}

// canonicalizeFilesystemUUID does the minimum amount of canonicalization
// required to make two valid equivalent UUIDs compare equal, but doesn't
// attempt to fully validate the UUID.
func canonicalizeFilesystemUUID(format, uuid string) string {
	uuid = strings.ToLower(uuid)
	if format == "vfat" {
		// FAT uses a 32-bit volume ID instead of a UUID. blkid
		// (and the rest of the world) formats it as A1B2-C3D4, but
		// mkfs.fat doesn't permit the dash, so strip it. Older
		// versions of Ignition would fail if the config included
		// the dash, so we need to support omitting it.
		if len(uuid) >= 5 && uuid[4] == '-' {
			uuid = uuid[0:4] + uuid[5:]
		}
	}
	return uuid
}
