package ignition

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/coreos/ignition/config/v2_1/types"
)

func TestIngnitionRaid(t *testing.T) {
	testIgnition(t, `
		data "ignition_raid" "foo" {
			name = "foo"
			level = "raid10"
			devices = ["/foo"]
			spares = 42
		}

		data "ignition_config" "test" {
			arrays = [
				"${data.ignition_raid.foo.id}",
			]
		}
	`, func(c *types.Config) error {
		if len(c.Storage.Raid) != 1 {
			return fmt.Errorf("arrays, found %d", len(c.Storage.Raid))
		}

		a := c.Storage.Raid[0]
		if a.Name != "foo" {
			return fmt.Errorf("name, found %q", a.Name)
		}

		if len(a.Devices) != 1 || a.Devices[0] != "/foo" {
			return fmt.Errorf("devices, found %v", a.Devices)
		}

		if a.Level != "raid10" {
			return fmt.Errorf("level, found %q", a.Level)
		}

		if a.Spares != 42 {
			return fmt.Errorf("spares, found %q", a.Spares)
		}

		return nil
	})
}

func TestIngnitionRaidInvalidLevel(t *testing.T) {
	testIgnitionError(t, `
		data "ignition_raid" "foo" {
			name = "foo"
			level = "foo"
			devices = ["/foo"]
			spares = 42
		}

		data "ignition_config" "test" {
			arrays = [
				"${data.ignition_raid.foo.id}",
			]
		}
	`, regexp.MustCompile("raid level"))
}

func TestIngnitionRaidInvalidDevices(t *testing.T) {
	testIgnitionError(t, `
		data "ignition_raid" "foo" {
			name = "foo"
			level = "raid10"
			devices = ["foo"]
			spares = 42
		}

		data "ignition_config" "test" {
			arrays = [
				"${data.ignition_raid.foo.id}",
			]
		}
	`, regexp.MustCompile("absolute"))
}
