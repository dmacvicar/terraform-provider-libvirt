---
layout: "libvirt"
page_title: "Libvirt: libvirt_pool"
sidebar_current: "docs-libvirt-pool"
description: |-
  Manages a storage pool in libvirt
---

# libvirt\_pool

Manages a storage pool in libvirt. Currently only directory-based storage pool are supported. For more information on
storage pools in libvirt, see [the official documentation](https://libvirt.org/formatstorage.html).

**WARNING:** This is experimental API and may change in the future.

## Example Usage

```hcl
# A pool for all cluster volumes
resource "libvirt_pool" "cluster" {
  name = "cluster"
  type = "dir"
  path = "/home/user/cluster_storage"
}

resource "libvirt_volume" "opensuse_leap" {
  name   = "opensuse_leap"
  pool   = libvirt_pool.cluster.name
  source = "http://download.opensuse.org/repositories/Cloud:/Images:/Leap_42.1/images/openSUSE-Leap-42.1-OpenStack.x86_64.qcow2"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the resource, required by libvirt.
* `type` - (Required) The type of the pool. Currently, only "dir" supported.
* `path` - (Optional) The directory where the pool will keep all its volumes. This is only relevant to (and required by)
                      the "dir" type pools.

### Altering libvirt's generated pool XML definition

The optional `xml` block relates to the generated pool XML.

Currently the following attributes are supported:

* `xslt`: specifies a XSLT stylesheet to transform the generated XML definition before creating the pool. This is used
  to support features the provider does not allow to set from the schema. It is not recommended to alter properties and
  settings that are exposed to the schema, as terraform will insist in changing them back to the known state.

See the domain option with the same name for more information and examples.

## Attributes Reference

* `id` - a unique identifier for the resource
