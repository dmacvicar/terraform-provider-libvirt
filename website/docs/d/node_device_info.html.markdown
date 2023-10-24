---
layout: "libvirt"
page_title: "Libvirt: libvirt_node_device_info"
sidebar_current: "docs-libvirt-node-device-info"
description: |-
  Use this data source to get information about a specific device on the current node
---

# Data Source: libvirt\_node\_device\_info

Retrieve information about a specific device on the current node

## Example Usage

```hcl
data "libvirt_node_device_info" "device" {
  name = "pci_0000_00_00_0"
}
```

## Argument Reference

* `name` - (Required) The name of the device name as expected by [libvirt](https://www.libvirt.org/manpages/virsh.html#nodedev-commands).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `capability` - A map of the various attributes for the device, depending on the type of device.
  Currently implemented are `pci`, `storage`, `usb`
  * `type` - Device type: `pci`, `storage`, `usb`

  For `pci` devices the attributes are:
  * `class` - Device PCI class
  * `domain` - Device PCI domain
  * `bus` - Device PCI bus
  * `slot` - Device PCI slot
  * `function` - Device PCI function
  * `product_id` - Device PCI product id
  * `product_name` - Device PCI product name
  * `vendor_id` - Device PCI product id
  * `vendor_name` - Device PCI product name
  * `iommu_group_number` - IOMMU Group number for the device

  For `storage` devices the attributes are:
  * `block`: Block device name
  * `drive_type`: Device drive type
  * `model`: Device model
  * `serial`: Device serial number
  * `size`: Device size in bytes
  * `logical_block_size`: Device logical block size
  * `num_blocks`: Number of blocks on the device

  For `usb` devices the attributes are:
  * `number`: Device number
  * `class`: Device class
  * `subclass`: Device subclass
  * `protocol`: Device protocol

* `parent` - The parent of this device in the hierarchy
* `path` - Full path of the device
* `xml` - The XML returned by the libvirt API call
