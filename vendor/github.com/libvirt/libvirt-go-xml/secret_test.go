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

var secretTestData = []struct {
	Object   *Secret
	Expected []string
}{
	{
		Object: &Secret{
			Description: "Demo",
		},
		Expected: []string{
			`<secret>`,
			`  <description>Demo</description>`,
			`</secret>`,
		},
	},
	{
		Object: &Secret{
			Description: "Demo",
			Private:     "yes",
			Ephemeral:   "yes",
			UUID:        "55806c7d-8e93-456f-829b-607d8c198367",
		},
		Expected: []string{
			`<secret ephemeral="yes" private="yes">`,
			`  <description>Demo</description>`,
			`  <uuid>55806c7d-8e93-456f-829b-607d8c198367</uuid>`,
			`</secret>`,
		},
	},
	{
		Object: &Secret{
			Description: "Demo",
			Private:     "yes",
			Ephemeral:   "yes",
			UUID:        "55806c7d-8e93-456f-829b-607d8c198367",
			Usage: &SecretUsage{
				Type:   "volume",
				Volume: "/var/lib/libvirt/images/puppyname.img",
			},
		},
		Expected: []string{
			`<secret ephemeral="yes" private="yes">`,
			`  <description>Demo</description>`,
			`  <uuid>55806c7d-8e93-456f-829b-607d8c198367</uuid>`,
			`  <usage type="volume">`,
			`    <volume>/var/lib/libvirt/images/puppyname.img</volume>`,
			`  </usage>`,
			`</secret>`,
		},
	},
	{
		Object: &Secret{
			Description: "Demo",
			Private:     "yes",
			Ephemeral:   "yes",
			UUID:        "55806c7d-8e93-456f-829b-607d8c198367",
			Usage: &SecretUsage{
				Type: "ceph",
				Name: "mycluster",
			},
		},
		Expected: []string{
			`<secret ephemeral="yes" private="yes">`,
			`  <description>Demo</description>`,
			`  <uuid>55806c7d-8e93-456f-829b-607d8c198367</uuid>`,
			`  <usage type="ceph">`,
			`    <name>mycluster</name>`,
			`  </usage>`,
			`</secret>`,
		},
	},
	{
		Object: &Secret{
			Description: "Demo",
			Private:     "yes",
			Ephemeral:   "yes",
			UUID:        "55806c7d-8e93-456f-829b-607d8c198367",
			Usage: &SecretUsage{
				Type:   "iscsi",
				Target: "libvirtiscsi",
			},
		},
		Expected: []string{
			`<secret ephemeral="yes" private="yes">`,
			`  <description>Demo</description>`,
			`  <uuid>55806c7d-8e93-456f-829b-607d8c198367</uuid>`,
			`  <usage type="iscsi">`,
			`    <target>libvirtiscsi</target>`,
			`  </usage>`,
			`</secret>`,
		},
	},
	{
		Object: &Secret{
			Description: "Demo",
			Private:     "yes",
			Ephemeral:   "yes",
			UUID:        "55806c7d-8e93-456f-829b-607d8c198367",
			Usage: &SecretUsage{
				Type: "tls",
				Name: "Demo cert",
			},
		},
		Expected: []string{
			`<secret ephemeral="yes" private="yes">`,
			`  <description>Demo</description>`,
			`  <uuid>55806c7d-8e93-456f-829b-607d8c198367</uuid>`,
			`  <usage type="tls">`,
			`    <name>Demo cert</name>`,
			`  </usage>`,
			`</secret>`,
		},
	},
}

func TestSecret(t *testing.T) {
	for _, test := range secretTestData {
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
