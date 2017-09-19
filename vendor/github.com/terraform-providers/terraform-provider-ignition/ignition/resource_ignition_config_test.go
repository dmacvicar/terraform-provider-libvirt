package ignition

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/coreos/ignition/config/v2_1/types"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestIngnitionFileReplace(t *testing.T) {
	testIgnition(t, `
		data "ignition_config" "test" {
			replace {
				source = "foo"
				verification = "sha512-0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
			}
		}
	`, func(c *types.Config) error {
		r := c.Ignition.Config.Replace
		if r == nil {
			return fmt.Errorf("unable to find replace config")
		}

		if r.Source != "foo" {
			return fmt.Errorf("config.replace.source, found %q", r.Source)
		}

		if *r.Verification.Hash != "sha512-0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" {
			return fmt.Errorf("config.replace.verification, found %q", *r.Verification.Hash)
		}

		return nil
	})
}

func TestIngnitionFileAppend(t *testing.T) {
	testIgnition(t, `
		data "ignition_config" "test" {
			append {
				source = "foo"
				verification = "sha512-0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
			}

		    append {
		    	source = "foo"
		    	verification = "sha512-0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
			}
		}
	`, func(c *types.Config) error {
		a := c.Ignition.Config.Append
		if len(a) != 2 {
			return fmt.Errorf("unable to find append config, expected 2")
		}

		if a[0].Source != "foo" {
			return fmt.Errorf("config.replace.source, found %q", a[0].Source)
		}

		if *a[0].Verification.Hash != "sha512-0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" {
			return fmt.Errorf("config.replace.verification, found %q", *a[0].Verification.Hash)
		}

		return nil
	})
}

func TestIngnitionFileReplaceNoVerification(t *testing.T) {
	testIgnition(t, `
		data "ignition_config" "test" {
			replace {
				source = "foo"
			}
		}
	`, func(c *types.Config) error {
		r := c.Ignition.Config.Replace
		if r == nil {
			return fmt.Errorf("unable to find replace config")
		}

		if r.Source != "foo" {
			return fmt.Errorf("config.replace.source, found %q", r.Source)
		}

		if r.Verification.Hash != nil {
			return fmt.Errorf("verification hash should be nil")
		}

		return nil
	})
}

func TestIngnitionFileAppendNoVerification(t *testing.T) {
	testIgnition(t, `
		data "ignition_config" "test" {
			append {
				source = "foo"
			}

			append {
				source = "foo"
			}
		}
	`, func(c *types.Config) error {
		a := c.Ignition.Config.Append
		if len(a) != 2 {
			return fmt.Errorf("unable to find append config, expected 2")
		}

		if a[0].Source != "foo" {
			return fmt.Errorf("config.replace.source, found %q", a[0].Source)
		}

		if a[0].Verification.Hash != nil {
			return fmt.Errorf("verification hash should be nil")
		}

		return nil
	})
}

func TestIgnitionConfigDisks(t *testing.T) {
	testIgnition(t, `
	variable "ignition_disk_ids" {
		type = "list"
		default = [""]
	}

	data "ignition_disk" "test" {
		device = "/dev/sda"
		partition {
			start = 2048
			size = 20480
		}
	 }

	data "ignition_config" "test" {
		disks = [
			"${data.ignition_disk.test.id}",
			"${var.ignition_disk_ids}",
		]
	}
	`, func(c *types.Config) error {
		f := c.Storage.Disks[0]
		if f.Device != "/dev/sda" {
			return fmt.Errorf("device, found %q", f.Device)
		}
		return nil
	})
}

func TestIgnitionConfigArrays(t *testing.T) {
	testIgnition(t, `
	variable "ignition_array_ids" {
		type = "list"
		default = [""]
	}

	data "ignition_raid" "md" {
		name = "data"
		level = "stripe"
		devices = [
			"/dev/disk/by-partlabel/raid.1.1",
			"/dev/disk/by-partlabel/raid.1.2"
		]
	}

	data "ignition_config" "test" {
		arrays = [
			"${data.ignition_raid.md.id}",
			"${var.ignition_array_ids}"
		]
	}
	`, func(c *types.Config) error {
		f := c.Storage.Raid[0]
		if f.Name != "data" {
			return fmt.Errorf("device, found %q", f.Name)
		}
		return nil
	})
}

func TestIgnitionConfigFilesystems(t *testing.T) {
	testIgnition(t, `
	variable "ignition_filesystem_ids" {
		type = "list"
		default = [""]
	}

	data "ignition_filesystem" "test" {
		name = "test"
		mount = {
			device = "/dev/sda"
			format = "ext4"
	 	}
	 }

	data "ignition_config" "test" {
		filesystems = [
			"${data.ignition_filesystem.test.id}",
			"${var.ignition_filesystem_ids}",
		]
	}
	`, func(c *types.Config) error {
		f := c.Storage.Filesystems[0]
		if f.Name != "test" {
			return fmt.Errorf("device, found %q", f.Name)
		}
		return nil
	})
}

func TestIgnitionConfigFiles(t *testing.T) {
	testIgnition(t, `
	variable "ignition_file_ids" {
		type = "list"
		default = [""]
	}

	data "ignition_file" "test" {
		filesystem = "foo"
		path = "/hello.text"
		content {
			content = "Hello World!"
		}
	 }

	data "ignition_config" "test" {
		files = [
			"${data.ignition_file.test.id}",
			"${var.ignition_file_ids}",
		]
	}
	`, func(c *types.Config) error {
		f := c.Storage.Files[0]
		if f.Filesystem != "foo" {
			return fmt.Errorf("device, found %q", f.Filesystem)
		}
		return nil
	})
}

func TestIgnitionConfigSystemd(t *testing.T) {
	testIgnition(t, `
	variable "ignition_systemd_ids" {
		type = "list"
		default = [""]
	}

	data "ignition_systemd_unit" "test" {
		name = "example.service"
		content = "[Service]\nType=oneshot\nExecStart=/usr/bin/echo Hello World\n\n[Install]\nWantedBy=multi-user.target"
	}

	data "ignition_config" "test" {
		systemd = [
			"${data.ignition_systemd_unit.test.id}",
			"${var.ignition_systemd_ids}",
		]
	}
	`, func(c *types.Config) error {
		f := c.Systemd.Units[0]
		if f.Name != "example.service" {
			return fmt.Errorf("device, found %q", f.Name)
		}
		return nil
	})
}

func TestIgnitionConfigNetworkd(t *testing.T) {
	testIgnition(t, `
	variable "ignition_networkd_ids" {
		type = "list"
		default = [""]
	}

	data "ignition_networkd_unit" "test" {
		name = "00-eth0.network"
		content = "[Match]\nName=eth0\n\n[Network]\nAddress=10.0.1.7"
	}

	data "ignition_config" "test" {
		networkd = [
			"${data.ignition_networkd_unit.test.id}",
			"${var.ignition_networkd_ids}",
		]
	}
	`, func(c *types.Config) error {
		f := c.Networkd.Units[0]
		if f.Name != "00-eth0.network" {
			return fmt.Errorf("device, found %q", f.Name)
		}
		return nil
	})
}

func TestIgnitionConfigUsers(t *testing.T) {
	testIgnition(t, `
	variable "ignition_user_ids" {
		type = "list"
		default = [""]
	}

	data "ignition_user" "test" {
		name = "foo"
		home_dir = "/home/foo/"
		shell = "/bin/bash"
	}

	data "ignition_config" "test" {
		users = [
			"${data.ignition_user.test.id}",
			"${var.ignition_user_ids}",
		]
	}
	`, func(c *types.Config) error {
		f := c.Passwd.Users[0]
		if f.Name != "foo" {
			return fmt.Errorf("device, found %q", f.Name)
		}
		return nil
	})
}

func TestIgnitionConfigGroupss(t *testing.T) {
	testIgnition(t, `
	variable "ignition_group_ids" {
		type = "list"
		default = [""]
	}

	data "ignition_group" "test" {
		name = "test"
	}

	data "ignition_config" "test" {
		groups = [
			"${data.ignition_group.test.id}",
			"${var.ignition_group_ids}",
		]
	}
	`, func(c *types.Config) error {
		f := c.Passwd.Groups[0]
		if f.Name != "test" {
			return fmt.Errorf("device, found %q", f.Name)
		}
		return nil
	})
}

func testIgnitionError(t *testing.T, input string, expectedErr *regexp.Regexp) {
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testProviders,
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(testTemplate, input),
				ExpectError: expectedErr,
			},
		},
	})
}

func testIgnition(t *testing.T, input string, assert func(*types.Config) error) {
	check := func(s *terraform.State) error {
		got := s.RootModule().Outputs["rendered"].Value.(string)

		c := &types.Config{}
		err := json.Unmarshal([]byte(got), c)
		if err != nil {
			return err
		}

		return assert(c)
	}

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testTemplate, input),
				Check:  check,
			},
		},
	})
}

var testTemplate = `
%s

output "rendered" {
	value = "${data.ignition_config.test.rendered}"
}

`
