// +build xmlroundtrip

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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var xmldirs = []string{
	"testdata/libvirt/tests/bhyveargv2xmldata",
	"testdata/libvirt/tests/bhyvexml2argvdata",
	"testdata/libvirt/tests/bhyvexml2xmloutdata",
	"testdata/libvirt/tests/capabilityschemadata",
	"testdata/libvirt/tests/cputestdata",
	"testdata/libvirt/tests/domaincapsschemadata",
	"testdata/libvirt/tests/domainconfdata",
	"testdata/libvirt/tests/domainschemadata",
	"testdata/libvirt/tests/domainsnapshotxml2xmlin",
	"testdata/libvirt/tests/domainsnapshotxml2xmlout",
	"testdata/libvirt/tests/genericxml2xmlindata",
	"testdata/libvirt/tests/genericxml2xmloutdata",
	"testdata/libvirt/tests/interfaceschemadata",
	"testdata/libvirt/tests/libxlxml2domconfigdata",
	"testdata/libvirt/tests/lxcconf2xmldata",
	"testdata/libvirt/tests/lxcxml2xmldata",
	"testdata/libvirt/tests/lxcxml2xmloutdata",
	"testdata/libvirt/tests/networkxml2confdata",
	"testdata/libvirt/tests/networkxml2firewalldata",
	"testdata/libvirt/tests/networkxml2xmlin",
	"testdata/libvirt/tests/networkxml2xmlout",
	"testdata/libvirt/tests/networkxml2xmlupdatein",
	"testdata/libvirt/tests/networkxml2xmlupdateout",
	"testdata/libvirt/tests/nodedevschemadata",
	"testdata/libvirt/tests/nwfilterxml2firewalldata",
	"testdata/libvirt/tests/nwfilterxml2xmlin",
	"testdata/libvirt/tests/nwfilterxml2xmlout",
	"testdata/libvirt/tests/qemuagentdata",
	"testdata/libvirt/tests/qemuargv2xmldata",
	"testdata/libvirt/tests/qemucapabilitiesdata",
	"testdata/libvirt/tests/qemucaps2xmldata",
	"testdata/libvirt/tests/qemuhotplugtestcpus",
	"testdata/libvirt/tests/qemuhotplugtestdevices",
	"testdata/libvirt/tests/qemuhotplugtestdomains",
	"testdata/libvirt/tests/qemumemlockdata",
	"testdata/libvirt/tests/qemuxml2argvdata",
	"testdata/libvirt/tests/qemuxml2xmloutdata",
	"testdata/libvirt/tests/secretxml2xmlin",
	"testdata/libvirt/tests/securityselinuxlabeldata",
	"testdata/libvirt/tests/sexpr2xmldata",
	"testdata/libvirt/tests/storagepoolschemadata",
	"testdata/libvirt/tests/storagepoolxml2xmlin",
	"testdata/libvirt/tests/storagepoolxml2xmlout",
	"testdata/libvirt/tests/storagevolschemadata",
	"testdata/libvirt/tests/storagevolxml2xmlin",
	"testdata/libvirt/tests/storagevolxml2xmlout",
	"testdata/libvirt/tests/vircaps2xmldata",
	"testdata/libvirt/tests/virstorageutildata",
	"testdata/libvirt/tests/vmx2xmldata",
	"testdata/libvirt/tests/xlconfigdata",
	"testdata/libvirt/tests/xmconfigdata",
	"testdata/libvirt/tests/xml2sexprdata",
	"testdata/libvirt/tests/xml2vmxdata",
}

var consoletype = "/domain[0]/devices[0]/console[0]/@type"
var volsrc = "/volume[0]/source[0]"

var blacklist = map[string]bool{
	// intentionally invalid xml
	"testdata/libvirt/tests/genericxml2xmlindata/chardev-unix-redirdev-missing-path.xml":  true,
	"testdata/libvirt/tests/genericxml2xmlindata/chardev-unix-rng-missing-path.xml":       true,
	"testdata/libvirt/tests/qemuxml2argvdata/virtio-rng-egd-crash.xml":                    true,
	"testdata/libvirt/tests/genericxml2xmlindata/chardev-unix-smartcard-missing-path.xml": true,
	"testdata/libvirt/tests/genericxml2xmlindata/chardev-tcp-multiple-source.xml":         true,
	"testdata/libvirt/tests/networkxml2xmlupdatein/dns-host-gateway-incomplete.xml":       true,
	"testdata/libvirt/tests/networkxml2xmlupdatein/host-new-incomplete.xml":               true,
	"testdata/libvirt/tests/networkxml2xmlupdatein/unparsable-dns-host.xml":               true,
	// udp source in different order
	"testdata/libvirt/tests/genericxml2xmlindata/chardev-udp.xml":                 true,
	"testdata/libvirt/tests/genericxml2xmlindata/chardev-udp-multiple-source.xml": true,
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-fv-serial-udp.xml":            true,
}

var extraActualNodes = map[string][]string{

	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-pv-vcpus.xml":              []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-pv.xml":                    []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-pv-bootloader.xml":         []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-pv-bootloader-cmdline.xml": []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-pci-devs.xml":              []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-net-routed.xml":            []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-net-e1000.xml":             []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-net-bridged.xml":           []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-fv-kernel.xml":             []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-escape.xml":                []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-disk-file.xml":             []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-disk-drv-loop.xml":         []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-disk-drv-blktap2.xml":      []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-disk-drv-blktap2-raw.xml":  []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-disk-drv-blktap.xml":       []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-disk-drv-blktap-raw.xml":   []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-disk-drv-blktap-qcow.xml":  []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-disk-drv-blkback.xml":      []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-disk-block.xml":            []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-disk-block-shareable.xml":  []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-bridge-ipaddr.xml":         []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-boot-grub.xml":             []string{consoletype},
	"testdata/libvirt/tests/xml2sexprdata/xml2sexpr-no-source-cdrom.xml": []string{
		"/domain[0]/devices[0]/disk[1]/@type",
	},

	"testdata/libvirt/tests/qemuxml2argvdata/fs9p-ccw.xml": []string{
		"/domain[0]/devices[0]/filesystem[1]/@type",
		"/domain[0]/devices[0]/filesystem[2]/@type",
	},
	"testdata/libvirt/tests/qemuxml2argvdata/fs9p.xml": []string{
		"/domain[0]/devices[0]/filesystem[1]/@type",
		"/domain[0]/devices[0]/filesystem[2]/@type",
	},
	"testdata/libvirt/tests/qemuxml2argvdata/disk-drive-discard.xml": []string{
		"/domain[0]/devices[0]/disk[0]/@type",
	},
	"testdata/libvirt/tests/genericxml2xmlindata/chardev-udp.xml": []string{
		"/domain[0]/devices[0]/channel[0]/source[0]/@mode",
	},
	"testdata/libvirt/tests/qemuxml2argvdata/disk-mirror-old.xml": []string{
		"/domain[0]/devices[0]/disk[0]/mirror[0]/@type",
		"/domain[0]/devices[0]/disk[0]/mirror[0]/source[0]",
		"/domain[0]/devices[0]/disk[2]/mirror[0]/@type",
		"/domain[0]/devices[0]/disk[2]/mirror[0]/format[0]",
		"/domain[0]/devices[0]/disk[2]/mirror[0]/source[0]",
	},

	"testdata/libvirt/tests/networkxml2xmlin/openvswitch-net.xml": []string{
		"/network[0]/virtualport[0]/parameters[0]",
	},
	"testdata/libvirt/tests/networkxml2xmlout/openvswitch-net.xml": []string{
		"/network[0]/virtualport[0]/parameters[0]",
	},
	"testdata/libvirt/tests/networkxml2xmlupdateout/openvswitch-net-modified.xml": []string{
		"/network[0]/virtualport[0]/parameters[0]",
	},
	"testdata/libvirt/tests/networkxml2xmlupdateout/openvswitch-net-more-portgroups.xml": []string{
		"/network[0]/virtualport[0]/parameters[0]",
	},
	"testdata/libvirt/tests/networkxml2xmlupdateout/openvswitch-net-without-alice.xml": []string{
		"/network[0]/virtualport[0]/parameters[0]",
	},
	"testdata/libvirt/tests/interfaceschemadata/bridge-vlan.xml": []string{
		"/interface[0]/bridge[0]/interface[0]/vlan[0]/interface[0]/@type",
	},
	"testdata/libvirt/tests/interfaceschemadata/vlan.xml": []string{
		"/interface[0]/vlan[0]/interface[0]/@type",
	},
	"testdata/libvirt/tests/domainsnapshotxml2xmlin/disk_driver_name_null.xml": []string{
		"/domainsnapshot[0]/disks[0]/disk[0]/@type",
	},
	"testdata/libvirt/tests/domainsnapshotxml2xmlin/disk_snapshot.xml": []string{
		"/domainsnapshot[0]/disks[0]/disk[0]/@type",
		"/domainsnapshot[0]/disks[0]/disk[1]/@type",
		"/domainsnapshot[0]/disks[0]/disk[2]/@type",
		"/domainsnapshot[0]/disks[0]/disk[3]/@type",
		"/domainsnapshot[0]/disks[0]/disk[4]/@type",
	},
	"testdata/libvirt/tests/domainsnapshotxml2xmlout/disk_snapshot.xml": []string{
		"/domainsnapshot[0]/disks[0]/disk[0]/@type",
		"/domainsnapshot[0]/disks[0]/disk[1]/@type",
		"/domainsnapshot[0]/disks[0]/disk[2]/@type",
	},
	"testdata/libvirt/tests/domainsnapshotxml2xmlout/disk_snapshot_redefine.xml": []string{
		"/domainsnapshot[0]/disks[0]/disk[0]/@type",
		"/domainsnapshot[0]/disks[0]/disk[1]/@type",
		"/domainsnapshot[0]/disks[0]/disk[2]/@type",
	},
	"testdata/libvirt/tests/domainsnapshotxml2xmlout/external_vm_redefine.xml": []string{
		"/domainsnapshot[0]/disks[0]/disk[0]/@type",
	},
}

var extraExpectNodes = map[string][]string{
	"testdata/libvirt/tests/genericxml2xmlindata/chardev-unix.xml": []string{
		"/domain[0]/devices[0]/channel[1]/source[0]",
	},
	"testdata/libvirt/tests/qemuxml2argvdata/usb-redir-filter.xml": []string{
		"/domain[0]/devices[0]/redirfilter[0]/usbdev[1]/@vendor",
		"/domain[0]/devices[0]/redirfilter[0]/usbdev[1]/@product",
		"/domain[0]/devices[0]/redirfilter[0]/usbdev[1]/@class",
		"/domain[0]/devices[0]/redirfilter[0]/usbdev[1]/@version",
	},
	"testdata/libvirt/tests/domainschemadata/domain-parallels-ct-simple.xml": []string{
		"/domain[0]/description[0]",
	},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-file-backing.xml":              []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-file-iso.xml":                  []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-file-naming.xml":               []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-file.xml":                      []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-gluster-dir-neg-uid.xml":       []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-gluster-dir.xml":               []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-logical-backing.xml":           []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-logical.xml":                   []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-luks-cipher.xml":               []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-luks-convert.xml":              []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-luks.xml":                      []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-partition.xml":                 []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-qcow2-0.10-lazy.xml":           []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-qcow2-1.1.xml":                 []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-qcow2-lazy.xml":                []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-qcow2-encryption.xml":          []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-qcow2-nobacking.xml":           []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-qcow2-nocapacity-backing.xml":  []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-qcow2-nocapacity.xml":          []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-qcow2-nocow.xml":               []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-qcow2.xml":                     []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlin/vol-sheepdog.xml":                  []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-file-backing.xml":             []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-file-iso.xml":                 []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-file-naming.xml":              []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-file.xml":                     []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-gluster-dir-neg-uid.xml":      []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-gluster-dir.xml":              []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-logical-backing.xml":          []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-logical.xml":                  []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-luks-cipher.xml":              []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-luks.xml":                     []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-partition.xml":                []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-qcow2-0.10-lazy.xml":          []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-qcow2-1.1.xml":                []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-qcow2-encryption.xml":         []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-qcow2-lazy.xml":               []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-qcow2-nobacking.xml":          []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-qcow2-nocapacity-backing.xml": []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-qcow2-nocapacity.xml":         []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-qcow2-nocow.xml":              []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-qcow2.xml":                    []string{volsrc},
	"testdata/libvirt/tests/storagevolxml2xmlout/vol-sheepdog.xml":                 []string{volsrc},
	"testdata/libvirt/tests/qemuhotplugtestdevices/qemuhotplug-qemu-agent.xml": []string{
		"/channel[0]/source[0]",
	},
	"testdata/libvirt/tests/domainsnapshotxml2xmlin/disk_snapshot.xml": []string{
		"/domainsnapshot[0]/disks[0]/disk[3]/source[0]",
	},
}

func trimXML(xml string) string {
	xml = strings.TrimSpace(xml)
	if strings.HasPrefix(xml, "<?xml") {
		end := strings.Index(xml, "?>")
		if end != -1 {
			xml = xml[end+2:]
			xml = strings.TrimSpace(xml)
		}
	}
	if strings.HasPrefix(xml, "<!--") {
		end := strings.Index(xml, "-->")
		if end != -1 {
			xml = xml[end+3:]
			xml = strings.TrimSpace(xml)
		}
	}
	return xml
}

func testRoundTrip(t *testing.T, xml string, filename string) {
	if strings.HasSuffix(filename, "-invalid.xml") {
		return
	}

	xml = trimXML(xml)

	var doc Document
	if strings.HasPrefix(xml, "<domain ") {
		doc = &Domain{}
	} else if strings.HasPrefix(xml, "<capabilities") {
		doc = &Caps{}
	} else if strings.HasPrefix(xml, "<network") {
		doc = &Network{}
	} else if strings.HasPrefix(xml, "<secret") {
		doc = &Secret{}
	} else if strings.HasPrefix(xml, "<device") {
		doc = &NodeDevice{}
	} else if strings.HasPrefix(xml, "<volume") {
		doc = &StorageVolume{}
	} else if strings.HasPrefix(xml, "<pool") {
		doc = &StoragePool{}
	} else if strings.HasPrefix(xml, "<cpuTest") || strings.HasPrefix(xml, "<cpudata") {
		// Not a public schema
		return
	} else if strings.HasPrefix(xml, "<cpu") {
		if strings.Contains(xml, "mode=") || strings.Contains(xml, "match=") {
			doc = &DomainCPU{}
		} else {
			doc = &CapsHostCPU{}
		}
	} else if strings.HasPrefix(xml, "<filter") {
		doc = &NWFilter{}
	} else if strings.HasPrefix(xml, "<interface") {
		if strings.Contains(filename, "networkxml") {
			doc = &NetworkForwardInterface{}
		} else {
			doc = &Interface{}
		}
	} else if strings.HasPrefix(xml, "<domainsnapshot") {
		doc = &DomainSnapshot{}
	} else if strings.HasPrefix(xml, "<domainCapabilities") {
		doc = &DomainCaps{}
	} else if strings.HasPrefix(xml, "<disk") {
		doc = &DomainDisk{}
	} else if strings.HasPrefix(xml, "<console") {
		doc = &DomainConsole{}
	} else if strings.HasPrefix(xml, "<channel") {
		doc = &DomainChannel{}
	} else if strings.HasPrefix(xml, "<watchdog") {
		doc = &DomainWatchdog{}
	} else if strings.HasPrefix(xml, "<shmem") {
		doc = &DomainShmem{}
	} else if strings.HasPrefix(xml, "<graphics") {
		doc = &DomainGraphic{}
	} else if strings.HasPrefix(xml, "<host") {
		if strings.Contains(xml, "mac=") {
			doc = &NetworkDHCPHost{}
		} else {
			doc = &NetworkDNSHost{}
		}
	} else if strings.HasPrefix(xml, "<portgroup") {
		doc = &NetworkPortGroup{}
	} else if strings.HasPrefix(xml, "<txt") {
		doc = &NetworkDNSTXT{}
	} else if strings.HasPrefix(xml, "<srv") {
		doc = &NetworkDNSSRV{}
	} else if strings.HasPrefix(xml, "<range") {
		doc = &NetworkDHCPRange{}
	} else if strings.HasPrefix(xml, "<qemuCaps") ||
		strings.HasPrefix(xml, "<sources") ||
		strings.HasPrefix(xml, "<cpudata") ||
		strings.HasPrefix(xml, "<cliOutput") {
		// Private libvirt internal XML schemas we don't
		// need public API coverage for
		return
	} else {
		t.Fatal(fmt.Errorf("Unexpected XML document schema in %s\n", filename))
	}
	err := doc.Unmarshal(xml)
	if err != nil {
		t.Fatal(fmt.Errorf("Cannot parse file %s: %s\n", filename, err))
	}

	newxml, err := doc.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	extraExpectNodes, _ := extraExpectNodes[filename]
	extraActualNodes, _ := extraActualNodes[filename]
	err = testCompareXML(filename, xml, newxml, extraExpectNodes, extraActualNodes)
	if err != nil {
		t.Fatal(err)
	}
}

func syncGit(t *testing.T) {
	_, err := os.Stat("testdata/libvirt/tests")
	if err != nil {
		if os.IsNotExist(err) {
			msg, err := exec.Command("git", "clone", "--depth", "1", "https://libvirt.org/git/libvirt.git", "testdata/libvirt").CombinedOutput()
			if err != nil {
				t.Fatal(fmt.Errorf("Unable to clone libvirt.git: %s: %s", err, msg))
			}
		} else {
			t.Fatal(err)
		}
	} else {
		here, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = os.Chdir("testdata/libvirt")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			os.Chdir(here)
		}()
		msg, err := exec.Command("git", "pull").CombinedOutput()
		if err != nil {
			t.Fatal(fmt.Errorf("Unable to update libvirt.git: %s: %s", err, msg))
		}
	}
}

func TestRoundTrip(t *testing.T) {
	syncGit(t)
	for _, xmldir := range xmldirs {
		xmlfiles, err := ioutil.ReadDir(xmldir)
		if err != nil {
			t.Fatal(err)
		}

		for _, xmlfile := range xmlfiles {
			if !xmlfile.IsDir() && strings.HasSuffix(xmlfile.Name(), ".xml") {
				fname := xmldir + "/" + xmlfile.Name()
				_, ok := blacklist[fname]
				if ok {
					continue
				}
				xml, err := ioutil.ReadFile(fname)
				if err != nil {
					t.Fatal(err)
				}
				testRoundTrip(t, string(xml), fname)
			}
		}
	}
}
