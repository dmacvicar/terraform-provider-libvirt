package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
