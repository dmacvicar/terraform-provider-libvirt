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
* `metadata` - (Optional) A string containing arbitrary data. This is going to be
  added to the final domain inside of the [metadata tag](https://libvirt.org/formatdomain.html#elementsMetadata).
  This can be used to integrate terraform into other tools by inspecting the
  the contents of the `terraform.tf` file.
* `cloudinit` - (Optional) The `libvirt_cloudinit` disk that has to be used by
  the domain. This is going to be attached as a CDROM ISO. Changing the
  cloud-init won't cause the domain to be recreated, however the change will
  have effect on the next reboot.

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
  base_volume_id = "${libvirt_volume.leap.id}"
}

resource "libvirt_domain" "domain1" {
  name = "domain1"
  disk {
    volume_id = "${libvirt_volume.mydisk.id}"
  }

  network_interface {
    network_id = "${libvirt_network.net1.id}"

    hostname = "master"
    address = "10.17.3.3"
    mac = "AA:BB:CC:11:22:22"
    wait_for_lease = 1
  }
}
```

The `network_interface` block supports:

* `network_name` - (Optional) The name of an _existing_ network to attach this interface to. The network will _NOT_ be managed by the Terraform/libvirt provider.
* `network_id` - (Optional) The ID of a network resource to attach this interface to. The network will be under the control of the Terraform/libvirt provider.
* `mac` - (Optional) The specific MAC address to use for this interface.
* `ip` - (Optional) An IP address for this domain in this network
* `hostname` - (Optional) A hostname that will be assigned to this domain resource in this network.
* `wait_for_lease`- (Optional) When creating the domain resource, wait until the network interface gets a DHCP lease from libvirt, so that the computed ip addresses will be available when the domain is up and the plan applied.
