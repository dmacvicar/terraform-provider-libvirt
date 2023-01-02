package libvirt

import (
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

func init() {
	spew.Config.Indent = "\t"
}

func TestSplitKernelCmdLine(t *testing.T) {
	e := []map[string]string{
		{"foo": "bar"},
		{
			"foo":  "bar",
			"key":  "val",
			"root": "UUID=aa52d618-a2c4-4aad-aeb7-68d9e3a2c91d"},
		{"_": "nosplash rw"}}
	r:= splitKernelCmdLine("foo=bar foo=bar key=val root=UUID=aa52d618-a2c4-4aad-aeb7-68d9e3a2c91d nosplash rw")
	if !reflect.DeepEqual(r, e) {
		t.Fatalf("got='%s' expected='%s'", spew.Sdump(r), spew.Sdump(e))
	}
}

func TestSplitKernelEmptyCmdLine(t *testing.T) {
	var e []map[string]string
	r := splitKernelCmdLine("")
	if !reflect.DeepEqual(r, e) {
		t.Fatalf("got='%s' expected='%s'", spew.Sdump(r), spew.Sdump(e))
	}
}

func TestGetHostArchitecture(t *testing.T) {
	skipIfAccDisabled(t)
	conn := testAccProvider.Meta().(*Client).libvirt

	arch, err := getHostArchitecture(conn)

	if err != nil {
		t.Errorf("error %v", err)
	}

	t.Logf("[DEBUG] arch - %s", arch)

	if arch == "" {
		t.Errorf("arch is blank.")
	}
}

func TestGetCanonicalMachineName(t *testing.T) {
	skipIfAccDisabled(t)
	conn := testAccProvider.Meta().(*Client).libvirt

	arch := "x86_64"
	virttype := "hvm"
	machine := "pc"

	caps, err := getHostCapabilities(conn)
	if err != nil {
		t.Error(err)
	}

	name, err := getCanonicalMachineName(caps, arch, virttype, machine)

	if err != nil {
		t.Errorf("Could not get canonical name for %s/%s", arch, machine)
		return
	}

	t.Logf("Canonical name for %s/%s = %s", arch, machine, name)
}

func TestGetOriginalMachineName(t *testing.T) {
	skipIfAccDisabled(t)
	conn := testAccProvider.Meta().(*Client).libvirt

	arch := "x86_64"
	virttype := "hvm"
	machine := "pc"

	caps, err := getHostCapabilities(conn)
	if err != nil {
		t.Error(err)
	}

	canonname, err := getCanonicalMachineName(caps, arch, virttype, machine)
	if err != nil {
		t.Error(err)
	}
	reversename, err := getOriginalMachineName(caps, arch, virttype, canonname)
	if err != nil {
		t.Error(err)
	}
	if reversename != machine {
		t.Errorf("Cannot reverse canonical machine lookup")
	}

	t.Logf("Reverse canonical lookup for %s is %s which matches %s", canonname, reversename, machine)
}

func TestGetHostCapabilties(t *testing.T) {
	skipIfAccDisabled(t)
	conn := testAccProvider.Meta().(*Client).libvirt

	start := time.Now()

	caps, err := getHostCapabilities(conn)
	if err != nil {
		t.Errorf("Can't get host capabilties")
	}
	if caps.Host.UUID == "" {
		t.Errorf("Host has no UUID!")
	}

	elapsed := time.Since(start)
	t.Logf("[DEBUG] Get host capabilities took %s", elapsed)
}
