---
layout: "libvirt"
page_title: "Libvirt: libvirt_ignition"
sidebar_current: "docs-libvirt-ignition"
description: |-
  Manages a CoreOS Ignition file to supply to a domain
---

# libvirt\_ignition

Manages a [CoreOS Ignition](https://coreos.com/ignition/docs/latest/supported-platforms.html)
file written as a volume to a libvirt storage pool that can be used to customize a CoreOS Domain during 1st
boot.

## Example Usage

```
resource "libvirt_ignition" "ignition" {
  name = "example.ign"
  content = <file-name or ignition object>
}

```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the resource, required by libvirt.
* `pool` - (Optional) The pool where the resource will be created.
  If not given, the `default` pool will be used.
* `content` - (Required) This points to the source of the Ignition configuration
  information that will be used to create the Ignition file in the libvirt
  storage pool.  The `content` can be
  * The name of file that contains Ignition configuration data, or its contents
  * A rendered Terraform Ignition object

Any change of the above fields will cause a new resource to be created.
