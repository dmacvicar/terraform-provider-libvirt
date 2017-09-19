package ignition

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/coreos/ignition/config/v2_1/types"
)

func TestIngnitionFile(t *testing.T) {
	testIgnition(t, `
		data "ignition_file" "foo" {
			filesystem = "foo"
			path = "/foo"
			content {
				content = "foo"
			}
			mode = 420
			uid = 42
			gid = 84
		}

		data "ignition_file" "qux" {
			filesystem = "qux"
			path = "/qux"
			source {
				source = "qux"
				compression = "gzip"
				verification = "sha512-0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
			}
		}

		data "ignition_file" "nop" {
			filesystem = "nop"
			path = "/nop"
			source {
				source = "nop"
				compression = "gzip"
			}
		}

		data "ignition_config" "test" {
			files = [
				"${data.ignition_file.foo.id}",
				"${data.ignition_file.qux.id}",
				"${data.ignition_file.nop.id}",
			]
		}
	`, func(c *types.Config) error {
		if len(c.Storage.Files) != 3 {
			return fmt.Errorf("arrays, found %d", len(c.Storage.Raid))
		}

		f := c.Storage.Files[0]
		if f.Filesystem != "foo" {
			return fmt.Errorf("filesystem, found %q", f.Filesystem)
		}

		if f.Path != "/foo" {
			return fmt.Errorf("path, found %q", f.Path)
		}

		if f.Contents.Source != "data:text/plain;charset=utf-8;base64,Zm9v" {
			return fmt.Errorf("contents.source, found %q", f.Contents.Source)
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

		f = c.Storage.Files[1]
		if f.Filesystem != "qux" {
			return fmt.Errorf("filesystem, found %q", f.Filesystem)
		}

		if f.Path != "/qux" {
			return fmt.Errorf("path, found %q", f.Path)
		}

		if f.Contents.Source != "qux" {
			return fmt.Errorf("contents.source, found %q", f.Contents.Source)
		}

		if f.Contents.Compression != "gzip" {
			return fmt.Errorf("contents.compression, found %q", f.Contents.Compression)
		}

		if *f.Contents.Verification.Hash != "sha512-0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" {
			return fmt.Errorf("config.replace.verification, found %q", *f.Contents.Verification.Hash)
		}

		f = c.Storage.Files[2]
		if f.Filesystem != "nop" {
			return fmt.Errorf("filesystem, found %q", f.Filesystem)
		}

		if f.Path != "/nop" {
			return fmt.Errorf("path, found %q", f.Path)
		}

		if f.Contents.Source != "nop" {
			return fmt.Errorf("contents.source, found %q", f.Contents.Source)
		}

		if f.Contents.Compression != "gzip" {
			return fmt.Errorf("contents.compression, found %q", f.Contents.Compression)
		}

		if f.Contents.Verification.Hash != nil {
			return fmt.Errorf("contents.verification should be nil, found %q", f.Contents.Verification.Hash)
		}

		return nil
	})
}

func TestIngnitionFileInvalidMode(t *testing.T) {
	testIgnitionError(t, `
		data "ignition_file" "foo" {
			filesystem = "foo"
			path = "/foo"
			mode = 999999
			content {
				content = "foo"
			}
		}

		data "ignition_config" "test" {
			files = [
				"${data.ignition_file.foo.id}",
			]
		}
	`, regexp.MustCompile("illegal file mode"))
}

func TestIngnitionFileInvalidPath(t *testing.T) {
	testIgnitionError(t, `
		data "ignition_file" "foo" {
			filesystem = "foo"
			path = "foo"
			mode = 999999
			content {
				content = "foo"
			}
		}

		data "ignition_config" "test" {
			files = [
				"${data.ignition_file.foo.id}",
			]
		}
	`, regexp.MustCompile("absolute"))
}
