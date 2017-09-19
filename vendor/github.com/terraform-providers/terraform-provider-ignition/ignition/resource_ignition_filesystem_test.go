package ignition

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/coreos/ignition/config/v2_1/types"
)

func TestIngnitionFilesystem(t *testing.T) {
	testIgnition(t, `
		data "ignition_filesystem" "foo" {
			name = "foo"
			path = "/foo"
		}

		data "ignition_filesystem" "qux" {
			name = "qux"
			mount {
				device = "/qux"
				format = "ext4"
			}
		}

		data "ignition_filesystem" "baz" {
			name = "baz"
			mount {
				device = "/baz"
				format = "ext4"
				wipe_filesystem = true
				label = "root"
				uuid = "qux"
				options = ["rw"]
			}
		}

		data "ignition_config" "test" {
			filesystems = [
				"${data.ignition_filesystem.foo.id}",
				"${data.ignition_filesystem.qux.id}",
				"${data.ignition_filesystem.baz.id}",
			]
		}
	`, func(c *types.Config) error {
		if len(c.Storage.Filesystems) != 3 {
			return fmt.Errorf("disks, found %d", len(c.Storage.Filesystems))
		}

		f := c.Storage.Filesystems[0]
		if f.Name != "foo" {
			return fmt.Errorf("name, found %q", f.Name)
		}

		if f.Mount != nil {
			return fmt.Errorf("mount, found %q", f.Mount.Device)
		}

		if string(*f.Path) != "/foo" {
			return fmt.Errorf("path, found %q", f.Path)
		}

		f = c.Storage.Filesystems[1]
		if f.Name != "qux" {
			return fmt.Errorf("name, found %q", f.Name)
		}

		if f.Mount.Device != "/qux" {
			return fmt.Errorf("mount.0.device, found %q", f.Mount.Device)
		}

		if f.Mount.Format != "ext4" {
			return fmt.Errorf("mount.0.format, found %q", f.Mount.Format)
		}

		if f.Mount.Create != nil {
			return fmt.Errorf("mount, create was found %#v", f.Mount.Create)
		}

		f = c.Storage.Filesystems[2]
		if f.Name != "baz" {
			return fmt.Errorf("name, found %q", f.Name)
		}

		if f.Mount.Device != "/baz" {
			return fmt.Errorf("mount.0.device, found %q", f.Mount.Device)
		}

		if f.Mount.Format != "ext4" {
			return fmt.Errorf("mount.0.format, found %q", f.Mount.Format)
		}

		if *f.Mount.Label != "root" {
			return fmt.Errorf("mount.0.label, found %q", *f.Mount.Label)
		}

		if *f.Mount.UUID != "qux" {
			return fmt.Errorf("mount.0.uuid, found %q", *f.Mount.UUID)
		}

		if f.Mount.WipeFilesystem != true {
			return fmt.Errorf("mount.0.wipe_filesystem, found %t", f.Mount.WipeFilesystem)
		}

		if len(f.Mount.Options) != 1 || f.Mount.Options[0] != "rw" {
			return fmt.Errorf("mount.0.options, found %q", f.Mount.Options)
		}

		return nil
	})
}

func TestIngnitionFilesystemInvalidPath(t *testing.T) {
	testIgnitionError(t, `
		data "ignition_filesystem" "foo" {
			name = "foo"
			path = "foo"
		}

		data "ignition_config" "test" {
			filesystems = [
				"${data.ignition_filesystem.foo.id}",
			]
		}
	`, regexp.MustCompile("absolute"))
}

func TestIngnitionFilesystemInvalidPathAndMount(t *testing.T) {
	testIgnitionError(t, `
		data "ignition_filesystem" "foo" {
			name = "foo"
			path = "/foo"
			mount {
				device = "/qux"
				format = "ext4"
			}
		}

		data "ignition_config" "test" {
			filesystems = [
				"${data.ignition_filesystem.foo.id}",
			]
		}
	`, regexp.MustCompile("mount and path"))
}
