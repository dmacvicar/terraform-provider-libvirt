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

```hcl
resource "libvirt_domain" "default" {
  name = "test"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the resource, required by libvirt.
  Changing this forces a new resource to be created.
* `cpu` - (Optional) Configures CPU mode. See [below](#cpu-mode) for more
  details.
* `vcpu` - (Optional) The amount of virtual CPUs. If not specified, a single CPU
  will be created.
* `memory` - (Optional) The amount of memory in MiB. If not specified the domain
  will be created with 512 MiB of memory be used.
* `running` - (Optional) Use `false` to turn off the instance. If not specified,
  true is assumed and the instance, if stopped, will be started at next apply.
* `disk` - (Optional) An array of one or more disks to attach to the domain. The
  `disk` object structure is documented [below](#handling-disks).
* `network_interface` - (Optional) An array of one or more network interfaces to
  attach to the domain. The `network_interface` object structure is documented
  [below](#handling-network-interfaces).
* `cloudinit` - (Optional) The `libvirt_cloudinit_disk` disk that has to be used by
  the domain. This is going to be attached as a CDROM ISO. Changing the
  cloud-init won't cause the domain to be recreated, however the change will
  have effect on the next reboot.
* `autostart` - (Optional) Set to `true` to start the domain on host boot up.
  If not specified `false` is assumed.
* `filesystem` - (Optional) An array of one or more host filesystems to attach to
  the domain. The `filesystem` object structure is documented
  [below](#sharing-filesystem-between-libvirt-host-and-guest).
* `coreos_ignition` - (Optional) The
  [libvirt_ignition](/docs/providers/libvirt/r/coreos_ignition.html) resource
  that is to be used by the CoreOS domain.
* `arch` - (Optional) The architecture for the VM (probably x86_64 or i686),
  you normally won't need to set this unless you are building a special VM
* `machine` - (Optional) The machine type,
  you normally won't need to set this unless you are running on a platform that
  defaults to the wrong machine type for your template
* `boot_device` - (Optional) A list of devices (dev) which defines boot order. Example
   [below](#define-boot-device-order).
* `emulator` - (Optional) The path of the emulator to use
* `qemu_agent` (Optional) By default is disabled, set to true for enabling it. More info [qemu-agent](https://wiki.libvirt.org/page/Qemu_guest_agent).
### Kernel and boot arguments

* `kernel` - (Optional) The path of the kernel to boot

If you are using a qcow2 volume, you can pass the id of the volume (eg. `${libvirt_volume.kernel.id}`)
as they are local to the hypervisor.

Given that you can define a volume from a remote http file, this means, you can also have remote kernels.

```hcl
resource "libvirt_volume" "kernel" {
  source = "http://download.opensuse.org/tumbleweed/repo/oss/boot/x86_64/loader/linux"
  name = "kernel"
  pool = "default"
  format = "raw"
}

resource "libvirt_domain" "domain-suse" {
  name = "suse"
  memory = "1024"
  vcpu = 1

  kernel = "${libvirt_volume.kernel.id}"

  // ...
}
```

* `initrd` - (Optional) The path of the initrd to boot.

You can use it in the same way as the kernel.

* `cmdline` - (Optional) Arguments to the kernel

```hcl
resource "libvirt_domain" "domain-suse" {
  name = "suse"
  memory = "1024"
  vcpu = 1

  kernel = "${libvirt_volume.kernel.id}"

  cmdline {
    arg1 = "value1"
    arg2 = "value2"
    "_" = "rw nosplash"
  }
}
```

Kernel params that don't have a keyword identifier can be specified using the
special `"_"` keyword. Multiple keyword-less params have to be specified using
the same `"_"` keyword, like in the example above.

Also note that the `cmd` block is actually a list of maps, so it is possible to
declare several of them by using either the literal list and map syntax as in
the following examples:

```hcl
resource "libvirt_domain" "my_machine" {
  //...

  cmdline {
    arg1 = "value1"
  }
  cmdline {
    arg2 = "value2"
  }
}
```

```hcl
resource "libvirt_domain" "my_machine" {
  ...
  cmdline = [
    {
      arg1 = "value1"
    },
    {
      arg2 = "value2"
    }
  ]
}
```
The kernel supports passing the same option multiple times. If you need this, use separate cmdline blocks.

```hcl
resource "libvirt_domain" "my_machine" {
  //...

  cmdline {
    arg1 = "value1"
  }
  cmdline {
    arg1 = "value2"
  }
}
```

~> **Note well:** `kernel` and `initrd` have to be specified at the same time; `cmdline`
   arguments can be specified only when `kernel` and `initrd` have been defined.
   Otherwise libvirt will refuse to start the domain.

### UEFI images

Some extra arguments are also provided for using UEFI images:

* `firmware` - (Optional) The UEFI rom images for exercising UEFI secure boot in a qemu
environment. Users should usually specify one of the standard _Open Virtual Machine
Firmware_ (_OVMF_) images available for their distributions. The file will be opened
read-only.
* `nvram` - (Optional) this block allows specifying the following attributes related to the _nvram_:
  * `file` - path to the file backing the NVRAM store for non-volatile variables. When provided,
  this file must be writable and specific to this domain, as it will be updated when running the
  domain. However, `libvirt` can  manage this automatically (and this is the recommended solution)
  if a mapping for the firmware to a _variables file_ exists in `/etc/libvirt/qemu.conf:nvram`.
  In that case, `libvirt` will copy that variables file into a file specific for this domain.
  * `template` - (Optional) path to the file used to override variables from the master NVRAM
  store.

So you should typically use the firmware as this,

```hcl
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

```hcl
nvram = [
   "/usr/share/qemu/ovmf-x86_64-code.bin:/usr/share/qemu/ovmf-x86_64-vars.bin"
]
```

In case you need (or want) to specify the path for the NVRAM store, the domain definition should
look like this:

```hcl
resource "libvirt_domain" "my_machine" {
  name = "my_machine"
  firmware = "/usr/share/qemu/ovmf-x86_64-code.bin"
  nvram {
    file = "/usr/local/share/qemu/custom-vars.bin"
  }
  memory = "2048"

  disk {
    volume_id = "${libvirt_volume.volume.id}"
  }
  ...
}

```

Finally, if you want the initial values for the NVRAM to be overridden by custom initial values
coming from a template, the domain definition should look like this:

```hcl
resource "libvirt_domain" "my_machine" {
  name = "my_machine"
  firmware = "/usr/share/qemu/ovmf-x86_64-code.bin"
  nvram {
    file = "/usr/local/share/qemu/custom-vars.bin"
    template = "/usr/local/share/qemu/template-vars.bin"
  }
  memory = "2048"

  disk {
    volume_id = "${libvirt_volume.volume.id}"
  }
  ...
}
```

### Handling disks

The `disk` block supports:

* `volume_id` - (Optional) The volume id to use for this disk.
* `url` - (Optional) The http url to use as the block device for this disk (read-only)
* `file` - (Optional) The filename to use as the block device for this disk (read-only)

While `volume_id`, `url` and `file` are optional, it is intended that you use one of them.

* `scsi` - (Optional, Boolean) Use a scsi controller for this disk.  The controller
model is set to `virtio-scsi`
* `wwn` - (Optional) Specify a WWN to use for the disk if the disk is using
a scsi controller, if not specified then a random wwn is generated for the disk


```hcl
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
    scsi = "true"
  }

  disk {
    url = "http://foo.com/install.iso"
  }

  disk {
    file = "/absolute/path/to/disk.iso"
  }
}
```

Also note that the `disk` block is actually a list of maps, so it is possible to
declare several of them by using either the literal list and map syntax as in
the following examples:

```hcl
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

```hcl
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

```hcl
resource "libvirt_domain" "my_machine" {
  ...
  disk = ["${var.disk_map_list}"]
}
```

### Handling network interfaces

The `network_interface` specifies a network interface that can be connected
either to a virtual network (the preferred method on hosts with
dynamic / wireless networking configs) or directly to a LAN.

```hcl
resource "libvirt_domain" "domain1" {
  name = "domain1"

  network_interface {
    network_id = "${libvirt_network.net1.id}"
    hostname = "master"
    addresses = ["10.17.3.3"]
    mac = "AA:BB:CC:11:22:22"
    wait_for_lease = true
  }
}
```

When using a virtual network, users can specify:

* `network_name` - (Optional) The name of an _existing_ network to attach this
  interface to. The network will _NOT_ be managed by the Terraform libvirt
  provider.
* `network_id` - (Optional) The ID of a network resource to attach this
  interface to. This is a
  [network resource](/website/docs/r/network.markdown) managed by the
  Terraform libvirt provider.
* `mac` - (Optional) The specific MAC address to use for this interface.
* `model` - Set the model of emulated network interface card, defaults to `virtio`.
* `addresses` - (Optional) An IP address for this domain in this network.
* `hostname` - (Optional) A hostname that will be assigned to this domain
  resource in this network.
* `wait_for_lease`- (Optional boolean) When creating the domain resource, wait until the
  network interface gets a DHCP lease from libvirt, so that the computed IP
  addresses will be available when the domain is up and the plan applied.

When connecting to a LAN, users can specify a target device with:

* `bridge` - Provides a bridge from the VM directly to the LAN. This assumes
  there is a bridge device on the host which has one or more of the hosts
  physical NICs enslaved. The guest VM will have an associated _tun_ device
  created and enslaved to the bridge. The IP range / network configuration is
  whatever is used on the LAN. This provides the guest VM full incoming &
  outgoing net access just like a physical machine.
* `vepa` - All VMs' packets are sent to the external bridge. Packets whose
  destination is a VM on the same host as where the packet originates from are
  sent back to the host by the VEPA capable bridge (today's bridges are
  typically not VEPA capable).
* `macvtap` - Packets whose destination is on the same host as where they
  originate from are directly delivered to the target macvtap device. Both
  origin and destination devices need to be in bridge mode for direct delivery.
  If either one of them is in vepa mode, a VEPA capable bridge is required.
* `passthrough` - This feature attaches a virtual function of a SRIOV capable
  NIC directly to a VM without losing the migration capability. All packets are
  sent to the VF/IF of the configured network device. Depending on the
  capabilities of the device additional prerequisites or limitations may apply;
  for example, on Linux this requires kernel 2.6.38 or newer.

Example of a `macvtap` interface:

```hcl
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

### Graphics devices

The optional `graphics` block allows you to override the default graphics
settings.

The block supports:

* `type` - the type of graphics emulation (default is "spice")
* `autoport` - defaults to "yes"
* `listen_type` - "listen type", defaults to "none"

On occasion we have found it necessary to set a `type` of `vnc` and a
`listen_type` of `address` with certain builds of QEMU.

The `graphics` block will look as follows:

```hcl
resource "libvirt_domain" "my_machine" {
  ...
  graphics {
    type = "vnc"
    listen_type = "address"
  }
}
```

~> **Note well:** the `graphics` block is ignored for the architectures
  `s390x` and `ppc64`.


### Console devices

The optional `console` block allows you to define a console for the domain.

The block looks as follows:

```hcl
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

~> **Note well:**
  <ul>
    <li>You can repeat the `console` block to create more than one console, in the
        same way that you can repeat `disk` blocks (see [above](#handling-disks)).</li>
    <li>The `target_type` is optional for the first console and defaults to `serial`.</li>
    <li>All subsequent `console` blocks must specify a `target_type` of `virtio`.</li>
    <li>The `source_path` is optional for all consoles.</li>
  </ul>

See [libvirt Domain XML Console element](https://libvirt.org/formatdomain.html#elementsConsole)
for more information.


### CPU mode

The optional `cpu` block allows to configure CPU mode. Example:

```hcl
resource "libvirt_domain" "my_machine" {
  ...
  cpu {
    mode = "host-passthrough"
  }
}
```

To start the domain on host boot up set `autostart` to `true` like so:
```
resource "libvirt_domain" "my_machine" {
  ...
  autostart = true
}
```

### Sharing filesystem between libvirt host and guest

The optional `filesystem` block allows to define one or more [filesytem](https://libvirt.org/formatdomain.html#elementsFilesystems)
entries to be added to the domain. This allows to share a directory of the libvirtd
host with the guest.

Currently the following attributes are supported:

  * `accessmode`: specifies the security mode for accessing the source. By default
    the `mapped` mode is chosen.
  * `source`: the directory of the host to be shared with the guest.
  * `target`: an arbitrary string tag that is exported to the guest as a hint for
     where to mount the source.
  * `readonly`: enables exporting filesystem as a readonly mount for guest, by
    default read-only access is given.

Example:

```hcl
filesystem {
  source = "/tmp"
  target = "tmp"
  readonly = false
}
filesystem {
  source = "/proc"
  target = "proc"
  readonly = true
}
```

The exported filesystems can be mounted inside of the guest in this way:

```hcl
sudo mount -t 9p -o trans=virtio,version=9p2000.L,rw tmp /host/tmp
sudo mount -t 9p -o trans=virtio,version=9p2000.L,r proc /host/proc
```

This can be automated inside of `/etc/fstab`:
```hcl
tmp /host/tmp 9p  trans=virtio,version=9p2000.L,rw  0 0
proc /host/proc  9p  trans=virtio,version=9p2000.L,r  0 0
```

### Define Boot Device Order

Set hd as default and fallback to network.

```hcl
boot_device {
  dev = [ "hd", "network"]
}
```

### Altering libvirt's generated domain XML definition

The optional `xml` block relates to the generated domain XML.

Currently the following attributes are supported:

* `xslt`: specifies a XSLT stylesheet to transform the generated XML definition before creating the domain.
  This is used to support features the provider does not allow to set from the schema.
  It is not recommended to alter properties and settings that are exposed to the schema, as terraform will insist in changing them back to the known state.

See https://github.com/dmacvicar/terraform-provider-libvirt/blob/master/examples/xslt/main.tf and https://github.com/dmacvicar/terraform-provider-libvirt/blob/master/examples/xslt/nicmodel.xsl for a working example that changes the NIC model.

## Attributes Reference

* `id` - a unique identifier for the resource.
* `network_interface.<N>.addresses.<M>` - M-th IP address assigned to the N-th
  network interface.
