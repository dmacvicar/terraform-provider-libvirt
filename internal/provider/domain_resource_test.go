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
