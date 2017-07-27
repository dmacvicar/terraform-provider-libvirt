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
[Ignition Provider](/docs/providers/ignition/index.html).

An example where a terraform ignition provider object is used:

```hcl
# Systemd unit resource containing the unit definition
resource "ignition_systemd_unit" "example" {
  name = "example.service"
  content = "[Service]\nType=oneshot\nExecStart=/usr/bin/echo Hello World\n\n[Install]\nWantedBy=multi-user.target"
}

# Ignition config include the previous defined systemd unit resource
resource "ignition_config" "example" {
  systemd = [
      "${ignition_systemd_unit.example.id}",
  ]
}

resource "libvirt_ignition" "ignition" {
  name = "ignition"
  content = "${ignition_config.example.rendered}"
}

resource "libvirt_domain" "my_machine" {
  coreos_ignition = "${libvirt_ignition.ignition.id}"
  ...
}
```
