/*
 * This file is part of the libvirt-go-xml project
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
 * Copyright (C) 2017 Red Hat, Inc.
 *
 */

package libvirtxml

import (
	"strings"
	"testing"
)

var storagePoolTestData = []struct {
	Object   *StoragePool
	Expected []string
}{
	{
		Object: &StoragePool{
			Type:       "dir",
			Name:       "pool",
			UUID:       "3e3fce45-4f53-4fa7-bb32-11f34168b82b",
			Allocation: &StoragePoolSize{Value: 1000000},
			Capacity:   &StoragePoolSize{Value: 5000000},
			Available:  &StoragePoolSize{Value: 3000000},
		},
		Expected: []string{
			`<pool type="dir">`,
			`  <name>pool</name>`,
			`  <uuid>3e3fce45-4f53-4fa7-bb32-11f34168b82b</uuid>`,
			`  <allocation>1000000</allocation>`,
			`  <capacity>5000000</capacity>`,
			`  <available>3000000</available>`,
			`</pool>`,
		},
	},
	{
		Object: &StoragePool{
			Type: "iscsi",
			Name: "pool",
			Source: &StoragePoolSource{
				Host: &StoragePoolSourceHost{
					Name: "host.example.com",
				},
				Device: &StoragePoolSourceDevice{
					Path: "pool.example.com:iscsi-pool",
				},
				Auth: &StoragePoolSourceAuth{
					Type:     "chap",
					Username: "username",
					Secret: &StoragePoolSourceAuthSecret{
						Usage: "cluster",
					},
				},
				Vendor: &StoragePoolSourceVendor{
					Name: "vendor",
				},
				Product: &StoragePoolSourceProduct{
					Name: "product",
				},
			},
		},
		Expected: []string{
			`<pool type="iscsi">`,
			`  <name>pool</name>`,
			`  <source>`,
			`    <host name="host.example.com"></host>`,
			`    <device path="pool.example.com:iscsi-pool"></device>`,
			`    <auth type="chap" username="username">`,
			`      <secret usage="cluster"></secret>`,
			`    </auth>`,
			`    <vendor name="vendor"></vendor>`,
			`    <product name="product"></product>`,
			`  </source>`,
			`</pool>`,
		},
	},
	{
		Object: &StoragePool{
			Type: "disk",
			Name: "pool",
			Source: &StoragePoolSource{
				Device: &StoragePoolSourceDevice{
					Path:          "/dev/mapper/pool",
					PartSeparator: "no",
				},
				Format: &StoragePoolSourceFormat{
					Type: "gpt",
				},
			},
		},
		Expected: []string{
			`<pool type="disk">`,
			`  <name>pool</name>`,
			`  <source>`,
			`    <device path="/dev/mapper/pool" part_separator="no"></device>`,
			`    <format type="gpt"></format>`,
			`  </source>`,
			`</pool>`,
		},
	},
	{
		Object: &StoragePool{
			Type: "scsi",
			Name: "pool",
			Source: &StoragePoolSource{
				Adapter: &StoragePoolSourceAdapter{
					Type: "scsi_host",
					Name: "scsi_host",
				},
			},
		},
		Expected: []string{
			`<pool type="scsi">`,
			`  <name>pool</name>`,
			`  <source>`,
			`    <adapter type="scsi_host" name="scsi_host"></adapter>`,
			`  </source>`,
			`</pool>`,
		},
	},
	{
		Object: &StoragePool{
			Type: "scsi",
			Name: "pool",
			Source: &StoragePoolSource{
				Adapter: &StoragePoolSourceAdapter{
					Type: "scsi_host",
					ParentAddr: &StoragePoolSourceAdapterParentAddr{
						UniqueID: 1,
						Address: &StoragePoolSourceAdapterParentAddrAddress{
							Domain: "0x0000",
							Bus:    "0x00",
							Slot:   "0x1f",
							Addr:   "0x2",
						},
					},
				},
			},
		},
		Expected: []string{
			`<pool type="scsi">`,
			`  <name>pool</name>`,
			`  <source>`,
			`    <adapter type="scsi_host">`,
			`      <parentaddr unique_id="1">`,
			`        <address domain="0x0000" bus="0x00" slot="0x1f" addr="0x2"></address>`,
			`      </parentaddr>`,
			`    </adapter>`,
			`  </source>`,
			`</pool>`,
		},
	},
	{
		Object: &StoragePool{
			Type: "fs",
			Name: "pool",
			Source: &StoragePoolSource{
				Adapter: &StoragePoolSourceAdapter{
					Type:   "fc_host",
					Parent: "scsi_parent",
					WWNN:   "20000000c9831b4b",
					WWPN:   "10000000c9831b4b",
				},
			},
		},
		Expected: []string{
			`<pool type="fs">`,
			`  <name>pool</name>`,
			`  <source>`,
			`    <adapter type="fc_host" parent="scsi_parent" wwnn="20000000c9831b4b" wwpn="10000000c9831b4b"></adapter>`,
			`  </source>`,
			`</pool>`,
		},
	},
	{
		Object: &StoragePool{
			Type: "dir",
			Name: "pool",
			Target: &StoragePoolTarget{
				Path: "/pool",
				Permissions: &StoragePoolTargetPermissions{
					Owner: "1",
					Group: "1",
					Mode:  "0744",
					Label: "pool",
				},
			},
		},
		Expected: []string{
			`<pool type="dir">`,
			`  <name>pool</name>`,
			`  <target>`,
			`    <path>/pool</path>`,
			`    <permissions>`,
			`      <owner>1</owner>`,
			`      <group>1</group>`,
			`      <mode>0744</mode>`,
			`      <label>pool</label>`,
			`    </permissions>`,
			`  </target>`,
			`</pool>`,
		},
	},
}

func TestStoragePool(t *testing.T) {
	for _, test := range storagePoolTestData {
		doc, err := test.Object.Marshal()
		if err != nil {
			t.Fatal(err)
		}

		expect := strings.Join(test.Expected, "\n")

		if doc != expect {
			t.Fatal("Bad xml:\n", string(doc), "\n does not match\n", expect, "\n")
		}
	}
}
