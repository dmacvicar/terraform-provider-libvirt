package libvirt

import (
	"encoding/xml"
	"testing"

	libvirt "github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
	"os"
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

func connect(t *testing.T) *libvirt.Connect {
	conn, err := libvirt.NewConnect(os.Getenv("LIBVIRT_DEFAULT_URI"))
	if err != nil {
		t.Fatalf("Cannot connect")
	}

	return conn
}

func TestGetHostArchitecture(t *testing.T) {

	conn := connect(t)
	defer conn.Close()

	arch, err := getHostArchitecture(conn)

	if err != nil {
		t.Errorf("error %v", err)
	}

	t.Logf("[DEBUG] arch - %s", arch)

	if arch == "" {
		t.Errorf("arch is blank.")
	}
}

func TestGetGuestMachines(t *testing.T) {
	conn := connect(t)
	defer conn.Close()

	machines, err := getGuestMachines(conn, "x86_64")

	if err != nil {
		t.Log(err)
		t.Fatalf("Could not get list of GuestMachines")
	}

	t.Logf("First is %s", machines[0].Name)
}

func TestGetCanonicalMachineName(t *testing.T) {

	conn := connect(t)
	defer conn.Close()
	arch := "x86_64"
	machine := "pc"
	name, err := getCanonicalMachineName(conn, arch, machine)

	if err != nil {
		t.Errorf("Could not get canonical name for %s/%s", arch, machine)
		return
	}

	t.Logf("Canonical name for %s/%s = %s", arch, machine, name)

}

func TestGetOriginalMachineName(t *testing.T) {
	conn := connect(t)
	defer conn.Close()
	arch := "x86_64"
	machine := "pc"
	canonname, err := getCanonicalMachineName(conn, arch, machine)
	if err != nil {
		t.Error(err)
	}
	reversename, err := getOriginalMachineName(conn, arch, canonname)
	if err != nil {
		t.Error(err)
	}
	if reversename != machine {
		t.Errorf("Cannot reverse canonical machine lookup")
	}

	t.Logf("Reverse canonical lookup for %s is %s which matches %s", canonname, reversename, machine)
}
