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

  # (Optional) DNS configuration
  dns {
    # (Optional, default false)
    # true: DNS requests under this domain will only be resolved by the
    # virtual network's own DNS server
    # false: Unresolved requests will be forwarded to the host's
    # upstream DNS server if the virtual network's DNS server does not
    # have an answer.
￼   local_only = true

    # (Optional) one or more DNS forwarder entries.  One or both of
    # "address" and "domain" must be specified.  The format is:
    # forwarders = [
    #   {
    #     address = "my address"
    #     domain = "my domain"
    #   }
    # ]

    # (Optional) one or more DNS host entries.  Both of
    # "ip" and "hostname" must be specified.  The format is:
    # hosts = [
    #   {
    #     hostname = "my_hostname"
    #     ip = "my.ip.address.1"
    #   },
    #   {
    #     hostname = "my_hostname"
    #     ip = "my.ip.address.2"
    #   },
    # ]
  }
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
* `dhcp` - DHCP configuration of network:
  * `enabled = true` enables DHCP server, `addresses` field is also used to define the
	address range served by the DHCP server.
  No DHCP server will be started if `addresses` is omitted.
  * `hosts`: list of static ip address assignments
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
* `dns` - (Optional) configuration of DNS specific settings for the network

Inside of `dns` section the following argument are supported:
* `local_only` - (Optional) true/false: true means 'do not forward unresolved requests for this domain to the part DNS server
* `forwarders` - (Optional) Either `address`, `domain`, or both must be set
* `srvs` - (Optional) a DNS SRV entry block. You can have one or more of these blocks
   in your DNS definition. You must specify `service` and `protocol`.
* `hosts` - (Optional) a DNS host entry block. You can have one or more of these
   blocks in your DNS definition. You must specify both `ip` and `hostname`.

An advanced example of round-robin DNS (using DNS host templates) follows:

```hcl
resource "libvirt_network" "my_network" {
  ...
  dns = {
    hosts = [ "${flatten(data.libvirt_network_dns_host_template.hosts.*.rendered)}" ]
  }
  ...
}

data "libvirt_network_dns_host_template" "hosts" {
  count = "${var.host_count}"
  ip = "${var.host_ips[count.index]}"
  hostname = "my_host"
}
```

An advanced example of setting up multiple SRV records using DNS SRV templates is:

```hcl
data "libvirt_network_dns_srv_template" "etcd_cluster" {
  count = "${var.etcd_count}"
  service = "etcd-server"
  protocol = "tcp"
  domain = "${discovery_domain}"
  target = "${var.cluster_name}-etcd-${count.index}.${discovery_domain}"
}

resource "libvirt_network" "k8snet" {
  ...
  dns = [{
    srvs = [ "${flatten(data.libvirt_network_dns_srv_template.etcd_cluster.*.rendered)}" ]
  }]
  ...
}
```

* `dhcp` - (Optional) DHCP configuration. 
   You need to use it in conjunction with the addresses variable.
  * `enabled` - (Optional) when false, disable the DHCP server
  * `hosts` - (Optional) list of static ip address assignments
```hcl
resource "libvirt_network" "test_net" {
	name      = "networktest"
	mode      = "nat"
	domain    = "k8s.local"
	addresses = ["10.17.3.0/24"]
	dhcp {
		enabled = true
		hosts = [
			{
				ip = "10.17.3.2",
				mac = "00:11:22:33:44:55",
			},
			{
				ip = "10.17.3.3",
				mac = "00:11:22:33:44:56",
				name = "hostname1"
			},
		]
	}
}
```
	Can be used with data source as well:
```hcl
data "libvirt_network_dhcp_host_template" "test_hosts" {
  count = "${var.hosts_count}"
  ip = "${var.hosts_ips[count.index]}"
  mac = "${var.hosts_macs[count.index]}"
  name = "${var.hosts_names[count.index]}"
}

resource "libvirt_network" "test_net" {
	name      = "networktest"
	mode      = "nat"
	domain    = "k8s.local"
	addresses = ["10.17.3.0/24"]
	dhcp {
		enabled = true
		hosts = [ "${flatten(data.libvirt_network_dns_srv_template.test_hosts.*.rendered)}" ]
	}
}
```

### Altering libvirt's generated network XML definition

The optional `xml` block relates to the generated network XML.

Currently the following attributes are supported:

* `xslt`: specifies a XSLT stylesheet to transform the generated XML definition before creating the network.
  This is used to support features the provider does not allow to set from the schema.
  It is not recommended to alter properties and settings that are exposed to the schema, as terraform will insist in changing them back to the known state.

See the domain option with the same name for more information and examples.

## Attributes Reference

* `id` - a unique identifier for the resource
