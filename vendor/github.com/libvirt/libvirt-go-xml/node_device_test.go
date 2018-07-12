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

var pciDomain uint = 1
var pciBus uint = 21
var pciSlot uint = 10
var pciFunc uint = 50

var NodeDeviceTestData = []struct {
	Object *NodeDevice
	XML    []string
}{
	{
		Object: &NodeDevice{
			Name:   "pci_0000_81_00_0",
			Parent: "pci_0000_80_01_0",
			Driver: &NodeDeviceDriver{
				Name: "ixgbe",
			},
			Capability: NodeDeviceCapability{
				PCI: &NodeDevicePCICapability{
					Domain:   &pciDomain,
					Bus:      &pciBus,
					Slot:     &pciSlot,
					Function: &pciFunc,
					Product: NodeDeviceIDName{
						ID:   "0x1528",
						Name: "Ethernet Controller 10-Gigabit X540-AT2",
					},
					Vendor: NodeDeviceIDName{
						ID:   "0x8086",
						Name: "Intel Corporation",
					},
					IOMMUGroup: &NodeDeviceIOMMUGroup{
						Number: 3,
					},
					NUMA: &NodeDeviceNUMA{
						Node: 1,
					},
					Capabilities: []NodeDevicePCISubCapability{
						NodeDevicePCISubCapability{
							VirtFunctions: &NodeDevicePCIVirtFunctionsCapability{
								MaxCount: 63,
							},
						},
					},
				},
			},
		},
		XML: []string{
			`<device>`,
			`  <name>pci_0000_81_00_0</name>`,
			`  <parent>pci_0000_80_01_0</parent>`,
			`  <driver>`,
			`    <name>ixgbe</name>`,
			`  </driver>`,
			`  <capability type="pci">`,
			`    <domain>1</domain>`,
			`    <bus>21</bus>`,
			`    <slot>10</slot>`,
			`    <function>50</function>`,
			`    <product id="0x1528">Ethernet Controller 10-Gigabit X540-AT2</product>`,
			`    <vendor id="0x8086">Intel Corporation</vendor>`,
			`    <iommuGroup number="3"></iommuGroup>`,
			`    <numa node="1"></numa>`,
			`    <capability type="virt_functions" maxCount="63"></capability>`,
			`  </capability>`,
			`</device>`,
		},
	},
}

func TestNodeDevice(t *testing.T) {
	for _, test := range NodeDeviceTestData {
		expect := strings.Join(test.XML, "\n")

		nodeDevice := NodeDevice{}
		err := nodeDevice.Unmarshal(expect)
		if err != nil {
			t.Fatal(err)
		}

		doc, err := nodeDevice.Marshal()
		if err != nil {
			t.Fatal(err)
		}

		if doc != expect {
			t.Fatal("Bad xml:\n", string(doc), "\n does not match\n", expect, "\n")
		}
	}
}
