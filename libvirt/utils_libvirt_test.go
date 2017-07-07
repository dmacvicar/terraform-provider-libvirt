package libvirt

import (
	"encoding/xml"
	"testing"

	"github.com/libvirt/libvirt-go-xml"
)

func TestGetHostXMLDesc(t *testing.T) {
	ip := "127.0.0.1"
	mac := "XX:YY:ZZ"
	name := "localhost"

	data := getHostXMLDesc(ip, mac, name)

	dd := libvirtxml.NetworkDHCPHost{}
	err := xml.Unmarshal([]byte(data), &dd)
	if err != nil {
		t.Errorf("error %v", err)
	}

	if dd.IP != ip {
		t.Errorf("expected ip %s, got %s", ip, dd.IP)
	}

	if dd.MAC != mac {
		t.Errorf("expected mac %s, got %s", mac, dd.MAC)
	}

	if dd.Name != name {
		t.Errorf("expected name %s, got %s", name, dd.Name)
	}
}
