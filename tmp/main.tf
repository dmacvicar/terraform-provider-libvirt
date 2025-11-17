terraform {
  required_providers {
    libvirt = {
      source = "dmacvicar/libvirt"
    }
  }
}

provider "libvirt" {
  uri = "qemu:///system"
}

# Base Alpine Linux cloud image stored in the default pool.
resource "libvirt_volume" "alpine_base" {
  name   = "alpine-3.22-base.qcow2"
  pool   = "default"
  format = "qcow2"

  create = {
    content = {
      url = "https://dl-cdn.alpinelinux.org/alpine/v3.22/releases/cloud/generic_alpine-3.22.2-x86_64-bios-cloudinit-r0.qcow2"
    }
  }
}

# Writable copy-on-write layer for the VM.
resource "libvirt_volume" "alpine_disk" {
  name     = "alpine-vm.qcow2"
  pool     = "default"
  format   = "qcow2"
  capacity = 2147483648

  backing_store = {
    path   = libvirt_volume.alpine_base.path
    format = "qcow2"
  }
}

# Cloud-init seed ISO.
resource "libvirt_cloudinit_disk" "alpine_seed" {
  name = "alpine-cloudinit"

  user_data = <<-EOF
    #cloud-config
    chpasswd:
      list: |
        root:password
      expire: false

    ssh_pwauth: true

    packages:
      - openssh-server
    timezone: UTC
  EOF

  meta_data = <<-EOF
    instance-id: alpine-001
    local-hostname: alpine-vm
  EOF

  network_config = <<-EOF
    version: 2
    ethernets:
      eth0:
        dhcp4: true
  EOF
}

# Upload the cloud-init ISO into the pool.
resource "libvirt_volume" "alpine_seed_volume" {
  name = "alpine-cloudinit.iso"
  pool = "default"

  create = {
    content = {
      url = libvirt_cloudinit_disk.alpine_seed.path
    }
  }
}

# Virtual machine definition.
resource "libvirt_domain" "alpine" {
  name   = "alpine-vm"
  memory = 1048576
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }

  features = {
    acpi = true
  }

  devices = {
    disks = [
      {
        source = {
          pool   = libvirt_volume.alpine_disk.pool
          volume = libvirt_volume.alpine_disk.name
        }
        target = {
          dev = "vda"
          bus = "virtio"
        }
      },
      {
        device = "cdrom"
        source = {
          pool   = libvirt_volume.alpine_seed_volume.pool
          volume = libvirt_volume.alpine_seed_volume.name
        }
        target = {
          dev = "sdb"
          bus = "sata"
        }
      }
    ]

    interfaces = [
      {
        type  = "network"
        model = "virtio"  # e1000 is more compatible than virtio for Alpine
        source = {
          network = "default"
        }
        # TODO: wait_for_ip not implemented yet (Phase 2)
        # This will wait during creation until the interface gets an IP
        wait_for_ip = {
          timeout = 300    # seconds, default 300
          source  = "any"  # "lease" (DHCP), "agent" (qemu-guest-agent), or "any" (try both)
        }
      }
    ]

    graphics = {
      vnc = {
        autoport = "yes"
        listen   = "127.0.0.1"
      }
    }
  }

  running = true
}

# Query the domain's interface addresses
# This data source can be used at any time to retrieve current IP addresses
# without blocking operations like Delete
data "libvirt_domain_interface_addresses" "alpine" {
  domain = libvirt_domain.alpine.name
  source = "lease" # optional: "lease" (DHCP), "agent" (qemu-guest-agent), or "any"
}

# Output all interface information
output "vm_interfaces" {
  description = "All network interfaces with their IP addresses"
  value       = data.libvirt_domain_interface_addresses.alpine.interfaces
}

# Output the first IP address found
output "vm_ip" {
  description = "First IP address of the VM"
  value = length(data.libvirt_domain_interface_addresses.alpine.interfaces) > 0 && length(data.libvirt_domain_interface_addresses.alpine.interfaces[0].addrs) > 0 ? data.libvirt_domain_interface_addresses.alpine.interfaces[0].addrs[0].addr : "No IP address found"
}

# Output all IP addresses across all interfaces
output "vm_all_ips" {
  description = "All IP addresses across all interfaces"
  value = flatten([
    for iface in data.libvirt_domain_interface_addresses.alpine.interfaces : [
      for addr in iface.addrs : addr.addr
    ]
  ])
}
