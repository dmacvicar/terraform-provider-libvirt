package libvirt

import (
	"encoding/xml"
	"fmt"
	libvirt "github.com/dmacvicar/libvirt-go"
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
	Name string `xml:"name,omitempty"`
}

type defNetworkDomain struct {
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
	Host []*struct {
		Ip       string   `xml:"ip,attr"`
		HostName []string `xml:"hostname"`
	} `xml:"host,omitempty"`
	Forwarder []*struct {
		Address string `xml:"addr,attr"`
	} `xml:"forwarder,omitempty"`
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

type NetworkHost struct {
	XMLName xml.Name `xml:"host"`
	Mac     string   `xml:"mac,attr,omitempty"`
	Name    string   `xml:"name,attr,omitempty"`
	IP      string   `xml:"ip,attr,omitempty"`
}

func getHostXMLDesc(ip, mac, name string) (string, error) {
	host := NetworkHost{
		Mac:  mac,
		Name: name,
		IP:   ip,
	}

	b, err := xml.Marshal(host)
	if err != nil {
		var virErr libvirt.VirError
		virErr.Code = libvirt.VIR_ERR_XML_ERROR
		virErr.Message = fmt.Sprintf("Invalid host entry definition: %s", err)

		return "", virErr
	}
	return string(b), nil
}

// Check if the network has a DHCP server managed by libvirt
func (net defNetwork) HasDHCP() bool {
	if net.Forward != nil {
		if net.Forward.Mode == "nat" {
			return true // TODO: all other modes
		}
	}
	return false
}

// Creates a network definition from a XML
func newNetworkDefFromXML(s string) (defNetwork, error) {
	var networkDef defNetwork
	err := xml.Unmarshal([]byte(s), &networkDef)
	if err != nil {
		return defNetwork{}, err
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
	if d, err := newNetworkDefFromXML(defNetworkXML); err != nil {
		panic(fmt.Sprint("Unexpected error while parsing default network definition: %s", err))
	} else {
		return d
	}
}
