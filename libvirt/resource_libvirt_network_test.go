package libvirt

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
)

func getNetworkDef(s *terraform.State, name string) (*libvirtxml.Network, error) {
	var network *libvirt.Network
	rs, ok := s.RootModule().Resources[name]
	if !ok {
		return nil, fmt.Errorf("Not found: %s", name)
	}
	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("No libvirt network ID is set")
	}
	virConn := testAccProvider.Meta().(*Client).libvirt
	network, err := virConn.LookupNetworkByUUIDString(rs.Primary.ID)
	if err != nil {
		return nil, err
	}
	networkDef, err := getXMLNetworkDefFromLibvirt(network)
	if err != nil {
		return nil, fmt.Errorf("Error reading libvirt network XML description: %s", err)
	}
	return &networkDef, nil
}

func TestAccCheckLibvirtNetwork_LocalOnly(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network" "test_net" {
					name      = "networktest"
					domain    = "k8s.local"
					addresses = ["10.17.3.0/24"]
					dns {
						local_only = true
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network.test_net", "dns.0.local_only", "true"),
					checkLocalOnly("libvirt_network.test_net", true),
				),
			},
		},
	})
}

func checkLocalOnly(name string, expectLocalOnly bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		networkDef, err := getNetworkDef(s, name)
		if err != nil {
			return err
		}
		if expectLocalOnly {
			if networkDef.Domain == nil || networkDef.Domain.LocalOnly != "yes" {
				return fmt.Errorf("networkDef.Domain.LocalOnly is not true")
			}
		} else {
			if networkDef.Domain != nil && networkDef.Domain.LocalOnly != "no" {
				return fmt.Errorf("networkDef.Domain.LocalOnly is true")
			}
		}
		return nil
	}
}

func TestAccCheckLibvirtNetwork_DNSForwarders(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network" "test_net" {
					name      = "networktest"
					domain    = "k8s.local"
					addresses = ["10.17.3.0/24"]
					dns {
						forwarders = [
						  {
						    address = "8.8.8.8",
					          },
						  {
						    address = "10.10.0.67",
						    domain = "my.domain.com",
						  },
						  {
						    domain = "hello.com",
						  },
						]
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network.test_net", "dns.0.forwarders.#", "3"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "dns.0.forwarders.0.address", "8.8.8.8"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "dns.0.forwarders.1.address", "10.10.0.67"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "dns.0.forwarders.1.domain", "my.domain.com"),
					resource.TestCheckResourceAttr("libvirt_network.test_net", "dns.0.forwarders.2.domain", "hello.com"),
					checkDNSForwarders("libvirt_network.test_net", []libvirtxml.NetworkDNSForwarder{
						{
							Addr: "8.8.8.8",
						},
						{
							Addr:   "10.10.0.67",
							Domain: "my.domain.com",
						},
						{
							Domain: "hello.com",
						},
					}),
				),
			},
		},
	})
}

func checkDNSForwarders(name string, expected []libvirtxml.NetworkDNSForwarder) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		networkDef, err := getNetworkDef(s, name)
		if err != nil {
			return err
		}
		if networkDef.DNS == nil {
			return fmt.Errorf("DNS block not found in networkDef")
		}
		actual := networkDef.DNS.Forwarders
		if len(expected) != len(actual) {
			return fmt.Errorf("len(expected): %d != len(actual): %d", len(expected), len(actual))
		}
		for _, e := range expected {
			found := false
			for _, a := range actual {
				if reflect.DeepEqual(a, e) {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("Unable to find %v in %v", e, actual)
			}
		}
		return nil
	}
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

func testAccCheckLibvirtNetworkDhcpStatus(name string, network *libvirt.Network, expectedDhcpStatus string) resource.TestCheckFunc {
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

		networkDef, err := getXMLNetworkDefFromLibvirt(network)
		if err != nil {
			return fmt.Errorf("Error reading libvirt network XML description: %s", err)
		}
		if expectedDhcpStatus == "disabled" {
			for _, ips := range networkDef.IPs {
				// &libvirtxml.NetworkDHCP{..} should be nil when dhcp is disabled
				if ips.DHCP != nil {
					fmt.Printf("%#v", ips.DHCP)
					return fmt.Errorf("the network should have DHCP disabled")
				}
			}
		}
		if expectedDhcpStatus == "enabled" {
			for _, ips := range networkDef.IPs {
				if ips.DHCP == nil {
					return fmt.Errorf("the network should have DHCP enabled")
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

func TestAccLibvirtNetwork_DhcpEnabled(t *testing.T) {
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
					resource.TestCheckResourceAttr("libvirt_network.test_net", "dhcp.0.enabled", "true"),
					testAccCheckLibvirtNetworkDhcpStatus("libvirt_network.test_net", &network1, "enabled"),
				),
			},
		},
	})
}

func TestAccLibvirtNetwork_DhcpDisabled(t *testing.T) {
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
