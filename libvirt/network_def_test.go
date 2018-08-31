package libvirt

import (
	"bytes"
	"encoding/xml"
	"errors"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/libvirt/libvirt-go-xml"
)

func init() {
	spew.Config.Indent = "\t"
}

func TestDefaultNetworkMarshall(t *testing.T) {
	b := newNetworkDef()
	buf := new(bytes.Buffer)
	enc := xml.NewEncoder(buf)
	enc.Indent("  ", "    ")
	if err := enc.Encode(b); err != nil {
		t.Fatalf("could not marshall this:\n%s", spew.Sdump(b))
	}
}

func TestNetworkDefUnmarshall(t *testing.T) {
	// some testing XML from the official docs (some unsupported attrs will be just ignored)
	text := `
		<network>
			<name>my-network</name>
			<bridge name="virbr0" stp="on" delay="5" macTableManager="libvirt"/>
			<mac address='00:16:3E:5D:C7:9E'/>
			<domain name="example.com" localOnly="no"/>
			<forward mode='nat'>
				<nat>
					<address start='1.2.3.4' end='1.2.3.10'/>
				</nat>
			</forward>
			<dns>
				<txt name="example" value="example value" />
				<forwarder addr="8.8.8.8"/>
				<forwarder addr="8.8.4.4"/>
				<srv service='name' protocol='tcp' domain='test-domain-name' target='.' port='1024' priority='10' weight='10'/>
				<host ip='192.168.122.2'>
					<hostname>myhost</hostname>
					<hostname>myhostalias</hostname>
				</host>
			</dns>
			<ip address="192.168.122.1" netmask="255.255.255.0">
				<dhcp>
					<range start="192.168.122.100" end="192.168.122.254" />
					<host mac="00:16:3e:77:e2:ed" name="foo.example.com" ip="192.168.122.10" />
					<host mac="00:16:3e:3e:a9:1a" name="bar.example.com" ip="192.168.122.11" />
				</dhcp>
			</ip>
			<ip family="ipv6" address="2001:db8:ca2:2::1" prefix="64" />
			<route family="ipv6" address="2001:db9:ca1:1::" prefix="64" gateway="2001:db8:ca2:2::2" />
  		</network>
	`

	b, err := newDefNetworkFromXML(text)
	if err != nil {
		t.Errorf("could not parse: %s", err)
	}
	if b.Name != "my-network" {
		t.Errorf("wrong network name: '%s'", b.Name)
	}
	if b.Domain.Name != "example.com" {
		t.Errorf("wrong domain name: '%s'", b.Domain.Name)
	}
	if b.Forward.Mode != "nat" {
		t.Errorf("wrong forward mode: '%s'", b.Forward.Mode)
	}
	if len(b.Forward.NAT.Addresses) == 0 {
		t.Errorf("wrong number of addresses: %s", b.Forward.NAT.Addresses)
	}
	if b.Forward.NAT.Addresses[0].Start != "1.2.3.4" {
		t.Errorf("wrong forward start address: %s", b.Forward.NAT.Addresses[0].Start)
	}
	if len(b.IPs) == 0 {
		t.Errorf("wrong number of IPs: %d", len(b.IPs))
	}
	_, err2 := xmlMarshallIndented(b)
	if err2 != nil {
		t.Fatalf("marshalling error\n%s", spew.Sdump(b))
	}
}

func TestBrokenNetworkDefUnmarshall(t *testing.T) {
	// Try unmarshalling some broken xml
	text := `
		<network>
	`

	_, err := newDefNetworkFromXML(text)
	if err == nil {
		t.Error("Unmarshal was supposed to fail")
	}
}

func TestHasDHCPNoForwardSet(t *testing.T) {
	net := libvirtxml.Network{}

	if HasDHCP(net) {
		t.Error("Expected to not have forward enabled")
	}
}

func TestHasDHCPForwardSet(t *testing.T) {
	createNet := func(mode string) libvirtxml.Network {
		return libvirtxml.Network{
			Forward: &libvirtxml.NetworkForward{
				Mode: mode,
			},
		}
	}

	for _, mode := range []string{"nat", "route", ""} {
		net := createNet(mode)
		if !HasDHCP(net) {
			t.Errorf(
				"Expected to have forward enabled with forward set to be '%s'",
				mode)
		}
	}
}

func TestNetworkFromLibvirtError(t *testing.T) {
	net := NetworkMock{
		GetXMLDescError: errors.New("boom"),
	}

	_, err := newDefNetworkfromLibvirt(net)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestNetworkFromLibvirtWrongResponse(t *testing.T) {
	net := NetworkMock{
		GetXMLDescReply: "wrong xml",
	}

	_, err := newDefNetworkfromLibvirt(net)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestNetworkFromLibvirt(t *testing.T) {
	net := NetworkMock{
		GetXMLDescReply: `
		<network>
		  <name>default</name>
		  <forward mode='nat'>
		    <nat>
		      <port start='1024' end='65535'/>
		    </nat>
		  </forward>
		</network>`,
	}

	dn, err := newDefNetworkfromLibvirt(net)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if dn.Forward.Mode != "nat" {
		t.Errorf("Wrong forward mode: %s", dn.Forward.Mode)
	}
}
