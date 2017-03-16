package libvirt

import (
	"encoding/xml"
	"log"

	libvirt "github.com/dmacvicar/libvirt-go"
)

func getHostXMLDesc(ip, mac, name string) string {
	dd := defNetworkIpDhcpHost{
		Ip:   ip,
		Mac:  mac,
		Name: name,
	}
	xml, err := xmlMarshallIndented(dd)
	if err != nil {
		panic("could not marshall host")
	}
	return xml
}

// addDHCPRange adds a DHCP range to the network
func addDHCPRange(n *libvirt.VirNetwork, start, end string) error {
	dr := defNetworkIpDhcpRange{
		Start: start,
		End:   end,
	}
	xml, err := xmlMarshallIndented(dr)
	if err != nil {
		panic("could not marshall dhcp range")
	}
	log.Printf("Adding dhcp range with XML:\n%s", xml)
	return n.UpdateXMLDesc(xml, libvirt.VIR_NETWORK_UPDATE_COMMAND_ADD_FIRST, libvirt.VIR_NETWORK_SECTION_IP_DHCP_RANGE)
}

// Adds a new static host to the network
func addHost(n *libvirt.VirNetwork, ip, mac, name string) error {
	xmlDesc := getHostXMLDesc(ip, mac, name)
	log.Printf("Adding host with XML:\n%s", xmlDesc)
	return n.UpdateXMLDesc(xmlDesc, libvirt.VIR_NETWORK_UPDATE_COMMAND_ADD_LAST, libvirt.VIR_NETWORK_SECTION_IP_DHCP_HOST)
}

// Removes a static host from the network
func removeHost(n *libvirt.VirNetwork, ip, mac, name string) error {
	xmlDesc := getHostXMLDesc(ip, mac, name)
	log.Printf("Removing host with XML:\n%s", xmlDesc)
	return n.UpdateXMLDesc(xmlDesc, libvirt.VIR_NETWORK_UPDATE_COMMAND_DELETE, libvirt.VIR_NETWORK_SECTION_IP_DHCP_HOST)
}

// Update a static host from the network
func updateHost(n *libvirt.VirNetwork, ip, mac, name string) error {
	xmlDesc := getHostXMLDesc(ip, mac, name)
	log.Printf("Updating host with XML:\n%s", xmlDesc)
	return n.UpdateXMLDesc(xmlDesc, libvirt.VIR_NETWORK_UPDATE_COMMAND_MODIFY, libvirt.VIR_NETWORK_SECTION_IP_DHCP_HOST)
}

func getHostArchitecture(virConn *libvirt.VirConnection) (string, error) {
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
