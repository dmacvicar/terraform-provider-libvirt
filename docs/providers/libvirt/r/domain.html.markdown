---
layout: "libvirt"
page_title: "Libvirt: libvirt_domain"
sidebar_current: "docs-libvirt-domain"
description: |-
  Manages a virtual machine (domain) in libvirt
---

# libvirt\_domain

Manages a VM domain resource within libvirt. For more information see
[the official documentation](https://libvirt.org/formatdomain.html).

## Example Usage

```
resource "libvirt_domain" "default" {
	name = "test"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the resource, required by libvirt.
   Changing this forces a new resource to be created.
* `memory` - (Optional) The amount of memory in MiB. If not specified the domain will be
   created with 512 MiB of memory
   be used.
* `vcpu` - (Optional) The amount of virtual CPUs. If not specified, a single CPU will be created.
* `disk` - (Optional) An array of one or more disks to attach to the domain. The `disk` object structure is documented below.
* `network_interface` - (Optional) An array of one or more network interfaces to attach to the domain. The `network_interface` object structure is documented below.

The `disk` block supports:

* `volume_id` - (Required) The volume id to use for this disk.

If you have a volume with a template image, create a second volume using the image as the backing volume, and then use the new volume as the volume for the disk. This way the image will not be modified.

```
resource "libvirt_volume" "leap" {
  name = "leap"
  source = "http://someurl/openSUSE_Leap-42.1.qcow2"
}

resource "libvirt_volume" "mydisk" {
  name = "mydisk"
  base_volume = "${libvirt_volume.mydisk.id}"
}

resource "libvirt_domain" "domain1" {
  name = "domain1"
  disk {
    volume_id = "${libvirt_volume.mydisk.id}"
  }
}
```

The `network_interface` block supports:

* `mac` - (Optional) The specific MAC address to use for this interface.
* `network` - (Optional) The network to attach this interface to.
* `wait_for_lease`- (Optional) When creating the domain resource, wait until the network interface gets a DHCP lease from libvirt, so that the computed ip addresses will be available when the domain is up and the plan applied.
