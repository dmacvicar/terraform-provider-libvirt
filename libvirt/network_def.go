package libvirt

import (
	"encoding/xml"
	"fmt"
	"log"
	"net"

	libvirt "github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
)

// HasDHCP checks if the network has a DHCP server managed by libvirt
func HasDHCP(net libvirtxml.Network) bool {
	if net.Forward != nil {
		if net.Forward.Mode == "nat" || net.Forward.Mode == "route" || net.Forward.Mode == "" {
			return true
		}
	} else {
		// isolated network
		return true
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

func getXMLNetworkDefFromLibvirt(network Network) (libvirtxml.Network, error) {
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
		panic(fmt.Sprintf("Unexpected error while parsing default network definition: %s", err))
	} else {
		return d
	}
}

func getHostXMLDesc(ip, mac, name string) string {
	dd := libvirtxml.NetworkDHCPHost{
		IP:   ip,
		MAC:  mac,
		Name: name,
	}
	tmp := struct {
		XMLName xml.Name `xml:"host"`
		libvirtxml.NetworkDHCPHost
	}{xml.Name{}, dd}
	xml, err := xmlMarshallIndented(tmp)
	if err != nil {
		panic("could not marshall host")
	}
	return xml
}

// Adds a new static host to the network
func addHost(n *libvirt.Network, ip, mac, name string, xmlIdx int) error {
	xmlDesc := getHostXMLDesc(ip, mac, name)
	log.Printf("Adding host with XML:\n%s", xmlDesc)
	// From https://libvirt.org/html/libvirt-libvirt-network.html#virNetworkUpdateFlags
	// Update live and config for hosts to make update permanent across reboots
	return n.Update(libvirt.NETWORK_UPDATE_COMMAND_ADD_LAST, libvirt.NETWORK_SECTION_IP_DHCP_HOST, xmlIdx, xmlDesc, libvirt.NETWORK_UPDATE_AFFECT_CONFIG|libvirt.NETWORK_UPDATE_AFFECT_LIVE)
}

// Update a static host from the network
func updateHost(n *libvirt.Network, ip, mac, name string, xmlIdx int) error {
	xmlDesc := getHostXMLDesc(ip, mac, name)
	log.Printf("Updating host with XML:\n%s", xmlDesc)
	// From https://libvirt.org/html/libvirt-libvirt-network.html#virNetworkUpdateFlags
	// Update live and config for hosts to make update permanent across reboots
	return n.Update(libvirt.NETWORK_UPDATE_COMMAND_MODIFY, libvirt.NETWORK_SECTION_IP_DHCP_HOST, xmlIdx, xmlDesc, libvirt.NETWORK_UPDATE_AFFECT_CONFIG|libvirt.NETWORK_UPDATE_AFFECT_LIVE)
}

// Get the network index of the target network
func getNetworkIdx(n *libvirtxml.Network, ip string) (int, error) {
	xmlIdx := -1

	if n == nil {
		return xmlIdx, fmt.Errorf("failed to convert to libvirt XML")
	}

	for idx, netIps := range n.IPs {
		_, netw, err := net.ParseCIDR(fmt.Sprintf("%s/%d", netIps.Address, netIps.Prefix))
		if err != nil {
			return xmlIdx, err
		}

		if netw.Contains(net.ParseIP(ip)) {
			xmlIdx = idx
			break
		}
	}

	return xmlIdx, nil
}

// Tries to update first, if that fails, it will add it
func updateOrAddHost(n *libvirt.Network, ip, mac, name string) error {
	xmlNet, _ := getXMLNetworkDefFromLibvirt(n)
	// We don't check the error above
	// if we can't parse the network to xml for some kind fo reasons
	// we will return the default '-1' value.
	xmlIdx, err := getNetworkIdx(&xmlNet, ip)
	if err == nil {
		log.Printf("Error during detecting network index: %s\nUsing default value: %d", err, xmlIdx)
	}

	err = updateHost(n, ip, mac, name, xmlIdx)
	if virErr, ok := err.(libvirt.Error); ok && virErr.Code == libvirt.ERR_OPERATION_INVALID && virErr.Domain == libvirt.FROM_NETWORK {
		return addHost(n, ip, mac, name, xmlIdx)
	}
	return err
}
