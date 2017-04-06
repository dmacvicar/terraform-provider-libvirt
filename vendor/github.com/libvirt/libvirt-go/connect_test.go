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

func buildTestConnection() *Connect {
	conn, err := NewConnect("test:///default")
	if err != nil {
		panic(err)
	}
	return conn
}

func TestVersion(t *testing.T) {
	version, err := GetVersion()
	if err != nil {
		t.Error(err)
		return
	}
	if version == 0 {
		t.Error("Version was 0")
		return
	}
}

func TestConnection(t *testing.T) {
	conn, err := NewConnect("test:///default")
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

func TestConnectionReadOnly(t *testing.T) {
	conn, err := NewConnectReadOnly("test:///default")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()

	_, err = conn.NetworkDefineXML(`<network>
    <name>` + time.Now().String() + `</name>
    <bridge name="testbr0"/>
    <forward/>
    <ip address="192.168.0.1" netmask="255.255.255.0">
    </ip>
    </network>`)
	if err == nil {
		t.Fatal("writing on a read only connection")
	}
}

func TestInvalidConnection(t *testing.T) {
	_, err := NewConnect("invalid_transport:///default")
	if err == nil {
		t.Error("Non-existent transport works")
	}
}

func TestGetType(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	tp, err := conn.GetType()
	if err != nil {
		t.Error(err)
		return
	}
	if strings.ToLower(tp) != "test" {
		t.Fatalf("type should have been \"test\" but got %q", tp)
		return
	}
}

func TestIsAlive(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	alive, err := conn.IsAlive()
	if err != nil {
		t.Error(err)
		return
	}
	if !alive {
		t.Fatal("Connection should be alive")
		return
	}
}

func TestIsEncryptedAndSecure(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	secure, err := conn.IsSecure()
	if err != nil {
		t.Log(err)
		return
	}
	enc, err := conn.IsEncrypted()
	if err != nil {
		t.Error(err)
		return
	}
	if !secure {
		t.Fatal("Test driver should be secure")
		return
	}
	if enc {
		t.Fatal("Test driver should not be encrypted")
		return
	}
}

func TestCapabilities(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	capabilities, err := conn.GetCapabilities()
	if err != nil {
		t.Error(err)
		return
	}
	if capabilities == "" {
		t.Error("Capabilities was empty")
		return
	}
}

func TestGetNodeInfo(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	ni, err := conn.GetNodeInfo()
	if err != nil {
		t.Error(err)
		return
	}
	if ni.Model != "i686" {
		t.Error("Expected i686 model in test transport")
		return
	}
}

func TestHostname(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	hostname, err := conn.GetHostname()
	if err != nil {
		t.Error(err)
		return
	}
	if hostname == "" {
		t.Error("Hostname was empty")
		return
	}
}

func TestLibVersion(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	version, err := conn.GetLibVersion()
	if err != nil {
		t.Error(err)
		return
	}
	if version == 0 {
		t.Error("Version was 0")
		return
	}
}

func TestListDefinedDomains(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	doms, err := conn.ListDefinedDomains()
	if err != nil {
		t.Error(err)
		return
	}
	if doms == nil {
		t.Fatal("ListDefinedDomains shouldn't be nil")
		return
	}
}

func TestListDomains(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	doms, err := conn.ListDomains()
	if err != nil {
		t.Error(err)
		return
	}
	if doms == nil {
		t.Fatal("ListDomains shouldn't be nil")
		return
	}
}

func TestListInterfaces(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	_, err := conn.ListInterfaces()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestListNetworks(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	_, err := conn.ListNetworks()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestListStoragePools(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	_, err := conn.ListStoragePools()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestLookupDomainById(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	ids, err := conn.ListDomains()
	if err != nil {
		t.Error(err)
		return
	}

	if len(ids) == 0 {
		t.Fatal("Length of ListDomains shouldn't be zero")
		return
	}
	dom, err := conn.LookupDomainById(ids[0])
	if err != nil {
		t.Error(err)
		return
	}
	defer dom.Free()
}

func TestLookupDomainByUUIDString(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	doms, err := conn.ListAllDomains(0)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		for _, dom := range doms {
			dom.Free()
		}
	}()

	if len(doms) == 0 {
		t.Fatal("Length of ListAllDomains shouldn't be empty")
		return
	}
	uuid, err := doms[0].GetUUIDString()
	if err != nil {
		t.Error(err)
		return
	}
	dom, err := conn.LookupDomainByUUIDString(uuid)
	if err != nil {
		t.Error(err)
		return
	}
	defer dom.Free()

	rawuuid, err := doms[0].GetUUID()
	if err != nil {
		t.Error(err)
		return
	}
	dom, err = conn.LookupDomainByUUID(rawuuid)
	if err != nil {
		t.Error(err)
		return
	}
	defer dom.Free()
}

func TestLookupInvalidDomainById(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	_, err := conn.LookupDomainById(12345)
	if err == nil {
		t.Error("Domain #12345 shouldn't exist in test transport")
		return
	}
}

func TestLookupDomainByName(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	dom, err := conn.LookupDomainByName("test")
	if err != nil {
		t.Error(err)
		return
	}
	defer dom.Free()
}

func TestLookupInvalidDomainByName(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	_, err := conn.LookupDomainByName("non_existent_domain")
	if err == nil {
		t.Error("Could find non-existent domain by name")
		return
	}
}

func TestDomainCreateXML(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	// Test a minimally valid xml
	defName := time.Now().String()
	xml := `<domain type="test">
		<name>` + defName + `</name>
		<memory unit="KiB">8192</memory>
		<os>
			<type>hvm</type>
		</os>
	</domain>`
	dom, err := conn.DomainCreateXML(xml, DOMAIN_NONE)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		dom.Destroy()
		dom.Free()
	}()
	name, err := dom.GetName()
	if err != nil {
		t.Error(err)
		return
	}
	if name != defName {
		t.Fatalf("Name was not '%s': %s", defName, name)
		return
	}

	// Destroy the domain: it should not be persistent
	if err := dom.Destroy(); err != nil {
		t.Error(err)
		return
	}

	testeddom, err := conn.LookupDomainByName(defName)
	if err == nil {
		testeddom.Free()
		t.Fatal("Created domain is persisting")
		return
	}
}

func TestDomainDefineXML(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	// Test a minimally valid xml
	defName := time.Now().String()
	xml := `<domain type="test">
		<name>` + defName + `</name>
		<memory unit="KiB">8192</memory>
		<os>
			<type>hvm</type>
		</os>
	</domain>`
	dom, err := conn.DomainDefineXML(xml)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		dom.Undefine()
		dom.Free()
	}()
	name, err := dom.GetName()
	if err != nil {
		t.Error(err)
		return
	}
	if name != defName {
		t.Fatalf("Name was not 'test': %s", name)
		return
	}
	// And an invalid one
	xml = `<domain type="test"></domain>`
	_, err = conn.DomainDefineXML(xml)
	if err == nil {
		t.Fatal("Should have had an error")
		return
	}
}

func TestListDefinedInterfaces(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	_, err := conn.ListDefinedInterfaces()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestListDefinedNetworks(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	_, err := conn.ListDefinedNetworks()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestListDefinedStoragePools(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	_, err := conn.ListDefinedStoragePools()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestNumOfDefinedInterfaces(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := conn.NumOfDefinedInterfaces(); err != nil {
		t.Error(err)
		return
	}
}

func TestNumOfDefinedNetworks(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := conn.NumOfDefinedNetworks(); err != nil {
		t.Error(err)
		return
	}
}

func TestNumOfDefinedStoragePools(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := conn.NumOfDefinedStoragePools(); err != nil {
		t.Error(err)
		return
	}
}

func TestNumOfDomains(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := conn.NumOfDomains(); err != nil {
		t.Error(err)
		return
	}
}

func TestNumOfInterfaces(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := conn.NumOfInterfaces(); err != nil {
		t.Error(err)
		return
	}
}

func TestNumOfNetworks(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := conn.NumOfNetworks(); err != nil {
		t.Error(err)
		return
	}
}

func TestNumOfNWFilters(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := conn.NumOfNWFilters(); err == nil {
		t.Fatalf("NumOfNWFilters should fail due to no support on test driver")
		return
	}
}

func TestNumOfSecrets(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := conn.NumOfSecrets(); err == nil {
		t.Fatalf("NumOfSecrets should fail due to no support on test driver")
		return
	}
}

func TestGetURI(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	uri, err := conn.GetURI()
	if err != nil {
		t.Error(err)
	}
	origUri := "test:///default"
	if uri != origUri {
		t.Fatalf("should be %s but got %s", origUri, uri)
	}
}

func TestGetMaxVcpus(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	_, err := conn.GetMaxVcpus("")
	if err != nil {
		t.Error(err)
	}
}

func TestInterfaceDefineXML(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	defName := "ethTest0"
	xml := `<interface type='ethernet' name='` + defName + `'><mac address='` + generateRandomMac() + `'/></interface>`
	iface, err := conn.InterfaceDefineXML(xml, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		iface.Undefine()
		iface.Free()
	}()
	name, err := iface.GetName()
	if err != nil {
		t.Error(err)
		return
	}
	if name != defName {
		t.Fatalf("Expected interface name: %s,got: %s", defName, name)
		return
	}
	// Invalid configuration
	xml = `<interface type="test"></interface>`
	_, err = conn.InterfaceDefineXML(xml, 0)
	if err == nil {
		t.Fatal("Should have had an error")
		return
	}
}

func TestLookupInterfaceByName(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	testEth := "eth1"
	iface, err := conn.LookupInterfaceByName(testEth)
	if err != nil {
		t.Error(err)
		return
	}
	defer iface.Free()
	var ifName string
	ifName, err = iface.GetName()
	if err != nil {
		t.Error(err)
		return
	}
	if ifName != testEth {
		t.Fatalf("expected interface name: %s ,got: %s", testEth, ifName)
	}
}

func TestLookupInterfaceByMACString(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	testMAC := "aa:bb:cc:dd:ee:ff"
	iface, err := conn.LookupInterfaceByMACString(testMAC)
	if err != nil {
		t.Error(err)
		return
	}
	defer iface.Free()
	var ifMAC string
	ifMAC, err = iface.GetMACString()
	if err != nil {
		t.Error(err)
		return
	}
	if ifMAC != testMAC {
		t.Fatalf("expected interface MAC: %s ,got: %s", testMAC, ifMAC)
	}
}

func TestStoragePoolDefineXML(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	defName := "default-pool-test-0"
	xml := `<pool type='dir'><name>default-pool-test-0</name><target>
            <path>/default-pool</path></target></pool>`
	pool, err := conn.StoragePoolDefineXML(xml, 0)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Free()
	defer pool.Undefine()
	name, err := pool.GetName()
	if err != nil {
		t.Error(err)
		return
	}
	if name != defName {
		t.Fatalf("Expected storage pool name: %s,got: %s", defName, name)
		return
	}
	// Invalid configuration
	xml = `<pool type='bad'></pool>`
	_, err = conn.StoragePoolDefineXML(xml, 0)
	if err == nil {
		t.Fatal("Should have had an error")
		return
	}
}

func TestLookupStoragePoolByName(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	testPool := "default-pool"
	pool, err := conn.LookupStoragePoolByName(testPool)
	if err != nil {
		t.Error(err)
		return
	}
	defer pool.Free()
	var poolName string
	poolName, err = pool.GetName()
	if err != nil {
		t.Error(err)
		return
	}
	if poolName != testPool {
		t.Fatalf("expected storage pool name: %s ,got: %s", testPool, poolName)
	}
}

func TestLookupStoragePoolByUUIDString(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	poolName := "default-pool"
	pool, err := conn.LookupStoragePoolByName(poolName)
	if err != nil {
		t.Error(err)
		return
	}
	defer pool.Free()
	var poolUUID string
	poolUUID, err = pool.GetUUIDString()
	if err != nil {
		t.Error(err)
		return
	}
	pool2, err := conn.LookupStoragePoolByUUIDString(poolUUID)
	if err != nil {
		t.Error(err)
		return
	}
	defer pool2.Free()
	name, err := pool2.GetName()
	if err != nil {
		t.Error(err)
	}
	if name != poolName {
		t.Fatalf("fetching by UUID: expected storage pool name: %s ,got: %s", name, poolName)
	}
	rawpoolUUID, err := pool.GetUUID()
	if err != nil {
		t.Error(err)
		return
	}
	pool2, err = conn.LookupStoragePoolByUUID(rawpoolUUID)
	if err != nil {
		t.Error(err)
		return
	}
	defer pool2.Free()
}

func TestLookupStorageVolByKey(t *testing.T) {
	pool, conn := buildTestStoragePool("")
	defer func() {
		pool.Undefine()
		pool.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := pool.Create(0); err != nil {
		t.Error(err)
		return
	}
	defer pool.Destroy()
	defPoolPath := "default-pool"
	defVolName := time.Now().String()
	defVolKey := "/" + defPoolPath + "/" + defVolName
	vol, err := pool.StorageVolCreateXML(testStorageVolXML(defVolName, defPoolPath), 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		vol.Delete(STORAGE_VOL_DELETE_NORMAL)
		vol.Free()
	}()
	vol2, err := conn.LookupStorageVolByKey(defVolKey)
	if err != nil {
		t.Error(err)
		return
	}
	defer vol2.Free()
	key, err := vol2.GetKey()
	if err != nil {
		t.Error(err)
		return
	}
	if key != defVolKey {
		t.Fatalf("expected storage volume key: %s ,got: %s", defVolKey, key)
	}
}

func TestLookupStorageVolByPath(t *testing.T) {
	pool, conn := buildTestStoragePool("")
	defer func() {
		pool.Undefine()
		pool.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := pool.Create(0); err != nil {
		t.Error(err)
		return
	}
	defer pool.Destroy()
	defPoolPath := "default-pool"
	defVolName := time.Now().String()
	defVolPath := "/" + defPoolPath + "/" + defVolName
	vol, err := pool.StorageVolCreateXML(testStorageVolXML(defVolName, defPoolPath), 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		vol.Delete(STORAGE_VOL_DELETE_NORMAL)
		vol.Free()
	}()
	vol2, err := conn.LookupStorageVolByPath(defVolPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer vol2.Free()
	path, err := vol2.GetPath()
	if err != nil {
		t.Error(err)
		return
	}
	if path != defVolPath {
		t.Fatalf("expected storage volume path: %s ,got: %s", defVolPath, path)
	}
}

func TestListAllDomains(t *testing.T) {
	conn := buildTestConnection()
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	doms, err := conn.ListAllDomains(CONNECT_LIST_DOMAINS_PERSISTENT)
	if err != nil {
		t.Error(err)
		return
	}
	if len(doms) == 0 {
		t.Fatal("length of []Domain shouldn't be 0")
	}
	testDomName := "test"
	found := false
	for _, dom := range doms {
		name, _ := dom.GetName()
		if name == testDomName {
			found = true
		}
		// not mandatory for the tests but lets make it in a proper way
		dom.Free()
	}
	if found == false {
		t.Fatalf("domain %s not found", testDomName)
	}
}

func TestListAllNetworks(t *testing.T) {
	testNetwork := time.Now().String()
	net, conn := buildTestNetwork(testNetwork)
	defer func() {
		// actually,no nicessaty to destroy as the network is being removed as soon as
		// the test connection is closed
		net.Destroy()
		net.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	nets, err := conn.ListAllNetworks(CONNECT_LIST_NETWORKS_INACTIVE)
	if err != nil {
		t.Fatal(err)
	}
	if len(nets) == 0 {
		t.Fatal("length of []Network shouldn't be 0")
	}
	found := false
	for _, n := range nets {
		name, _ := n.GetName()
		if name == testNetwork {
			found = true
		}
		n.Free()
	}
	if found == false {
		t.Fatalf("network %s not found", testNetwork)
	}
}

func TestListAllStoragePools(t *testing.T) {
	testStoragePool := "default-pool-test-1"
	pool, conn := buildTestStoragePool(testStoragePool)
	defer func() {
		pool.Undefine()
		pool.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	pools, err := conn.ListAllStoragePools(CONNECT_LIST_STORAGE_POOLS_INACTIVE)
	if err != nil {
		t.Fatal(err)
	}
	if len(pools) == 0 {
		t.Fatal("length of []StoragePool shouldn't be 0")
	}
	found := false
	for _, p := range pools {
		name, _ := p.GetName()
		if name == testStoragePool {
			found = true
		}
		p.Free()
	}
	if found == false {
		t.Fatalf("storage pool %s not found", testStoragePool)
	}
}
