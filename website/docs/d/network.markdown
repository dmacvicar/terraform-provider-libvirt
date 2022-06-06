---
layout: "libvirt"
page_title: "Libvirt: libvirt_network"
sidebar_current: "docs-libvirt-network"
description: |-
  Manages data source of virtual machine (network) in libvirt
---

# libvirt\_network

Feed your configuration with information of existing virtual network.

## Example Usage

```hcl
data "libvirt_network" "default" {
  name = "default"
}
```

```hcl
resource "libvirt_domain" "domain1" {
  name = "domain1"

  network_interface {
    network_id     = data.libvirt_network.default.id
    hostname       = "master"
    addresses      = ["10.17.3.3"]
    mac            = "AA:BB:CC:11:22:22"
    wait_for_lease = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the virtual network to use.
