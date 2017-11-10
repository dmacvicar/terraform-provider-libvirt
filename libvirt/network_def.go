package libvirt

import (
	"encoding/xml"
	"fmt"

	"github.com/libvirt/libvirt-go-xml"
)

// HasDHCP Check if the network has a DHCP server managed by libvirt
func HasDHCP(net libvirtxml.Network) bool {
	if net.Forward != nil {
		if net.Forward.Mode == "nat" || net.Forward.Mode == "route" || net.Forward.Mode == "" {
			return true
		}
	}
	return false
}

// Creates a network definition from a XML
func newDefNetworkFromXML(s string) (libvirtxml.Network, error) {
	var networkDef libvirtxml.Network
	err := xml.Unmarshal([]byte(s), &networkDef)
	if err != nil {
		return libvirtxml.Network{}, err
	}
	return networkDef, nil
}

func newDefNetworkfromLibvirt(network LibVirtNetwork) (libvirtxml.Network, error) {
	networkXMLDesc, err := network.GetXMLDesc(0)
	if err != nil {
		return libvirtxml.Network{}, fmt.Errorf("Error retrieving libvirt domain XML description: %s", err)
	}
	networkDef := libvirtxml.Network{}
	err = xml.Unmarshal([]byte(networkXMLDesc), &networkDef)
	if err != nil {
		return libvirtxml.Network{}, fmt.Errorf("Error reading libvirt network XML description: %s", err)
	}
	return networkDef, nil
}

// Creates a network definition with the defaults the provider uses
func newNetworkDef() libvirtxml.Network {
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
