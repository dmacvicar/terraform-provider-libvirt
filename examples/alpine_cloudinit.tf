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
  target = {
    format = {
      type = "qcow2"
    }
  }
  create = {
    content = {
      url = "https://dl-cdn.alpinelinux.org/alpine/v3.22/releases/cloud/generic_alpine-3.22.2-x86_64-bios-cloudinit-r0.qcow2"
    }
  }
}

# Writable copy-on-write layer for the VM.
resource "libvirt_volume" "alpine_disk" {
  name   = "alpine-vm.qcow2"
  pool   = "default"
  target = {
    format = {
      type = "qcow2"
    }
  }
  capacity = 2147483648

  backing_store = {
    path   = libvirt_volume.alpine_base.path
    format = {
      type = "qcow2"
    }
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
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    disks = [
      {
        source = {
          volume = {
            pool   = libvirt_volume.alpine_disk.pool
            volume = libvirt_volume.alpine_disk.name
          }
        }
        target = {
          dev = "vda"
          bus = "virtio"
        }
        driver = {
          type = "qcow2"
        }
      },
      {
        device = "cdrom"
        source = {
          volume = {
            pool   = libvirt_volume.alpine_seed_volume.pool
            volume = libvirt_volume.alpine_seed_volume.name
          }
        }
        target = {
          dev = "sda"
          bus = "sata"
        }
      }
    ]

    interfaces = [
      {
        type  = "network"
        model = { type = "virtio" }
        source = {
          network = {
            network = "default"
          }
        }
      }
    ]

    graphics = [
      {
        type = "vnc"
        vnc = {
          autoport = true
          listen   = "127.0.0.1"
        }
      }
    ]
  }

  running = true
}
