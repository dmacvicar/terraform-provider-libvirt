package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	golibvirt "github.com/digitalocean/go-libvirt"
	libvirtclient "github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func init() {
	resource.AddTestSweepers("libvirt_domain", &resource.Sweeper{
		Name: "libvirt_domain",
		F: func(uri string) error {
			ctx := context.Background()
			client, err := libvirtclient.NewClient(ctx, uri)
			if err != nil {
				return fmt.Errorf("failed to create libvirt client: %w", err)
			}
			defer func() { _ = client.Close() }()

			// List all domains
			domains, _, err := client.Libvirt().ConnectListAllDomains(1, 0)
			if err != nil {
				return fmt.Errorf("failed to list domains: %w", err)
			}

			// Delete test domains (prefix: test-)
			for _, domain := range domains {
				if strings.HasPrefix(domain.Name, "test-") || strings.HasPrefix(domain.Name, "test_") {
					// Try to destroy if running
					_ = client.Libvirt().DomainDestroy(domain)
					// Undefine the domain
					if err := client.Libvirt().DomainUndefine(domain); err != nil {
						// Log but don't fail - domain might already be gone
						fmt.Printf("Warning: failed to undefine domain %s: %v\n", domain.Name, err)
					}
				}
			}

			return nil
		},
	})
}

func testAccCheckDomainDestroy(s *terraform.State) error {
	ctx := context.Background()
	client, err := libvirtclient.NewClient(ctx, testAccLibvirtURI())
	if err != nil {
		return fmt.Errorf("failed to create libvirt client: %w", err)
	}
	defer func() { _ = client.Close() }()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "libvirt_domain" {
			continue
		}

		uuid := rs.Primary.Attributes["uuid"]
		if uuid == "" {
			continue
		}

		// Try to find the domain - it should not exist
		_, err := client.LookupDomainByUUID(uuid)
		if err == nil {
			return fmt.Errorf("domain %s still exists after destroy", uuid)
		}
	}

	return nil
}

func TestAccDomainResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDomainResourceConfigBasic("test-domain-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-basic"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "memory", "512"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "memory_unit", "MiB"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "vcpu", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "type", "kvm"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "os.type", "hvm"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "os.type_arch", "x86_64"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "os.type_machine", "q35"),
					resource.TestCheckResourceAttrSet("libvirt_domain.test", "uuid"),
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
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigUEFI("test-domain-uefi"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-uefi"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "os.firmware", "efi"),
					resource.TestCheckResourceAttrSet("libvirt_domain.test", "os.loader"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigBasic(name string) string {
	return fmt.Sprintf(`

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }
}
`, name)
}

func testAccDomainResourceConfigBasicUpdated(name string) string {
	return fmt.Sprintf(`

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 1024
  memory_unit   = "MiB"
  vcpu   = 2
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }
}
`, name)
}

func testAccDomainResourceConfigUEFI(name string) string {
	return fmt.Sprintf(`

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 1024
  memory_unit   = "MiB"
  vcpu   = 2
  type   = "kvm"

  os = {
    type        = "hvm"
    type_arch        = "x86_64"
    type_machine     = "q35"
    firmware    = "efi"
    loader = "/usr/share/edk2/x64/OVMF_CODE.fd"
  }
}
`, name)
}

func TestAccDomainResource_metadata(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigMetadata("test-domain-metadata"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-metadata"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "metadata.xml", "<app1:test xmlns:app1=\"http://test.example.com/app1/\"><app1:foo>bar</app1:foo></app1:test>"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigMetadata(name string) string {
	return fmt.Sprintf(`

resource "libvirt_domain" "test" {
  name     = %[1]q
  memory   = 512
  memory_unit     = "MiB"
  vcpu     = 1
  type     = "kvm"
  metadata = {
    xml = "<app1:test xmlns:app1=\"http://test.example.com/app1/\"><app1:foo>bar</app1:foo></app1:test>"
  }

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }
}
`, name)
}

func TestAccDomainResource_features(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigFeatures("test-domain-features"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-features"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "features.pae", "true"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "features.acpi", "true"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "features.apic.eoi", "on"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "features.hap.state", "on"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "features.pmu.state", "off"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigFeatures(name string) string {
	return fmt.Sprintf(`

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  features = {
    pae  = true
    acpi = true
    apic = {
      eoi = "on"
    }
    hap = {
      state = "on"
    }
    pmu = {
      state = "off"
    }
  }
}
`, name)
}

func TestAccDomainResource_cpu(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
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

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 2
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  cpu = {
    mode = "host-passthrough"
  }
}
`, name)
}

func TestAccDomainResource_clock(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
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

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  clock = {
    offset = "utc"
  }
}
`, name)
}

func TestAccDomainResource_lifecycle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
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

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  on_poweroff = "destroy"
  on_reboot   = "restart"
  on_crash    = "restart"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }
}
`, name)
}

func TestAccDomainResource_running(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
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
		CheckDestroy:             testAccCheckDomainDestroy,
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
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigClockTimers("test-domain-timers"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-timers"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.offset", "utc"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.#", "3"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.0.name", "rtc"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.0.tick_policy", "catchup"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.1.name", "pit"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.1.tick_policy", "delay"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.2.name", "hpet"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.2.present", "no"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigClockTimers(name string) string {
	return fmt.Sprintf(`

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  clock = {
    offset = "utc"

    timer = [
      {
        name       = "rtc"
        tick_policy = "catchup"
      },
      {
        name       = "pit"
        tick_policy = "delay"
      },
      {
        name    = "hpet"
        present = "no"
      }
    ]
  }
}
`, name)
}

func TestAccDomainResource_clockTimerCatchup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigClockTimerCatchup("test-domain-timer-catchup"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-timer-catchup"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.0.name", "rtc"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.0.tick_policy", "catchup"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.0.catch_up.threshold", "123"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.0.catch_up.slew", "120"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "clock.timer.0.catch_up.limit", "10000"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigClockTimerCatchup(name string) string {
	return fmt.Sprintf(`

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  clock = {
    offset = "utc"

    timer = [
      {
        name       = "rtc"
        tick_policy = "catchup"

        catch_up = {
          threshold = 123
          slew      = 120
          limit     = 10000
        }
      }
    ]
  }
}
`, name)
}

func testAccDomainResourceConfigRunning(name string) string {
	return fmt.Sprintf(`

resource "libvirt_domain" "test" {
  name    = %[1]q
  memory  = 512
  memory_unit    = "MiB"
  vcpu    = 1
  type    = "kvm"
  running = true

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }
}
`, name)
}

func testAccCheckDomainIsRunning(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()
		client, err := libvirtclient.NewClient(ctx, testAccLibvirtURI())
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
		client, err := libvirtclient.NewClient(ctx, testAccLibvirtURI())
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
		client, err := libvirtclient.NewClient(ctx, testAccLibvirtURI())
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
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigNetwork("test-domain-network"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-network"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.#", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.model.type", "virtio"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.source.network.network", "default"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.model.type", "virtio"),
				),
			},
		},
	})
}

// TestAccDomainResource_bridgeNetworkInterface tests using a bridge from a
// libvirt_network resource in a domain interface. According to libvirt RNG schema,
// bridge-type interfaces only support 'bridge' attribute, not 'mode'.
func TestAccDomainResource_bridgeNetworkInterface(t *testing.T) {
	config := `
resource "libvirt_network" "public_bridge" {
  name = "public_bridge"

  # Bridge mode network
  forward = {
    mode = "bridge"
  }

  # Bridge configuration
  bridge = {
    name = "virbr0"
  }
}

resource "libvirt_domain" "worker_node" {
  name   = "test-domain-bridge-network"
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    interfaces = [
      {
        model = {
          type = "virtio"
        }
        # Bridge type interface with bridge attribute from network resource
        source = {
          bridge = {
            bridge = libvirt_network.public_bridge.bridge.name
          }
        }
      }
    ]
  }
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.worker_node", "name", "test-domain-bridge-network"),
					resource.TestCheckResourceAttr("libvirt_domain.worker_node", "devices.interfaces.#", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.worker_node", "devices.interfaces.0.model.type", "virtio"),
					resource.TestCheckResourceAttr("libvirt_domain.worker_node", "devices.interfaces.0.source.bridge.bridge", "virbr0"),
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
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    interfaces = [
      {
        model = {
          type = "virtio"
        }
        source = {
          network = {
            network = "default"
          }
        }
      }
    ]
  }
}
`, name)
}

func TestAccDomainResource_directNetworkBridge(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigDirectNetworkBridge("test-domain-direct-bridge"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-direct-bridge"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.#", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.model.type", "virtio"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.source.direct.dev", "eth0"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.source.direct.mode", "bridge"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigDirectNetworkBridge(name string) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    interfaces = [
      {
        model = {
          type = "virtio"
        }
        source = {
          direct = {
            dev  = "eth0"
            mode = "bridge"
          }
        }
      }
    ]
  }
}
`, name)
}

func TestAccDomainResource_directNetworkVEPA(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigDirectNetworkVEPA("test-domain-direct-vepa"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-direct-vepa"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.#", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.model.type", "virtio"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.source.direct.dev", "eth0"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.source.direct.mode", "vepa"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigDirectNetworkVEPA(name string) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    interfaces = [
      {
        model = {
          type = "virtio"
        }
        source = {
          direct = {
            dev  = "eth0"
            mode = "vepa"
          }
        }
      }
    ]
  }
}
`, name)
}

func TestAccDomainResource_directNetworkPrivate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigDirectNetworkPrivate("test-domain-direct-private"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-direct-private"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.#", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.model.type", "virtio"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.source.direct.dev", "eth0"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.source.direct.mode", "private"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigDirectNetworkPrivate(name string) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    interfaces = [
      {
        model = {
          type = "virtio"
        }
        source = {
          direct = {
            dev  = "eth0"
            mode = "private"
          }
        }
      }
    ]
  }
}
`, name)
}

func TestAccDomainResource_directNetworkPassthrough(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigDirectNetworkPassthrough("test-direct-passthrough"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-direct-passthrough"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.#", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.model.type", "virtio"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.source.direct.dev", "eth0"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.source.direct.mode", "passthrough"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigDirectNetworkPassthrough(name string) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    interfaces = [
      {
        model = {
          type = "virtio"
        }
        source = {
          direct = {
            dev  = "eth0"
            mode = "passthrough"
          }
        }
      }
    ]
  }
}
`, name)
}

func TestAccDomainResource_directNetworkMacvtap(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigDirectNetworkMacvtap("test-direct-macvtap"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-direct-macvtap"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.#", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.model.type", "virtio"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.source.direct.dev", "eth0"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.interfaces.0.source.direct.mode", "bridge"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigDirectNetworkMacvtap(name string) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    interfaces = [
      {
        model = {
          type = "virtio"
        }
        source = {
          direct = {
            dev  = "eth0"
            mode = "bridge"
          }
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
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigGraphicsVNC("test-domain-graphics"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-graphics"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.graphics.#", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.graphics.0.vnc.auto_port", "true"),
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
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    graphics = [
      {
        vnc = {
          auto_port = true
          listen    = "0.0.0.0"
        }
      }
    ]
  }
}
`, name)
}

func TestAccDomainResource_video(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigVideo("test-domain-video"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-video"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.videos.#", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.videos.0.model.type", "vga"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.videos.0.model.primary", "yes"),
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
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    videos = [
      {
        model = {
          type    = "vga"
          primary = "yes"
          heads   = 1
          vram    = 16384
        }
      }
    ]
  }
}
`, name)
}

func TestAccDomainResource_emulator(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
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
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
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
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigConsole("test-domain-console"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-console"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.consoles.#", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.consoles.0.source.file.path", "/tmp/test-domain-console.log"),
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
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    consoles = [
      {
        source = {
          file = {
            path = "/tmp/%[1]s.log"
          }
        }
        target = {
          type = "serial"
        }
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
		CheckDestroy:             testAccCheckDomainDestroy,
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
  memory_unit      = "MiB"
  vcpu      = 1
  type      = "kvm"
  autostart = %[2]t

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
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
		client, err := libvirtclient.NewClient(ctx, testAccLibvirtURI())
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
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigMinimal("test-domain-minimal"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify fields user didn't set are NOT in state (even though libvirt defaults them)
					resource.TestCheckNoResourceAttr("libvirt_domain.test", "on_poweroff"),
					resource.TestCheckNoResourceAttr("libvirt_domain.test", "on_reboot"),
					resource.TestCheckNoResourceAttr("libvirt_domain.test", "on_crash"),
					resource.TestCheckNoResourceAttr("libvirt_domain.test", "autostart"),
					resource.TestCheckNoResourceAttr("libvirt_domain.test", "memory_unit"),
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
  type   = "kvm"

  os = {
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
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigFilesystem("test-domain-fs", sharedDir, dataDir),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-fs"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.#", "2"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.0.source.mount.dir", sharedDir),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.0.target.dir", "shared"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.0.access_mode", "mapped"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.0.read_only", "true"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.1.source.mount.dir", dataDir),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.1.target.dir", "data"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.1.access_mode", "passthrough"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.filesystems.1.read_only", "false"),
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
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    filesystems = [
      {
        source = {
          mount = {
            dir = %[2]q
          }
        }
        target = {
          dir = "shared"
        }
        access_mode = "mapped"
        read_only   = true
      },
      {
        source = {
          mount = {
            dir = %[3]q
          }
        }
        target = {
          dir = "data"
        }
        access_mode = "passthrough"
        read_only   = false
      }
    ]
  }
}
`, name, sharedDir, dataDir)
}

func TestAccDomainResource_rng(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigRNG("test-domain-rng"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-rng"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.rngs.#", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.rngs.0.model", "virtio"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.rngs.0.backend.random", "/dev/urandom"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigRNG(name string) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    rngs = [
      {
        model = "virtio"
        backend = {
          random = "/dev/urandom"
        }
      }
    ]
  }
}
`, name)
}

func TestAccDomainResource_diskWWN(t *testing.T) {
	const domainName = "test-domain-wwn"
	const volumeName = domainName + ".qcow2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigDiskWWN(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", domainName),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.disks.#", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.disks.0.source.volume.pool", domainName),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.disks.0.source.volume.volume", volumeName),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.disks.0.target.dev", "sda"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.disks.0.target.bus", "scsi"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.disks.0.wwn", "5000c50015ea71ad"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigDiskWWN(name string) string {
	return fmt.Sprintf(`
resource "libvirt_pool" "test" {
  name = %[1]q
  type = "dir"
  target = {
    path = "/tmp/terraform-provider-libvirt-pool-%[1]s"
  }
}

resource "libvirt_volume" "test" {
  name     = "%[1]s.qcow2"
  pool     = libvirt_pool.test.name
  capacity = 1073741824
  target = {
    format = {
      type = "qcow2"
    }
  }
}

resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    disks = [
      {
        source = {
          volume = {
            pool   = libvirt_pool.test.name
            volume = libvirt_volume.test.name
          }
        }
        target = {
          dev = "sda"
          bus = "scsi"
        }
        wwn = "5000c50015ea71ad"
      }
    ]
  }
}
`, name)
}

func TestAccDomainResource_diskBlockDevice(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigDiskBlockDevice("test-domain-blockdev"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-blockdev"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.disks.#", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.disks.0.source.block.dev", "/dev/null"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.disks.0.target.dev", "vda"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.disks.0.target.bus", "virtio"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigDiskBlockDevice(name string) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    disks = [
      {
        source = {
          block = {
            dev = "/dev/null"
          }
        }
        target = {
          dev = "vda"
          bus = "virtio"
        }
      }
    ]
  }
}
`, name)
}

func TestAccDomainResource_nvramTemplate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigNvramTemplate("test-domain-nvram-$(date +%s)"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "os.nv_ram.nv_ram", "/tmp/test-domain-nvram-$(date +%s).fd"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "os.nv_ram.template", "/usr/share/edk2/x64/OVMF_VARS.4m.fd"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigNvramTemplate(name string) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type        = "hvm"
    type_arch        = "x86_64"
    type_machine     = "q35"
    loader = "/usr/share/edk2/x64/OVMF_CODE.fd"
    nv_ram = {
      nv_ram   = "/tmp/%[1]s.fd"
      template = "/usr/share/edk2/x64/OVMF_VARS.4m.fd"
    }
  }
}
`, name)
}

// TPM support requires swtpm to be installed on the host
func TestAccDomainResource_tpm(t *testing.T) {
	t.Skip("TPM emulator (swtpm) not available in test environment")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigTPM("test-domain-tpm"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-tpm"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.tpms.#", "1"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.tpms.0.model", "tpm-tis"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.tpms.0.backend_type", "emulator"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigTPM(name string) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    tpms = [
      {
        model        = "tpm-tis"
        backend_type = "emulator"
      }
    ]
  }
}
`, name)
}

func TestAccDomainResource_inputs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfigInputs("test-domain-inputs"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-inputs"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.inputs.0.type", "tablet"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.inputs.0.bus", "usb"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.inputs.1.type", "keyboard"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.inputs.1.bus", "virtio"),
				),
			},
		},
	})
}

func testAccDomainResourceConfigInputs(name string) string {
	return fmt.Sprintf(`
resource "libvirt_domain" "test" {
  name   = %[1]q
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    inputs = [
      {
        type = "tablet"
        bus  = "usb"
      },
      {
        type = "keyboard"
        bus  = "virtio"
      }
    ]
  }
}
`, name)
}
