package libvirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/libvirt/libvirt-go"
)

func TestNetworkAutostart(t *testing.T) {
	var network libvirt.Network
	randomNetworkName := acctest.RandString(10)
	randomDomainName := acctest.RandString(8)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network" "test_net" {
					name      = "%s"
					mode      = "nat"
					domain    = "%s"
					addresses = ["10.17.3.0/24"]
					autostart = true
				}`, randomNetworkName, randomDomainName),
				Check: resource.ComposeTestCheckFunc(
					networkExists("libvirt_network.test_net", &network),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "autostart", "true"),
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network" "test_net" {
					name      = "%s"
					mode      = "nat"
					domain    = "%s"
					addresses = ["10.17.3.0/24"]
					autostart = false
				}`, randomNetworkName, randomDomainName),
				Check: resource.ComposeTestCheckFunc(
					networkExists("libvirt_network.test_net", &network),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "autostart", "false"),
				),
			},
		},
	})
}

func networkExists(n string, network *libvirt.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No libvirt network ID is set")
		}

		virConn := testAccProvider.Meta().(*Client).libvirt

		networkRetrived, err := virConn.LookupNetworkByUUIDString(rs.Primary.ID)
		if err != nil {
			return err
		}

		realID, err := networkRetrived.GetUUIDString()
		if err != nil {
			return err
		}

		if realID != rs.Primary.ID {
			return fmt.Errorf("Libvirt network not found")
		}

		*network = *networkRetrived

		return nil
	}
}

func testAccCheckLibvirtNetworkDestroy(s *terraform.State) error {
	virtConn := testAccProvider.Meta().(*Client).libvirt

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "libvirt_network" {
			continue
		}
		_, err := virtConn.LookupNetworkByUUIDString(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf(
				"Error waiting for network (%s) to be destroyed: %s",
				rs.Primary.ID, err)
		}
	}

	return nil
}

func TestAccLibvirtNetwork_Import(t *testing.T) {
	var network libvirt.Network
	randomNetworkName := acctest.RandString(8)
	randomDomainName := acctest.RandString(8)
	fmt.Printf("RANDOMNET %s, RANDOMDOMAIN %s", randomNetworkName, randomDomainName)
	config := fmt.Sprintf(`
	resource "libvirt_network" "test_net" {
		name      = "%s"
		mode      = "nat"
		domain    = "%s"
		addresses = ["10.17.3.0/24"]
	}`, randomNetworkName, randomDomainName)

	resourceName := "libvirt_network.test_net"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: config,
			},
			resource.TestStep{
				ResourceName: resourceName,
				ImportState:  true,
				Check: resource.ComposeTestCheckFunc(
					networkExists("libvirt_network.test_net", &network),
				),
			},
		},
	})
}
