/*
 * This file is part of the libvirt-go-xml project
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 *
 * Copyright (C) 2017 Lian Duan <blazeblue@gmail.com>
 *
 */

package libvirtxml

import (
	"strings"
	"testing"
)

var nicInAverage uint = 1000
var nicInBurst uint = 10000
var nicInPeak uint = 2000
var nicInFloor uint = 500
var nicOutAverage uint = 2000

var netFwdDomain uint = 0
var netFwdBus uint = 3
var netFwdSlot uint = 0
var netFwdFunc1 uint = 1
var netFwdFunc2 uint = 2

var networkTestData = []struct {
	Object   *Network
	Expected []string
}{
	{
		Object: &Network{
			Name: "test",
			IPv6: "yes",
			Metadata: &NetworkMetadata{
				XML: "<myvalue xmlns='http://myapp.com/schemeas/my/1.0'><widget name='foo'/></myvalue>" +
					"<myothervalue xmlns='http://myotherapp.com/schemeas/my/1.0'><gizmo name='foo'/></myothervalue>",
			},
		},
		Expected: []string{
			`<network ipv6="yes">`,
			`  <name>test</name>`,
			`  <metadata>` +
				`<myvalue xmlns='http://myapp.com/schemeas/my/1.0'><widget name='foo'/></myvalue>` +
				`<myothervalue xmlns='http://myotherapp.com/schemeas/my/1.0'><gizmo name='foo'/></myothervalue>` +
				`</metadata>`,
			`</network>`,
		},
	},
	{
		Object: &Network{
			Name: "test",
			MTU: &NetworkMTU{
				Size: 1400,
			},
			Domain: &NetworkDomain{
				Name: "example.com",
			},
			Forward: &NetworkForward{
				Mode:    "hostdev",
				Managed: "yes",
				Driver: &NetworkForwardDriver{
					Name: "vfio",
				},
				Addresses: []NetworkForwardAddress{
					NetworkForwardAddress{
						PCI: &NetworkForwardAddressPCI{
							Domain:   &netFwdDomain,
							Bus:      &netFwdBus,
							Slot:     &netFwdSlot,
							Function: &netFwdFunc1,
						},
					},
					NetworkForwardAddress{
						PCI: &NetworkForwardAddressPCI{
							Domain:   &netFwdDomain,
							Bus:      &netFwdBus,
							Slot:     &netFwdSlot,
							Function: &netFwdFunc2,
						},
					},
				},
				PFs: []NetworkForwardPF{
					NetworkForwardPF{
						Dev: "eth2",
					},
				},
			},
			PortGroups: []NetworkPortGroup{
				NetworkPortGroup{
					Name:    "main",
					Default: "yes",
					VLAN: &NetworkVLAN{
						Trunk: "yes",
						Tags: []NetworkVLANTag{
							NetworkVLANTag{
								ID:         123,
								NativeMode: "tagged",
							},
							NetworkVLANTag{
								ID: 444,
							},
						},
					},
				},
			},
		},
		Expected: []string{
			`<network>`,
			`  <name>test</name>`,
			`  <forward mode="hostdev" managed="yes">`,
			`    <driver name="vfio"></driver>`,
			`    <pf dev="eth2"></pf>`,
			`    <address type="pci" domain="0x0000" bus="0x03" slot="0x00" function="0x1"></address>`,
			`    <address type="pci" domain="0x0000" bus="0x03" slot="0x00" function="0x2"></address>`,
			`  </forward>`,
			`  <mtu size="1400"></mtu>`,
			`  <domain name="example.com"></domain>`,
			`  <portgroup name="main" default="yes">`,
			`    <vlan trunk="yes">`,
			`      <tag id="123" nativeMode="tagged"></tag>`,
			`      <tag id="444"></tag>`,
			`    </vlan>`,
			`  </portgroup>`,
			`</network>`,
		},
	},
	{
		Object: &Network{
			Name: "test",
			Bridge: &NetworkBridge{
				Name: "virbr0",
			},
			VirtualPort: &NetworkVirtualPort{
				Params: &NetworkVirtualPortParams{
					OpenVSwitch: &NetworkVirtualPortParamsOpenVSwitch{
						InterfaceID: "09b11c53-8b5c-4eeb-8f00-d84eaa0aaa4f",
					},
				},
			},
			Forward: &NetworkForward{
				Mode: "nat",
				NAT: &NetworkForwardNAT{
					Addresses: []NetworkForwardNATAddress{
						NetworkForwardNATAddress{
							Start: "1.2.3.4",
							End:   "1.2.3.10",
						},
					},
					Ports: []NetworkForwardNATPort{
						NetworkForwardNATPort{
							Start: 500,
							End:   1000,
						},
					},
				},
				Interfaces: []NetworkForwardInterface{
					NetworkForwardInterface{
						Dev: "eth0",
					},
				},
			},
			IPs: []NetworkIP{
				NetworkIP{
					Address: "192.168.122.1",
					Netmask: "255.255.255.0",
					DHCP: &NetworkDHCP{
						Ranges: []NetworkDHCPRange{
							NetworkDHCPRange{
								Start: "192.168.122.2",
								End:   "192.168.122.254",
							},
						},
						Hosts: []NetworkDHCPHost{
							NetworkDHCPHost{
								MAC:  "00:16:3e:77:e2:ed",
								Name: "foo.example.com",
								IP:   "192.168.122.10",
							},
						},
						Bootp: []NetworkBootp{
							NetworkBootp{
								File:   "pxelinux.0",
								Server: "192.168.122.1",
							},
						},
					},
					TFTP: &NetworkTFTP{
						Root: "/var/lib/tftp",
					},
				},
				NetworkIP{
					Family:  "ipv6",
					Address: "2001:db8:ca2:2::1",
					Prefix:  64,
					DHCP: &NetworkDHCP{
						Hosts: []NetworkDHCPHost{
							NetworkDHCPHost{
								IP:   "2001:db8:ca2:2:3::1",
								Name: "paul",
							},
							NetworkDHCPHost{
								ID: "0:1:0:1:18:aa:62:fe:0:16:3e:44:55:66",
								IP: "2001:db8:ca2:2:3::2",
							},
						},
					},
				},
			},
			DNS: &NetworkDNS{
				Enable:            "yes",
				ForwardPlainNames: "no",
				Forwarders: []NetworkDNSForwarder{
					NetworkDNSForwarder{
						Addr: "8.8.8.8",
					},
					NetworkDNSForwarder{
						Domain: "example.com",
						Addr:   "8.8.4.4",
					},
					NetworkDNSForwarder{
						Domain: "www.example.com",
					},
				},
				TXTs: []NetworkDNSTXT{
					NetworkDNSTXT{
						Name:  "example",
						Value: "example value",
					},
				},
				Host: []NetworkDNSHost{
					NetworkDNSHost{
						IP: "192.168.122.2",
						Hostnames: []NetworkDNSHostHostname{
							NetworkDNSHostHostname{
								Hostname: "myhost",
							},
							NetworkDNSHostHostname{
								Hostname: "myhostalias",
							},
						},
					},
				},
				SRVs: []NetworkDNSSRV{
					NetworkDNSSRV{
						Service:  "name",
						Protocol: "tcp",
						Domain:   "test-domain-name",
						Target:   ".",
						Port:     1024,
						Priority: 10,
						Weight:   10,
					},
				},
			},
			Bandwidth: &NetworkBandwidth{
				Inbound: &NetworkBandwidthParams{
					Average: &nicInAverage,
					Peak:    &nicInPeak,
					Burst:   &nicInBurst,
					Floor:   &nicInFloor,
				},
				Outbound: &NetworkBandwidthParams{
					Average: &nicOutAverage,
				},
			},
			Routes: []NetworkRoute{
				NetworkRoute{
					Address: "192.168.222.0",
					Netmask: "255.255.255.0",
					Gateway: "192.168.122.10",
				},
				NetworkRoute{
					Family:  "ipv6",
					Address: "2001:db8:ac10:fc00::",
					Prefix:  64,
					Gateway: "2001:db8:ac10:fd01::1:24",
				},
			},
			VLAN: &NetworkVLAN{
				Tags: []NetworkVLANTag{
					NetworkVLANTag{
						ID: 49,
					},
					NetworkVLANTag{
						ID:         52,
						NativeMode: "tagged",
					},
				},
			},
		},
		Expected: []string{
			`<network>`,
			`  <name>test</name>`,
			`  <forward mode="nat">`,
			`    <nat>`,
			`      <address start="1.2.3.4" end="1.2.3.10"></address>`,
			`      <port start="500" end="1000"></port>`,
			`    </nat>`,
			`    <interface dev="eth0"></interface>`,
			`  </forward>`,
			`  <bridge name="virbr0"></bridge>`,
			`  <dns enable="yes" forwardPlainNames="no">`,
			`    <forwarder addr="8.8.8.8"></forwarder>`,
			`    <forwarder domain="example.com" addr="8.8.4.4"></forwarder>`,
			`    <forwarder domain="www.example.com"></forwarder>`,
			`    <txt name="example" value="example value"></txt>`,
			`    <host ip="192.168.122.2">`,
			`      <hostname>myhost</hostname>`,
			`      <hostname>myhostalias</hostname>`,
			`    </host>`,
			`    <srv service="name" protocol="tcp" target="." port="1024" priority="10" weight="10" domain="test-domain-name"></srv>`,
			`  </dns>`,
			`  <vlan>`,
			`    <tag id="49"></tag>`,
			`    <tag id="52" nativeMode="tagged"></tag>`,
			`  </vlan>`,
			`  <bandwidth>`,
			`    <inbound average="1000" peak="2000" burst="10000" floor="500"></inbound>`,
			`    <outbound average="2000"></outbound>`,
			`  </bandwidth>`,
			`  <ip address="192.168.122.1" netmask="255.255.255.0">`,
			`    <dhcp>`,
			`      <range start="192.168.122.2" end="192.168.122.254"></range>`,
			`      <host mac="00:16:3e:77:e2:ed" name="foo.example.com" ip="192.168.122.10"></host>`,
			`      <bootp file="pxelinux.0" server="192.168.122.1"></bootp>`,
			`    </dhcp>`,
			`    <tftp root="/var/lib/tftp"></tftp>`,
			`  </ip>`,
			`  <ip address="2001:db8:ca2:2::1" family="ipv6" prefix="64">`,
			`    <dhcp>`,
			`      <host name="paul" ip="2001:db8:ca2:2:3::1"></host>`,
			`      <host id="0:1:0:1:18:aa:62:fe:0:16:3e:44:55:66" ip="2001:db8:ca2:2:3::2"></host>`,
			`    </dhcp>`,
			`  </ip>`,
			`  <route address="192.168.222.0" netmask="255.255.255.0" gateway="192.168.122.10"></route>`,
			`  <route family="ipv6" address="2001:db8:ac10:fc00::" prefix="64" gateway="2001:db8:ac10:fd01::1:24"></route>`,
			`  <virtualport type="openvswitch">`,
			`    <parameters interfaceid="09b11c53-8b5c-4eeb-8f00-d84eaa0aaa4f"></parameters>`,
			`  </virtualport>`,
			`</network>`,
		},
	},
}

func TestNetwork(t *testing.T) {
	for _, test := range networkTestData {
		doc, err := test.Object.Marshal()
		if err != nil {
			t.Fatal(err)
		}

		expect := strings.Join(test.Expected, "\n")

		if doc != expect {
			t.Fatal("Bad xml:\n", string(doc), "\n does not match\n", expect, "\n")
		}
	}
}
