package libvirt

import (
	"encoding/xml"
	"fmt"
)

type defNetworkIpDhcpRange struct {
	XMLName xml.Name `xml:"range,omitempty"`

	Start string `xml:"start,attr,omitempty"`
	End   string `xml:"end,attr,omitempty"`
}

type defNetworkIpDhcpHost struct {
	XMLName xml.Name `xml:"host,omitempty"`

	Ip   string `xml:"ip,attr,omitempty"`
	Mac  string `xml:"mac,attr,omitempty"`
	Name string `xml:"name,attr,omitempty"`
}

type defNetworkIpDhcp struct {
	XMLName xml.Name `xml:"dhcp,omitempty"`

	Ranges []*defNetworkIpDhcpRange `xml:"range,omitempty"`
	Hosts  []*defNetworkIpDhcpHost  `xml:"host,omitempty"`
}

type defNetworkIp struct {
	XMLName xml.Name `xml:"ip,omitempty"`

	Address string            `xml:"address,attr"`
	Netmask string            `xml:"netmask,attr,omitempty"`
	Prefix  int               `xml:"prefix,attr,omitempty"`
	Family  string            `xml:"family,attr,omitempty"`
	Dhcp    *defNetworkIpDhcp `xml:"dhcp,omitempty"`
}

type defNetworkBridge struct {
	XMLName xml.Name `xml:"bridge,omitempty"`

	Name string `xml:"name,attr,omitempty"`
	Stp  string `xml:"stp,attr,omitempty"`
}

type defNetworkDomain struct {
	XMLName xml.Name `xml:"domain,omitempty"`

	Name      string `xml:"name,attr,omitempty"`
	LocalOnly string `xml:"localOnly,attr,omitempty"`
}

type defNetworkForward struct {
	Mode   string `xml:"mode,attr"`
	Device string `xml:"dev,attr,omitempty"`
	Nat    *struct {
		Addresses []*struct {
			Start string `xml:"start,attr"`
			End   string `xml:"end,attr"`
		} `xml:"address,omitempty"`
		Ports []*struct {
			Start string `xml:"start,attr"`
			End   string `xml:"end,attr"`
		} `xml:"port,omitempty"`
	} `xml:"nat,omitempty"`
}

type defNetworkDns struct {
	Host      []*defDnsHost      `xml:"host,omitempty"`
	Forwarder []*defDnsForwarder `xml:"forwarder,omitempty"`
}

type defDnsHost struct {
	Ip       string   `xml:"ip,attr"`
	HostName []string `xml:"hostname"`
}

type defDnsForwarder struct {
	Domain  string `xml:"domain,attr,omitempty"`
	Address string `xml:"addr,attr,omitempty"`
}

// network definition in XML, compatible with what libvirt expects
// note: we have to use pointers or otherwise golang's XML will not properly detect
//       empty values and generate things like "<bridge></bridge>" that
//       make libvirt crazy...
type defNetwork struct {
	XMLName xml.Name `xml:"network"`

	Name    string             `xml:"name,omitempty"`
	Domain  *defNetworkDomain  `xml:"domain,omitempty"`
	Bridge  *defNetworkBridge  `xml:"bridge,omitempty"`
	Forward *defNetworkForward `xml:"forward,omitempty"`
	Ips     []*defNetworkIp    `xml:"ip,omitempty"`
	Dns     *defNetworkDns     `xml:"dns,omitempty"`
}

// Check if the network has a DHCP server managed by libvirt
func (net defNetwork) HasDHCP() bool {
	if net.Forward != nil {
		if net.Forward.Mode == "nat" || net.Forward.Mode == "route" || net.Forward.Mode == "" {
			return true
		}
	}
	return false
}

// Creates a network definition from a XML
func newDefNetworkFromXML(s string) (defNetwork, error) {
	var networkDef defNetwork
	err := xml.Unmarshal([]byte(s), &networkDef)
	if err != nil {
		return defNetwork{}, err
	}
	return networkDef, nil
}

func newDefNetworkfromLibvirt(network LibVirtNetwork) (defNetwork, error) {
	networkXmlDesc, err := network.GetXMLDesc(0)
	if err != nil {
		return defNetwork{}, fmt.Errorf("Error retrieving libvirt domain XML description: %s", err)
	}
	networkDef := defNetwork{}
	err = xml.Unmarshal([]byte(networkXmlDesc), &networkDef)
	if err != nil {
		return defNetwork{}, fmt.Errorf("Error reading libvirt network XML description: %s", err)
	}
	return networkDef, nil
}

// Creates a network definition with the defaults the provider uses
func newNetworkDef() defNetwork {
	const defNetworkXML = `
		<network>
		  <name>default</name>
		  <forward mode='nat'>
		    <nat>
		      <port start='1024' end='65535'/>
		    </nat>
		  </forward>
		</network>`
	if d, err := newDefNetworkFromXML(defNetworkXML); err != nil {
		panic(fmt.Sprint("Unexpected error while parsing default network definition: %s", err))
	} else {
		return d
	}
}
