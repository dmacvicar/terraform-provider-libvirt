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
  (like `local_hostname`, `ssh_authorized_key`, etc). The contents of
  `user_data` will take precedence over the ones defined by the other keys.
* `network_config` - (Optional) Raw NoCloud network-config data.
  The content of this will be placed into the network-config file directly
  in the root of the iso. This allows you to specify static networking
  configurations (such as static IPs and DNS) that are applied from early
  boot.  See:
  http://cloudinit.readthedocs.io/en/latest/topics/datasources/nocloud.html#datasource-nocloud

Any change of the above fields will cause a new resource to be created.
