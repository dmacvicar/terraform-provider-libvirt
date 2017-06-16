// +build integration

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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

func buildTestQEMUConnection() *Connect {
	conn, err := NewConnect("qemu:///system")
	if err != nil {
		panic(err)
	}
	return conn
}

func buildTestQEMUDomain(transient bool, name string) (*Domain, *Connect) {
	conn := buildTestQEMUConnection()
	fullname := fmt.Sprintf("libvirt-go-test-%s", name)

	dom, err := conn.LookupDomainByName(fullname)
	if err == nil {
		dom.Destroy()
		dom.Undefine()
		dom.Free()
	}

	xml := fmt.Sprintf(`<domain type="qemu">
		<name>libvirt-go-test-%s</name>
		<memory unit="KiB">128</memory>
                <features>
                  <acpi/>
                  <apic/>
                </features>
		<os>
			<type>hvm</type>
		</os>
	</domain>`, name)
	if transient {
		if dom, err = conn.DomainCreateXML(xml, 0); err != nil {
			panic(err)
		}
	} else {
		if dom, err = conn.DomainDefineXML(xml); err != nil {
			panic(err)
		}
	}
	return dom, conn
}

func getDefaultStoragePool(conn *Connect) *StoragePool {
	pool, err := conn.LookupStoragePoolByName("default")
	if err == nil {
		return pool
	}

	pool, err = conn.StoragePoolDefineXML(`<pool type='dir'>
                                                 <name>default</name>
                                                 <target>
                                                   <path>/var/lib/libvirt/images</path>
                                                 </target>
                                               </pool>`, 0)
	if err := pool.Create(0); err != nil {
		pool.Undefine()
		panic(err)
	}

	return pool
}

func getOrCreateStorageVol(pool *StoragePool, name string, size int64) *StorageVol {
	vol, err := pool.LookupStorageVolByName(name)
	if err == nil {
		return vol
	}

	vol, err = pool.StorageVolCreateXML(fmt.Sprintf(
		`<volume type="file">
                   <name>%s</name>
                   <allocation unit="b">%d</allocation>
                   <capacity unit="b">%d</capacity>
                   </volume>`, name, size, size), 0)
	if err != nil {
		panic(err)
	}

	return vol
}

func TestMultipleCloseCallback(t *testing.T) {
	nbCall1 := 0
	nbCall2 := 0
	nbCall3 := 0
	conn := buildTestQEMUConnection()
	defer func() {
		res, _ := conn.Close()
		// Blacklist versions of libvirt which had a ref counting
		// bug wrt close callbacks
		if VERSION_NUMBER <= 1002019 || VERSION_NUMBER >= 1003003 {
			if res != 0 {
				t.Errorf("Close() == %d, expected 0", res)
			}
		}
		if nbCall1 != 0 || nbCall2 != 0 || nbCall3 != 1 {
			t.Errorf("Wrong number of calls to callback, got %v, expected %v",
				[]int{nbCall1, nbCall2, nbCall3},
				[]int{0, 0, 1})
		}
	}()

	callback := func(conn *Connect, reason ConnectCloseReason) {
		if reason != CONNECT_CLOSE_REASON_KEEPALIVE {
			t.Errorf("Expected close reason to be %d, got %d",
				CONNECT_CLOSE_REASON_KEEPALIVE, reason)
		}
	}
	err := conn.RegisterCloseCallback(func(conn *Connect, reason ConnectCloseReason) {
		callback(conn, reason)
		nbCall1++
	})
	if err != nil {
		t.Fatalf("Unable to register close callback: %+v", err)
	}
	err = conn.RegisterCloseCallback(func(conn *Connect, reason ConnectCloseReason) {
		callback(conn, reason)
		nbCall2++
	})
	if err != nil {
		t.Fatalf("Unable to register close callback: %+v", err)
	}
	err = conn.RegisterCloseCallback(func(conn *Connect, reason ConnectCloseReason) {
		callback(conn, reason)
		nbCall3++
	})
	if err != nil {
		t.Fatalf("Unable to register close callback: %+v", err)
	}

	// To trigger a disconnect, we use a keepalive
	if err := conn.SetKeepAlive(1, 1); err != nil {
		t.Fatalf("Unable to enable keeplive: %+v", err)
	}
	EventRunDefaultImpl()
	time.Sleep(2 * time.Second)
	EventRunDefaultImpl()
}

func TestUnregisterCloseCallback(t *testing.T) {
	nbCall := 0
	conn := buildTestQEMUConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
		if nbCall != 0 {
			t.Errorf("Expected no call to close callback, got %d", nbCall)
		}
	}()

	callback := func(conn *Connect, reason ConnectCloseReason) {
		nbCall++
	}
	err := conn.RegisterCloseCallback(callback)
	if err != nil {
		t.Fatalf("Unable to register close callback: %+v", err)
	}
	err = conn.UnregisterCloseCallback()
	if err != nil {
		t.Fatalf("Unable to unregister close callback: %+v", err)
	}
}

func TestSetKeepalive(t *testing.T) {
	EventRegisterDefaultImpl()        // We need the event loop for keepalive
	conn := buildTestQEMUConnection() // The test driver doesn't support keepalives
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := conn.SetKeepAlive(1, 1); err != nil {
		t.Error(err)
		return
	}

	// It should block until we have a keepalive message
	done := make(chan struct{})
	timeout := time.After(5 * time.Second)
	go func() {
		EventRunDefaultImpl()
		close(done)
	}()
	select {
	case <-done: // OK!
	case <-timeout:
		t.Fatalf("timeout reached while waiting for keepalive")
	}
}

func TestConnectionWithAuth(t *testing.T) {
	callback := func(creds []*ConnectCredential) {
		for _, cred := range creds {
			if cred.Type == CRED_AUTHNAME {
				cred.Result = "user"
				cred.ResultLen = len(cred.Result)
			} else if cred.Type == CRED_PASSPHRASE {
				cred.Result = "pass"
				cred.ResultLen = len(cred.Result)
			}
		}
	}
	auth := &ConnectAuth{
		CredType: []ConnectCredentialType{
			CRED_AUTHNAME, CRED_PASSPHRASE,
		},
		Callback: callback,
	}
	conn, err := NewConnectWithAuth("test+tcp://127.0.0.1/default", auth, 0)
	if err != nil {
		t.Error(err)
		return
	}
	res, err := conn.Close()
	if err != nil {
		t.Error(err)
		return
	}
	if res != 0 {
		t.Errorf("Close() == %d, expected 0", res)
	}
}

func TestConnectionWithWrongCredentials(t *testing.T) {
	callback := func(creds []*ConnectCredential) {
		for _, cred := range creds {
			if cred.Type == CRED_AUTHNAME {
				cred.Result = "user"
				cred.ResultLen = len(cred.Result)
			} else if cred.Type == CRED_PASSPHRASE {
				cred.Result = "wrongpass"
				cred.ResultLen = len(cred.Result)
			}
		}
	}
	auth := &ConnectAuth{
		CredType: []ConnectCredentialType{
			CRED_AUTHNAME, CRED_PASSPHRASE,
		},
		Callback: callback,
	}
	conn, err := NewConnectWithAuth("test+tcp://127.0.0.1/default", auth, 0)
	if err == nil {
		conn.Close()
		t.Error(err)
		return
	}
}

func TestQemuMonitorCommand(t *testing.T) {
	dom, conn := buildTestQEMUDomain(false, "monitor")
	defer func() {
		dom.Destroy()
		dom.Undefine()
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()

	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}

	if _, err := dom.QemuMonitorCommand("{\"execute\" : \"query-cpus\"}", DOMAIN_QEMU_MONITOR_COMMAND_DEFAULT); err != nil {
		t.Error(err)
		return
	}

	if _, err := dom.QemuMonitorCommand("info cpus", DOMAIN_QEMU_MONITOR_COMMAND_HMP); err != nil {
		t.Error(err)
		return
	}
}

func TestDomainCreateWithFlags(t *testing.T) {
	dom, conn := buildTestQEMUDomain(false, "create")
	defer func() {
		dom.Destroy()
		dom.Undefine()
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()

	if err := dom.CreateWithFlags(DOMAIN_START_PAUSED); err != nil {
		state, reason, err := dom.GetState()
		if err != nil {
			t.Error(err)
			return
		}

		if state != DOMAIN_PAUSED {
			t.Fatalf("Domain should be paused")
			return
		}
		if DomainPausedReason(reason) != DOMAIN_PAUSED_STARTING_UP {
			t.Fatal("Domain reason should be starting up")
			return
		}
	}
}

func defineTestLxcDomain(conn *Connect, name string) (*Domain, error) {
	fullname := "libvirt-go-test-" + name
	dom, err := conn.LookupDomainByName(fullname)
	if err == nil {
		dom.Destroy()
		dom.Undefine()
		dom.Free()
	}

	xml := `<domain type='lxc'>
	  <name>` + fullname + `</name>
	  <title>` + name + `</title>
	  <memory>102400</memory>
	  <os>
	    <type>exe</type>
	    <init>/bin/sh</init>
	  </os>
	  <devices>
	    <console type='pty'/>
	  </devices>
	</domain>`
	dom, err = conn.DomainDefineXML(xml)
	return dom, err
}

// Integration tests are run against LXC using Libvirt 1.2.x
// on Debian Wheezy (libvirt from wheezy-backports)
//
// To run,
// 		go test -tags integration

func TestIntegrationGetMetadata(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	dom, err := defineTestLxcDomain(conn, "meta")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		dom.Undefine()
		dom.Destroy()
		dom.Free()
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}
	v, err := dom.GetMetadata(DOMAIN_METADATA_TITLE, "", 0)
	if err != nil {
		t.Error(err)
		return
	}
	if v != "meta" {
		t.Fatalf("title didnt match: expected meta, got %s", v)
		return
	}
}

func TestIntegrationSetMetadata(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	dom, err := defineTestLxcDomain(conn, "meta")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		dom.Undefine()
		dom.Free()
	}()
	const domTitle = "newtitle"
	if err := dom.SetMetadata(DOMAIN_METADATA_TITLE, domTitle, "", "", 0); err != nil {
		t.Error(err)
		return
	}
	v, err := dom.GetMetadata(DOMAIN_METADATA_TITLE, "", 0)
	if err != nil {
		t.Error(err)
		return
	}
	if v != domTitle {
		t.Fatalf("DOMAIN_METADATA_TITLE should have been %s, not %s", domTitle, v)
		return
	}
}

func TestIntegrationGetSysinfo(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	info, err := conn.GetSysinfo(0)
	if err != nil {
		t.Error(err)
		return
	}
	if strings.Index(info, "<sysinfo") != 0 {
		t.Fatalf("Sysinfo not valid: %s", info)
		return
	}
}

func testNWFilterXML(name, chain string) string {
	defName := name
	if defName == "" {
		defName = time.Now().String()
	}
	return `<filter name='` + defName + `' chain='` + chain + `'>
            <rule action='drop' direction='out' priority='500'>
            <ip match='no' srcipaddr='$IP'/>
            </rule>
			</filter>`
}

func TestIntergrationDefineUndefineNWFilterXML(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	filter, err := conn.NWFilterDefineXML(testNWFilterXML("", "ipv4"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := filter.Undefine(); err != nil {
			t.Fatal(err)
		}
		filter.Free()
	}()
	_, err = conn.NWFilterDefineXML(testNWFilterXML("", "bad"))
	if err == nil {
		t.Fatal("Should have had an error")
	}
}

func TestIntegrationNWFilterGetName(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	filter, err := conn.NWFilterDefineXML(testNWFilterXML("", "ipv4"))
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		filter.Undefine()
		filter.Free()
	}()
	if _, err := filter.GetName(); err != nil {
		t.Error(err)
	}
}

func TestIntegrationNWFilterGetUUID(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	filter, err := conn.NWFilterDefineXML(testNWFilterXML("", "ipv4"))
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		filter.Undefine()
		filter.Free()
	}()
	if _, err := filter.GetUUID(); err != nil {
		t.Error(err)
	}
}

func TestIntegrationNWFilterGetUUIDString(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	filter, err := conn.NWFilterDefineXML(testNWFilterXML("", "ipv4"))
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		filter.Undefine()
		filter.Free()
	}()
	if _, err := filter.GetUUIDString(); err != nil {
		t.Error(err)
	}
}

func TestIntegrationNWFilterGetXMLDesc(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	filter, err := conn.NWFilterDefineXML(testNWFilterXML("", "ipv4"))
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		filter.Undefine()
		filter.Free()
	}()
	if _, err := filter.GetXMLDesc(0); err != nil {
		t.Error(err)
	}
}

func TestIntegrationLookupNWFilterByName(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	origName := time.Now().String()
	filter, err := conn.NWFilterDefineXML(testNWFilterXML(origName, "ipv4"))
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		filter.Undefine()
		filter.Free()
	}()
	filter2, err := conn.LookupNWFilterByName(origName)
	if err != nil {
		t.Error(err)
		return
	}
	defer filter2.Free()
	var newName string
	newName, err = filter2.GetName()
	if err != nil {
		t.Error(err)
		return
	}
	if newName != origName {
		t.Fatalf("expected filter name: %s ,got: %s", origName, newName)
	}
}

func TestIntegrationLookupNWFilterByUUIDString(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	origName := time.Now().String()
	filter, err := conn.NWFilterDefineXML(testNWFilterXML(origName, "ipv4"))
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		filter.Undefine()
		filter.Free()
	}()
	filter2, err := conn.LookupNWFilterByName(origName)
	if err != nil {
		t.Error(err)
		return
	}
	defer filter2.Free()
	var filterUUID string
	filterUUID, err = filter2.GetUUIDString()
	if err != nil {
		t.Error(err)
		return
	}
	filter3, err := conn.LookupNWFilterByUUIDString(filterUUID)
	defer filter3.Free()
	if err != nil {
		t.Error(err)
		return
	}
	name, err := filter3.GetName()
	if err != nil {
		t.Error(err)
		return
	}
	if name != origName {
		t.Fatalf("fetching by UUID: expected filter name: %s ,got: %s", name, origName)
	}
}

func TestIntegrationDomainAttachDetachDevice(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()

	dom, err := defineTestLxcDomain(conn, "attach")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		dom.Undefine()
		dom.Free()
	}()
	const nwXml = `<interface type='network'>
		<mac address='52:54:00:37:aa:c7'/>
		<source network='default'/>
		<model type='virtio'/>
		</interface>`
	if err := dom.AttachDeviceFlags(nwXml, DOMAIN_DEVICE_MODIFY_CONFIG); err != nil {
		t.Error(err)
		return
	}
	if err := dom.DetachDeviceFlags(nwXml, DOMAIN_DEVICE_MODIFY_CONFIG); err != nil {
		t.Error(err)
		return
	}
}

func TestStorageVolResize(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()

	poolPath, err := ioutil.TempDir("", "default-pool-test-1")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(poolPath)
	pool, err := conn.StoragePoolDefineXML(`<pool type='dir'>
                                          <name>default-pool-test-1</name>
                                          <target>
                                          <path>`+poolPath+`</path>
                                          </target>
                                          </pool>`, 0)
	defer func() {
		pool.Undefine()
		pool.Free()
	}()
	if err := pool.Create(0); err != nil {
		t.Error(err)
		return
	}
	defer pool.Destroy()
	vol, err := pool.StorageVolCreateXML(testStorageVolXML("", poolPath), 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		vol.Delete(STORAGE_VOL_DELETE_NORMAL)
		vol.Free()
	}()
	const newCapacityInBytes = 12582912
	if err := vol.Resize(newCapacityInBytes, STORAGE_VOL_RESIZE_ALLOCATE); err != nil {
		t.Fatal(err)
	}
}

func TestStorageVolWipe(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()

	poolPath, err := ioutil.TempDir("", "default-pool-test-1")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(poolPath)
	pool, err := conn.StoragePoolDefineXML(`<pool type='dir'>
                                          <name>default-pool-test-1</name>
                                          <target>
                                          <path>`+poolPath+`</path>
                                          </target>
                                          </pool>`, 0)
	defer func() {
		pool.Undefine()
		pool.Free()
	}()
	if err := pool.Create(0); err != nil {
		t.Error(err)
		return
	}
	defer pool.Destroy()
	vol, err := pool.StorageVolCreateXML(testStorageVolXML("", poolPath), 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		vol.Delete(STORAGE_VOL_DELETE_NORMAL)
		vol.Free()
	}()
	if err := vol.Wipe(0); err != nil {
		t.Fatal(err)
	}
}

func TestStorageVolWipePattern(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()

	poolPath, err := ioutil.TempDir("", "default-pool-test-1")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(poolPath)
	pool, err := conn.StoragePoolDefineXML(`<pool type='dir'>
                                          <name>default-pool-test-1</name>
                                          <target>
                                          <path>`+poolPath+`</path>
                                          </target>
                                          </pool>`, 0)
	defer func() {
		pool.Undefine()
		pool.Free()
	}()
	if err := pool.Create(0); err != nil {
		t.Error(err)
		return
	}
	defer pool.Destroy()
	vol, err := pool.StorageVolCreateXML(testStorageVolXML("", poolPath), 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		vol.Delete(STORAGE_VOL_DELETE_NORMAL)
		vol.Free()
	}()
	if err := vol.WipePattern(STORAGE_VOL_WIPE_ALG_ZERO, 0); err != nil {
		t.Fatal(err)
	}
}

func testSecretTypeCephFromXML(name string) string {
	var setName string
	if name == "" {
		setName = time.Now().String()
	} else {
		setName = name
	}
	return `<secret ephemeral='no' private='no'>
            <usage type='ceph'>
            <name>` + setName + `</name>
            </usage>
            </secret>`
}

func TestIntegrationSecretDefineUndefine(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	sec, err := conn.SecretDefineXML(testSecretTypeCephFromXML(""), 0)
	if err != nil {
		t.Fatal(err)
	}
	defer sec.Free()

	if err := sec.Undefine(); err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationSecretGetUUID(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	sec, err := conn.SecretDefineXML(testSecretTypeCephFromXML(""), 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		sec.Undefine()
		sec.Free()
	}()
	if _, err := sec.GetUUID(); err != nil {
		t.Error(err)
	}
}

func TestIntegrationSecretGetUUIDString(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	sec, err := conn.SecretDefineXML(testSecretTypeCephFromXML(""), 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		sec.Undefine()
		sec.Free()
	}()
	if _, err := sec.GetUUIDString(); err != nil {
		t.Error(err)
	}
}

func TestIntegrationSecretGetXMLDesc(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	sec, err := conn.SecretDefineXML(testSecretTypeCephFromXML(""), 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		sec.Undefine()
		sec.Free()
	}()
	if _, err := sec.GetXMLDesc(0); err != nil {
		t.Error(err)
	}
}

func TestIntegrationSecretGetUsageType(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	sec, err := conn.SecretDefineXML(testSecretTypeCephFromXML(""), 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		sec.Undefine()
		sec.Free()
	}()
	uType, err := sec.GetUsageType()
	if err != nil {
		t.Error(err)
		return
	}
	if uType != SECRET_USAGE_TYPE_CEPH {
		t.Fatal("unexpected usage type.Expected usage type is Ceph")
	}
}

func TestIntegrationSecretGetUsageID(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	setUsageID := time.Now().String()
	sec, err := conn.SecretDefineXML(testSecretTypeCephFromXML(setUsageID), 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		sec.Undefine()
		sec.Free()
	}()
	recUsageID, err := sec.GetUsageID()
	if err != nil {
		t.Error(err)
		return
	}
	if recUsageID != setUsageID {
		t.Fatalf("exepected usage ID: %s, got: %s", setUsageID, recUsageID)
	}
}

func TestIntegrationLookupSecretByUsage(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	usageID := time.Now().String()
	sec, err := conn.SecretDefineXML(testSecretTypeCephFromXML(usageID), 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		sec.Undefine()
		sec.Free()
	}()
	sec2, err := conn.LookupSecretByUsage(SECRET_USAGE_TYPE_CEPH, usageID)
	if err != nil {
		t.Fatal(err)
	}
	sec2.Free()
}

func TestIntegrationGetDomainCPUStats(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	dom, err := defineTestLxcDomain(conn, "cpustats")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		dom.Undefine()
		dom.Free()
	}()

	if err := dom.Create(); err != nil {
		t.Fatal(err)
	}
	defer dom.Destroy()

	stats, err := dom.GetCPUStats(0, 0, 0)
	if err != nil {
		lverr, ok := err.(Error)
		if ok && lverr.Code == ERR_NO_SUPPORT {
			return
		}
		t.Fatal(err)
	}

	if len(stats) < 1 {
		t.Errorf("Expected stats for at least one CPU")
	}

	if !stats[0].CpuTimeSet {
		t.Errorf("Expected CpuTime to be set")
	}
}

// Not supported on libvirt driver, so no integration test
// func TestGetInterfaceParameters(t *testing.T) {
// 	dom, conn := buildTestDomain()
// 	defer func() {
// 		dom.Undefine()
// 		dom.Free()
// 		if res, _ := conn.Close(); res != 0 {
// 			t.Errorf("Close() == %d, expected 0", res)
// 		}
// 	}()
// 	iface := "either mac or path to interface"
// 	nparams := int(0)
// 	if _, err := dom.GetInterfaceParameters(iface, nil, &nparams, 0); err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	var params VirTypedParameters
// 	if _, err := dom.GetInterfaceParameters(iface, &params, &nparams, 0); err != nil {
// 		t.Error(err)
// 		return
// 	}
// }

func TestIntegrationListAllInterfaces(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	ifaces, err := conn.ListAllInterfaces(0)
	if err != nil {
		t.Fatal(err)
	}
	lookingFor := "lo"
	found := false
	for _, iface := range ifaces {
		name, err := iface.GetName()
		if err != nil {
			t.Fatal(err)
		}
		if name == lookingFor {
			found = true
		}
		iface.Free()
	}
	if found == false {
		t.Fatalf("interface %s not found", lookingFor)
	}
}

func TestIntergrationListAllNWFilters(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()

	testNWFilterName := time.Now().String()
	filter, err := conn.NWFilterDefineXML(testNWFilterXML(testNWFilterName, "ipv4"))
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		filter.Undefine()
		filter.Free()
	}()

	filters, err := conn.ListAllNWFilters(0)
	if len(filters) == 0 {
		t.Fatal("length of []NWFilter shouldn't be 0")
	}

	found := false
	for _, f := range filters {
		name, _ := f.GetName()
		if name == testNWFilterName {
			found = true
		}
		f.Free()
	}
	if found == false {
		t.Fatalf("NWFilter %s not found", testNWFilterName)
	}
}

func TestIntegrationDomainInterfaceStats(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()

	net, err := conn.LookupNetworkByName("default")
	if err != nil {
		return
	}

	defer net.Free()

	active, err := net.IsActive()
	if err != nil {
		t.Fatal(err)
	}

	if !active {
		return
	}

	dom, err := defineTestLxcDomain(conn, "ifacestats")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		dom.Undefine()
		dom.Free()
	}()
	const nwXml = `<interface type='network'>
		<mac address='52:54:00:37:aa:c7'/>
		<source network='default'/>
		<model type='virtio'/>
                <target dev="lvgotest0"/>
		</interface>`
	if err := dom.AttachDeviceFlags(nwXml, DOMAIN_DEVICE_MODIFY_CONFIG); err != nil {
		t.Fatal(err)
	}

	if err := dom.Create(); err != nil {
		t.Fatal(err)
	}

	if _, err := dom.InterfaceStats("lvgotest0"); err != nil {
		t.Error(err)
	}

	if err := dom.Destroy(); err != nil {
		t.Fatal(err)
	}

	if err := dom.DetachDeviceFlags(nwXml, DOMAIN_DEVICE_MODIFY_CONFIG); err != nil {
		t.Fatal(err)
	}
}

func TestStorageVolUploadDownload(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()

	poolPath, err := ioutil.TempDir("", "default-pool-test-1")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(poolPath)
	pool, err := conn.StoragePoolDefineXML(`<pool type='dir'>
                                          <name>default-pool-test-1</name>
                                          <target>
                                          <path>`+poolPath+`</path>
                                          </target>
                                          </pool>`, 0)
	defer func() {
		pool.Undefine()
		pool.Free()
	}()
	if err := pool.Create(0); err != nil {
		t.Error(err)
		return
	}
	defer pool.Destroy()
	vol, err := pool.StorageVolCreateXML(testStorageVolXML("", poolPath), 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		vol.Delete(STORAGE_VOL_DELETE_NORMAL)
		vol.Free()
	}()

	data := []byte{1, 2, 3, 4, 5, 6}

	// write above data to the vol
	// 1. create a stream
	stream, err := conn.NewStream(0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		stream.Free()
	}()

	// 2. set it up to upload from stream
	if err := vol.Upload(stream, 0, uint64(len(data)), 0); err != nil {
		stream.Abort()
		t.Fatal(err)
	}

	// 3. do the actual writing
	if n, err := stream.Send(data); err != nil || n != len(data) {
		t.Fatal(err, n)
	}

	// 4. finish!
	if err := stream.Finish(); err != nil {
		t.Fatal(err)
	}

	// read back the data
	// 1. create a stream
	downStream, err := conn.NewStream(0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		downStream.Free()
	}()

	// 2. set it up to download from stream
	if err := vol.Download(downStream, 0, uint64(len(data)), 0); err != nil {
		downStream.Abort()
		t.Fatal(err)
	}

	// 3. do the actual reading
	buf := make([]byte, 1024)
	if n, err := downStream.Recv(buf); err != nil || n != len(data) {
		t.Fatal(err, n)
	}

	t.Logf("read back: %#v", buf[:len(data)])

	// 4. finish!
	if err := downStream.Finish(); err != nil {
		t.Fatal(err)
	}
}

func TestStorageVolUploadDownloadCallbacks(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()

	poolPath, err := ioutil.TempDir("", "default-pool-test-1")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(poolPath)
	pool, err := conn.StoragePoolDefineXML(`<pool type='dir'>
                                          <name>default-pool-test-1</name>
                                          <target>
                                          <path>`+poolPath+`</path>
                                          </target>
                                          </pool>`, 0)
	defer func() {
		pool.Undefine()
		pool.Free()
	}()
	if err := pool.Create(0); err != nil {
		t.Error(err)
		return
	}
	defer pool.Destroy()
	vol, err := pool.StorageVolCreateXML(testStorageVolXML("", poolPath), 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		vol.Delete(STORAGE_VOL_DELETE_NORMAL)
		vol.Free()
	}()

	input := make([]byte, 1024*1024)
	for i := 0; i < len(input); i++ {
		input[i] = (byte)(((i % 256) ^ (i / 256)) % 256)
	}

	// write above data to the vol
	// 1. create a stream
	stream, err := conn.NewStream(0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		stream.Free()
	}()

	// 2. set it up to upload from stream
	if err := vol.Upload(stream, 0, uint64(len(input)), 0); err != nil {
		stream.Abort()
		t.Fatal(err)
	}

	sent := 0
	source := func(stream *Stream, nbytes int) ([]byte, error) {
		tosend := nbytes
		if tosend > (len(input) - sent) {
			tosend = len(input) - sent
		}

		if tosend == 0 {
			return []byte{}, nil
		}

		data := input[sent : sent+tosend]
		sent += tosend
		return data, nil
	}

	// 3. do the actual writing
	if err := stream.SendAll(source); err != nil {
		t.Fatal(err)
	}

	if sent != len(input) {
		t.Fatal("Wanted %d but only sent %d bytes",
			len(input), sent)
	}

	// 4. finish!
	if err := stream.Finish(); err != nil {
		t.Fatal(err)
	}

	// read back the data
	// 1. create a stream
	downStream, err := conn.NewStream(0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		downStream.Free()
	}()

	// 2. set it up to download from stream
	if err := vol.Download(downStream, 0, uint64(len(input)), 0); err != nil {
		downStream.Abort()
		t.Fatal(err)
	}

	// 3. do the actual reading
	output := make([]byte, len(input))

	got := 0
	sink := func(st *Stream, data []byte) (int, error) {
		toget := len(data)
		if (got + toget) > len(output) {
			toget = len(output) - got
		}
		if toget == 0 {
			return 0, fmt.Errorf("Output buffer is full")
		}

		target := output[got : got+toget]
		copied := copy(target, data)
		got += copied

		return copied, nil
	}

	if err := downStream.RecvAll(sink); err != nil {
		t.Fatal(err)
	}

	if got != len(input) {
		t.Fatalf("Wanted %d but only received %d bytes",
			len(input), got)
	}

	// 4. finish!
	if err := downStream.Finish(); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(input, output) {
		t.Fatal("Input and output arrays are different")
	}
}

/*func TestDomainMemoryStats(t *testing.T) {
	conn, err := NewConnect("lxc:///")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()

	dom, err := defineTestLxcDomain(conn, "memstats")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		dom.Undefine()
		dom.Free()
	}()
	if err := dom.Create(); err != nil {
		t.Fatal(err)
	}
	defer dom.Destroy()

	ms, err := dom.MemoryStats(1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) != 1 {
		t.Fatal("Should have got one result, got", len(ms))
	}
}*/

func TestDomainListAllInterfaceAddresses(t *testing.T) {
	dom, conn := buildTestQEMUDomain(false, "ifaces")
	defer func() {
		dom.Destroy()
		dom.Undefine()
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}

	ifaces, err := dom.ListAllInterfaceAddresses(0)
	if err != nil {
		lverr, ok := err.(Error)
		if ok && lverr.Code == ERR_NO_SUPPORT {
			return
		}
		t.Fatal(err)
	}

	if len(ifaces) != 0 {
		t.Fatal("should have 0 interfaces", len(ifaces))
	}
}

func TestDomainGetAllStats(t *testing.T) {
	dom, conn := buildTestQEMUDomain(false, "stats")
	defer func() {
		dom.Destroy()
		dom.Undefine()
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := dom.Create(); err != nil {
		t.Error(err)
		return
	}

	stats, err := conn.GetAllDomainStats([]*Domain{}, DOMAIN_STATS_STATE|DOMAIN_STATS_CPU_TOTAL|DOMAIN_STATS_INTERFACE|DOMAIN_STATS_BALLOON|DOMAIN_STATS_BLOCK|DOMAIN_STATS_PERF|DOMAIN_STATS_VCPU, 0)

	if err != nil {
		lverr, ok := err.(Error)
		if ok && lverr.Code == ERR_NO_SUPPORT {
			return
		}
		t.Error(err)
		return
	}

	for _, stat := range stats {
		stat.Domain.Free()
	}
}

func TestDomainBlockCopy(t *testing.T) {
	if VERSION_NUMBER < 1002008 {
		return
	}
	dom, conn := buildTestQEMUDomain(true, "blockcopy")
	defer func() {
		dom.Destroy()
		dom.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()

	pool := getDefaultStoragePool(conn)
	defer pool.Free()

	srcVol := getOrCreateStorageVol(pool, "libvirt-go-test-block-copy-src.img", 1024*1024*10)
	defer func() {
		srcVol.Delete(0)
		srcVol.Free()
	}()

	dstVol := getOrCreateStorageVol(pool, "libvirt-go-test-block-copy-dst.img", 1024*1024*10)
	defer func() {
		dstVol.Delete(0)
		dstVol.Free()
	}()

	srcPath, err := srcVol.GetPath()
	if err != nil {
		t.Error(err)
		return
	}
	dstPath, err := dstVol.GetPath()
	if err != nil {
		t.Error(err)
		return
	}

	params := DomainBlockCopyParameters{
		BandwidthSet:   true,
		Bandwidth:      2147483648,
		GranularitySet: true,
		Granularity:    512,
	}

	srcXML := fmt.Sprintf(`<disk type="file">
                                 <driver type="raw"/>
                                 <source file="%s"/>
                                 <target dev="vda"/>
                               </disk>`, srcPath)

	err = dom.AttachDeviceFlags(srcXML, DOMAIN_DEVICE_MODIFY_LIVE)
	if err != nil {
		t.Error(err)
		return
	}

	dstXML := fmt.Sprintf(`<disk type='file'>
                                  <driver type='raw'/>
                                  <source file='%s'/>
                               </disk>`, dstPath)

	err = dom.BlockCopy(srcPath, dstXML, &params, DOMAIN_BLOCK_COPY_REUSE_EXT)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestNodeGetMemoryStats(t *testing.T) {

	c := buildTestQEMUConnection()
	defer c.Close()

	stats, err := c.GetMemoryStats(0, 0)

	if err != nil {
		t.Error(err)
		return
	}

	if stats.TotalSet && stats.Total == 0 {
		t.Error("Expected non-zero total memory")
	}
	if stats.FreeSet && stats.Free == 0 {
		t.Error("Expected non-zero free memory")
	}
}

func TestNodeGetCPUStats(t *testing.T) {

	c := buildTestQEMUConnection()
	defer c.Close()

	stats, err := c.GetCPUStats(0, 0)

	if err != nil {
		t.Error(err)
		return
	}

	if stats.KernelSet && stats.Kernel == 0 {
		t.Error("Expected non-zero kernel time")
	}
	if stats.UserSet && stats.User == 0 {
		t.Error("Expected non-zero user time")
	}
}
