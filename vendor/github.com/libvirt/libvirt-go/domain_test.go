/*
 * This file is part of the libvirt-go project
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 *
 * Copyright (c) 2013 Alex Zorin
 * Copyright (C) 2016 Red Hat, Inc.
 *
 */

package libvirt

import (
	"strings"
	"testing"
	"time"
)

func buildTestDomain() (*Domain, *Connect) {
	conn := buildTestConnection()
	dom, err := conn.DomainDefineXML(`<domain type="test">
		<name>` + time.Now().String() + `</name>
		<memory unit="KiB">8192</memory>
		<os>
			<type>hvm</type>
		</os>
	</domain>`)
	if err != nil {
		panic(err)
	}
	return dom, conn
}

func buildSMPTestDomain() (*Domain, *Connect) {
	conn := buildTestConnection()
	dom, err := conn.DomainDefineXML(`<domain type="test">
		<name>` + time.Now().String() + `</name>
		<memory unit="KiB">8192</memory>
		<vcpu>8</vcpu>
  		<os>
			<type>hvm</type>
		</os>
	</domain>`)
	if err != nil {
		panic(err)
	}
	return dom, conn
}

func buildTransientTestDomain() (*Domain, *Connect) {
	conn := buildTestConnection()
	dom, err := conn.DomainCreateXML(`<domain type="test">
		<name>`+time.Now().String()+`</name>
		<memory unit="KiB">8192</memory>
		<os>
			<type>hvm</type>
		</os>
	</domain>`, DOMAIN_NONE)
	if err != nil {
		panic(err)
	}
	return dom, conn
}

func TestUndefineDomain(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	name, err := dom.GetName()
	if err != nil {
		t.Error(err)
		return
	}
	if err := dom.Undefine(); err != nil {
		t.Error(err)
		return
	}
	if _, err := conn.LookupDomainByName(name); err == nil {
		t.Fatal("Shouldn't have been able to find domain")
		return
	}
}

func TestGetDomainName(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Undefine()
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := dom.GetName(); err != nil {
		t.Error(err)
		return
	}
}

func TestGetDomainState(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	state, reason, err := dom.GetState()
	if err != nil {
		t.Error(err)
		return
	}
	if state != DOMAIN_SHUTOFF {
		t.Error("Domain state in test transport should be shutoff")
		return
	}
	if DomainShutoffReason(reason) != DOMAIN_SHUTOFF_UNKNOWN {
		t.Error("Domain reason in test transport should be unknown")
		return
	}
}

func TestGetDomainID(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()

	if err := dom.Create(); err != nil {
		t.Error("Failed to create domain")
	}

	if id, err := dom.GetID(); id == ^uint(0) || err != nil {
		dom.Destroy()
		t.Error("Couldn't get domain ID")
		return
	}
	dom.Destroy()
}

func TestGetDomainUUID(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	_, err := dom.GetUUID()
	// how to test uuid validity?
	if err != nil {
		t.Error(err)
		return
	}
}

func TestGetDomainUUIDString(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	_, err := dom.GetUUIDString()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestGetDomainInfo(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	_, err := dom.GetInfo()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestGetDomainXMLDesc(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	_, err := dom.GetXMLDesc(0)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestCreateDomainSnapshotXML(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	ss, err := dom.CreateSnapshotXML(`
		<domainsnapshot>
			<description>Test snapshot that will fail because its unsupported</description>
		</domainsnapshot>
	`, 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer ss.Free()
}

func TestSaveDomain(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}
	// get the name so we can get a handle on it later
	domName, err := dom.GetName()
	if err != nil {
		t.Error(err)
		return
	}
	const tmpFile = "/tmp/libvirt-go-test.tmp"
	if err := dom.Save(tmpFile); err != nil {
		t.Error(err)
		return
	}
	if err := conn.DomainRestore(tmpFile); err != nil {
		t.Error(err)
		return
	}
	if dom2, err := conn.LookupDomainByName(domName); err != nil {
		t.Error(err)
		return
	} else {
		dom2.Free()
	}
}

func TestSaveDomainFlags(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}
	const srcFile = "/tmp/libvirt-go-test.tmp"
	if err := dom.SaveFlags(srcFile, "", 0); err == nil {
		t.Fatal("expected xml modification unsupported")
		return
	}
}

func TestCreateDestroyDomain(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}
	state, reason, err := dom.GetState()
	if err != nil {
		t.Error(err)
		return
	}
	if state != DOMAIN_RUNNING {
		t.Fatal("Domain should be running")
		return
	}
	if DomainRunningReason(reason) != DOMAIN_RUNNING_BOOTED {
		t.Fatal("Domain reason should be booted")
		return
	}
	if err = dom.Destroy(); err != nil {
		t.Error(err)
		return
	}
	state, reason, err = dom.GetState()
	if err != nil {
		t.Error(err)
		return
	}
	if state != DOMAIN_SHUTOFF {
		t.Fatal("Domain should be destroyed")
		return
	}
	if DomainShutoffReason(reason) != DOMAIN_SHUTOFF_DESTROYED {
		t.Fatal("Domain reason should be destroyed")
		return
	}
}

func TestShutdownDomain(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}
	if err := dom.Shutdown(); err != nil {
		t.Error(err)
		return
	}
	state, reason, err := dom.GetState()
	if err != nil {
		t.Error(err)
		return
	}
	if state != DOMAIN_SHUTOFF {
		t.Error("Domain state in test transport should be shutoff")
		return
	}
	if DomainShutoffReason(reason) != DOMAIN_SHUTOFF_SHUTDOWN {
		t.Error("Domain reason in test transport should be shutdown")
		return
	}
}

func TestShutdownReboot(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}
	if err := dom.Reboot(0); err != nil {
		t.Error(err)
		return
	}
}

func TestDomainAutostart(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	as, err := dom.GetAutostart()
	if err != nil {
		t.Error(err)
		return
	}
	if as {
		t.Fatal("autostart should be false")
		return
	}
	if err := dom.SetAutostart(true); err != nil {
		t.Error(err)
		return
	}
	as, err = dom.GetAutostart()
	if err != nil {
		t.Error(err)
		return
	}
	if !as {
		t.Fatal("autostart should be true")
		return
	}
}

func TestDomainIsActive(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Log(err)
		return
	}
	active, err := dom.IsActive()
	if err != nil {
		t.Error(err)
		return
	}
	if !active {
		t.Fatal("Domain should be active")
		return
	}
	if err := dom.Destroy(); err != nil {
		t.Error(err)
		return
	}
	active, err = dom.IsActive()
	if err != nil {
		t.Error(err)
		return
	}
	if active {
		t.Fatal("Domain should be inactive")
		return
	}
}

func TestDomainIsPersistent(t *testing.T) {
	dom, conn := buildTransientTestDomain()
	dom2, conn2 := buildTestDomain()
	defer func() {
		dom.Free()
		dom2.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
		if res, _ := conn2.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	persistent, err := dom.IsPersistent()
	if err != nil {
		t.Error(err)
		return
	}
	if persistent {
		t.Fatal("Domain shouldn't be persistent")
		return
	}
	persistent, err = dom2.IsPersistent()
	if err != nil {
		t.Error(err)
		return
	}
	if !persistent {
		t.Fatal("Domain should be persistent")
		return
	}
}

func TestDomainSetMaxMemory(t *testing.T) {
	const mem = 8192 * 100
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.SetMaxMemory(mem); err != nil {
		t.Error(err)
		return
	}
}

func TestDomainSetMemory(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}
	if err := dom.SetMemory(1024); err != nil {
		t.Error(err)
		return
	}
}

func TestDomainSetVcpus(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}
	if err := dom.SetVcpus(1); err != nil {
		t.Error(err)
		return
	}
	if err := dom.SetVcpusFlags(1, DOMAIN_VCPU_LIVE); err != nil {
		t.Error(err)
		return
	}
}

func TestDomainFree(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Free(); err != nil {
		t.Error(err)
		return
	}
}

func TestDomainSuspend(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}
	defer dom.Destroy()
	if err := dom.Suspend(); err != nil {
		t.Error(err)
		return
	}
	defer dom.Resume()
}

func TesDomainShutdownFlags(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}
	if err := dom.ShutdownFlags(DOMAIN_SHUTDOWN_SIGNAL); err != nil {
		t.Error(err)
		return
	}
	state, reason, err := dom.GetState()
	if err != nil {
		t.Error(err)
		return
	}
	if state != DOMAIN_SHUTOFF {
		t.Error("Domain state in test transport should be shutoff")
		return
	}
	if DomainShutoffReason(reason) != DOMAIN_SHUTOFF_SHUTDOWN {
		t.Error("Domain reason in test transport should be shutdown")
		return
	}
}

func TesDomainDestoryFlags(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}
	if err := dom.DestroyFlags(DOMAIN_DESTROY_GRACEFUL); err != nil {
		t.Error(err)
		return
	}
	state, reason, err := dom.GetState()
	if err != nil {
		t.Error(err)
		return
	}
	if state != DOMAIN_SHUTOFF {
		t.Error("Domain state in test transport should be shutoff")
		return
	}
	if DomainShutoffReason(reason) != DOMAIN_SHUTOFF_SHUTDOWN {
		t.Error("Domain reason in test transport should be shutdown")
		return
	}
}

func TestDomainScreenshot(t *testing.T) {
	if VERSION_NUMBER == 2005000 {
		/* 2.5.0 broke screenshot for test:///default driver */
		return
	}
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}
	stream, err := conn.NewStream(0)
	if err != nil {
		t.Fatalf("failed to create new stream: %s", err)
	}
	defer stream.Free()
	mime, err := dom.Screenshot(stream, 0, 0)
	if err != nil {
		t.Fatalf("failed to take screenshot: %s", err)
	}
	if strings.Index(mime, "image/") != 0 {
		t.Fatalf("Wanted image/*, got %s", mime)
	}
}

func TestDomainGetVcpus(t *testing.T) {
	dom, conn := buildSMPTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}
	defer dom.Destroy()

	stats, err := dom.GetVcpus()
	if err != nil {
		t.Fatal(err)
	}

	if len(stats) != 8 {
		t.Fatal("should have 1 cpu")
	}

	if stats[0].State != 1 {
		t.Fatal("state should be 1")
	}
}

func TestDomainGetVcpusFlags(t *testing.T) {
	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}
	defer dom.Destroy()

	num, err := dom.GetVcpusFlags(0)
	if err != nil {
		t.Fatal(err)
	}

	if num != 1 {
		t.Fatal("should have 1 cpu", num)
	}
}

func TestDomainPinVcpu(t *testing.T) {
	dom, conn := buildSMPTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}
	defer dom.Destroy()

	err := dom.PinVcpu(2, []bool{false, true, false, false,
		true, false, false, false})
	if err != nil {
		t.Fatal(err)
	}
}

type CPUStringData struct {
	cpustr string
	cpumap []bool
	err    bool
}

func TestParserCPUString(t *testing.T) {
	testDataList := []CPUStringData{
		CPUStringData{
			"0-1,4,7-9,^8",
			[]bool{
				true, true, false, false,
				true, false, false, true,
				false, true,
			},
			false,
		},
		CPUStringData{
			"0-0,1,^1",
			[]bool{true, false},
			false,
		},
		CPUStringData{
			"0-3,,",
			[]bool{},
			true,
		},
		CPUStringData{
			"0-3,,",
			[]bool{},
			true,
		},
		CPUStringData{
			"3-0",
			[]bool{},
			true,
		},
		CPUStringData{
			"!0-3",
			[]bool{},
			true,
		},
	}

	for _, testData := range testDataList {
		actual, err := parseCPUString(testData.cpustr)

		if testData.err {
			if err == nil {
				t.Errorf("Expected parse error from %s",
					testData.cpustr)
				return
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected parse error from %s",
					testData.cpustr)
				return
			}
		}

		if len(actual) != len(testData.cpumap) {
			t.Errorf("Expected %s got %s",
				actual, testData.cpumap)
			return
		}

		for idx, val := range actual {
			if val != testData.cpumap[idx] {
				t.Errorf("Expected %s got %s",
					actual, testData.cpumap)
				return
			}
		}
	}
}

func TestSetMetadata(t *testing.T) {
	xmlns := "http://libvirt.org/xmlns/libvirt-go/test"
	xmlprefix := "test"
	meta := "<blob/>"

	dom, conn := buildTestDomain()
	defer func() {
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}
	defer dom.Destroy()

	data, err := dom.GetMetadata(DOMAIN_METADATA_ELEMENT, xmlns, DOMAIN_AFFECT_LIVE)
	if err == nil {
		t.Errorf("Expected an error for missing metadata")
		return
	}

	err = dom.SetMetadata(DOMAIN_METADATA_ELEMENT, meta, xmlprefix, xmlns, DOMAIN_AFFECT_LIVE)
	if err != nil {
		t.Error(err)
		return
	}

	data, err = dom.GetMetadata(DOMAIN_METADATA_ELEMENT, xmlns, DOMAIN_AFFECT_LIVE)
	if err != nil {
		t.Errorf("Unexpected an error for metadata")
		return
	}

	if data != meta {
		t.Errorf("Metadata %s doesn't match %s", data, meta)
		return
	}

	err = dom.SetMetadata(DOMAIN_METADATA_ELEMENT, "", "", xmlns, DOMAIN_AFFECT_LIVE)
	if err != nil {
		t.Error(err)
		return
	}

	data, err = dom.GetMetadata(DOMAIN_METADATA_ELEMENT, xmlns, DOMAIN_AFFECT_LIVE)
	if err == nil {
		t.Errorf("Expected an error for deleted metadata")
		return
	}

}
