package libvirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/libvirt/libvirt-go"
)

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

func testAccCheckLibvirtNetworkDhcpStatus(name string, network *libvirt.Network, DhcpStatus string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No libvirt network ID is set")
		}

		virConn := testAccProvider.Meta().(*Client).libvirt

		network, err := virConn.LookupNetworkByUUIDString(rs.Primary.ID)
		if err != nil {
			return err
		}

		networkDef, err := newDefNetworkfromLibvirt(network)
		if err != nil {
			return fmt.Errorf("Error reading libvirt network XML description: %s", err)
		}
		if DhcpStatus == "disabled" {
			for _, ips := range networkDef.IPs {
				fmt.Printf("%#v", ips.DHCP)
				// &libvirtxml.NetworkDHCP{..} should be nil when dhcp is disabled
				if ips.DHCP != nil {
					return fmt.Errorf("the network should have DHCP disabled")
				}
			}
		}
		if DhcpStatus == "enabled" {
			for _, ips := range networkDef.IPs {
				fmt.Printf("%#v", ips.DHCP)
				if ips.DHCP == nil {
					return fmt.Errorf("the network should have DHCP disabled")
				}
			}
		}
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

	const config = `
	resource "libvirt_network" "test_net" {
		name      = "networktest"
		mode      = "nat"
		domain    = "k8s.local"
		addresses = ["10.17.3.0/24"]
	}`

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

func TestAccLibvirtNetwork_dhcpEnabled(t *testing.T) {
	var network1 libvirt.Network
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{
			{
				// dhcp is enabled true by default.
				Config: fmt.Sprintf(`
				resource "libvirt_network" "test_net" {
					name      = "networktest"
					mode      = "nat"
					domain    = "k8s.local"
					addresses = ["10.17.3.0/24"]
					dhcp {
						enabled = true
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtNetworkDhcpStatus("libvirt_network.test_net", &network1, "enabled"),
				),
			},
		},
	})
}

func TestAccLibvirtNetwork_dhcpDisabled(t *testing.T) {
	var network1 libvirt.Network
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network" "test_net" {
					name      = "networktest"
					mode      = "nat"
					domain    = "k8s.local"
					addresses = ["10.17.3.0/24"]
					dhcp {
						enabled = false
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network.test_net", "dhcp.0.enabled", "false"),
					testAccCheckLibvirtNetworkDhcpStatus("libvirt_network.test_net", &network1, "disabled"),
				),
			},
		},
	})
}
func TestAccLibvirtNetwork_Autostart(t *testing.T) {
	var network libvirt.Network
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network" "test_net" {
					name      = "networktest"
					mode      = "nat"
					domain    = "k8s.local"
					addresses = ["10.17.3.0/24"]
					autostart = true
				}`),
				Check: resource.ComposeTestCheckFunc(
					networkExists("libvirt_network.test_net", &network),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "autostart", "true"),
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network" "test_net" {
					name      = "networktest"
					mode      = "nat"
					domain    = "k8s.local"
					addresses = ["10.17.3.0/24"]
					autostart = false
				}`),
				Check: resource.ComposeTestCheckFunc(
					networkExists("libvirt_network.test_net", &network),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "autostart", "false"),
				),
			},
		},
	})
}
