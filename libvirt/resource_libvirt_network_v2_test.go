package libvirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccLibvirtNetworkV2_Basic(t *testing.T) {
	skipIfPrivilegedDisabled(t)

	randomNetworkResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomNetworkResourceFull := "libvirt_network_v2." + randomNetworkResource
	randomNetworkName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_network_v2" "%s" {
					name = "%s"
					domain {
                      name = "k8s.local"
                      local_only = true
                    }
					ip {
					  address = "192.168.179.0"
					  netmask = "255.255.255.0"
                      dhcp {
                        range {
                          start = "192.168.179.100"
                          end = "192.168.179.150"
                        }
                        range {
                          start = "192.168.179.200"
                          end = "192.168.179.210"
                        }
                      }
					}

                    dns {
                      forwarder {
                        domain = "foo.bar"
                        addr = "154.34.77.11"
                      }
                      host {
                        ip = "1.3.4.5"
                        hostname = "bim.bam"
                      }
                      host {
                        ip = "1.3.4.6"
                        hostname = "bim.bom"
                      }
                    }

                    bridge {
                      name = "vbr-0"
                    }

					ip {
					  address = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
                      family = "ipv6"
					}

                    forward {
                      mode = "nat"
                    }
				}`, randomNetworkResource, randomNetworkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"domain.0.name", "k8s.local"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"domain.0.local_only", "true"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"dns.0.forwarder.0.domain", "foo.bar"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"dns.0.forwarder.0.addr", "154.34.77.11"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"dns.0.host.0.ip", "1.3.4.5"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"dns.0.host.0.hostname", "bim.bam"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"dns.0.host.1.ip", "1.3.4.6"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"dns.0.host.1.hostname", "bim.bom"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"ip.0.address", "192.168.179.0"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"ip.0.netmask", "255.255.255.0"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"ip.0.dhcp.0.range.0.start", "192.168.179.100"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"ip.0.dhcp.0.range.0.end", "192.168.179.150"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"ip.0.dhcp.0.range.1.start", "192.168.179.200"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"ip.0.dhcp.0.range.1.end", "192.168.179.210"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"ip.1.address", "2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"ip.1.family", "ipv6"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"forward.0.mode", "nat"),
					resource.TestCheckResourceAttr(randomNetworkResourceFull,
						"bridge.0.name", "vbr-0"),
				),
			},
		},
	})
}
