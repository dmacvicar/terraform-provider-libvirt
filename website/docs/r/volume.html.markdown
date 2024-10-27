---
layout: "libvirt"
page_title: "Libvirt: libvirt_volume"
sidebar_current: "docs-libvirt-volume"
description: |-
  Manages a storage volume in libvirt
---

# libvirt\_volume

Manages a storage volume in libvirt. For more information see
[the official documentation](https://libvirt.org/formatstorage.html).

## Example Usage

```hcl
# Base OS image to use to create a cluster of different
# nodes
resource "libvirt_volume" "opensuse_leap" {
  name   = "opensuse_leap"
  source = "http://download.opensuse.org/repositories/Cloud:/Images:/Leap_42.1/images/openSUSE-Leap-42.1-OpenStack.x86_64.qcow2"
}

# volume to attach to the "master" domain as main disk
resource "libvirt_volume" "master" {
  name           = "master.qcow2"
  base_volume_id = libvirt_volume.opensuse_leap.id
}

# volumes to attach to the "workers" domains as main disk
resource "libvirt_volume" "worker" {
  name           = "worker_${count.index}.qcow2"
  base_volume_id = libvirt_volume.opensuse_leap.id
  count          = var.workers_count
}
```

~> **Tip:** when provisioning multiple domains using the same base image, create
a `libvirt_volume` for the base image and then define the domain specific ones
as based on it. This way the image will not be modified and no extra disk space
is going to be used for the base image.

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the resource, required by libvirt.
  Changing this forces a new resource to be created.
* `pool` - (Optional) The storage pool where the resource will be created.
  If not given, the `default` storage pool will be used.
* `source` - (Optional) If specified, the image will be uploaded into libvirt
  storage pool. It's possible to specify the path to a local (relative to the
  machine running the `terraform` command) image or a remote one. Remote images
  have to be specified using HTTP(S) urls for now.
* `size` - (Optional) The size of the volume in bytes (if you don't like this,
  help fix [this issue](https://github.com/hashicorp/terraform/issues/3287).
  If `source` is specified and `size` is omitted, it will be set to the source image file size.
  `size` can also be omitted if `base_volume_id` or `base_volume_name` is specified. `size` will then be set to the base volume size.
  If `size` is specified to be bigger than `base_volume_id`, `base_volume_name` or `source` size, you can use [cloudinit](https://cloudinit.readthedocs.io) if your OS supports it, with `libvirt_cloudinit_disk` and the [growpart](https://cloudinit.readthedocs.io/en/latest/topics/modules.html#growpart) module to resize the partition.
* `base_volume_id` - (Optional) The backing volume (CoW) to use for this volume.
* `base_volume_name` - (Optional) The name of the backing volume (CoW) to use
  for this volume. Note well: when `base_volume_pool` is not specified the
  volume is going to be searched inside of `pool`.
* `base_volume_pool` - (Optional) The name of the storage pool containing the
  volume defined by `base_volume_name`.

### Altering libvirt's generated volume XML definition

The optional `xml` block relates to the generated volume XML.

Currently the following attributes are supported:

* `xslt`: specifies a XSLT stylesheet to transform the generated XML definition before creating the volume.
  This is used to support features the provider does not allow to set from the schema.
  It is not recommended to alter properties and settings that are exposed to the schema, as terraform will insist in changing them back to the known state.

See the domain option with the same name for more information and examples.

## Attributes Reference

* `id` - a unique identifier for the resource
