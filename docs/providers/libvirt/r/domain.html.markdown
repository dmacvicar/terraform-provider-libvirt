---
layout: "libvirt"
page_title: "Libvirt: libvirt_domain"
sidebar_current: "docs-libvirt-domain"
description: |-
  Manages a virtual machine (domain) in libvirt
---

# libvirt\_domain

Manages a VM domain resource within libvirt. For more information see
[the official documentation](https://libvirt.org/formatdomain.html).

## Example Usage

```
resource "libvirt_domain" "default" {
	name = "test"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the resource, required by libvirt.
   Changing this forces a new resource to be created.
* `memory` - (Optional) The amount of memory in MiB. If not specified the domain will be
   created with 512 MiB of memory
   be used.
* `vcpu` - (Optional) The amount of virtual CPUs. If not specified, a single CPU will be created.
* `running` - (Optional) Use `false` to turn off the instance. If not specified, true is assumed and the instance, if stopped, will be started at next apply.
* `disk` - (Optional) An array of one or more disks to attach to the domain. The `disk` object structure is documented below.
* `network_interface` - (Optional) An array of one or more network interfaces to attach to the domain. The `network_interface` object structure is documented below.
* `metadata` - (Optional) A string containing arbitrary data. This is going to be
  added to the final domain inside of the [metadata tag](https://libvirt.org/formatdomain.html#elementsMetadata).
  This can be used to integrate terraform into other tools by inspecting the
  the contents of the `terraform.tf` file.
* `cloudinit` - (Optional) The `libvirt_cloudinit` disk that has to be used by
  the domain. This is going to be attached as a CDROM ISO. Changing the
  cloud-init won't cause the domain to be recreated, however the change will
  have effect on the next reboot.

The optional `coreos_ignition` block is provided for CoreOS images.  This block contains
the following parameters:

* `content` - (Required) This can be set to the name of an existing ignition
file or alternatively can be set to the rendered value of a Terraform ignition provider object.
* `generated_ignition_file_dir` - (Optional) If `content` is set to the rendered value
of a Terraform ignition provider object you can specify where to locate the resulting
ignition file that will be loaded into the CoreOS vm.  If not set, "/tmp" is used.  These
generated ignition files will be removed when the domain is destroyed.

An example where a Terraform ignition provider object is used:
```
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

resource "libvirt_domain" "my_machine" {
  coreos_ignition {
      content = "${ignition_config.example.rendered}"
      generated_ignition_file_dir = "/my/ignition/file/directory" (Optional)
  }
  ...
}
```

Note that to make use of Ignition files with CoreOS the host must be running
QEMU v2.6 or greater.

Some extra arguments are also provided for using UEFI images:

* `firmware` - (Optional) The UEFI rom images for exercising UEFI secure boot in a qemu
environment. Users should usually specify one of the standard _Open Virtual Machine
Firmware_ (_OVMF_) images available for their distributions. The file will be opened
read-only.
* `nvram` - (Optional) the _nvram_ variables file corresponding to the firmware. When provided,
this file must be writable and specific to this domain, as it will be updated when running
the domain. However, `libvirt` can manage this automatically (and this is the recommended solution)
if a mapping for the firmware to a _variables file_ exists in `/etc/libvirt/qemu.conf:nvram`.
In that case, `libvirt` will copy that variables file into a file specific for this domain.

So you should typically use the firmware as this,

```

resource "libvirt_domain" "my_machine" {
  name = "my_machine"
  firmware = "/usr/share/qemu/ovmf-x86_64-code.bin"
  memory = "2048"

  disk {
    volume_id = "${libvirt_volume.volume.id}"
  }
  ...
}
```

and `/etc/libvirt/qemu.conf` should contain:

```
nvram = [
   "/usr/share/qemu/ovmf-x86_64-code.bin:/usr/share/qemu/ovmf-x86_64-vars.bin"
]
```

The `disk` block supports:

* `volume_id` - (Required) The volume id to use for this disk.

If you have a volume with a template image, create a second volume using the image as the backing volume, and then use the new volume as the volume for the disk. This way the image will not be modified.

```
resource "libvirt_volume" "leap" {
  name = "leap"
  source = "http://someurl/openSUSE_Leap-42.1.qcow2"
}

resource "libvirt_volume" "mydisk" {
  name = "mydisk"
  base_volume_id = "${libvirt_volume.leap.id}"
}

resource "libvirt_domain" "domain1" {
  name = "domain1"
  disk {
    volume_id = "${libvirt_volume.mydisk.id}"
  }

  network_interface {
    network_id = "${libvirt_network.net1.id}"
    hostname = "master"
    addresses = ["10.17.3.3"]
    mac = "AA:BB:CC:11:22:22"
    wait_for_lease = 1
  }
}
```

Also note that the `disk` block is actually a list of maps, so it is possible to declare several, use the literal list and map syntax or variables and interpolations, as in the following examples.

```
resource "libvirt_domain" "my_machine" {
  ...
  disk {
    volume_id = "${libvirt_volume.volume1.id}"
  }
  disk {
    volume_id = "${libvirt_volume.volume2.id}"
  }
}
```

```
resource "libvirt_domain" "my_machine" {
  ...
  disk = [
    {
      volume_id = "${libvirt_volume.volume1.id}"
    },
    {
      volume_id = "${libvirt_volume.volume2.id}"
    }
  ]
}
```

```
resource "libvirt_domain" "my_machine" {
  ...
  disk = ["${var.disk_map_list}"]
}
```


The `network_interface` specifies a network interface that can be connected either to
a virtual network (the preferred method on hosts with dynamic / wireless networking
configs) or directly to a LAN.

When using a virtual network, users can specify:

* `network_name` - (Optional) The name of an _existing_ network to attach this interface to.
The network will _NOT_ be managed by the Terraform/libvirt provider.
* `network_id` - (Optional) The ID of a network resource to attach this interface to.
The network will be under the control of the Terraform/libvirt provider.
* `mac` - (Optional) The specific MAC address to use for this interface.
* `addresses` - (Optional) An IP address for this domain in this network
* `hostname` - (Optional) A hostname that will be assigned to this domain resource in this network.
* `wait_for_lease`- (Optional) When creating the domain resource, wait until the network
interface gets a DHCP lease from libvirt, so that the computed ip addresses will be available
when the domain is up and the plan applied.

When connecting to a LAN, users can specify a target device with:

* `bridge` - Provides a bridge from the VM directly to the LAN. This assumes there is a bridge
device on the host which has one or more of the hosts physical NICs enslaved. The guest VM
will have an associated _tun_ device created and enslaved to the bridge. The IP range / network
configuration is whatever is used on the LAN. This provides the guest VM full incoming &
outgoing net access just like a physical machine.
* `vepa` - All VMs' packets are sent to the external bridge. Packets whose destination is a
VM on the same host as where the packet originates from are sent back to the host by the VEPA
capable bridge (today's bridges are typically not VEPA capable).
* `macvtap` - Packets whose destination is on the same host as where they originate from are
directly delivered to the target macvtap device. Both origin and destination devices need to
be in bridge mode for direct delivery. If either one of them is in vepa mode, a VEPA capable
bridge is required.
* `passthrough` - This feature attaches a virtual function of a SRIOV capable NIC directly to
a VM without losing the migration capability. All packets are sent to the VF/IF of the
configured network device. Depending on the capabilities of the device additional prerequisites
or limitations may apply; for example, on Linux this requires kernel 2.6.38 or newer.

Example of a `macvtap` interface:

```
resource "libvirt_domain" "my-domain" {
  name = "master"
  ...
  network_interface {
    macvtap = "eth0"
  }
}
```

**Warning:** the [Qemu guest agent](http://wiki.libvirt.org/page/Qemu_guest_agent)
must be installed and running inside of the domain in order to discover the IP
addresses of all the network interfaces attached to a LAN.

The optional `graphics` block allows you to override the default graphics settings.  The
block supports:

* `type` - the type of graphics emulation (default is "spice")
* `autoport` - defaults to "yes"
* `listen_type` - "listen type", defaults to "none"

On occasion we have found it necessary to set a `type` of "vnc" and a `listen_type` of "address"
with certain builds of QEMU.

The `graphics` block will look as follows:
```
resource "libvirt_domain" "my_machine" {
  ...
  graphics {
    type = "vnc"
    listen_type = "address"
  }
}
```

The optional `console` block allows you to define a console for the domain.  The block
looks as follows:
```
resource "libvirt_domain" "my_machine" {
  ...
  console {
    type = "pty"
    target_port = "0"
    target_type = <"serial" or "virtio">
    source_path = "/dev/pts/4"
  }
}
```

Note the following:
* You can repeat the `console` block to create more than one console, in the same way
that you can repeat `disk` blocks (see above)
* The `target_type` is optional for the first console
* All subsequent `console` blocks must specify a `target_type` of `virtio`
* The `source_path` is optional for all consoles

See [libvirt Domain XML Console element](https://libvirt.org/formatdomain.html#elementsConsole)
for more information.

## Attributes Reference

* `id` - a unique identifier for the resource
* `network_interface.<N>.addresses.<M>` - M-th IP address assigned to the N-th network interface
