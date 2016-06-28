---
layout: "libvirt"
page_title: "Libvirt: libvirt_network"
sidebar_current: "docs-libvirt-network"
description: |-
  Manages a virtual machine (network) in libvirt
---

# libvirt\_network

Manages a VM network resource within libvirt. For more information see
[the official documentation](https://libvirt.org/formatnetwork.html).

## Example Usage

```
resource "libvirt_network" "network1" {
  # the name used by libvirt
  name = "k8snet"

  # mode can be: "nat", "isolated"
  mode = "nat"

  #  the domain used by the DNS server in this network
  domain = "k8s.local"

  # the addresses allowed for domains connected and served by the DHCP server
  addresses = ["10.17.3.0/24"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the resource, required by libvirt.
   Changing this forces a new resource to be created.
* `domain` - The domain used by the DNS server
* `addresses` - A list of (0 or 1) ipv4 and (0 or 1) ipv6 ranges for being
   served by the DHCP server.

