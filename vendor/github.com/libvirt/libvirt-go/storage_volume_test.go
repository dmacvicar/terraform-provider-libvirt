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

func testStorageVolXML(volName, poolPath string) string {
	defName := volName
	if defName == "" {
		defName = time.Now().String()
	}
	return `<volume>
        <name>` + defName + `</name>
        <allocation>0</allocation>
        <capacity unit="M">10</capacity>
        <target>
          <path>` + "/" + poolPath + "/" + defName + `</path>
          <permissions>
            <owner>107</owner>
            <group>107</group>
            <mode>0744</mode>
            <label>testLabel0</label>
          </permissions>
        </target>
      </volume>`
}

func TestStorageVolGetInfo(t *testing.T) {
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
		t.Error(err)
		return
	}
	defer func() {
		vol.Delete(STORAGE_VOL_DELETE_NORMAL)
		vol.Free()
	}()
	if _, err := vol.GetInfo(); err != nil {
		t.Fatal(err)
	}
}

func TestStorageVolGetKey(t *testing.T) {
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
		t.Error(err)
		return
	}
	defer func() {
		vol.Delete(STORAGE_VOL_DELETE_NORMAL)
		vol.Free()
	}()
	if _, err := vol.GetKey(); err != nil {
		t.Fatal(err)
	}
}

func TestStorageVolGetName(t *testing.T) {
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
		t.Error(err)
		return
	}
	defer func() {
		vol.Delete(STORAGE_VOL_DELETE_NORMAL)
		vol.Free()
	}()
	if _, err := vol.GetName(); err != nil {
		t.Fatal(err)
	}
}

func TestStorageVolGetPath(t *testing.T) {
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
		t.Error(err)
		return
	}
	defer func() {
		vol.Delete(STORAGE_VOL_DELETE_NORMAL)
		vol.Free()
	}()
	if _, err := vol.GetPath(); err != nil {
		t.Fatal(err)
	}
}

func TestStorageVolGetXMLDesc(t *testing.T) {
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
		t.Error(err)
		return
	}
	defer func() {
		vol.Delete(STORAGE_VOL_DELETE_NORMAL)
		vol.Free()
	}()
	if _, err := vol.GetXMLDesc(0); err != nil {
		t.Fatal(err)
	}
}

func TestPoolLookupByVolume(t *testing.T) {
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
		t.Error(err)
		return
	}
	defer func() {
		vol.Delete(STORAGE_VOL_DELETE_NORMAL)
		vol.Free()
	}()

	retPool, err := vol.LookupPoolByVolume()
	if err != nil {
		t.Fatal(err)
	}
	defer retPool.Free()

	poolUUID, err := pool.GetUUIDString()
	if err != nil {
		t.Fatal(err)
	}

	retPoolUUID, err := retPool.GetUUIDString()
	if err != nil {
		t.Fatal(err)
	}

	if retPoolUUID != poolUUID {
		t.Fail()
	}
}
