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

* `uri` - (Required) The [connection URI](https://libvirt.org/uri.html) used
  to connect to the libvirt host.

## Environment variables

The libvirt connection URI can also be specified with the `LIBVIRT_DEFAULT_URI`
shell environment variable.

```hcl
$ export LIBVIRT_DEFAULT_URI="qemu+ssh://root@192.168.1.100/system"
$ terraform plan
```
