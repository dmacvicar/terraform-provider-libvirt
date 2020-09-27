package libvirt

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	libvirt "github.com/libvirt/libvirt-go"
)

const qemuGetIfaceWait = "qemu-agent-wait"
const qemuGetIfaceDone = "qemu-agent-done"

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

func qemuAgentInterfacesRefreshFunc(domain Domain, wait4ipv4 bool) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		var interfaces []libvirt.DomainInterface

		log.Printf("[DEBUG] sending command to qemu-agent")
		result, err := domain.QemuAgentCommand(
			"{\"execute\":\"guest-network-get-interfaces\"}",
			libvirt.DOMAIN_QEMU_AGENT_COMMAND_DEFAULT,
			0)
		if err != nil {
			log.Printf("[DEBUG] command error: %s", err)
			return interfaces, qemuGetIfaceWait, nil
		}

		log.Printf("[DEBUG] qemu-agent response: %s", result)

		response := QemuAgentInterfacesResponse{}
		if err := json.Unmarshal([]byte(result), &response); err != nil {
			log.Printf("[DEBUG] Error converting qemu-agent response about domain interfaces: %s", err)
			log.Printf("[DEBUG] Original message: %+v", response)
			log.Print("[DEBUG] Returning an empty list of interfaces")
			return interfaces, "", nil
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

		log.Printf("[DEBUG] Interfaces obtained via qemu-agent: %+v", interfaces)
		return interfaces, qemuGetIfaceDone, nil
	}
}

// Retrieve all the interfaces attached to a domain and their addresses. Only
// the interfaces with at least an IP address are returned.
// When wait4ipv4 is turned on the code will not report interfaces that don't
// have a ipv4 address set. This is useful when a domain gets the ipv6 address
// before the ipv4 one.
func qemuAgentGetInterfacesInfo(domain Domain, wait4ipv4 bool) []libvirt.DomainInterface {

	qemuAgentQuery := &resource.StateChangeConf{
		Pending:    []string{qemuGetIfaceWait},
		Target:     []string{qemuGetIfaceDone},
		Refresh:    qemuAgentInterfacesRefreshFunc(domain, wait4ipv4),
		MinTimeout: 1 * time.Minute,
		Delay:      30 * time.Second, // Wait this time before starting checks
		Timeout:    5 * time.Minute,
	}

	interfaces, err := qemuAgentQuery.WaitForState()
	if err != nil {
		return []libvirt.DomainInterface{}
	}

	return interfaces.([]libvirt.DomainInterface)
}
