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

func networkXML(netName string) string {
	var name string
	if netName == "" {
		name = time.Now().String()
	} else {
		name = netName
	}

	return `<network>
    <name>` + name + `</name>
    <bridge name="testbr0"/>
    <forward/>
    <ip address="192.168.0.1" netmask="255.255.255.0">
    </ip>
    </network>`
}

func buildTestNetwork(netName string) (*Network, *Connect) {
	conn := buildTestConnection()
	networkXML := networkXML(netName)
	net, err := conn.NetworkDefineXML(networkXML)
	if err != nil {
		panic(err)
	}
	return net, conn
}

func TestGetNetworkName(t *testing.T) {
	net, conn := buildTestNetwork("")
	defer func() {
		net.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := net.GetName(); err != nil {
		t.Fatal(err)
		return
	}
}

func TestGetNetworkUUID(t *testing.T) {
	net, conn := buildTestNetwork("")
	defer func() {
		net.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	_, err := net.GetUUID()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestGetNetworkUUIDString(t *testing.T) {
	net, conn := buildTestNetwork("")
	defer func() {
		net.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	_, err := net.GetUUIDString()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestGetNetworkXMLDesc(t *testing.T) {
	net, conn := buildTestNetwork("")
	defer func() {
		net.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if _, err := net.GetXMLDesc(0); err != nil {
		t.Error(err)
		return
	}
}

func TestCreateDestroyNetwork(t *testing.T) {
	net, conn := buildTestNetwork("")
	defer func() {
		net.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := net.Create(); err != nil {
		t.Error(err)
		return
	}

	if err := net.Destroy(); err != nil {
		t.Error(err)
		return
	}
}

func TestNetworkAutostart(t *testing.T) {
	net, conn := buildTestNetwork("")
	defer func() {
		net.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	as, err := net.GetAutostart()
	if err != nil {
		t.Error(err)
		return
	}
	if as {
		t.Fatal("autostart should be false")
		return
	}
	if err := net.SetAutostart(true); err != nil {
		t.Error(err)
		return
	}
	as, err = net.GetAutostart()
	if err != nil {
		t.Error(err)
		return
	}
	if !as {
		t.Fatal("autostart should be true")
		return
	}
}

func TestNetworkIsActive(t *testing.T) {
	net, conn := buildTestNetwork("")
	defer func() {
		net.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := net.Create(); err != nil {
		t.Log(err)
		return
	}
	active, err := net.IsActive()
	if err != nil {
		t.Error(err)
		return
	}
	if !active {
		t.Fatal("Network should be active")
		return
	}
	if err := net.Destroy(); err != nil {
		t.Error(err)
		return
	}
	active, err = net.IsActive()
	if err != nil {
		t.Error(err)
		return
	}
	if active {
		t.Fatal("Network should be inactive")
		return
	}
}

func TestNetworkGetBridgeName(t *testing.T) {
	net, conn := buildTestNetwork("")
	defer func() {
		net.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := net.Create(); err != nil {
		t.Error(err)
		return
	}
	brName := "testbr0"
	br, err := net.GetBridgeName()
	if err != nil {
		t.Errorf("got %s but expected %s", br, brName)
	}
}

func TestNetworkFree(t *testing.T) {
	net, conn := buildTestNetwork("")
	defer func() {
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()
	if err := net.Free(); err != nil {
		t.Error(err)
		return
	}
}

func TestNetworkCreateXML(t *testing.T) {
	conn := buildTestConnection()
	networkXML := networkXML("")
	net, err := conn.NetworkCreateXML(networkXML)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		net.Free()
		if res, _ := conn.Close(); res != 0 {
			t.Errorf("Close() == %d, expected 0", res)
		}
	}()

	if is_active, err := net.IsActive(); err != nil {
		t.Error(err)
	} else {
		if !is_active {
			t.Error("Network should be active")
		}
	}
	if is_persistent, err := net.IsPersistent(); err != nil {
		t.Error(err)
	} else {
		if is_persistent {
			t.Error("Network should not be persistent")
		}
	}
}
