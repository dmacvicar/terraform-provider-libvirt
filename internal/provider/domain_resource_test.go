package provider

import (
	"context"
	"fmt"
	"testing"

	golibvirt "github.com/digitalocean/go-libvirt"
	libvirtclient "github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDomainResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDomainResourceConfigBasic("test-domain-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-basic"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "memory", "512"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "unit", "MiB"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "vcpu", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "type", "kvm"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "os.type", "hvm"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "os.arch", "x86_64"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "os.machine", "q35"),
					resource.TestCheckResourceAttrSet("libvirt_domain.test", "uuid"),
					resource.TestCheckResourceAttrSet("libvirt_domain.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccDomainResourceConfigBasicUpdated("test-domain-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-basic"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "memory", "1024"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "vcpu", "2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TODO: UEFI firmware configuration needs more work
// The firmware attribute is not directly mapped in older libvirt XML
// We need to handle this through the loader element properly
func TestAccDomainResource_uefi(t *testing.T) {
	t.Skip("UEFI firmware configuration needs additional XML mapping work")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigUEFI("test-domain-uefi"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-uefi"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "os.firmware", "efi"),
					resource.TestCheckResourceAttrSet("libvirt_domain.test", "os.loader_path"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigBasic(name string) string {
	return fmt.Sprintf(`
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }
}
`, name)
}

func testAccDomainResourceConfigBasicUpdated(name string) string {
	return fmt.Sprintf(`
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 1024
  unit   = "MiB"
  vcpu   = 2
  type   = "kvm"

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }
}
`, name)
}

func testAccDomainResourceConfigUEFI(name string) string {
	return fmt.Sprintf(`
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 1024
  unit   = "MiB"
  vcpu   = 2
  type   = "kvm"

  os {
    type        = "hvm"
    arch        = "x86_64"
    machine     = "q35"
    firmware    = "efi"
    loader_path = "/usr/share/edk2/x64/OVMF_CODE.fd"
  }
}
`, name)
}

func TestAccDomainResource_metadata(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigMetadata("test-domain-metadata"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-metadata"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "title", "Test Domain"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "description", "A test domain with metadata"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigMetadata(name string) string {
	return fmt.Sprintf(`
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_domain" "test" {
  name        = %[1]q
  title       = "Test Domain"
  description = "A test domain with metadata"
  memory      = 512
  unit        = "MiB"
  vcpu        = 1
  type        = "kvm"

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }
}
`, name)
}

func TestAccDomainResource_features(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigFeatures("test-domain-features"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-features"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "features.pae", "true"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "features.acpi", "true"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "features.apic", "true"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "features.hap", "on"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "features.pmu", "off"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigFeatures(name string) string {
	return fmt.Sprintf(`
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }

  features {
    pae  = true
    acpi = true
    apic = true
    hap  = "on"
    pmu  = "off"
  }
}
`, name)
}

func TestAccDomainResource_cpu(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigCPU("test-domain-cpu"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-cpu"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "cpu.mode", "host-passthrough"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigCPU(name string) string {
	return fmt.Sprintf(`
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  unit   = "MiB"
  vcpu   = 2
  type   = "kvm"

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }

  cpu {
    mode = "host-passthrough"
  }
}
`, name)
}

func TestAccDomainResource_clock(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigClock("test-domain-clock"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-clock"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.offset", "utc"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigClock(name string) string {
	return fmt.Sprintf(`
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }

  clock {
    offset = "utc"
  }
}
`, name)
}

func TestAccDomainResource_lifecycle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigLifecycle("test-domain-lifecycle"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-lifecycle"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "on_poweroff", "destroy"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "on_reboot", "restart"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "on_crash", "restart"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigLifecycle(name string) string {
	return fmt.Sprintf(`
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  on_poweroff = "destroy"
  on_reboot   = "restart"
  on_crash    = "restart"

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }
}
`, name)
}

func TestAccDomainResource_running(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigRunning("test-domain-running"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-running"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "running", "true"),
					testAccCheckDomainIsRunning("test-domain-running"),
				),
			},
		},
	})
}

func TestAccDomainResource_updateWithRunning(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create domain
			{
				Config: testAccDomainResourceConfigBasic("test-domain-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-update"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "memory", "512"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "vcpu", "1"),
					testAccCheckDomainStart("test-domain-update"),
				),
			},
			// Update while running
			{
				Config: testAccDomainResourceConfigBasicUpdated("test-domain-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-update"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "memory", "1024"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "vcpu", "2"),
				),
			},
		},
	})
}

func TestAccDomainResource_clockTimers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigClockTimers("test-domain-timers"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-timers"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.offset", "utc"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.#", "3"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.0.name", "rtc"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.0.tickpolicy", "catchup"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.1.name", "pit"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.1.tickpolicy", "delay"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.2.name", "hpet"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.2.present", "no"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigClockTimers(name string) string {
	return fmt.Sprintf(`
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }

  clock {
    offset = "utc"

    timer {
      name       = "rtc"
      tickpolicy = "catchup"
    }

    timer {
      name       = "pit"
      tickpolicy = "delay"
    }

    timer {
      name    = "hpet"
      present = "no"
    }
  }
}
`, name)
}

func TestAccDomainResource_clockTimerCatchup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigClockTimerCatchup("test-domain-timer-catchup"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-timer-catchup"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.0.name", "rtc"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.0.tickpolicy", "catchup"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.0.catchup.threshold", "123"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.0.catchup.slew", "120"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.0.catchup.limit", "10000"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigClockTimerCatchup(name string) string {
	return fmt.Sprintf(`
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }

  clock {
    offset = "utc"

    timer {
      name       = "rtc"
      tickpolicy = "catchup"

      catchup {
        threshold = 123
        slew      = 120
        limit     = 10000
      }
    }
  }
}
`, name)
}

func testAccDomainResourceConfigRunning(name string) string {
	return fmt.Sprintf(`
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_domain" "test" {
  name    = %[1]q
  memory  = 512
  unit    = "MiB"
  vcpu    = 1
  type    = "kvm"
  running = true

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }
}
`, name)
}

func testAccCheckDomainIsRunning(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()
		client, err := libvirtclient.NewClient(ctx, "qemu:///system")
		if err != nil {
			return fmt.Errorf("failed to create libvirt client: %w", err)
		}
		defer func() { _ = client.Close() }()

		domains, _, err := client.Libvirt().ConnectListAllDomains(1, 0)
		if err != nil {
			return fmt.Errorf("failed to list domains: %w", err)
		}

		var targetDomain *golibvirt.Domain
		for _, d := range domains {
			if d.Name == name {
				targetDomain = &d
				break
			}
		}

		if targetDomain == nil {
			return fmt.Errorf("domain %s not found", name)
		}

		state, _, err := client.Libvirt().DomainGetState(*targetDomain, 0)
		if err != nil {
			return fmt.Errorf("failed to get domain state: %w", err)
		}

		if uint32(state) != uint32(golibvirt.DomainRunning) {
			return fmt.Errorf("domain is not running, state = %d", state)
		}

		return nil
	}
}

func testAccCheckDomainStart(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()
		client, err := libvirtclient.NewClient(ctx, "qemu:///system")
		if err != nil {
			return fmt.Errorf("failed to create libvirt client: %w", err)
		}
		defer func() { _ = client.Close() }()

		domains, _, err := client.Libvirt().ConnectListAllDomains(1, 0)
		if err != nil {
			return fmt.Errorf("failed to list domains: %w", err)
		}

		var targetDomain *golibvirt.Domain
		for _, d := range domains {
			if d.Name == name {
				targetDomain = &d
				break
			}
		}

		if targetDomain == nil {
			return fmt.Errorf("domain %s not found", name)
		}

		err = client.Libvirt().DomainCreate(*targetDomain)
		if err != nil {
			return fmt.Errorf("failed to start domain %s: %w", name, err)
		}

		state, _, err := client.Libvirt().DomainGetState(*targetDomain, 0)
		if err != nil {
			return fmt.Errorf("failed to get domain state: %w", err)
		}

		if uint32(state) != uint32(golibvirt.DomainRunning) {
			return fmt.Errorf("domain is not running, state = %d", state)
		}

		return nil
	}
}

//nolint:unused
func testAccCheckDomainCanStart(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Create a new libvirt client for testing
		ctx := context.Background()
		client, err := libvirtclient.NewClient(ctx, "qemu:///system")
		if err != nil {
			return fmt.Errorf("failed to create libvirt client: %w", err)
		}
		defer func() { _ = client.Close() }()

		// Look up the domain by name using the raw libvirt API
		domains, _, err := client.Libvirt().ConnectListAllDomains(1, 0)
		if err != nil {
			return fmt.Errorf("failed to list domains: %w", err)
		}

		var targetDomain *golibvirt.Domain
		for _, d := range domains {
			if d.Name == name {
				targetDomain = &d
				break
			}
		}

		if targetDomain == nil {
			return fmt.Errorf("domain %s not found", name)
		}

		// Try to start it
		err = client.Libvirt().DomainCreate(*targetDomain)
		if err != nil {
			return fmt.Errorf("failed to start domain %s: %w", name, err)
		}

		// Verify it's running
		state, _, err := client.Libvirt().DomainGetState(*targetDomain, 0)
		if err != nil {
			return fmt.Errorf("failed to get domain state: %w", err)
		}

		if uint32(state) != uint32(golibvirt.DomainRunning) {
			return fmt.Errorf("domain is not running, state = %d", state)
		}

		// Stop it so cleanup works
		err = client.Libvirt().DomainDestroy(*targetDomain)
		if err != nil {
			return fmt.Errorf("failed to stop domain: %w", err)
		}

		return nil
	}
}

func TestAccDomainResource_network(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigNetwork("test-domain-network"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-network"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.#", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.type", "network"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.source.network", "default"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.model", "virtio"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigNetwork(name string) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }

  devices = {
    interfaces = [
      {
        type  = "network"
        model = "virtio"
        source = {
          network = "default"
        }
      }
    ]
  }
}
`, name)
}

func TestAccDomainResource_graphics(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigGraphicsVNC("test-domain-graphics"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-graphics"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.graphics.vnc.autoport", "yes"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigGraphicsVNC(name string) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }

  devices = {
    graphics = {
      vnc = {
        autoport = "yes"
        listen   = "0.0.0.0"
      }
    }
  }
}
`, name)
}

func TestAccDomainResource_video(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigVideo("test-domain-video"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-video"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.video.type", "vga"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigVideo(name string) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }

  devices = {
    video = {
      type = "vga"
    }
  }
}
`, name)
}

func TestAccDomainResource_emulator(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigEmulator("test-domain-emulator"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-emulator"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.emulator", "/usr/bin/qemu-system-x86_64"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigEmulator(name string) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }

  devices = {
    emulator = "/usr/bin/qemu-system-x86_64"
  }
}
`, name)
}

func TestAccDomainResource_console(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigConsole("test-domain-console"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-console"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.consoles.#", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.consoles.0.type", "pty"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigConsole(name string) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }

  devices = {
    consoles = [
      {
        type        = "pty"
        target_type = "serial"
        target_port = 0
      }
    ]
    serials = [
      {
        type        = "pty"
        target_type = "isa-serial"
        target_port = 0
      }
    ]
  }
}
`, name)
}

func TestAccDomainResource_autostart(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with autostart=true
			{
				Config: testAccDomainResourceConfigAutostart("test-domain-autostart", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-autostart"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "autostart", "true"),
					testAccCheckDomainAutostart("libvirt_domain.test", true),
				),
			},
			// Update to autostart=false
			{
				Config: testAccDomainResourceConfigAutostart("test-domain-autostart", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "autostart", "false"),
					testAccCheckDomainAutostart("libvirt_domain.test", false),
				),
			},
			// Update back to autostart=true
			{
				Config: testAccDomainResourceConfigAutostart("test-domain-autostart", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "autostart", "true"),
					testAccCheckDomainAutostart("libvirt_domain.test", true),
				),
			},
		},
	})
}

func testAccDomainResourceConfigAutostart(name string, autostart bool) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name      = %[1]q
  memory    = 512
  unit      = "MiB"
  vcpu      = 1
  type      = "kvm"
  autostart = %[2]t

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }
}
`, name, autostart)
}

func testAccCheckDomainAutostart(resourceName string, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		uuid := rs.Primary.Attributes["uuid"]
		if uuid == "" {
			return fmt.Errorf("no UUID is set")
		}

		// Get libvirt client
		ctx := context.Background()
		client, err := libvirtclient.NewClient(ctx, "qemu:///system")
		if err != nil {
			return fmt.Errorf("failed to create libvirt client: %w", err)
		}
		defer func() { _ = client.Close() }()

		domain, err := client.LookupDomainByUUID(uuid)
		if err != nil {
			return fmt.Errorf("failed to lookup domain: %w", err)
		}

		autostart, err := client.Libvirt().DomainGetAutostart(domain)
		if err != nil {
			return fmt.Errorf("failed to get autostart: %w", err)
		}

		actualAutostart := (autostart == 1)
		if actualAutostart != expected {
			return fmt.Errorf("expected autostart=%v, got %v", expected, actualAutostart)
		}

		return nil
	}
}

func TestAccDomainResource_preserveUserIntent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigMinimal("test-domain-minimal"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify fields user didn't set are NOT in state (even though libvirt defaults them)
					resource.TestCheckNoResourceAttr("libvirt_domain.test", "on_poweroff"),
					resource.TestCheckNoResourceAttr("libvirt_domain.test", "on_reboot"),
					resource.TestCheckNoResourceAttr("libvirt_domain.test", "on_crash"),
					resource.TestCheckNoResourceAttr("libvirt_domain.test", "autostart"),
					resource.TestCheckNoResourceAttr("libvirt_domain.test", "unit"),
					resource.TestCheckNoResourceAttr("libvirt_domain.test", "type"),
					resource.TestCheckNoResourceAttr("libvirt_domain.test", "current_memory"),
					resource.TestCheckNoResourceAttr("libvirt_domain.test", "os.boot_devices.#"),
					// Verify required fields ARE in state
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-minimal"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "memory", "512"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "vcpu", "1"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigMinimal(name string) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  vcpu   = 1

  os {
    type = "hvm"
  }
}
`, name)
}

func TestAccDomainResource_filesystem(t *testing.T) {
	// Create temporary directories for testing
	sharedDir := t.TempDir()
	dataDir := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigFilesystem("test-domain-fs", sharedDir, dataDir),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-fs"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.#", "2"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.0.source", sharedDir),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.0.target", "shared"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.0.accessmode", "mapped"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.0.readonly", "true"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.1.source", dataDir),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.1.target", "data"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.1.accessmode", "passthrough"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.1.readonly", "false"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigFilesystem(name, sharedDir, dataDir string) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }

  devices = {
    filesystems = [
      {
        source     = %[2]q
        target     = "shared"
        accessmode = "mapped"
        readonly   = true
      },
      {
        source     = %[3]q
        target     = "data"
        accessmode = "passthrough"
        readonly   = false
      }
    ]
  }
}
`, name, sharedDir, dataDir)
}
