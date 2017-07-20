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

var domainSnapshotTestData = []struct {
	Object   *DomainSnapshot
	Expected []string
}{
	{
		Object: &DomainSnapshot{
			Description: "Snapshot",
			Disks: &DomainSnapshotDisks{
				[]DomainSnapshotDisk{
					DomainSnapshotDisk{
						Name: "/old",
						Source: &DomainSnapshotDiskSource{
							File: "/new",
						},
					},
					DomainSnapshotDisk{
						Name:     "vdb",
						Snapshot: "no",
					},
				},
			},
		},
		Expected: []string{
			`<domainsnapshot>`,
			`  <description>Snapshot</description>`,
			`  <disks>`,
			`    <disk name="/old">`,
			`      <source file="/new"></source>`,
			`    </disk>`,
			`    <disk name="vdb" snapshot="no"></disk>`,
			`  </disks>`,
			`</domainsnapshot>`,
		},
	},
	{
		Object: &DomainSnapshot{
			Name:         "1270477159",
			Description:  "Snapshot of OS install and updates",
			State:        "running",
			CreationTime: "1270477159",
			Parent: &DomainSnapshotParent{
				Name: "bare-os-install",
			},
			Memory: &DomainSnapshotMemory{
				Snapshot: "no",
			},
			Disks: &DomainSnapshotDisks{
				Disks: []DomainSnapshotDisk{
					DomainSnapshotDisk{
						Name:     "vda",
						Snapshot: "external",
						Driver: &DomainSnapshotDiskDriver{
							Type: "qcow2",
						},
						Source: &DomainSnapshotDiskSource{
							File: "/path/to/new",
						},
					},
					DomainSnapshotDisk{
						Name:     "vdb",
						Snapshot: "no",
					},
				},
			},
			Domain: &Domain{
				Name: "fedora",
				Memory: &DomainMemory{
					Value: 1048576,
				},
				Devices: &DomainDeviceList{
					Disks: []DomainDisk{
						DomainDisk{
							Type:   "file",
							Device: "disk",
							Driver: &DomainDiskDriver{
								Name: "qemu",
								Type: "raw",
							},
							Source: &DomainDiskSource{
								File: "/path/to/old",
							},
							Target: &DomainDiskTarget{
								Dev: "vda",
								Bus: "virtio",
							},
						},
						DomainDisk{
							Type:     "file",
							Device:   "disk",
							Snapshot: "external",
							Driver: &DomainDiskDriver{
								Name: "qemu",
								Type: "raw",
							},
							Source: &DomainDiskSource{
								File: "/path/to/old2",
							},
							Target: &DomainDiskTarget{
								Dev: "vdb",
								Bus: "virtio",
							},
						},
					},
				},
			},
		},
		Expected: []string{
			`<domainsnapshot>`,
			`  <name>1270477159</name>`,
			`  <description>Snapshot of OS install and updates</description>`,
			`  <state>running</state>`,
			`  <creationTime>1270477159</creationTime>`,
			`  <parent>`,
			`    <name>bare-os-install</name>`,
			`  </parent>`,
			`  <memory snapshot="no"></memory>`,
			`  <disks>`,
			`    <disk name="vda" snapshot="external">`,
			`      <driver type="qcow2"></driver>`,
			`      <source file="/path/to/new"></source>`,
			`    </disk>`,
			`    <disk name="vdb" snapshot="no"></disk>`,
			`  </disks>`,
			`  <domain>`,
			`    <name>fedora</name>`,
			`    <memory>1048576</memory>`,
			`    <devices>`,
			`      <disk type="file" device="disk">`,
			`        <driver name="qemu" type="raw"></driver>`,
			`        <source file="/path/to/old"></source>`,
			`        <target dev="vda" bus="virtio"></target>`,
			`      </disk>`,
			`      <disk type="file" device="disk" snapshot="external">`,
			`        <driver name="qemu" type="raw"></driver>`,
			`        <source file="/path/to/old2"></source>`,
			`        <target dev="vdb" bus="virtio"></target>`,
			`      </disk>`,
			`    </devices>`,
			`  </domain>`,
			`</domainsnapshot>`,
		},
	},
}

func TestDomainSnapshot(t *testing.T) {
	for _, test := range domainSnapshotTestData {
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
