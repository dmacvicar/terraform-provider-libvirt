package libvirt

import (
	"encoding/xml"
	"errors"
	"testing"
)

func TestGetHostXMLDesc(t *testing.T) {
	ip := "127.0.0.1"
	mac := "XX:YY:ZZ"
	name := "localhost"

	data := getHostXMLDesc(ip, mac, name)

	dd := defNetworkIpDhcpHost{}
	err := xml.Unmarshal([]byte(data), &dd)
	if err != nil {
		t.Errorf("error %v", err)
	}

	if dd.Ip != ip {
		t.Errorf("expected ip %s, got %s", ip, dd.Ip)
	}

	if dd.Mac != mac {
		t.Errorf("expected mac %s, got %s", mac, dd.Mac)
	}

	if dd.Name != name {
		t.Errorf("expected name %s, got %s", name, dd.Name)
	}
}

func TestAddDHCPRange(t *testing.T) {
	network := &LibVirtNetworkMock{
		UpdateXMLDescError: nil,
	}
	if err := addDHCPRange(network, "10.0.0.2", "10.0.0.254"); err != nil {
		t.Errorf("error %v", err)
	}
}

func TestAddDHCPRangeError(t *testing.T) {
	network := &LibVirtNetworkMock{
		UpdateXMLDescError: errors.New("boom"),
	}
	if err := addDHCPRange(network, "invalid", "cidr"); err == nil {
		t.Error("Expected error")
	}
}
