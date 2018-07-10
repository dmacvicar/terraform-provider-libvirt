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

  # the addresses allowed for domains connected and served by the DHCP server
  addresses = ["10.17.3.0/24", "2001:db8:ca2:2::1/64"]

  # (optional) the bridge device defines the name of a bridge device
  # which will be used to construct the virtual network.
  # (only necessary in "bridge" mode)
  # bridge = "br7"

  # (Optional) DNS configuration
  dns {
    local_only = true

    # (Optional) one or more DNS forwarder entries.  One or both of
    # "address" and "domain" must be specified.  The format is:
    # forwarder {
    #   address = "my address"
    #   domain = "my domain"
    # }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the resource, required by libvirt.
  Changing this forces a new resource to be created.
* `domain` - (Optional) The domain used by the DNS server.
* `addresses` - (Optional) A list of (0 or 1) IPv4 and (0 or 1) IPv6 subnets in CIDR notation
  format for being served by the DHCP server. Address of subnet should be used.
  No DHCP server will be started if this attributed is omitted.
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
* `autostart` - (Optional) Set to `true` to start the network on host boot up.
  If not specified `false` is assumed.
* `dns` - (Optional) DNS configuration
  * `enabled` - (Optional) when false, disable the DNS server
  * `local_only` - (Optional) when set, then DNS requests for this domain will
  only be resolved by the virtual network's own DNS server  (they will not be
  forwarded to the host's upstream DNS server)
  * `host` - (Optional) the host element within DNS is the definition of DNS hosts
  to be passed to the DNS service. The IP address is identified by the `address` attribute
  and the names for that IP address are identified in the `name` sub-elements of
  the host element.
```hcl
resource "libvirt_network" "my_network" {
  ...
  dns {
    host {
      address = "10.17.3.2"
      name = ["server1.com", "server2.com"]
    }
  }
}
```
  * `forwarder` - (Optional) a list of DNS forwarders, with entries following
  the `[Domain ->] Domain|IP` format. Each forwarder element defines an alternate DNS
  server to use for some, or all, DNS requests sent to this network's DNS server.
  There are two attributes: a `Domain` and/or an `IP` (at least one of these must be specified).
    - If both `Domain` and `IP` are specified, then all requests that match the given `Domain` will
    be forwarded to the DNS server at `IP`.
    - If only `Domain` is specified, then all matching
    domains will be resolved locally (or via the host's standard DNS forwarding if they can't
    be resolved locally)
    - If an `IP` is specified by itself, then all DNS requests to the
    network's DNS server will be forwarded to the DNS server at that address with no
    exceptions.
  For example:
:
```hcl
resource "libvirt_network" "my_network" {
  ...
  dns {
    forwarders = ["8.8.8.8", "my.domain.com -> 10.10.0.67"]  
  }
}
```
* `dhcp` - (Optional) DHCP configuration
  * `enabled` - (Optional) when false, disable the DHCP server
* `routes` - (Optional) List of static routes, as a list of `CIDR -> gateway`. For example:
```hcl
resource "libvirt_network" "my_network" {
  ...
  routes = [
     "192.168.7.0/24 -> 127.0.0.1",
     "192.168.9.1/24 -> 127.0.0.1",
     "192.168.17.1/32 -> 127.0.0.1",
     "2001:db9:4:1::/64 -> 2001:db8:ca2:2::3"
  ]
}
```


## Attributes Reference

* `id` - a unique identifier for the resource
