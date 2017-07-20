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

var storageVolumeTestData = []struct {
	Object   *StorageVolume
	Expected []string
}{
	{
		Object: &StorageVolume{
			Type: "file",
			Name: "file.img",
			Key:  "/file.img",
			Allocation: &StorageVolumeSize{
				Value: 0,
			},

			Capacity: &StorageVolumeSize{
				Unit:  "T",
				Value: 1,
			},
		},
		Expected: []string{
			`<volume type="file">`,
			`  <name>file.img</name>`,
			`  <key>/file.img</key>`,
			`  <allocation>0</allocation>`,
			`  <capacity unit="T">1</capacity>`,
			`</volume>`,
		},
	},
	{
		Object: &StorageVolume{
			Type: "file",
			Name: "file.img",
			Target: &StorageVolumeTarget{
				Path: "/file.img",
				Format: &StorageVolumeTargetFormat{
					Type: "qcow2",
				},
				Permissions: &StorageVolumeTargetPermissions{
					Owner: "107",
					Group: "107",
					Mode:  "0744",
					Label: "image",
				},
				Timestamps: &StorageVolumeTargetTimestamps{
					Atime: "1341933637.273190990",
					Mtime: "1341930622.047245868",
					Ctime: "1341930622.047245868",
				},
				Compat: "1.1",
				NoCOW:  &struct{}{},
				Features: []StorageVolumeTargetFeature{
					StorageVolumeTargetFeature{
						LazyRefcounts: &struct{}{},
					},
				},
			},
		},
		Expected: []string{
			`<volume type="file">`,
			`  <name>file.img</name>`,
			`  <target>`,
			`    <path>/file.img</path>`,
			`    <format type="qcow2"></format>`,
			`    <permissions>`,
			`      <owner>107</owner>`,
			`      <group>107</group>`,
			`      <mode>0744</mode>`,
			`      <label>image</label>`,
			`    </permissions>`,
			`    <timestamps>`,
			`      <atime>1341933637.273190990</atime>`,
			`      <mtime>1341930622.047245868</mtime>`,
			`      <ctime>1341930622.047245868</ctime>`,
			`    </timestamps>`,
			`    <compat>1.1</compat>`,
			`    <nocow></nocow>`,
			`    <features>`,
			`      <lazy_refcounts></lazy_refcounts>`,
			`    </features>`,
			`  </target>`,
			`</volume>`,
		},
	},
	{
		Object: &StorageVolume{
			Type: "file",
			Name: "file.img",
			BackingStore: &StorageVolumeBackingStore{
				Path: "/master.img",
				Format: &StorageVolumeTargetFormat{
					Type: "raw",
				},
				Permissions: &StorageVolumeTargetPermissions{
					Owner: "107",
					Group: "107",
					Mode:  "0744",
					Label: "label",
				},
			},
		},
		Expected: []string{
			`<volume type="file">`,
			`  <name>file.img</name>`,
			`  <backingStore>`,
			`    <path>/master.img</path>`,
			`    <format type="raw"></format>`,
			`    <permissions>`,
			`      <owner>107</owner>`,
			`      <group>107</group>`,
			`      <mode>0744</mode>`,
			`      <label>label</label>`,
			`    </permissions>`,
			`  </backingStore>`,
			`</volume>`,
		},
	},
	{
		Object: &StorageVolume{
			Name: "luks.img",
			Capacity: &StorageVolumeSize{
				Unit:  "G",
				Value: 5,
			},
			Target: &StorageVolumeTarget{
				Path: "/luks.img",
				Format: &StorageVolumeTargetFormat{
					Type: "raw",
				},
				Encryption: &StorageEncryption{
					Format: "luks",
					Secret: &StorageEncryptionSecret{
						Type: "passphrase",
						UUID: "f52a81b2-424e-490c-823d-6bd4235bc572",
					},
				},
			},
		},
		Expected: []string{
			`<volume>`,
			`  <name>luks.img</name>`,
			`  <capacity unit="G">5</capacity>`,
			`  <target>`,
			`    <path>/luks.img</path>`,
			`    <format type="raw"></format>`,
			`    <encryption format="luks">`,
			`      <secret type="passphrase" uuid="f52a81b2-424e-490c-823d-6bd4235bc572"></secret>`,
			`    </encryption>`,
			`  </target>`,
			`</volume>`,
		},
	},
	{
		Object: &StorageVolume{
			Name: "twofish",
			Capacity: &StorageVolumeSize{
				Unit:  "G",
				Value: 5,
			},
			Target: &StorageVolumeTarget{
				Path: "/twofish.luks",
				Format: &StorageVolumeTargetFormat{
					Type: "raw",
				},
				Encryption: &StorageEncryption{
					Format: "luks",
					Secret: &StorageEncryptionSecret{
						Type: "passphrase",
						UUID: "f52a81b2-424e-490c-823d-6bd4235bc572",
					},
					Cipher: &StorageEncryptionCipher{
						Name: "twofish",
						Size: 256,
						Mode: "cbc",
						Hash: "sha256",
					},
					Ivgen: &StorageEncryptionIvgen{
						Name: "plain64",
						Hash: "sha256",
					},
				},
			},
		},
		Expected: []string{
			`<volume>`,
			`  <name>twofish</name>`,
			`  <capacity unit="G">5</capacity>`,
			`  <target>`,
			`    <path>/twofish.luks</path>`,
			`    <format type="raw"></format>`,
			`    <encryption format="luks">`,
			`      <secret type="passphrase" uuid="f52a81b2-424e-490c-823d-6bd4235bc572"></secret>`,
			`      <cipher name="twofish" size="256" mode="cbc" hash="sha256"></cipher>`,
			`      <ivgen name="plain64" hash="sha256"></ivgen>`,
			`    </encryption>`,
			`  </target>`,
			`</volume>`,
		},
	},
}

func TestStorageVolume(t *testing.T) {
	for _, test := range storageVolumeTestData {
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
