package libvirt

import (
	"encoding/json"
	"log"
	"strings"

	libvirt "github.com/libvirt/libvirt-go"
)

// QemuAgentInterfacesResponse type
type QemuAgentInterfacesResponse struct {
	Interfaces []QemuAgentInterface `json:"return"`
}

// QemuAgentInterface type
type QemuAgentInterface struct {
	Name        string                        `json:"name"`
	Hwaddr      string                        `json:"hardware-address"`
	IPAddresses []QemuAgentInterfaceIPAddress `json:"ip-addresses"`
}

// QemuAgentInterfaceIPAddress type
type QemuAgentInterfaceIPAddress struct {
	Type    string `json:"ip-address-type"`
	Address string `json:"ip-address"`
	Prefix  uint   `json:"prefix"`
}

// Retrieve all the interfaces attached to a domain and their addresses. Only
// the interfaces with at least an IP address are returned.
// When wait4ipv4 is turned on the code will not report interfaces that don't
// have a ipv4 address set. This is useful when a domain gets the ipv6 address
// before the ipv4 one.
func getDomainInterfacesViaQemuAgent(domain Domain, wait4ipv4 bool) []libvirt.DomainInterface {
	log.Print("[DEBUG] get network interfaces using qemu agent")

	var interfaces []libvirt.DomainInterface

	result, err := domain.QemuAgentCommand(
		"{\"execute\":\"guest-network-get-interfaces\"}",
		libvirt.DOMAIN_QEMU_AGENT_COMMAND_DEFAULT,
		0)
	if err != nil {
		return interfaces
	}

	log.Printf("[DEBUG] qemu-agent response: %s", result)

	response := QemuAgentInterfacesResponse{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		log.Printf("[DEBUG] Error converting Qemu agent response about domain interfaces: %s", err)
		log.Printf("[DEBUG] Original message: %s", response)
		log.Print("[DEBUG] Returning an empty list of interfaces")
		return interfaces
	}
	log.Printf("[DEBUG] Parsed response %+v", response)

	for _, iface := range response.Interfaces {
		if iface.Name == "lo" {
			// ignore loopback interface
			continue
		}

		libVirtIface := libvirt.DomainInterface{
			Name:   iface.Name,
			Hwaddr: iface.Hwaddr}

		ipv4Assigned := false
		for _, addr := range iface.IPAddresses {
			if addr.Address == "" {
				// ignore interfaces without an address (eg. waiting for dhcp lease)
				continue
			}

			libVirtAddr := libvirt.DomainIPAddress{
				Addr:   addr.Address,
				Prefix: addr.Prefix,
			}

			switch strings.ToLower(addr.Type) {
			case "ipv4":
				libVirtAddr.Type = int(libvirt.IP_ADDR_TYPE_IPV4)
				ipv4Assigned = true
			case "ipv6":
				libVirtAddr.Type = int(libvirt.IP_ADDR_TYPE_IPV6)
			default:
				log.Printf("[ERROR] Cannot handle unknown address type %s", addr.Type)
				continue
			}
			libVirtIface.Addrs = append(libVirtIface.Addrs, libVirtAddr)
		}
		if len(libVirtIface.Addrs) > 0 && (ipv4Assigned || !wait4ipv4) {
			interfaces = append(interfaces, libVirtIface)
		}
	}

	log.Printf("[DEBUG] Interfaces obtained via qemu Agent: %+v", interfaces)

	return interfaces
}
