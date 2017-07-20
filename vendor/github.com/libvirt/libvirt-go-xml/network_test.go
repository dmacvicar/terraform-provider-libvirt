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

var networkTestData = []struct {
	Object   *Network
	Expected []string
}{
	{
		Object: &Network{
			Name: "test",
			IPv6: "yes",
		},
		Expected: []string{
			`<network ipv6="yes">`,
			`  <name>test</name>`,
			`</network>`,
		},
	},
	{
		Object: &Network{
			Name: "test",
			Domain: &NetworkDomain{
				Name: "example.com",
			},
		},
		Expected: []string{
			`<network>`,
			`  <name>test</name>`,
			`  <domain name="example.com"></domain>`,
			`</network>`,
		},
	},
	{
		Object: &Network{
			Name: "test",
			Bridge: &NetworkBridge{
				Name: "virbr0",
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
					},
				},
				NetworkIP{
					Family:  "ipv6",
					Address: "2001:db8:ca2:2::1",
					Prefix:  "64",
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
				Host: &NetworkDNSHost{
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
		},
		Expected: []string{
			`<network>`,
			`  <name>test</name>`,
			`  <bridge name="virbr0"></bridge>`,
			`  <forward mode="nat">`,
			`    <nat>`,
			`      <address start="1.2.3.4" end="1.2.3.10"></address>`,
			`      <port start="500" end="1000"></port>`,
			`    </nat>`,
			`  </forward>`,
			`  <ip address="192.168.122.1" netmask="255.255.255.0">`,
			`    <dhcp>`,
			`      <range start="192.168.122.2" end="192.168.122.254"></range>`,
			`      <host mac="00:16:3e:77:e2:ed" name="foo.example.com" ip="192.168.122.10"></host>`,
			`    </dhcp>`,
			`  </ip>`,
			`  <ip address="2001:db8:ca2:2::1" family="ipv6" prefix="64">`,
			`    <dhcp>`,
			`      <host name="paul" ip="2001:db8:ca2:2:3::1"></host>`,
			`      <host id="0:1:0:1:18:aa:62:fe:0:16:3e:44:55:66" ip="2001:db8:ca2:2:3::2"></host>`,
			`    </dhcp>`,
			`  </ip>`,
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
