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
	"crypto/rand"
	"fmt"
	"testing"
)

func buildTestInterface(mac string) (*Interface, *Connect) {
	conn := buildTestConnection()
	xml := `<interface type='ethernet' name='ethTest0'><mac address='` + mac + `'/></interface>`
	iface, err := conn.InterfaceDefineXML(xml, 0)
	if err != nil {
		panic(err)
	}
	return iface, conn
}

func generateRandomMac() string {
	macBuf := make([]byte, 3)
	if _, err := rand.Read(macBuf); err != nil {
		panic(err)
	}
	return fmt.Sprintf("aa:bb:cc:%02x:%02x:%02x", macBuf[0], macBuf[1], macBuf[2])
}

func TestCreateDestroyInterface(t *testing.T) {
	iface, conn := buildTestInterface(generateRandomMac())
	defer func() {
		iface.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := iface.Create(0); err != nil {
		t.Error(err)
		return
	}
	if err := iface.Destroy(0); err != nil {
		t.Error(err)
	}
}

func TestUndefineInterface(t *testing.T) {
	iface, conn := buildTestInterface(generateRandomMac())
	defer func() {
		iface.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	name, err := iface.GetName()
	if err != nil {
		t.Error(err)
		return
	}
	if err := iface.Undefine(); err != nil {
		t.Error(err)
		return
	}
	if _, err := conn.LookupInterfaceByName(name); err == nil {
		t.Fatal("Shouldn't have been able to find interface")
	}
}

func TestGetInterfaceName(t *testing.T) {
	iface, conn := buildTestInterface(generateRandomMac())
	defer func() {
		iface.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := iface.GetName(); err != nil {
		t.Fatal(err)
	}
}

func TestInterfaceIsActive(t *testing.T) {
	iface, conn := buildTestInterface(generateRandomMac())
	defer func() {
		iface.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := iface.Create(0); err != nil {
		t.Log(err)
		return
	}
	active, err := iface.IsActive()
	if err != nil {
		t.Error(err)
		return
	}
	if !active {
		t.Fatal("Interface should be active")
	}
	if err := iface.Destroy(0); err != nil {
		t.Error(err)
		return
	}
	active, err = iface.IsActive()
	if err != nil {
		t.Error(err)
		return
	}
	if active {
		t.Fatal("Interface should be inactive")
	}
}

func TestGetMACString(t *testing.T) {
	origMac := generateRandomMac()
	iface, conn := buildTestInterface(origMac)
	defer func() {
		iface.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	mac, err := iface.GetMACString()
	if err != nil {
		t.Error(err)
		return
	}
	if mac != origMac {
		t.Fatalf("expected MAC: %s , got: %s", origMac, mac)
	}
}

func TestGetInterfaceXMLDesc(t *testing.T) {
	iface, conn := buildTestInterface(generateRandomMac())
	defer func() {
		iface.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := iface.GetXMLDesc(0); err != nil {
		t.Error(err)
	}
}

func TestInterfaceFree(t *testing.T) {
	iface, conn := buildTestInterface(generateRandomMac())
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := iface.Free(); err != nil {
		t.Error(err)
		return
	}
}
