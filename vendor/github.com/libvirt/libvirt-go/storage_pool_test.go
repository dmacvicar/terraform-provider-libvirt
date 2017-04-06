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
	"testing"
	"time"
)

func buildTestStoragePool(poolName string) (*StoragePool, *Connect) {
	conn := buildTestConnection()
	var name string
	if poolName == "" {
		name = "default-pool-test-1"
	} else {
		name = poolName
	}
	pool, err := conn.StoragePoolDefineXML(`<pool type='dir'>
  <name>`+name+`</name>
  <target>
  <path>/default-pool</path>
  </target>
  </pool>`, 0)
	if err != nil {
		panic(err)
	}
	return pool, conn
}

func TestStoragePoolBuild(t *testing.T) {
	pool, conn := buildTestStoragePool("")
	defer func() {
		pool.Undefine()
		pool.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := pool.Build(STORAGE_POOL_BUILD_NEW); err != nil {
		t.Fatal(err)
	}
}

func TestUndefineStoragePool(t *testing.T) {
	pool, conn := buildTestStoragePool("")
	defer func() {
		pool.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	name, err := pool.GetName()
	if err != nil {
		t.Error(err)
		return
	}
	if err := pool.Undefine(); err != nil {
		t.Error(err)
		return
	}
	if _, err := conn.LookupStoragePoolByName(name); err == nil {
		t.Fatal("Shouldn't have been able to find storage pool")
		return
	}
}

func TestGetStoragePoolName(t *testing.T) {
	pool, conn := buildTestStoragePool("")
	defer func() {
		pool.Undefine()
		pool.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := pool.GetName(); err != nil {
		t.Error(err)
	}
}

func TestGetStoragePoolUUID(t *testing.T) {
	pool, conn := buildTestStoragePool("")
	defer func() {
		pool.Undefine()
		pool.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := pool.GetUUID(); err != nil {
		t.Error(err)
	}
}

func TestGetStoragePoolUUIDString(t *testing.T) {
	pool, conn := buildTestStoragePool("")
	defer func() {
		pool.Undefine()
		pool.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := pool.GetUUIDString(); err != nil {
		t.Error(err)
	}
}

func TestGetStoragePoolInfo(t *testing.T) {
	pool, conn := buildTestStoragePool("")
	defer func() {
		pool.Undefine()
		pool.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := pool.GetInfo(); err != nil {
		t.Error(err)
	}
}

func TestGetStoragePoolXMLDesc(t *testing.T) {
	pool, conn := buildTestStoragePool("")
	defer func() {
		pool.Undefine()
		pool.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := pool.GetXMLDesc(0); err != nil {
		t.Error(err)
	}
}

func TestStoragePoolRefresh(t *testing.T) {
	pool, conn := buildTestStoragePool("")
	defer func() {
		pool.Destroy()
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
	if err := pool.Refresh(0); err != nil {
		t.Error(err)
	}
}

func TestCreateDestroyStoragePool(t *testing.T) {
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
	info, err := pool.GetInfo()
	if err != nil {
		t.Error(err)
		return
	}
	if info.State != STORAGE_POOL_RUNNING {
		t.Fatal("Storage pool should be running")
	}
	if err = pool.Destroy(); err != nil {
		t.Error(err)
		return
	}

	info, err = pool.GetInfo()
	if err != nil {
		t.Error(err)
		return
	}
	if info.State != STORAGE_POOL_INACTIVE {
		t.Fatal("Storage pool should be inactive")
	}
}

func TestStoragePoolAutostart(t *testing.T) {
	pool, conn := buildTestStoragePool("")
	defer func() {
		pool.Undefine()
		pool.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	as, err := pool.GetAutostart()
	if err != nil {
		t.Error(err)
		return
	}
	if as {
		t.Fatal("autostart should be false")
	}
	if err := pool.SetAutostart(true); err != nil {
		t.Error(err)
		return
	}
	as, err = pool.GetAutostart()
	if err != nil {
		t.Error(err)
		return
	}
	if !as {
		t.Fatal("autostart should be true")
	}
}

func TestStoragePoolIsActive(t *testing.T) {
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
	active, err := pool.IsActive()
	if err != nil {
		t.Error(err)
		return
	}
	if !active {
		t.Fatal("Storage pool should be active")
	}
	if err := pool.Destroy(); err != nil {
		t.Error(err)
		return
	}
	active, err = pool.IsActive()
	if err != nil {
		t.Error(err)
		return
	}
	if active {
		t.Fatal("Storage pool should be inactive")
	}
}

func TestStorageVolCreateDelete(t *testing.T) {
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
	vol, err := pool.StorageVolCreateXML(testStorageVolXML("", "default-pool"), 0)
	if err != nil {
		t.Fatal(err)
	}
	defer vol.Free()
	if err := vol.Delete(STORAGE_VOL_DELETE_NORMAL); err != nil {
		t.Fatal(err)
	}
}

func TestStorageVolCreateFromDelete(t *testing.T) {
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
	vol, err := pool.StorageVolCreateXML(testStorageVolXML("", "default-pool"), 0)
	if err != nil {
		t.Fatal(err)
	}
	defer vol.Free()
	clonexml := `
	<volume>
		<name>clone-test</name>
		<capacity unit="KiB">128</capacity>
		<format type="qcow2"/>
	</volume>
	`
	clone, err := pool.StorageVolCreateXMLFrom(clonexml, vol, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer clone.Free()
	if err := clone.Delete(STORAGE_VOL_DELETE_NORMAL); err != nil {
		t.Fatal(err)
	}
	if err := vol.Delete(STORAGE_VOL_DELETE_NORMAL); err != nil {
		t.Fatal(err)
	}
}

func TestLookupStorageVolByName(t *testing.T) {
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
	defVolName := time.Now().String()
	vol, err := pool.StorageVolCreateXML(testStorageVolXML(defVolName, "default-pool"), 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		vol.Delete(STORAGE_VOL_DELETE_NORMAL)
		vol.Free()
	}()
	vol2, err := pool.LookupStorageVolByName(defVolName)
	if err != nil {
		t.Error(err)
		return
	}
	defer vol2.Free()
	name, err := vol2.GetName()
	if err != nil {
		t.Error(err)
		return
	}
	if name != defVolName {
		t.Fatalf("expected storage volume name: %s ,got: %s", defVolName, name)
	}
}
