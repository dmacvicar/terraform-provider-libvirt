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

```hcl
resource "libvirt_network" "kube_network" {
  # the name used by libvirt
  name = "k8snet"

  # mode can be: "nat" (default), "none", "route", "bridge"
  mode = "nat"

  #  the domain used by the DNS server in this network
  domain = "k8s.local"

  #  list of subnets the addresses allowed for domains connected
  # also derived to define the host addresses
  # also derived to define the addresses served by the DHCP server
  addresses = ["10.17.3.0/24", "2001:db8:ca2:2::1/64"]

  # (optional) the bridge device defines the name of a bridge device
  # which will be used to construct the virtual network.
  # (only necessary in "bridge" mode)
  # bridge = "br7"

  # (Optional) one or more DNS forwarder entries.  One or both of
  # "address" and "domain" must be specified.  The format is:
  # dns_forwarder {
  #   address = "my address"
  #   domain = "my domain"
  # }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the resource, required by libvirt.
  Changing this forces a new resource to be created.
* `domain` - (Optional) The domain used by the DNS server.
* `addresses` - (Optional) A list of (0 or 1) IPv4 and (0 or 1) IPv6 subnets in
  CIDR notation.  This defines the subnets associated to that network.
  This argument is also used to define the address on the real host.
  If `dhcp {  enabled = true }` addresses is also used to define the address range served by
  the DHCP server.
  No DHCP server will be started if `addresses` is omitted.
* `mode` -  One of:
    - `none`: the guests can talk to each other and the host OS, but cannot reach
    any other machines on the LAN.
    - `nat`: it is the default network mode. This is a configuration that
    allows guest OS to get outbound connectivity regardless of whether the host
    uses ethernet, wireless, dialup, or VPN networking without requiring any
    specific admin configuration. In the absence of host networking, it at
    least allows guests to talk directly to each other.
    - `route`: this is a variant on the default network which routes traffic from
    the virtual network to the LAN **without applying any NAT**. It requires that
    the IP address range be pre-configured in the routing tables of the router
    on the host network.
    - `bridge`: use a pre-existing host bridge. The guests will effectively be
    directly connected to the physical network (i.e. their IP addresses will
    all be on the subnet of the physical network, and there will be no
    restrictions on inbound or outbound connections). The `bridge` network
    attribute is mandatory in this case.
* `bridge` - (Optional) The bridge device defines the name of a bridge
   device which will be used to construct the virtual network (when not provided,
   it will be automatically obtained by libvirt in `none`, `nat` and `route` modes).
*  `dns_forwarder` - (Optional) a DNS forwarder entry block.  You can have
   one or mode of these blocks in your network definition.  You must specify one or
   both of `address` and `domain`.  You can use either of the forms below to
   specify dns_forwarders:
* `autostart` - (Optional) Set to `true` to start the network on host boot up.
  If not specified `false` is assumed.

```hcl
resource "libvirt_network" "my_network" {
  ...
  dns_forwarder {
    address = "my address"
  }
  dns_forwarder {
    address = "my address 1"
    domain = "my domain"
  }
}
```

```hcl
resource "libvirt_network" "my_network" {
  ...
  dns_forwarder = [
    {
      address = "my address"
    },
    {
      address = "my address 1"
      domain = "my domain
    }
  ]
}
```

* `dhcp` - (Optional) DHCP configuration. 
   You need to use it in conjuction with the adresses variable.
  * `enabled` - (Optional) when false, disable the DHCP server
```hcl
				resource "libvirt_network" "test_net" {
					name      = "networktest"
					mode      = "nat"
					domain    = "k8s.local"
					addresses = ["10.17.3.0/24"]
					dhcp {
						enabled = true
					}
```
## Attributes Reference

* `id` - a unique identifier for the resource
