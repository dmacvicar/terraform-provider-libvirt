package libvirt

import (
	"testing"
)

func TestEnableDHCP(t *testing.T) {
	network := &LibVirtNetworkMock{
		GetXMLDescReply: `
            <network>
              <name>test-net</name>
              <forward mode='nat'>
                <nat>
                  <port start='1024' end='65535'/>
                </nat>
              </forward>
              <bridge name='testbr' stp='on' delay='0'/>
              <mac address='41:d6:45:0b:94:38'/>
              <ip family='ipv4' address='10.0.0.1' prefix='24'>
              </ip>
            </network>`,
		UpdateXMLDescError: nil,
	}

	if err := enableDHCP(network); err != nil {
		t.Errorf("error %v", err)
	}
	if !network.UpdateXMLDescCalled {
		t.Error("Expected update of the xml description to enable DHCP")
	}
}

func TestEnableDHCPAlreadyEnabled(t *testing.T) {
	network := &LibVirtNetworkMock{
		GetXMLDescReply: `
            <network>
              <name>test-net</name>
              <forward mode='nat'>
                <nat>
                  <port start='1024' end='65535'/>
                </nat>
              </forward>
              <bridge name='testbr' stp='on' delay='0'/>
              <mac address='41:d6:45:0b:94:38'/>
              <ip family='ipv4' address='10.0.0.1' prefix='24'>
                <dhcp>
                  <range start='10.0.0.2' end='10.0.0.254'/>
                </dhcp>
              </ip>
            </network>`,
		UpdateXMLDescError: nil,
	}

	if err := enableDHCP(network); err != nil {
		t.Errorf("error %v", err)
	}
	if network.UpdateXMLDescCalled {
		t.Error("Expected no update of the xml description (DHCP already enabled)")
	}
}
