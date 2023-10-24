---
layout: "libvirt"
page_title: "Libvirt: libvirt_node_devices"
sidebar_current: "docs-libvirt-node-devices"
description: |-
  Use this data source to get information about the devices present on current node
---

# Data Source: libvirt\_node\_devices

Retrieve information about the devices present on the current node

## Example Usage

```hcl
data "libvirt_node_devices" "node" {
  capability = "pci"
}
```

## Argument Reference

* `capability` - (Optional) The type of device, used to filter the output by capability type.
  Can be one of `system`, `pci`, `usb_device`,  `usb`,  `net`,  `scsi_host`,
  `scsi_target`,  `scsi`,  `storage`, `fc_host`,  `vports`, `scsi_generic`, `drm`,
  `mdev`, `mdev_types`, `ccw`, `css`, `ap_card`, `ap_queue`, `ap_matrix`.
  Defaults to all active devices.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `devices` - A list of devices that match the selected capability
