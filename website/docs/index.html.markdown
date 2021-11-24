---
layout: "libvirt"
page_title: "Provider: libvirt"
sidebar_current: "docs-libvirt-index"
description: |-
  The Libvirt provider is used to interact with Linux KVM/libvirt hypervisors. The provider needs to be configured with the proper connection information before it can be used.
---

# Libvirt Provider

The Libvirt provider is used to interact with Linux
[libvirt](https://libvirt.org) hypervisors.

The provider needs to be configured with the proper connection information
before it can be used.

~> **Note:** while libvirt can be used with several types of hypervisors, this
provider focuses on [KVM](http://libvirt.org/drvqemu.html). Other drivers may not be
working and haven't been tested.

## The connection URI

The provider understands [connection URIs](https://libvirt.org/uri.html). The supported transports are:

* `tcp` (non-encrypted connection)
* `unix` (UNIX domain socket)
* `tls` (See [here](https://www.libvirt.org/tlscerts.html) for information how to setup certificates)
* `ssh` (Secure shell)

Unlike the original libvirt, the `ssh` transport is not implemented using the ssh command and therefore does not require `nc` (netcat) on the server side.

Additionally, the `ssh` URI supports passwords using the `driver+ssh://[username:PASSWORD@][hostname][:port]/[path]?sshauth=ssh-password` syntax.

As the provider does not use libvirt on the client side, not all connection URI options are supported or apply.

## Example Usage

```hcl
# Configure the Libvirt provider
provider "libvirt" {
  uri = "qemu:///system"
}

# Create a new domain
resource "libvirt_domain" "test1" {
  ...
}
```

## Configuration Reference

The following keys can be used to configure the provider.

* `uri` - (Optional) The [connection URI](https://libvirt.org/uri.html) used
  to connect to the libvirt host.
* `host` - (Optional) An element describing a libvirt host. One may add a list of `host` to describe a full cluster.

At least one `uri` or a `host` element is required.

When using a `host`, users can specify:

* `region` - (Optional) The region the libvirt instance is associated to.
* `az` - (Optional) The availability-zone the libvirt instance is associated to.
* `uri` - (Optional) The [connection URI](https://libvirt.org/uri.html) used
  to connect to the libvirt host.

# Multiple Hosts Configuration

As Terraform is not able to iterate over providers, one can specify multiple hosts with associated
regions and availability zones to allow resources to interate over those and use the correct libvirt connection.

```hcl
# Configure the Libvirt provider with multiple hosts
provider "libvirt" {
  host {
    region = "us-west"
    az     = "1"
    uri    = "qemu+tcp://libvirt-us-west-1.acme.com/system"
  }
  host {
    region = "us-east"
    az     = "1"
    uri    = "qemu+tcp://libvirt-us-east-1.acme.com/system"
  }
}

locals {
  instances = {
    "1" = { region = "us-west", az = "1" },
    "2" = { region = "us-east", az = "2" },
  }
}

# Create a new domain with the same configuration on all regions
resource "libvirt_domain" "test" {
  for_each = local.instances
  region = each.value.region
  ...
}
```

## Environment variables

The libvirt connection URI can also be specified with the `LIBVIRT_DEFAULT_URI`
shell environment variable.

```hcl
$ export LIBVIRT_DEFAULT_URI="qemu+ssh://root@192.168.1.100/system"
$ terraform plan
```
