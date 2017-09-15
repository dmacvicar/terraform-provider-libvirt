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
 * Copyright (C) 2016 Red Hat, Inc.
 *
 */

package libvirtxml

import (
	"strings"
	"testing"
)

type Address struct {
	Domain   HexUint
	Bus      HexUint
	Slot     HexUint
	Function HexUint
}

var uhciIndex uint = 0
var uhciAddr = Address{0, 0, 1, 2}

var diskAddr = Address{0, 0, 3, 0}
var ifaceAddr = Address{0, 0, 4, 0}
var videoAddr = Address{0, 0, 5, 0}
var fsAddr = Address{0, 0, 6, 0}
var balloonAddr = Address{0, 0, 7, 0}
var duplexAddr = Address{0, 0, 8, 0}

var serialPort uint = 0
var tabletBus HexUint = 0
var tabletPort uint = 1

var nicAverage int = 1000
var nicBurst int = 10000

var domainTestData = []struct {
	Object   Document
	Expected []string
}{
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			Devices: &DomainDeviceList{
				Disks: []DomainDisk{
					DomainDisk{
						Type:   "file",
						Device: "cdrom",
						Driver: &DomainDiskDriver{
							Name: "qemu",
							Type: "qcow2",
						},
						Source: &DomainDiskSource{
							File: "/var/lib/libvirt/images/demo.qcow2",
						},
						Target: &DomainDiskTarget{
							Dev: "vda",
							Bus: "virtio",
						},
						Serial: "fishfood",
						Boot: &DomainDeviceBoot{
							Order: 1,
						},
					},
					DomainDisk{
						Type:   "block",
						Device: "disk",
						Driver: &DomainDiskDriver{
							Name: "qemu",
							Type: "raw",
						},
						Source: &DomainDiskSource{
							Device: "/dev/sda1",
						},
						Target: &DomainDiskTarget{
							Dev: "vdb",
							Bus: "virtio",
						},
						Address: &DomainAddress{
							Type:     "pci",
							Domain:   &diskAddr.Domain,
							Bus:      &diskAddr.Bus,
							Slot:     &diskAddr.Slot,
							Function: &diskAddr.Function,
						},
					},
					DomainDisk{
						Type:   "network",
						Device: "disk",
						Auth: &DomainDiskAuth{
							Username: "fred",
							Secret: &DomainDiskSecret{
								Type: "ceph",
								UUID: "e49f09c9-119e-43fd-b5a9-000d41e65493",
							},
						},
						Source: &DomainDiskSource{
							Protocol: "rbd",
							Name:     "somepool/somevol",
							Hosts: []DomainDiskSourceHost{
								DomainDiskSourceHost{
									Transport: "tcp",
									Name:      "rbd1.example.com",
									Port:      "3000",
								},
								DomainDiskSourceHost{
									Transport: "tcp",
									Name:      "rbd2.example.com",
									Port:      "3000",
								},
							},
						},
						Target: &DomainDiskTarget{
							Dev: "vdc",
							Bus: "virtio",
						},
					},
					DomainDisk{
						Type:   "network",
						Device: "disk",
						Source: &DomainDiskSource{
							Protocol: "nbd",
							Hosts: []DomainDiskSourceHost{
								DomainDiskSourceHost{
									Transport: "unix",
									Socket:    "/var/run/nbd.sock",
								},
							},
						},
						Target: &DomainDiskTarget{
							Dev: "vdd",
							Bus: "virtio",
						},
						Shareable: &DomainDiskShareable{},
					},
					DomainDisk{
						Type:   "volume",
						Device: "cdrom",
						Driver: &DomainDiskDriver{
							Cache:       "none",
							IO:          "native",
							ErrorPolicy: "stop",
						},
						Source: &DomainDiskSource{
							Pool:   "default",
							Volume: "myvolume",
						},
						Target: &DomainDiskTarget{
							Dev: "vde",
							Bus: "virtio",
						},
						ReadOnly: &DomainDiskReadOnly{},
					},
				},
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <devices>`,
			`    <disk type="file" device="cdrom">`,
			`      <driver name="qemu" type="qcow2"></driver>`,
			`      <source file="/var/lib/libvirt/images/demo.qcow2"></source>`,
			`      <target dev="vda" bus="virtio"></target>`,
			`      <serial>fishfood</serial>`,
			`      <boot order="1"></boot>`,
			`    </disk>`,
			`    <disk type="block" device="disk">`,
			`      <driver name="qemu" type="raw"></driver>`,
			`      <source dev="/dev/sda1"></source>`,
			`      <target dev="vdb" bus="virtio"></target>`,
			`      <address type="pci" domain="0" bus="0" slot="3" function="0"></address>`,
			`    </disk>`,
			`    <disk type="network" device="disk">`,
			`      <auth username="fred">`,
			`        <secret type="ceph" uuid="e49f09c9-119e-43fd-b5a9-000d41e65493"></secret>`,
			`      </auth>`,
			`      <source protocol="rbd" name="somepool/somevol">`,
			`        <host transport="tcp" name="rbd1.example.com" port="3000"></host>`,
			`        <host transport="tcp" name="rbd2.example.com" port="3000"></host>`,
			`      </source>`,
			`      <target dev="vdc" bus="virtio"></target>`,
			`    </disk>`,
			`    <disk type="network" device="disk">`,
			`      <source protocol="nbd">`,
			`        <host transport="unix" socket="/var/run/nbd.sock"></host>`,
			`      </source>`,
			`      <target dev="vdd" bus="virtio"></target>`,
			`      <shareable></shareable>`,
			`    </disk>`,
			`    <disk type="volume" device="cdrom">`,
			`      <driver cache="none" io="native" error_policy="stop"></driver>`,
			`      <source pool="default" volume="myvolume"></source>`,
			`      <target dev="vde" bus="virtio"></target>`,
			`      <readonly></readonly>`,
			`    </disk>`,
			`  </devices>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			Devices: &DomainDeviceList{
				Inputs: []DomainInput{
					DomainInput{
						Type: "tablet",
						Bus:  "usb",
						Address: &DomainAddress{
							Type: "usb",
							Bus:  &tabletBus,
							Port: &tabletPort,
						},
					},
					DomainInput{
						Type: "keyboard",
						Bus:  "ps2",
					},
				},
				Videos: []DomainVideo{
					DomainVideo{
						Model: DomainVideoModel{
							Type:   "cirrus",
							Heads:  1,
							Ram:    4096,
							VRam:   8192,
							VGAMem: 256,
						},
						Address: &DomainAddress{
							Type:     "pci",
							Domain:   &videoAddr.Domain,
							Bus:      &videoAddr.Bus,
							Slot:     &videoAddr.Slot,
							Function: &videoAddr.Function,
						},
					},
				},
				Graphics: []DomainGraphic{
					DomainGraphic{
						Type: "vnc",
					},
				},
				MemBalloon: &DomainMemBalloon{
					Model: "virtio",
					Address: &DomainAddress{
						Type:     "pci",
						Domain:   &balloonAddr.Domain,
						Bus:      &balloonAddr.Bus,
						Slot:     &balloonAddr.Slot,
						Function: &balloonAddr.Function,
					},
				},
				Consoles: []DomainConsole{
					DomainConsole{
						Type: "pty",
						Target: &DomainConsoleTarget{
							Type: "virtio",
							Port: &serialPort,
						},
					},
				},
				Serials: []DomainSerial{
					DomainSerial{
						Type: "pty",
						Target: &DomainSerialTarget{
							Type: "isa",
							Port: &serialPort,
						},
					},
					DomainSerial{
						Type: "file",
						Source: &DomainChardevSource{
							Path:   "/tmp/serial.log",
							Append: "off",
						},
						Target: &DomainSerialTarget{
							Port: &serialPort,
						},
					},
				},
				Channels: []DomainChannel{
					DomainChannel{
						Type: "pty",
						Target: &DomainChannelTarget{
							Type:  "virtio",
							Name:  "org.redhat.spice",
							State: "connected",
						},
					},
				},
				Sounds: []DomainSound{
					DomainSound{
						Model: "ich6",
						Codec: &DomainSoundCodec{
							Type: "duplex",
						},
						Address: &DomainAddress{
							Type:     "pci",
							Domain:   &duplexAddr.Domain,
							Bus:      &duplexAddr.Bus,
							Slot:     &duplexAddr.Slot,
							Function: &duplexAddr.Function,
						},
					},
				},
				RNGs: []DomainRNG{
					DomainRNG{
						Model: "virtio",
						Rate: &DomainRNGRate{
							Period: 2000,
							Bytes:  1234,
						},
						Backend: &DomainRNGBackend{
							Model: "egd",
							Type:  "udp",
							Sources: []DomainInterfaceSource{
								DomainInterfaceSource{
									Mode:    "bind",
									Service: "1234",
								},
								DomainInterfaceSource{
									Mode:    "connect",
									Host:    "1.2.3.4",
									Service: "1234",
								},
							},
						},
					},
				},
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <devices>`,
			`    <serial type="pty">`,
			`      <target type="isa" port="0"></target>`,
			`    </serial>`,
			`    <serial type="file">`,
			`      <source path="/tmp/serial.log" append="off"></source>`,
			`      <target port="0"></target>`,
			`    </serial>`,
			`    <console type="pty">`,
			`      <target type="virtio" port="0"></target>`,
			`    </console>`,
			`    <input type="tablet" bus="usb">`,
			`      <address type="usb" bus="0" port="1"></address>`,
			`    </input>`,
			`    <input type="keyboard" bus="ps2"></input>`,
			`    <graphics type="vnc"></graphics>`,
			`    <video>`,
			`      <model type="cirrus" heads="1" ram="4096" vram="8192" vgamem="256"></model>`,
			`      <address type="pci" domain="0" bus="0" slot="5" function="0"></address>`,
			`    </video>`,
			`    <channel type="pty">`,
			`      <target type="virtio" name="org.redhat.spice" state="connected"></target>`,
			`    </channel>`,
			`    <memballoon model="virtio">`,
			`      <address type="pci" domain="0" bus="0" slot="7" function="0"></address>`,
			`    </memballoon>`,
			`    <sound model="ich6">`,
			`      <codec type="duplex"></codec>`,
			`      <address type="pci" domain="0" bus="0" slot="8" function="0"></address>`,
			`    </sound>`,
			`    <rng model="virtio">`,
			`      <rate bytes="1234" period="2000"></rate>`,
			`      <backend model="egd" type="udp">`,
			`        <source mode="bind" service="1234"></source>`,
			`        <source mode="connect" service="1234" host="1.2.3.4"></source>`,
			`      </backend>`,
			`    </rng>`,
			`  </devices>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			Memory: &DomainMemory{
				Unit:  "KiB",
				Value: 8192,
			},
			CurrentMemory: &DomainMemory{
				Unit:  "KiB",
				Value: 4096,
			},
			MaximumMemory: &DomainMaxMemory{
				Unit:  "KiB",
				Value: 16384,
				Slots: 2,
			},
			OS: &DomainOS{
				Type: &DomainOSType{
					Arch:    "x86_64",
					Machine: "pc",
					Type:    "hvm",
				},
				BootDevices: []DomainBootDevice{
					DomainBootDevice{
						Dev: "hd",
					},
				},
				Loader: &DomainLoader{
					Readonly: "yes",
					Secure:   "no",
					Type:     "rom",
					Path:     "/loader",
				},
				SMBios: &DomainSMBios{
					Mode: "sysinfo",
				},
				BIOS: &DomainBIOS{
					UseSerial:     "yes",
					RebootTimeout: "0",
				},
				Init: "/bin/systemd",
				InitArgs: []string{
					"--unit",
					"emergency.service",
				},
			},
			SysInfo: &DomainSysInfo{
				Type: "smbios",
				BIOS: []DomainSysInfoEntry{
					DomainSysInfoEntry{
						Name:  "vendor",
						Value: "vendor",
					},
				},
				System: []DomainSysInfoEntry{
					DomainSysInfoEntry{
						Name:  "manufacturer",
						Value: "manufacturer",
					},
					DomainSysInfoEntry{
						Name:  "product",
						Value: "product",
					},
					DomainSysInfoEntry{
						Name:  "version",
						Value: "version",
					},
				},
				BaseBoard: []DomainSysInfoEntry{
					DomainSysInfoEntry{
						Name:  "manufacturer",
						Value: "manufacturer",
					},
					DomainSysInfoEntry{
						Name:  "product",
						Value: "product",
					},
					DomainSysInfoEntry{
						Name:  "version",
						Value: "version",
					},
					DomainSysInfoEntry{
						Name:  "serial",
						Value: "serial",
					},
				},
			},
			Clock: &DomainClock{
				Offset:     "variable",
				Basis:      "utc",
				Adjustment: 28794,
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <memory unit="KiB">8192</memory>`,
			`  <currentMemory unit="KiB">4096</currentMemory>`,
			`  <maxMemory unit="KiB" slots="2">16384</maxMemory>`,
			`  <sysinfo type="smbios">`,
			`    <system>`,
			`      <entry name="manufacturer">manufacturer</entry>`,
			`      <entry name="product">product</entry>`,
			`      <entry name="version">version</entry>`,
			`    </system>`,
			`    <bios>`,
			`      <entry name="vendor">vendor</entry>`,
			`    </bios>`,
			`    <baseBoard>`,
			`      <entry name="manufacturer">manufacturer</entry>`,
			`      <entry name="product">product</entry>`,
			`      <entry name="version">version</entry>`,
			`      <entry name="serial">serial</entry>`,
			`    </baseBoard>`,
			`  </sysinfo>`,
			`  <os>`,
			`    <type arch="x86_64" machine="pc">hvm</type>`,
			`    <loader readonly="yes" secure="no" type="rom">/loader</loader>`,
			`    <boot dev="hd"></boot>`,
			`    <smbios mode="sysinfo"></smbios>`,
			`    <bios useserial="yes" rebootTimeout="0"></bios>`,
			`    <init>/bin/systemd</init>`,
			`    <initarg>--unit</initarg>`,
			`    <initarg>emergency.service</initarg>`,
			`  </os>`,
			`  <clock offset="variable" basis="utc" adjustment="28794"></clock>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			OS: &DomainOS{
				NVRam: &DomainNVRam{
					Template: "/t.fd",
					NVRam:    "/vars.fd",
				},
				BootMenu: &DomainBootMenu{
					Enabled: "yes",
					Timeout: "3000",
				},
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <os>`,
			`    <nvram template="/t.fd">/vars.fd</nvram>`,
			`    <bootmenu enabled="yes" timeout="3000"></bootmenu>`,
			`  </os>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			OS: &DomainOS{
				Kernel:     "/vmlinuz",
				Initrd:     "/initrd",
				KernelArgs: "arg",
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <os>`,
			`    <kernel>/vmlinuz</kernel>`,
			`    <initrd>/initrd</initrd>`,
			`    <cmdline>arg</cmdline>`,
			`  </os>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			Resource: &DomainResource{
				Partition: "/machines/production",
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <resource>`,
			`    <partition>/machines/production</partition>`,
			`  </resource>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			VCPU: &DomainVCPU{
				Placement: "static",
				CPUSet:    "1-4,^3,6",
				Current:   "1",
				Value:     2,
			},
			Devices: &DomainDeviceList{
				Interfaces: []DomainInterface{
					DomainInterface{
						Type: "network",
						MAC: &DomainInterfaceMAC{
							Address: "00:11:22:33:44:55",
						},
						Model: &DomainInterfaceModel{
							Type: "virtio",
						},
						Virtualport: &DomainInterfaceVirtualport{
							Type: "openvswitch",
						},
					},
				},
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <vcpu placement="static" cpuset="1-4,^3,6" current="1">2</vcpu>`,
			`  <devices>`,
			`    <interface type="network">`,
			`      <mac address="00:11:22:33:44:55"></mac>`,
			`      <model type="virtio"></model>`,
			`      <virtualport type="openvswitch"></virtualport>`,
			`    </interface>`,
			`  </devices>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			CPU: &DomainCPU{
				Match: "exact",
				Model: &DomainCPUModel{
					Fallback: "allow",
					Value:    "core2duo",
				},
				Vendor: "Intel",
				Topology: &DomainCPUTopology{
					Sockets: 1,
					Cores:   2,
					Threads: 1,
				},
				Features: []DomainCPUFeature{
					DomainCPUFeature{Policy: "disable", Name: "lahf_lm"},
				},
			},
			Devices: &DomainDeviceList{
				Emulator: "/bin/qemu-kvm",
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <cpu match="exact">`,
			`    <model fallback="allow">core2duo</model>`,
			`    <vendor>Intel</vendor>`,
			`    <topology sockets="1" cores="2" threads="1"></topology>`,
			`    <feature policy="disable" name="lahf_lm"></feature>`,
			`  </cpu>`,
			`  <devices>`,
			`    <emulator>/bin/qemu-kvm</emulator>`,
			`  </devices>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			Devices: &DomainDeviceList{
				Interfaces: []DomainInterface{
					DomainInterface{
						Type: "bridge",
						MAC: &DomainInterfaceMAC{
							Address: "06:39:b4:00:00:46",
						},
						Model: &DomainInterfaceModel{
							Type: "virtio",
						},
						Source: &DomainInterfaceSource{
							Bridge: "private",
						},
						Target: &DomainInterfaceTarget{
							Dev: "vnet3",
						},
						Alias: &DomainInterfaceAlias{
							Name: "net1",
						},
					},
				},
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <devices>`,
			`    <interface type="bridge">`,
			`      <mac address="06:39:b4:00:00:46"></mac>`,
			`      <model type="virtio"></model>`,
			`      <source bridge="private"></source>`,
			`      <target dev="vnet3"></target>`,
			`      <alias name="net1"></alias>`,
			`    </interface>`,
			`  </devices>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			Devices: &DomainDeviceList{
				Interfaces: []DomainInterface{
					DomainInterface{
						Type: "network",
						MAC: &DomainInterfaceMAC{
							Address: "52:54:00:39:97:ac",
						},
						Model: &DomainInterfaceModel{
							Type: "e1000",
						},
						Source: &DomainInterfaceSource{
							Network: "default",
						},
					},
				},
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <devices>`,
			`    <interface type="network">`,
			`      <mac address="52:54:00:39:97:ac"></mac>`,
			`      <model type="e1000"></model>`,
			`      <source network="default"></source>`,
			`    </interface>`,
			`  </devices>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			Devices: &DomainDeviceList{
				Interfaces: []DomainInterface{
					DomainInterface{
						Type: "direct",
						MAC: &DomainInterfaceMAC{
							Address: "52:54:00:39:97:ac",
						},
						Model: &DomainInterfaceModel{
							Type: "e1000",
						},
						Source: &DomainInterfaceSource{
							Dev:  "eth0",
							Mode: "bridge",
						},
					},
				},
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <devices>`,
			`    <interface type="direct">`,
			`      <mac address="52:54:00:39:97:ac"></mac>`,
			`      <model type="e1000"></model>`,
			`      <source dev="eth0" mode="bridge"></source>`,
			`    </interface>`,
			`  </devices>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			Devices: &DomainDeviceList{
				Interfaces: []DomainInterface{
					DomainInterface{
						Type: "user",
						MAC: &DomainInterfaceMAC{
							Address: "52:54:00:39:97:ac",
						},
						Model: &DomainInterfaceModel{
							Type: "virtio",
						},
						Link: &DomainInterfaceLink{
							State: "up",
						},
						Boot: &DomainDeviceBoot{
							Order: 1,
						},
						Driver: &DomainInterfaceDriver{
							Name:   "vhost",
							Queues: 5,
						},
					},
				},
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <devices>`,
			`    <interface type="user">`,
			`      <mac address="52:54:00:39:97:ac"></mac>`,
			`      <model type="virtio"></model>`,
			`      <link state="up"></link>`,
			`      <boot order="1"></boot>`,
			`      <driver name="vhost" queues="5"></driver>`,
			`    </interface>`,
			`  </devices>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			Devices: &DomainDeviceList{
				Interfaces: []DomainInterface{
					DomainInterface{
						Type: "server",
						MAC: &DomainInterfaceMAC{
							Address: "52:54:00:39:97:ac",
						},
						Model: &DomainInterfaceModel{
							Type: "virtio",
						},
						Source: &DomainInterfaceSource{
							Address: "127.0.0.1",
							Port:    1234,
						},
					},
				},
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <devices>`,
			`    <interface type="server">`,
			`      <mac address="52:54:00:39:97:ac"></mac>`,
			`      <model type="virtio"></model>`,
			`      <source address="127.0.0.1" port="1234"></source>`,
			`    </interface>`,
			`  </devices>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			Devices: &DomainDeviceList{
				Interfaces: []DomainInterface{
					DomainInterface{
						Type: "ethernet",
						MAC: &DomainInterfaceMAC{
							Address: "52:54:00:39:97:ac",
						},
						Model: &DomainInterfaceModel{
							Type: "virtio",
						},
						Script: &DomainInterfaceScript{
							Path: "/etc/qemu-ifup",
						},
						Address: &DomainAddress{
							Type:     "pci",
							Domain:   &ifaceAddr.Domain,
							Bus:      &ifaceAddr.Bus,
							Slot:     &ifaceAddr.Slot,
							Function: &ifaceAddr.Function,
						},
					},
				},
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <devices>`,
			`    <interface type="ethernet">`,
			`      <mac address="52:54:00:39:97:ac"></mac>`,
			`      <model type="virtio"></model>`,
			`      <script path="/etc/qemu-ifup"></script>`,
			`      <address type="pci" domain="0" bus="0" slot="4" function="0"></address>`,
			`    </interface>`,
			`  </devices>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			Devices: &DomainDeviceList{
				Interfaces: []DomainInterface{
					DomainInterface{
						Type: "vhostuser",
						MAC: &DomainInterfaceMAC{
							Address: "52:54:00:39:97:ac",
						},
						Model: &DomainInterfaceModel{
							Type: "virtio",
						},
						Source: &DomainInterfaceSource{
							Type: "unix",
							Path: "/tmp/vhost0.sock",
							Mode: "server",
						},
					},
				},
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <devices>`,
			`    <interface type="vhostuser">`,
			`      <mac address="52:54:00:39:97:ac"></mac>`,
			`      <model type="virtio"></model>`,
			`      <source type="unix" path="/tmp/vhost0.sock" mode="server"></source>`,
			`    </interface>`,
			`  </devices>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			Devices: &DomainDeviceList{
				Interfaces: []DomainInterface{
					DomainInterface{
						Type: "vhostuser",
						MAC: &DomainInterfaceMAC{
							Address: "52:54:00:39:97:ac",
						},
						Model: &DomainInterfaceModel{
							Type: "virtio",
						},
						Bandwidth: &DomainInterfaceBandwidth{
							Inbound: &DomainInterfaceBandwidthParams{
								Average: &nicAverage,
								Burst:   &nicBurst,
							},
							Outbound: &DomainInterfaceBandwidthParams{
								Average: new(int),
								Burst:   new(int),
							},
						},
					},
				},
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <devices>`,
			`    <interface type="vhostuser">`,
			`      <mac address="52:54:00:39:97:ac"></mac>`,
			`      <model type="virtio"></model>`,
			`      <bandwidth>`,
			`        <inbound average="1000" burst="10000"></inbound>`,
			`        <outbound average="0" burst="0"></outbound>`,
			`      </bandwidth>`,
			`    </interface>`,
			`  </devices>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			Devices: &DomainDeviceList{
				Filesystems: []DomainFilesystem{
					DomainFilesystem{
						Type:       "mount",
						AccessMode: "mapped",
						Driver: &DomainFilesystemDriver{
							Type:     "path",
							WRPolicy: "immediate",
						},
						Source: &DomainFilesystemSource{
							Dir: "/home/user/test",
						},
						Target: &DomainFilesystemTarget{
							Dir: "user-test-mount",
						},
						Address: &DomainAddress{
							Type:     "pci",
							Domain:   &fsAddr.Domain,
							Bus:      &fsAddr.Bus,
							Slot:     &fsAddr.Slot,
							Function: &fsAddr.Function,
						},
					},
					DomainFilesystem{
						Type:       "file",
						AccessMode: "passthrough",
						Driver: &DomainFilesystemDriver{
							Name: "loop",
							Type: "raw",
						},
						Source: &DomainFilesystemSource{
							File: "/home/user/test.img",
						},
						Target: &DomainFilesystemTarget{
							Dir: "user-file-test-mount",
						},
					},
				},
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <devices>`,
			`    <filesystem type="mount" accessmode="mapped">`,
			`      <driver type="path" wrpolicy="immediate"></driver>`,
			`      <source dir="/home/user/test"></source>`,
			`      <target dir="user-test-mount"></target>`,
			`      <address type="pci" domain="0" bus="0" slot="6" function="0"></address>`,
			`    </filesystem>`,
			`    <filesystem type="file" accessmode="passthrough">`,
			`      <driver type="raw" name="loop"></driver>`,
			`      <source file="/home/user/test.img"></source>`,
			`      <target dir="user-file-test-mount"></target>`,
			`    </filesystem>`,
			`  </devices>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			Features: &DomainFeatureList{
				PAE:     &DomainFeature{},
				ACPI:    &DomainFeature{},
				APIC:    &DomainFeatureAPIC{},
				HAP:     &DomainFeatureState{},
				PrivNet: &DomainFeature{},
				HyperV: &DomainFeatureHyperV{
					Relaxed: &DomainFeatureState{State: "on"},
					VAPIC:   &DomainFeatureState{State: "on"},
					Spinlocks: &DomainFeatureHyperVSpinlocks{
						DomainFeatureState{State: "on"}, 4096,
					},
					VPIndex: &DomainFeatureState{State: "on"},
					Runtime: &DomainFeatureState{State: "on"},
					Synic:   &DomainFeatureState{State: "on"},
					Reset:   &DomainFeatureState{State: "on"},
					VendorId: &DomainFeatureHyperVVendorId{
						DomainFeatureState{State: "on"}, "KVM Hv",
					},
				},
				KVM: &DomainFeatureKVM{
					Hidden: &DomainFeatureState{State: "on"},
				},
				PVSpinlock: &DomainFeatureState{State: "on"},
				GIC:        &DomainFeatureGIC{Version: "2"},
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <features>`,
			`    <pae></pae>`,
			`    <acpi></acpi>`,
			`    <apic></apic>`,
			`    <hap></hap>`,
			`    <privnet></privnet>`,
			`    <hyperv>`,
			`      <relaxed state="on"></relaxed>`,
			`      <vapic state="on"></vapic>`,
			`      <spinlocks state="on" retries="4096"></spinlocks>`,
			`      <vpindex state="on"></vpindex>`,
			`      <runtime state="on"></runtime>`,
			`      <synic state="on"></synic>`,
			`      <reset state="on"></reset>`,
			`      <vendor_id state="on" value="KVM Hv"></vendor_id>`,
			`    </hyperv>`,
			`    <kvm>`,
			`      <hidden state="on"></hidden>`,
			`    </kvm>`,
			`    <pvspinlock state="on"></pvspinlock>`,
			`    <gic version="2"></gic>`,
			`  </features>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			Devices: &DomainDeviceList{
				Controllers: []DomainController{
					DomainController{
						Type:  "usb",
						Index: &uhciIndex,
						Model: "piix3-uhci",
						Address: &DomainAddress{
							Type:     "pci",
							Domain:   &uhciAddr.Domain,
							Bus:      &uhciAddr.Bus,
							Slot:     &uhciAddr.Slot,
							Function: &uhciAddr.Function,
						},
					},
					DomainController{
						Type:  "usb",
						Index: nil,
						Model: "ehci",
					},
				},
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <devices>`,
			`    <controller type="usb" index="0" model="piix3-uhci">`,
			`      <address type="pci" domain="0" bus="0" slot="1" function="2"></address>`,
			`    </controller>`,
			`    <controller type="usb" model="ehci"></controller>`,
			`  </devices>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "qemu",
			Name: "test",
			QEMUCommandline: &DomainQEMUCommandline{
				Args: []DomainQEMUCommandlineArg{
					DomainQEMUCommandlineArg{Value: "-newarg"},
					DomainQEMUCommandlineArg{Value: "-oldarg"},
				},
				Envs: []DomainQEMUCommandlineEnv{
					DomainQEMUCommandlineEnv{Name: "QEMU_ENV", Value: "VAL"},
					DomainQEMUCommandlineEnv{Name: "QEMU_VAR", Value: "VAR"},
				},
			},
		},
		Expected: []string{
			`<domain type="qemu">`,
			`  <name>test</name>`,
			`  <commandline xmlns="http://libvirt.org/schemas/domain/qemu/1.0">`,
			`    <arg value="-newarg"></arg>`,
			`    <arg value="-oldarg"></arg>`,
			`    <env name="QEMU_ENV" value="VAL"></env>`,
			`    <env name="QEMU_VAR" value="VAR"></env>`,
			`  </commandline>`,
			`</domain>`,
		},
	},

	/* Tests for sub-documents that can be hotplugged */
	{
		Object: &DomainController{
			Type:  "usb",
			Index: &uhciIndex,
			Model: "piix3-uhci",
			Address: &DomainAddress{
				Type:     "pci",
				Domain:   &uhciAddr.Domain,
				Bus:      &uhciAddr.Bus,
				Slot:     &uhciAddr.Slot,
				Function: &uhciAddr.Function,
			},
		},
		Expected: []string{
			`<controller type="usb" index="0" model="piix3-uhci">`,
			`  <address type="pci" domain="0" bus="0" slot="1" function="2"></address>`,
			`</controller>`,
		},
	},
	{
		Object: &DomainDisk{
			Type:   "file",
			Device: "cdrom",
			Driver: &DomainDiskDriver{
				Name: "qemu",
				Type: "qcow2",
			},
			Source: &DomainDiskSource{
				File: "/var/lib/libvirt/images/demo.qcow2",
			},
			Target: &DomainDiskTarget{
				Dev: "vda",
				Bus: "virtio",
			},
			Serial: "fishfood",
			WWN:    "0123456789abcdef",
		},
		Expected: []string{
			`<disk type="file" device="cdrom">`,
			`  <driver name="qemu" type="qcow2"></driver>`,
			`  <source file="/var/lib/libvirt/images/demo.qcow2"></source>`,
			`  <target dev="vda" bus="virtio"></target>`,
			`  <serial>fishfood</serial>`,
			`  <wwn>0123456789abcdef</wwn>`,
			`</disk>`,
		},
	},
	{
		Object: &DomainFilesystem{
			Type:       "mount",
			AccessMode: "mapped",
			Driver: &DomainFilesystemDriver{
				Type:     "path",
				WRPolicy: "immediate",
			},
			Source: &DomainFilesystemSource{
				Dir: "/home/user/test",
			},
			Target: &DomainFilesystemTarget{
				Dir: "user-test-mount",
			},
			Address: &DomainAddress{
				Type:     "pci",
				Domain:   &fsAddr.Domain,
				Bus:      &fsAddr.Bus,
				Slot:     &fsAddr.Slot,
				Function: &fsAddr.Function,
			},
		},

		Expected: []string{
			`<filesystem type="mount" accessmode="mapped">`,
			`  <driver type="path" wrpolicy="immediate"></driver>`,
			`  <source dir="/home/user/test"></source>`,
			`  <target dir="user-test-mount"></target>`,
			`  <address type="pci" domain="0" bus="0" slot="6" function="0"></address>`,
			`</filesystem>`,
		},
	},
	{
		Object: &DomainInterface{
			Type: "network",
			MAC: &DomainInterfaceMAC{
				Address: "00:11:22:33:44:55",
			},
			Model: &DomainInterfaceModel{
				Type: "virtio",
			},
		},
		Expected: []string{
			`<interface type="network">`,
			`  <mac address="00:11:22:33:44:55"></mac>`,
			`  <model type="virtio"></model>`,
			`</interface>`,
		},
	},
	{
		Object: &DomainSerial{
			Type: "pty",
			Target: &DomainSerialTarget{
				Type: "isa",
				Port: &serialPort,
			},
		},

		Expected: []string{
			`<serial type="pty">`,
			`  <target type="isa" port="0"></target>`,
			`</serial>`,
		},
	},
	{
		Object: &DomainConsole{
			Type: "pty",
			Target: &DomainConsoleTarget{
				Type: "virtio",
				Port: &serialPort,
			},
		},

		Expected: []string{
			`<console type="pty">`,
			`  <target type="virtio" port="0"></target>`,
			`</console>`,
		},
	},
	{
		Object: &DomainInput{
			Type: "tablet",
			Bus:  "usb",
			Address: &DomainAddress{
				Type: "usb",
				Bus:  &tabletBus,
				Port: &tabletPort,
			},
		},

		Expected: []string{
			`<input type="tablet" bus="usb">`,
			`  <address type="usb" bus="0" port="1"></address>`,
			`</input>`,
		},
	},
	{
		Object: &DomainVideo{
			Model: DomainVideoModel{
				Type:   "cirrus",
				Heads:  1,
				Ram:    4096,
				VRam:   8192,
				VGAMem: 256,
			},
			Address: &DomainAddress{
				Type:     "pci",
				Domain:   &videoAddr.Domain,
				Bus:      &videoAddr.Bus,
				Slot:     &videoAddr.Slot,
				Function: &videoAddr.Function,
			},
		},

		Expected: []string{
			`<video>`,
			`  <model type="cirrus" heads="1" ram="4096" vram="8192" vgamem="256"></model>`,
			`  <address type="pci" domain="0" bus="0" slot="5" function="0"></address>`,
			`</video>`,
		},
	},
	{
		Object: &DomainChannel{
			Type: "pty",
			Target: &DomainChannelTarget{
				Type:  "virtio",
				Name:  "org.redhat.spice",
				State: "connected",
			},
		},

		Expected: []string{
			`<channel type="pty">`,
			`  <target type="virtio" name="org.redhat.spice" state="connected"></target>`,
			`</channel>`,
		},
	},
	{
		Object: &DomainMemBalloon{
			Model: "virtio",
			Address: &DomainAddress{
				Type:     "pci",
				Domain:   &balloonAddr.Domain,
				Bus:      &balloonAddr.Bus,
				Slot:     &balloonAddr.Slot,
				Function: &balloonAddr.Function,
			},
		},

		Expected: []string{
			`<memballoon model="virtio">`,
			`  <address type="pci" domain="0" bus="0" slot="7" function="0"></address>`,
			`</memballoon>`,
		},
	},
	{
		Object: &DomainSound{
			Model: "ich6",
			Codec: &DomainSoundCodec{
				Type: "duplex",
			},
			Address: &DomainAddress{
				Type:     "pci",
				Domain:   &duplexAddr.Domain,
				Bus:      &duplexAddr.Bus,
				Slot:     &duplexAddr.Slot,
				Function: &duplexAddr.Function,
			},
		},

		Expected: []string{
			`<sound model="ich6">`,
			`  <codec type="duplex"></codec>`,
			`  <address type="pci" domain="0" bus="0" slot="8" function="0"></address>`,
			`</sound>`,
		},
	},
	{
		Object: &DomainRNG{
			Model: "virtio",
			Rate: &DomainRNGRate{
				Period: 2000,
				Bytes:  1234,
			},
			Backend: &DomainRNGBackend{
				Device: "/dev/random",
				Model:  "random",
			},
		},

		Expected: []string{
			`<rng model="virtio">`,
			`  <rate bytes="1234" period="2000"></rate>`,
			`  <backend model="random">/dev/random</backend>`,
			`</rng>`,
		},
	},
	{
		Object: &DomainRNG{
			Model: "virtio",
			Rate: &DomainRNGRate{
				Period: 2000,
				Bytes:  1234,
			},
			Backend: &DomainRNGBackend{
				Model: "egd",
				Type:  "udp",
				Sources: []DomainInterfaceSource{
					DomainInterfaceSource{
						Mode:    "bind",
						Service: "1234",
					},
					DomainInterfaceSource{
						Mode:    "connect",
						Host:    "1.2.3.4",
						Service: "1234",
					},
				},
			},
		},

		Expected: []string{
			`<rng model="virtio">`,
			`  <rate bytes="1234" period="2000"></rate>`,
			`  <backend model="egd" type="udp">`,
			`    <source mode="bind" service="1234"></source>`,
			`    <source mode="connect" service="1234" host="1.2.3.4"></source>`,
			`  </backend>`,
			`</rng>`,
		},
	},
}

func TestDomain(t *testing.T) {
	for _, test := range domainTestData {
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
