package ignition

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/coreos/ignition/config/v2_1/types"
)

func TestIngnitionDirectory(t *testing.T) {
	testIgnition(t, `
		data "ignition_directory" "foo" {
			filesystem = "foo"
			path = "/foo"
			mode = 420
			uid = 42
			gid = 84
		}

		data "ignition_config" "test" {
			directories = [
				"${data.ignition_directory.foo.id}",
			]
		}
	`, func(c *types.Config) error {
		if len(c.Storage.Directories) != 1 {
			return fmt.Errorf("arrays, found %d", len(c.Storage.Raid))
		}

		f := c.Storage.Directories[0]
		if f.Filesystem != "foo" {
			return fmt.Errorf("filesystem, found %q", f.Filesystem)
		}

		if f.Path != "/foo" {
			return fmt.Errorf("path, found %q", f.Path)
		}

		if f.Mode != 420 {
			return fmt.Errorf("mode, found %q", f.Mode)
		}

		if *f.User.ID != 42 {
			return fmt.Errorf("uid, found %q", *f.User.ID)
		}

		if *f.Group.ID != 84 {
			return fmt.Errorf("gid, found %q", *f.Group.ID)
		}

		return nil
	})
}

func TestIngnitionDirectoryInvalidMode(t *testing.T) {
	testIgnitionError(t, `
		data "ignition_directory" "foo" {
			filesystem = "foo"
			path = "/foo"
			mode = 999999
		}

		data "ignition_config" "test" {
			directories = [
				"${data.ignition_directory.foo.id}",
			]
		}
	`, regexp.MustCompile("illegal file mode"))
}

func TestIngnitionDirectoryInvalidPath(t *testing.T) {
	testIgnitionError(t, `
		data "ignition_directory" "foo" {
			filesystem = "foo"
			path = "foo"
			mode = 999999
		}

		data "ignition_config" "test" {
			directories = [
				"${data.ignition_directory.foo.id}",
			]
		}
	`, regexp.MustCompile("absolute"))
}
