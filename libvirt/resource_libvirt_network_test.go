package libvirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/libvirt/libvirt-go"
)

func TestNetworkAutostart(t *testing.T) {
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

func TestNetworkDNS(t *testing.T) {
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
					dns {
						enabled = true
						local_only = true

						forwarders = [
							"8.8.8.8",
							"my.domain.com -> 10.10.0.67",
							"hello.com"
						]

						host {
							address = "10.17.3.2"
							name = ["server1.com", "server2.com"]
						}
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network.test_net", "dns.0.enabled", "true"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "dns.0.local_only", "true"),

					resource.TestCheckResourceAttr("libvirt_network.test_net", "dns.0.forwarders.#", "3"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "dns.0.forwarders.0", "8.8.8.8"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "dns.0.forwarders.1", "my.domain.com -> 10.10.0.67"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "dns.0.forwarders.2", "hello.com"),

					resource.TestCheckResourceAttr("libvirt_network.test_net", "dns.0.host.0.address", "10.17.3.2"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "dns.0.host.0.name.#", "2"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "dns.0.host.0.name.0", "server1.com"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "dns.0.host.0.name.1", "server2.com"),
				),
			},
		},
	})
}

func TestNetworkDHCP(t *testing.T) {
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
				),
			},
		},
	})
}

func TestNetworkStaticRoutes(t *testing.T) {
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
					routes = [
						"192.168.7.0/24 -> 127.0.0.1",
						"192.168.9.1/24 -> 127.0.0.1",
						"192.168.17.1/32 -> 127.0.0.1",
						"2001:db9:4:1::/64 -> ::1"
					]
				}`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network.test_net", "route.0.address", "192.168.7.0"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "route.0.prefix", "24"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "route.0.family", "ipv4"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "route.0.gateway", "127.0.0.1"),

					resource.TestCheckResourceAttr("libvirt_network.test_net", "route.1.address", "192.168.9.0"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "route.1.prefix", "24"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "route.1.family", "ipv4"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "route.1.gateway", "127.0.0.1"),

					resource.TestCheckResourceAttr("libvirt_network.test_net", "route.2.address", "192.168.17.1"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "route.2.prefix", "32"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "route.2.family", "ipv4"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "route.2.gateway", "127.0.0.1"),
				),
			},
		},
	})
}

/*************************
 * tests helpers
 ************************/

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
