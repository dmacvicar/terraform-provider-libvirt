package libvirt

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
)

func TestAccCheckLibvirtNetwork_LocalOnly(t *testing.T) {
	randomNetworkResource := acctest.RandString(10)
	randomNetworkName := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network" "%s" {
					name      = "%s"
					domain    = "k8s.local"
					addresses = ["10.17.3.0/24"]
					dns {
						local_only = true
					}
				}`, randomNetworkResource, randomNetworkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.local_only", "true"),
					checkLocalOnly("libvirt_network."+randomNetworkResource, true),
				),
			},
		},
	})
}

func checkLocalOnly(name string, expectLocalOnly bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		virConn := testAccProvider.Meta().(*Client).libvirt

		networkDef, err := getNetworkDef(s, name, *virConn)
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
	randomNetworkResource := acctest.RandString(10)
	randomNetworkName := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network" "%s" {
					name      = "%s"
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
				}`, randomNetworkResource, randomNetworkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.forwarders.#", "3"),
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.forwarders.0.address", "8.8.8.8"),
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.forwarders.1.address", "10.10.0.67"),
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.forwarders.1.domain", "my.domain.com"),
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.forwarders.2.domain", "hello.com"),
					checkDNSForwarders("libvirt_network."+randomNetworkResource, []libvirtxml.NetworkDNSForwarder{
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

		virConn := testAccProvider.Meta().(*Client).libvirt

		networkDef, err := getNetworkDef(s, name, *virConn)
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

func TestAccLibvirtNetwork_DNSHosts(t *testing.T) {
	randomNetworkResource := acctest.RandString(10)
	randomNetworkName := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network" "%s" {
					name      = "%s"
					domain    = "k8s.local"
					addresses = ["10.17.3.0/24"]
					dns {
						hosts = [
						  {
							  hostname = "myhost1",
							  ip = "1.1.1.1",
						  },
						  {
							  hostname = "myhost1",
							  ip = "1.1.1.2",
						  },
						  {
							  hostname = "myhost2",
							  ip = "1.1.1.1",
						  },
						]
					}
				}`, randomNetworkResource, randomNetworkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.hosts.0.hostname", "myhost1"),
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.hosts.0.ip", "1.1.1.1"),
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.hosts.1.hostname", "myhost1"),
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.hosts.1.ip", "1.1.1.2"),
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.hosts.2.hostname", "myhost2"),
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.hosts.2.ip", "1.1.1.1"),
					checkDNSHosts("libvirt_network."+randomNetworkResource, []libvirtxml.NetworkDNSHost{
						{
							IP: "1.1.1.1",
							Hostnames: []libvirtxml.NetworkDNSHostHostname{
								{Hostname: "myhost1"},
								{Hostname: "myhost2"},
							},
						},
						{
							IP: "1.1.1.2",
							Hostnames: []libvirtxml.NetworkDNSHostHostname{
								{Hostname: "myhost1"},
							},
						},
					}),
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network" "%s" {
					name      = "%s"
					domain    = "k8s.local"
					addresses = ["10.17.3.0/24"]
					dns {
						hosts = [
						  {
							  hostname = "myhost1",
							  ip = "1.1.1.1",
						  },
						]
					}
				}`, randomNetworkResource, randomNetworkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.hosts.0.hostname", "myhost1"),
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.hosts.0.ip", "1.1.1.1"),
					checkDNSHosts("libvirt_network."+randomNetworkResource, []libvirtxml.NetworkDNSHost{
						{
							IP: "1.1.1.1",
							Hostnames: []libvirtxml.NetworkDNSHostHostname{
								{Hostname: "myhost1"},
							},
						},
					}),
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network" "%s" {
					name      = "%s"
					domain    = "k8s.local"
					addresses = ["10.17.3.0/24"]
					dns {
						hosts = [
						  {
							  hostname = "myhost1",
							  ip = "1.1.1.1",
						  },
# Without https:#www.redhat.com/archives/libvir-list/2018-November/msg00231.html, this raises:
#
#   update DNS hosts: add {{ } 1.1.1.2 [{myhost1}]}: virError(Code=55, Domain=19, Message='Requested operation is not valid: there is already at least one DNS HOST record with a matching field in network fo64d9y6w9')
#						  {
#							  hostname = "myhost1",
#							  ip = "1.1.1.2",
#						  },
						  {
							  hostname = "myhost2",
							  ip = "1.1.1.1",
						  },
						]
					}
				}`, randomNetworkResource, randomNetworkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.hosts.0.hostname", "myhost1"),
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.hosts.0.ip", "1.1.1.1"),
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.hosts.1.hostname", "myhost2"),
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dns.0.hosts.1.ip", "1.1.1.1"),
					checkDNSHosts("libvirt_network."+randomNetworkResource, []libvirtxml.NetworkDNSHost{
						{
							IP: "1.1.1.1",
							Hostnames: []libvirtxml.NetworkDNSHostHostname{
								{Hostname: "myhost1"},
								{Hostname: "myhost2"},
							},
						},
					}),
				),
			},
		},
	})
}

func checkDNSHosts(name string, expected []libvirtxml.NetworkDNSHost) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		virConn := testAccProvider.Meta().(*Client).libvirt
		networkDef, err := getNetworkDef(s, name, *virConn)
		if err != nil {
			return err
		}
		if networkDef.DNS == nil {
			return fmt.Errorf("DNS block not found in networkDef")
		}
		actual := networkDef.DNS.Host
		if len(expected) != len(actual) {
			return fmt.Errorf("len(expected): %d != len(actual): %d", len(expected), len(actual))
		}
		for _, e := range expected {
			found := false
			for _, a := range actual {
				if reflect.DeepEqual(a.IP, e.IP) && reflect.DeepEqual(a.Hostnames, e.Hostnames) {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("Unable to find:%v in: %v", e, actual)
			}
		}
		return nil
	}
}

func checkDHCPHosts(name string, expected []libvirtxml.NetworkDHCPHost) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt
		networkDef, err := getNetworkDef(s, name, *virConn)
		if err != nil {
			return err
		}
		if networkDef.IPs == nil {
			return fmt.Errorf("IP block not found in networkDef")
		}
		for _, ips := range networkDef.IPs {
			if ips.DHCP == nil {
				return fmt.Errorf("DHCP block not found in networkDef")
			}
			if ips.DHCP.Hosts == nil {
				return fmt.Errorf("DHCP Hosts block not found in networkDef DHCP")
			}
			actual := ips.DHCP.Hosts
			if len(expected) != len(actual) {
				return fmt.Errorf("len(expected): %d != len(actual): %d", len(expected), len(actual))
			}
			for _, e := range expected {
				found := false
				for _, a := range actual {
					if reflect.DeepEqual(a.IP, e.IP) && reflect.DeepEqual(a.Name, e.Name) && reflect.DeepEqual(a.MAC, e.MAC) {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("Unable to find:%v in: %v", e, actual)
				}
			}
		}
		return nil
	}
}

func networkExists(name string, network *libvirt.Network) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		rs, err := getResourceFromTerraformState(name, state)
		if err != nil {
			return err
		}

		virConn := testAccProvider.Meta().(*Client).libvirt
		fmt.Printf("%s", virConn)
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

func testAccCheckLibvirtNetworkDhcpStatus(name string, expectedDhcpStatus string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt
		networkDef, err := getNetworkDef(s, name, *virConn)
		if err != nil {
			return err
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

func TestAccLibvirtNetwork_Import(t *testing.T) {
	var network libvirt.Network
	randomNetworkResource := acctest.RandString(10)
	randomNetworkName := acctest.RandString(10)
	resourceName := "libvirt_network." + randomNetworkResource

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network" "%s" {
					name      = "%s"
					mode      = "nat"
					domain    = "k8s.local"
					addresses = ["10.17.3.0/24"]
				}`, randomNetworkResource, randomNetworkName),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				Check: resource.ComposeTestCheckFunc(
					networkExists("libvirt_network."+randomNetworkResource, &network),
				),
			},
		},
	})
}

func TestAccLibvirtNetwork_DhcpEnabled(t *testing.T) {
	randomNetworkResource := acctest.RandString(10)
	randomNetworkName := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network" "%s" {
					name      = "%s"
					mode      = "nat"
					domain    = "k8s.local"
					addresses = ["10.17.3.0/24"]
					dhcp {
						enabled = true
						hosts = [
							{ name = "host1", ip = "10.17.3.1", mac = "00:11:22:33:44:55" },
							{ ip = "10.17.3.2", mac = "00:11:22:33:44:56" }
						]
					}
				}`, randomNetworkResource, randomNetworkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dhcp.0.enabled", "true"),
					testAccCheckLibvirtNetworkDhcpStatus("libvirt_network."+randomNetworkResource, "enabled"),
					checkDHCPHosts("libvirt_network."+randomNetworkResource, []libvirtxml.NetworkDHCPHost{
						{
							IP:   "10.17.3.1",
							Name: "host1",
							MAC:  "00:11:22:33:44:55",
						},
						{
							IP:  "10.17.3.2",
							MAC: "00:11:22:33:44:56",
						},
					}),
				),
			},
		},
	})
}

func TestAccLibvirtNetwork_DhcpDisabled(t *testing.T) {
	randomNetworkResource := acctest.RandString(10)
	randomNetworkName := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network" "%s" {
					name      = "%s"
					mode      = "nat"
					domain    = "k8s.local"
					addresses = ["10.17.3.0/24"]
					dhcp {
						enabled = false
					}
				}`, randomNetworkResource, randomNetworkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dhcp.0.enabled", "false"),
					testAccCheckLibvirtNetworkDhcpStatus("libvirt_network."+randomNetworkResource, "disabled"),
				),
			},
		},
	})
}

func checkBridge(resourceName string, bridgeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt
		networkDef, err := getNetworkDef(s, resourceName, *virConn)
		if err != nil {
			return err
		}

		if networkDef.Bridge == nil {
			return fmt.Errorf("Bridge type of network should be not nil")
		}

		if networkDef.Bridge.Name != bridgeName || networkDef.Bridge.STP != "on" {
			fmt.Printf("%#v", networkDef)
			return fmt.Errorf("fail: network brigde property were not set correctly")
		}

		return nil
	}
}

func TestAccLibvirtNetwork_BridgedMode(t *testing.T) {
	randomNetworkName := acctest.RandString(10)
	randomBridgeName := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "libvirt_network" "%s" {
	  				name        = "%s"
	  				mode        = "bridge"
	  			  bridge      = "vbr-%s"
	     	}`, randomNetworkName, randomNetworkName, randomBridgeName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkName, "mode", "bridge"),
					checkBridge("libvirt_network."+randomNetworkName, "vbr-"+randomBridgeName),
				),
			},
		},
	})
}

func TestAccLibvirtNetwork_Autostart(t *testing.T) {
	var network libvirt.Network
	randomNetworkResource := acctest.RandString(10)
	randomNetworkName := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network" "%s" {
					name      = "%s"
					mode      = "nat"
					domain    = "k8s.local"
					addresses = ["10.17.3.0/24"]
					autostart = true
				}`, randomNetworkResource, randomNetworkName),
				Check: resource.ComposeTestCheckFunc(
					networkExists("libvirt_network."+randomNetworkResource, &network),
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "autostart", "true"),
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network" "%s" {
					name      = "%s"
					mode      = "nat"
					domain    = "k8s.local"
					addresses = ["10.17.3.0/24"]
					autostart = false
				}`, randomNetworkResource, randomNetworkName),
				Check: resource.ComposeTestCheckFunc(
					networkExists("libvirt_network."+randomNetworkResource, &network),
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "autostart", "false"),
				),
			},
		},
	})
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
