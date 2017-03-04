package libvirt

import (
	"encoding/json"
	"log"
	"strings"

	libvirt "github.com/dmacvicar/libvirt-go"
)

type QemuAgentInterfacesResponse struct {
	Interfaces []QemuAgentInterface `json:"return"`
}

type QemuAgentInterface struct {
	Name        string                        `json:"name"`
	Hwaddr      string                        `json:"hardware-address"`
	IpAddresses []QemuAgentInterfaceIpAddress `json:"ip-addresses"`
}

type QemuAgentInterfaceIpAddress struct {
	Type    string `json:"ip-address-type"`
	Address string `json:"ip-address"`
	Prefix  uint   `json:"prefix"`
}

// Retrieve all the interfaces attached to a domain and their addresses. Only
// the interfaces with at least an IP address are returned.
// When wait4ipv4 is turned on the code will not report interfaces that don't
// have a ipv4 address set. This is useful when a domain gets the ipv6 address
// before the ipv4 one.
func getDomainInterfacesViaQemuAgent(domain LibVirtDomain, wait4ipv4 bool) []libvirt.VirDomainInterface {
	log.Print("[DEBUG] get network interfaces using qemu agent")

	var interfaces []libvirt.VirDomainInterface

	result := domain.QemuAgentCommand(
		"{\"execute\":\"guest-network-get-interfaces\"}",
		libvirt.VIR_DOMAIN_QEMU_AGENT_COMMAND_DEFAULT,
		0)

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

		libVirtIface := libvirt.VirDomainInterface{
			Name:   iface.Name,
			Hwaddr: iface.Hwaddr}

		ipv4Assigned := false
		for _, addr := range iface.IpAddresses {
			if addr.Address == "" {
				// ignore interfaces without an address (eg. waiting for dhcp lease)
				continue
			}

			libVirtAddr := libvirt.VirDomainIPAddress{
				Addr:   addr.Address,
				Prefix: addr.Prefix,
			}

			switch strings.ToLower(addr.Type) {
			case "ipv4":
				libVirtAddr.Type = libvirt.VIR_IP_ADDR_TYPE_IPV4
				ipv4Assigned = true
			case "ipv6":
				libVirtAddr.Type = libvirt.VIR_IP_ADDR_TYPE_IPV6
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
