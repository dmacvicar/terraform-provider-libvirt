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
	"encoding/xml"
)

type NetworkBridge struct {
	Name            string `xml:"name,attr,omitempty"`
	STP             string `xml:"stp,attr,omitempty"`
	Delay           string `xml:"delay,attr,omitempty"`
	MACTableManager string `xml:"macTableManager,attr,omitempty"`
}

type NetworkDomain struct {
	Name      string `xml:"name,attr,omitempty"`
	LocalOnly string `xml:"localOnly,attr,omitempty"`
}

type NetworkForwardNATAddress struct {
	Start string `xml:"start,attr"`
	End   string `xml:"end,attr"`
}

type NetworkForwardNATPort struct {
	Start uint `xml:"start,attr"`
	End   uint `xml:"end,attr"`
}

type NetworkForwardNAT struct {
	Addresses []NetworkForwardNATAddress `xml:"address"`
	Ports     []NetworkForwardNATPort    `xml:"port"`
}

type NetworkForward struct {
	Mode string             `xml:"mode,attr,omitempty"`
	Dev  string             `xml:"dev,attr,omitempty"`
	NAT  *NetworkForwardNAT `xml:"nat"`
}

type NetworkMAC struct {
	Address string `xml:"address,attr,omitempty"`
}

type NetworkDHCPRange struct {
	Start string `xml:"start,attr,omitempty"`
	End   string `xml:"end,attr,omitempty"`
}

type NetworkDHCPHost struct {
	ID   string `xml:"id,attr,omitempty"`
	MAC  string `xml:"mac,attr,omitempty"`
	Name string `xml:"name,attr,omitempty"`
	IP   string `xml:"ip,attr,omitempty"`
}

type NetworkDHCP struct {
	Ranges []NetworkDHCPRange `xml:"range"`
	Hosts  []NetworkDHCPHost  `xml:"host"`
}

type NetworkIP struct {
	Address  string       `xml:"address,attr,omitempty"`
	Family   string       `xml:"family,attr,omitempty"`
	Netmask  string       `xml:"netmask,attr,omitempty"`
	Prefix   string       `xml:"prefix,attr,omitempty"`
	LocalPtr string       `xml:"localPtr,attr,omitempty"`
	DHCP     *NetworkDHCP `xml:"dhcp"`
}

type NetworkRoute struct {
	Address string `xml:"address,attr,omitempty"`
	Family  string `xml:"family,attr,omitempty"`
	Prefix  string `xml:"prefix,attr,omitempty"`
	Metric  string `xml:"metric,attr,omitempty"`
	Gateway string `xml:"gateway,attr,omitempty"`
}

type NetworkDNSForwarder struct {
	Domain string `xml:"domain,attr,omitempty"`
	Addr   string `xml:"addr,attr,omitempty"`
}

type NetworkDNSTXT struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type NetworkDNSHostHostname struct {
	Hostname string `xml:",chardata"`
}

type NetworkDNSHost struct {
	IP        string                   `xml:"ip,attr"`
	Hostnames []NetworkDNSHostHostname `xml:"hostname"`
}

type NetworkDNSSRV struct {
	Service  string `xml:"service,attr"`
	Protocol string `xml:"protocol,attr"`
	Target   string `xml:"target,attr,omitempty"`
	Port     uint   `xml:"port,attr,omitempty"`
	Priority uint   `xml:"priority,attr,omitempty"`
	Weight   uint   `xml:"weight,attr,omitempty"`
	Domain   string `xml:"domain,attr,omitempty"`
}

type NetworkDNS struct {
	Enable            string                `xml:"enable,attr,omitempty"`
	ForwardPlainNames string                `xml:"forwardPlainNames,attr,omitempty"`
	Forwarders        []NetworkDNSForwarder `xml:"forwarder"`
	TXTs              []NetworkDNSTXT       `xml:"txt"`
	Host              *NetworkDNSHost       `xml:"host"`
	SRVs              []NetworkDNSSRV       `xml:"srv"`
}

type Network struct {
	XMLName             xml.Name        `xml:"network"`
	IPv6                string          `xml:"ipv6,attr,omitempty"`
	TrustGuestRxFilters string          `xml:"trustGuestRxFilters,attr,omitempty"`
	Name                string          `xml:"name,omitempty"`
	UUID                string          `xml:"uuid,omitempty"`
	MAC                 *NetworkMAC     `xml:"mac"`
	Bridge              *NetworkBridge  `xml:"bridge"`
	Forward             *NetworkForward `xml:"forward"`
	Domain              *NetworkDomain  `xml:"domain"`
	IPs                 []NetworkIP     `xml:"ip"`
	Routes              []NetworkRoute  `xml:"route"`
	DNS                 *NetworkDNS     `xml:"dns"`
}

func (s *Network) Unmarshal(doc string) error {
	return xml.Unmarshal([]byte(doc), s)
}

func (s *Network) Marshal() (string, error) {
	doc, err := xml.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}
	return string(doc), nil
}
