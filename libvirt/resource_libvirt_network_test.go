package libvirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
)

func TestAccLibvirtNetwork_Addresses(t *testing.T) {
	randomNetworkResource := acctest.RandString(10)
	randomNetworkResourceFull := "libvirt_network." + randomNetworkResource
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
				}`, randomNetworkResource, randomNetworkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"addresses.0", "10.17.3.0/24"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"mode", "nat"),
				),
			},
		},
	})
}

func TestAccLibvirtNetwork_LocalOnly(t *testing.T) {
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
					testAccCheckLibvirtNetworkLocalOnly("libvirt_network."+randomNetworkResource, true),
				),
			},
		},
	})
}

func TestAccLibvirtNetwork_DNSForwarders(t *testing.T) {
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
					testAccCheckLibvirtNetworkDNSForwarders("libvirt_network."+randomNetworkResource, []libvirtxml.NetworkDNSForwarder{
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
					testAccCheckDNSHosts("libvirt_network."+randomNetworkResource, []libvirtxml.NetworkDNSHost{
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
					testAccCheckDNSHosts("libvirt_network."+randomNetworkResource, []libvirtxml.NetworkDNSHost{
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
					testAccCheckDNSHosts("libvirt_network."+randomNetworkResource, []libvirtxml.NetworkDNSHost{
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
					testAccCheckNetworkExists("libvirt_network."+randomNetworkResource, &network),
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
					}
				}`, randomNetworkResource, randomNetworkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "dhcp.0.enabled", "true"),
					testAccCheckLibvirtNetworkDhcpStatus("libvirt_network."+randomNetworkResource, "enabled"),
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
					testAccCheckLibvirtNetworkBridge("libvirt_network."+randomNetworkName, "vbr-"+randomBridgeName),
				),
			},
		},
	})
}

func TestAccLibvirtNetwork_StaticRoutes(t *testing.T) {

	checkRoutes := func(resourceName string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			virConn := testAccProvider.Meta().(*Client).libvirt
			networkDef, err := getNetworkDef(s, resourceName, *virConn)
			if err != nil {
				return err
			}

			if len(networkDef.Routes) != 1 {
				return fmt.Errorf("Network should have one route but it has %d", len(networkDef.Routes))
			}

			if networkDef.Routes[0].Address != "10.18.0.0" {
				return fmt.Errorf("Unexpected network address '%s'", networkDef.Routes[0].Address)
			}

			if !(networkDef.Routes[0].Family == "" || networkDef.Routes[0].Family == "ipv6") {
				return fmt.Errorf("Unexpected network family '%s'", networkDef.Routes[0].Family)
			}

			if networkDef.Routes[0].Prefix != 16 {
				return fmt.Errorf("Unexpected network prefix '%d'", networkDef.Routes[0].Prefix)
			}

			if networkDef.Routes[0].Gateway != "10.17.3.2" {
				return fmt.Errorf("Unexpected gateway '%s'", networkDef.Routes[0].Gateway)
			}

			return nil
		}
	}

	randomNetworkName := acctest.RandString(10)
	config := fmt.Sprintf(`
					resource "libvirt_network" "%s" {
					name      = "%s"
					mode      = "route"
					domain    = "k8s.local"
					addresses = ["10.17.3.0/24"]
					dhcp {
						enabled = false
					}
					routes = [
						{
							cidr = "10.18.0.0/16"
							gateway = "10.17.3.2"
						},
					]}`,
		randomNetworkName, randomNetworkName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					checkRoutes("libvirt_network." + randomNetworkName),
				),
			},
			// when we apply 2 times with same conf, we should not have a diff
			{
				Config:             config,
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
				Check: resource.ComposeTestCheckFunc(
					checkRoutes("libvirt_network." + randomNetworkName),
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
					testAccCheckNetworkExists("libvirt_network."+randomNetworkResource, &network),
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
					testAccCheckNetworkExists("libvirt_network."+randomNetworkResource, &network),
					resource.TestCheckResourceAttr("libvirt_network."+randomNetworkResource, "autostart", "false"),
				),
			},
		},
	})
}
