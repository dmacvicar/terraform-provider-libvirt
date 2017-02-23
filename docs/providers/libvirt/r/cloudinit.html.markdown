---
layout: "libvirt"
page_title: "Libvirt: libvirt_cloudinit"
sidebar_current: "docs-libvirt-cloudinit"
description: |-
  Manages a cloud-init ISO to attach to a domain
---

# libvirt\_cloudinit

Manages a [cloud-init](http://cloudinit.readthedocs.io/) ISO disk that can be used to customize a Domain during 1st
boot.

## Example Usage

```
resource "libvirt_cloudinit" "commoninit" {
  name = "commoninit.iso"
  local_hostname = "node"
}

```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the resource, required by libvirt.
* `pool` - (Optional) The pool where the resource will be created.
  If not given, the `default` pool will be used.
* `local_hostname` - (Optional) If specified this is going to be the hostname of
  the domain.
* `ssh_authorized_key` - (Optional) A public ssh key that will be accepted by
  the `root` user.
* `user_data` - (Optional) Raw cloud-init user data. This content will
be merged automatically with the values specified in other arguments
(like `local_hostname`, `ssh_authorized_key`, etc), but they cannot be
specified in both places at the same time.

Any change of the above fields will cause a new resource to be created.
