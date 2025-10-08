package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetworkResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNetworkResourceConfigBasic("test-network-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network.test", "name", "test-network-basic"),
					resource.TestCheckResourceAttr("libvirt_network.test", "mode", "nat"),
					resource.TestCheckResourceAttr("libvirt_network.test", "addresses.#", "1"),
					resource.TestCheckResourceAttr("libvirt_network.test", "addresses.0", "10.17.3.0/24"),
					resource.TestCheckResourceAttrSet("libvirt_network.test", "uuid"),
					resource.TestCheckResourceAttrSet("libvirt_network.test", "id"),
					resource.TestCheckResourceAttrSet("libvirt_network.test", "bridge"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccNetworkResource_isolated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfigIsolated("test-network-isolated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network.test", "name", "test-network-isolated"),
					resource.TestCheckResourceAttr("libvirt_network.test", "mode", "none"),
					resource.TestCheckResourceAttr("libvirt_network.test", "addresses.#", "1"),
				),
			},
		},
	})
}

func testAccNetworkResourceConfigBasic(name string) string {
	return fmt.Sprintf(`
resource "libvirt_network" "test" {
  name      = %[1]q
  mode      = "nat"
  addresses = ["10.17.3.0/24"]
  autostart = false
}
`, name)
}

func testAccNetworkResourceConfigIsolated(name string) string {
	return fmt.Sprintf(`
resource "libvirt_network" "test" {
  name      = %[1]q
  mode      = "none"
  addresses = ["192.168.100.0/24"]
}
`, name)
}
