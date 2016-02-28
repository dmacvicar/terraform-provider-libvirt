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

