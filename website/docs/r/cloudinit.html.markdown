---
layout: "libvirt"
page_title: "Libvirt: libvirt_cloudinit"
sidebar_current: "docs-libvirt-cloudinit"
description: |-
  Manages a cloud-init ISO to attach to a domain
---

# libvirt\_cloudinit

Manages a [cloud-init](http://cloudinit.readthedocs.io/) ISO disk that can be
used to customize a domain during first boot.

## Example Usage

```hcl
resource "libvirt_cloudinit" "commoninit" {
  name = "commoninit.iso"
}

```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the resource, required by libvirt.
* `pool` - (Optional) The pool where the resource will be created.
  If not given, the `default` pool will be used.
* `user_data` - (Optional)  cloud-init user data.
* `meta_data` - (Optional)  cloud-init user data.
# `network_config` - (Optional) cloud-init network-config data.

For user_data, network_config and meta_data parameters have a look at:
 http://cloudinit.readthedocs.io/en/latest/topics/datasources/nocloud.html#datasource-nocloud
