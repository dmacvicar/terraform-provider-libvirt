package libvirt

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	libvirt "github.com/libvirt/libvirt-go"
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

type libvirtIfacesLst []libvirt.DomainInterface

// Retrieve all the interfaces attached to a domain and their addresses. Only
// the interfaces with at least an IP address are returned.
// When wait4ipv4 is turned on the code will not report interfaces that don't
// have a ipv4 address set. This is useful when a domain gets the ipv6 address
// before the ipv4 one.
func getDomainInterfacesViaQemuAgent(domain LibVirtDomain, wait4ipv4 bool) libvirtIfacesLst {

	qemuAgentInterfacesRefreshFunc := func() resource.StateRefreshFunc {
		return func() (interface{}, string, error) {

			var interfaces libvirtIfacesLst

			log.Print("[DEBUG] sending command to qemu-agent")
			result, err := domain.QemuAgentCommand(
				"{\"execute\":\"guest-network-get-interfaces\"}",
				libvirt.DOMAIN_QEMU_AGENT_COMMAND_DEFAULT,
				0)
			if err != nil {
				log.Printf("[DEBUG] command error: %s", err)
				return interfaces, "failed", nil
			}

			log.Printf("[DEBUG] qemu-agent response: %s", result)

			response := QemuAgentInterfacesResponse{}
			if err := json.Unmarshal([]byte(result), &response); err != nil {
				log.Printf("[DEBUG] Error converting qemu-agent response about domain interfaces: %s", err)
				log.Printf("[DEBUG] Original message: %s", response)
				log.Print("[DEBUG] Returning an empty list of interfaces")
				return interfaces, "fatal", nil
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
				for _, addr := range iface.IpAddresses {
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

			log.Printf("[DEBUG] Interfaces obtained via qemu-agent: %+v", interfaces)
			return interfaces, "success", nil
		}
	}

	qemuAgentQuery := &resource.StateChangeConf{
		Pending:    []string{"failed"},
		Target:     []string{"success"},
		Refresh:    qemuAgentInterfacesRefreshFunc(),
		MinTimeout: 4 * time.Second,
		Delay:      4 * time.Second, // Wait this time before starting checks
		Timeout:    16 * time.Second,
	}

	interfaces, err := qemuAgentQuery.WaitForState()
	if err != nil {
		return libvirtIfacesLst{}
	}

	return interfaces.(libvirtIfacesLst)
}
