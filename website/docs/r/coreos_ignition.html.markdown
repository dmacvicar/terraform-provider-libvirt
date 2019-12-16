---
layout: "libvirt"
page_title: "Libvirt: libvirt_ignition"
sidebar_current: "docs-libvirt-ignition"
description: |-
  Manages a CoreOS Ignition file to supply to a domain
---

# libvirt\_ignition

Manages a [CoreOS Ignition](https://coreos.com/ignition/docs/latest/supported-platforms.html)
file written as a volume to a libvirt storage pool that can be used to customize
a CoreOS Domain during first boot.

~> **Note:** to make use of Ignition files with CoreOS the host must be running QEMU v2.6 or greater.

~> **Note:** On the i686 (x86), x86\_64 (AMD64), and aarch64 (ARM64) platforms,
the QEMU firmware config device (`-fw_cfg`) is used to pass Ignition configs
into the guests. On the s390x (IBM Z) platform however, the QEMU firmware
config device is not supported. As alternative, the OpenStack config-drive
mechanism is instead used. The `mkisofs` utility must be installed into the
path and Ignition must be run with the `--provider=openstack` flag.

## Example Usage

```hcl
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
  * A rendered terraform Ignition object

Any change of the above fields will cause a new resource to be created.

## Integration with Ignition provider

The `libvirt_ignition` resource can be integrated with terraform
[Ignition Provider](https://www.terraform.io/docs/providers/ignition/index.html).

An example where a terraform ignition provider object is used:

```hcl
# Systemd unit data source containing the unit definition
data "ignition_systemd_unit" "example" {
  name = "example.service"
  content = "[Service]\nType=oneshot\nExecStart=/usr/bin/echo Hello World\n\n[Install]\nWantedBy=multi-user.target"
}

# Ignition config include the previous defined systemd unit data source
data "ignition_config" "example" {
  systemd = [
      "data.ignition_systemd_unit.example.rendered",
  ]
}

resource "libvirt_ignition" "ignition" {
  name = "ignition"
  content = data.ignition_config.example.rendered
}

resource "libvirt_domain" "my_machine" {
  coreos_ignition = libvirt_ignition.ignition.id
  ...
}
```
