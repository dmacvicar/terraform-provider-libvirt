package libvirt

import (
	"net"
	"testing"
)

func TestRandomMACAddress(t *testing.T) {
	mac, err := randomMACAddress()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = net.ParseMAC(mac)

	if err != nil {
		t.Errorf("Invalid MAC address generated: %s - %v", mac, err)
	}
}
