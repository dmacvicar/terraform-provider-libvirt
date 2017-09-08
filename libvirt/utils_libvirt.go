package libvirt

import (
	"encoding/xml"
	"log"

	libvirt "github.com/libvirt/libvirt-go"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

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
func addHost(n *libvirt.Network, ip, mac, name string) error {
	xmlDesc := getHostXMLDesc(ip, mac, name)
	log.Printf("Adding host with XML:\n%s", xmlDesc)
	return n.Update(libvirt.NETWORK_UPDATE_COMMAND_ADD_LAST, libvirt.NETWORK_SECTION_IP_DHCP_HOST, -1, xmlDesc, libvirt.NETWORK_UPDATE_AFFECT_CURRENT)
}

// Removes a static host from the network
func removeHost(n *libvirt.Network, ip, mac, name string) error {
	xmlDesc := getHostXMLDesc(ip, mac, name)
	log.Printf("Removing host with XML:\n%s", xmlDesc)
	return n.Update(libvirt.NETWORK_UPDATE_COMMAND_DELETE, libvirt.NETWORK_SECTION_IP_DHCP_HOST, -1, xmlDesc, libvirt.NETWORK_UPDATE_AFFECT_CURRENT)
}

// Update a static host from the network
func updateHost(n *libvirt.Network, ip, mac, name string) error {
	xmlDesc := getHostXMLDesc(ip, mac, name)
	log.Printf("Updating host with XML:\n%s", xmlDesc)
	return n.Update(libvirt.NETWORK_UPDATE_COMMAND_MODIFY, libvirt.NETWORK_SECTION_IP_DHCP_HOST, -1, xmlDesc, libvirt.NETWORK_UPDATE_AFFECT_CURRENT)
}

func getHostArchitecture(virConn *libvirt.Connect) (string, error) {
	type HostCapabilities struct {
		XMLName xml.Name `xml:"capabilities"`
		Host    struct {
			XMLName xml.Name `xml:"host"`
			CPU     struct {
				XMLName xml.Name `xml:"cpu"`
				Arch    string   `xml:"arch"`
			}
		}
	}

	info, err := virConn.GetCapabilities()
	if err != nil {
		return "", err
	}

	capabilities := HostCapabilities{}
	xml.Unmarshal([]byte(info), &capabilities)

	return capabilities.Host.CPU.Arch, nil
}

func getHostCapabilities(virConn *libvirt.Connect) (libvirtxml.Caps, error) {
	// We should perhaps think of storing this on the connect object
	// on first call to avoid the back and forth
	caps := libvirtxml.Caps{}
	capsXML, err := virConn.GetCapabilities()
	if err != nil {
		return caps, err
	}
	xml.Unmarshal([]byte(capsXML), &caps)
	log.Printf("[TRACE] Capabilities of host \n %+v", caps)
	return caps, nil
}
