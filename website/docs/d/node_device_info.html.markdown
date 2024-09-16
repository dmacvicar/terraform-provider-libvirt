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
  * `devnode`: For devices with a `/dev` special file. A mandatory attribute type specify the kind of file `path`, which may be either dev for the main name, or `link` for additional symlinks.

  For `pci` devices the additional attributes are:
  * `class` - Device PCI class
  * `domain` - Device PCI domain
  * `bus` - Device PCI bus
  * `slot` - Device PCI slot
  * `function` - Device PCI function
  * `product` - Device PCI product map of `id` and `name`
  * `vendor` - Device PCI vendor map of `id` and `name`
  * `iommu_group` - Structure that holds IOMMU Group `number` and the list of devices that are part of the group

  For `storage` devices the additional attributes are:
  * `block` - Block device name
  * `drive_type` - Device drive type
  * `model` - Device model
  * `serial` - Device serial number
  * `size` - Device size in bytes
  * `logical_block_size` - Device logical block size
  * `num_blocks` - Number of blocks on the device

  For `usb` devices the additional attributes are:
  * `number` - Device number
  * `class` - Device class
  * `subclass` - Device subclass
  * `protocol` - Device protocol

  For `usb_device` devices the additional attributes are:
  * `bus` - Which bus the device belongs to
  * `device` - Which device within the \
  * `product` - If present, the product `id` and `name` from the device ROM
  * `vendor` - If present, the vendor `id` and `name` from the device ROM

  For `net` devices the additional attributes are:
  * `interface` - The interface name tied to this device
  * `address` - If present, the MAC address of the device
  * `link` - Optional to reflect the status of the link via `speed` and `state` keys
  * `feature` - List of features supported by the network interface
  * `capability` - Holds key `type` that describes the type of network interface: `80203` for IEEE 802.3 or `80211` for IEEE 802.11

  For `scsi_host` devices the additional attributes are:
  * `host` - The SCSI host number
  * `unique_id` - This optionally provides the value from the 'unique_id' file found in the scsi_host's directory

  For `scsi` devices the additional attributes are:
  * `host` - The SCSI host containing the device
  * `bus` - The bus within the host
  * `target` - The target within the bus
  * `lun` - The lun within the target
  * `scsi_type` - The type of SCSI device

  For `drm` devices the additional attributes are:
  * `drm_type` - Type of DRM device: `render` or `card`

* `parent` - The parent of this device in the hierarchy
* `path` - Full path of the device
* `xml` - The XML returned by the libvirt API call
* `devnode` - For type `drm` holds the `path` and `link` that point to the device
