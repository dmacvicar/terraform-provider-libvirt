package libvirt

import (
	"encoding/json"
	"testing"

	libvirt "github.com/libvirt/libvirt-go"
)

func TestGetDomainInterfacesViaQemuAgentInvalidResponse(t *testing.T) {
	domain := DomainMock{}

	interfaces := qemuAgentGetInterfacesInfo(domain, false)

	if len(interfaces) != 0 {
		t.Errorf("wrong number of interfaces: %d instead of 0", len(interfaces))
	}
}

func TestGetDomainInterfacesViaQemuAgentNoInterfaces(t *testing.T) {

	response := QemuAgentInterfacesResponse{
		Interfaces: []QemuAgentInterface{}}
	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	domain := DomainMock{
		QemuAgentCommandResponse: string(data),
	}

	interfaces := qemuAgentGetInterfacesInfo(domain, false)
	if len(interfaces) != 0 {
		t.Errorf("wrong number of interfaces: %d instead of 0", len(interfaces))
	}
}

func TestGetDomainInterfacesViaQemuAgentIgnoreLoopbackDevice(t *testing.T) {
	response := QemuAgentInterfacesResponse{
		Interfaces: []QemuAgentInterface{
			{
				Name:   "lo",
				Hwaddr: "ho:me",
				IPAddresses: []QemuAgentInterfaceIPAddress{
					{
						Type:    "ipv4",
						Address: "127.0.0.1",
						Prefix:  1,
					},
				},
			},
		},
	}
	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	domain := DomainMock{
		QemuAgentCommandResponse: string(data),
	}

	interfaces := qemuAgentGetInterfacesInfo(domain, false)

	if len(interfaces) != 0 {
		t.Errorf("wrong number of interfaces)")
	}
}

func TestGetDomainInterfacesViaQemuAgentIgnoreDevicesWithoutAddress(t *testing.T) {
	response := QemuAgentInterfacesResponse{
		Interfaces: []QemuAgentInterface{
			{
				Name:   "eth1",
				Hwaddr: "xy:yy:zz",
				IPAddresses: []QemuAgentInterfaceIPAddress{
					{
						Type:    "ipv4",
						Address: "",
						Prefix:  1,
					},
				},
			},
		},
	}
	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	domain := DomainMock{
		QemuAgentCommandResponse: string(data),
	}

	interfaces := qemuAgentGetInterfacesInfo(domain, false)

	if len(interfaces) != 0 {
		t.Errorf("wrong number of interfaces")
	}
}

func TestGetDomainInterfacesViaQemuAgentUnknownIpAddressType(t *testing.T) {
	response := QemuAgentInterfacesResponse{
		Interfaces: []QemuAgentInterface{
			{
				Name:   "eth2",
				Hwaddr: "zy:yy:zz",
				IPAddresses: []QemuAgentInterfaceIPAddress{
					{
						Type:    "ipv8",
						Address: "i don't exist",
						Prefix:  1,
					},
				},
			},
		},
	}
	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	domain := DomainMock{
		QemuAgentCommandResponse: string(data),
	}

	interfaces := qemuAgentGetInterfacesInfo(domain, false)

	if len(interfaces) != 0 {
		t.Errorf("wrong number of interfaces: %d instead of 1", len(interfaces))
	}
}

func TestGetDomainInterfacesViaQemuAgent(t *testing.T) {
	device := "eth0"
	mac := "xx:yy:zz"
	ipv4Addr := "192.168.1.1"
	ipv6Addr := "2001:0db8:0000:0000:0000:ff00:0042:8329"

	response := QemuAgentInterfacesResponse{
		Interfaces: []QemuAgentInterface{
			{
				Name:   device,
				Hwaddr: mac,
				IPAddresses: []QemuAgentInterfaceIPAddress{
					{
						Type:    "ipv4",
						Address: ipv4Addr,
						Prefix:  1,
					},
					{
						Type:    "ipv6",
						Address: ipv6Addr,
						Prefix:  1,
					},
				},
			},
		},
	}
	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	domain := DomainMock{
		QemuAgentCommandResponse: string(data),
	}

	interfaces := qemuAgentGetInterfacesInfo(domain, false)

	if len(interfaces) != 1 {
		t.Errorf("wrong number of interfaces: %d instead of 1", len(interfaces))
	}

	if interfaces[0].Name != device {
		t.Errorf("wrong interface name: %s", interfaces[0].Name)
	}

	if interfaces[0].Hwaddr != mac {
		t.Errorf("wrong interface name: %s", interfaces[0].Hwaddr)
	}

	if len(interfaces[0].Addrs) != 2 {
		t.Errorf("wrong number of addresses: %d", len(interfaces[0].Addrs))
	}

	for _, addr := range interfaces[0].Addrs {
		var expected string

		if addr.Type == int(libvirt.IP_ADDR_TYPE_IPV4) {
			expected = ipv4Addr
		} else {
			expected = ipv6Addr
		}

		if expected != addr.Addr {
			t.Errorf("Expected %s, got %s", expected, addr.Addr)
		}
	}
}
