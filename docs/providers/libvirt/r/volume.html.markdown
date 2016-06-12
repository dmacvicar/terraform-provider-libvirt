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

```
resource "libvirt_volume" "opensuse_leap" {
  name = "opensuse_leap"
  source = "http://download.opensuse.org/repositories/Cloud:/Images:/Leap_42.1/images/openSUSE-Leap-42.1-OpenStack.x86_64.qcow2"
}

```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the resource, required by libvirt.
   Changing this forces a new resource to be created.
* `pool` - (Optional) The pool where the resource will be created.
   If not given, the `default` pool will be used.
* `source` - (Optional) If specified, the image will be retrieved from this URL and uploaded into
   libvirt. Only remote HTTP urls are supported for now.
* `size` - (Optional) The size of the volume in bytes (if you don't like this, help fix [this issue](https://github.com/hashicorp/terraform/issues/3287).
   If `source` is specified, `size` will be set to the source image file size.
* `base_volume_id` - (Optional) The backing volume (CoW) to use for this volume.
* `base_volume_name` - (Optional) The name of the backing volume (CoW) to use for this volume. Note well: when `base_volume_pool` is not specified the volume is going to be searched inside of `pool`.
* `base_volume_pool` - (Optional) The name of the pool containing the volume defined by `base_volume_name`.
