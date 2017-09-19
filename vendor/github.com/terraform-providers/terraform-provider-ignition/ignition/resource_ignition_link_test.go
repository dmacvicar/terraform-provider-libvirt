package ignition

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/coreos/ignition/config/v2_1/types"
)

func TestIngnitionLink(t *testing.T) {
	testIgnition(t, `
		data "ignition_link" "foo" {
			filesystem = "foo"
			path = "/foo"
			target = "/bar"
			hard = true
			uid = 42
			gid = 84
		}

		data "ignition_config" "test" {
			links = [
				"${data.ignition_link.foo.id}",
			]
		}
	`, func(c *types.Config) error {
		if len(c.Storage.Links) != 1 {
			return fmt.Errorf("arrays, found %d", len(c.Storage.Raid))
		}

		f := c.Storage.Links[0]
		if f.Filesystem != "foo" {
			return fmt.Errorf("filesystem, found %q", f.Filesystem)
		}

		if f.Path != "/foo" {
			return fmt.Errorf("path, found %q", f.Path)
		}

		if f.Target != "/bar" {
			return fmt.Errorf("target, found %q", f.Target)
		}

		if f.Hard != true {
			return fmt.Errorf("hard, found %v", f.Hard)
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

func TestIngnitionLinkInvalidPath(t *testing.T) {
	testIgnitionError(t, `
		data "ignition_link" "foo" {
			filesystem = "foo"
			path = "foo"
			target = "bar"
		}

		data "ignition_config" "test" {
			links = [
				"${data.ignition_link.foo.id}",
			]
		}
	`, regexp.MustCompile("absolute"))
}
