# Example: Single Ubuntu 24.04 LTS VM on Apple Silicon (ARM64)
#
# This example demonstrates provisioning an Ubuntu 24.04 VM with:
# - ARM64 Architecture support optimized for Apple Silicon (hvf/aarch64)
# - UEFI boot configuration using QEMU edk2 firmware
# - Automatic user configuration (ubuntu:password) via cloud-init
# - Copy-on-Write (CoW) storage using a base cloud image
# - Post-provisioning SSH port forwarding (Host:2222 -> Guest:22)

terraform {
  required_providers {
    libvirt = {
      source = "dmacvicar/libvirt"
    }
  }
}

provider "libvirt" {
  uri = "qemu:///session?socket=${pathexpand("~/.cache/libvirt/libvirt-sock")}"
}

# Download Ubuntu 24.04 (Noble) ARM64 cloud image
# Ubuntu cloud images provide native cloud-init support for automated setup
resource "libvirt_volume" "ubuntu_base" {
  name = "ubuntu-24.04-base.qcow2"
  pool = "default"

  target = {
    format = {
      type = "qcow2"
    }
  }

  create = {
    content = {
      url = "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-arm64.img"
    }
  }
}

# Create boot disk for the VM (uses base image as backing store)
# This implements a Copy-on-Write (CoW) layer to save space and time
resource "libvirt_volume" "ubuntu_disk" {
  name     = "ubuntu-vm.qcow2"
  pool     = "default"
  capacity = 10737418240

  target = {
    format = {
      type = "qcow2"
    }
  }

  backing_store = {
    path = libvirt_volume.ubuntu_base.path
    format = {
      type = "qcow2"
    }
  }
}

# Cloud-init configuration for the Ubuntu VM
# Defines user credentials, SSH server availability, and timezone
resource "libvirt_cloudinit_disk" "ubuntu_seed" {
  name = "ubuntu-cloudinit"

  user_data = <<-EOF
    #cloud-config
    users:
      - name: ubuntu
        sudo: ALL=(ALL) NOPASSWD:ALL
        shell: /bin/bash
        lock_passwd: false
        ssh_authorized_keys:
          - ssh-ed25519 ... Add your public ssh key here ...
    chpasswd:
      list: |
        ubuntu:password
      expire: false
    ssh_pwauth: true
    packages:
      - openssh-server
    timezone: UTC
  EOF

  meta_data = <<-EOF
    instance-id: ubuntu-001
    local-hostname: ubuntu-vm
  EOF

  # Network config: Specific interface naming (enp1s0) for Virtio on KVM/Ubuntu
  network_config = <<-EOF
    version: 2
    ethernets:
      enp1s0:
        dhcp4: true
  EOF
}

# Upload cloud-init ISO into the storage pool
# This volume acts as the seed drive for the cloud-init process
resource "libvirt_volume" "ubuntu_seed_volume" {
  name = "ubuntu-cloudinit.iso"
  pool = "default"

  create = {
    content = {
      url = libvirt_cloudinit_disk.ubuntu_seed.path
    }
  }
}

# Virtual Machine Definition (Apple Silicon / ARM64 optimized)
# Configures CPU passthrough, UEFI firmware paths, and virtual hardware
resource "libvirt_domain" "ubuntu_vm" {
  name        = "ubuntu-vm"
  memory      = 4096
  memory_unit = "MiB"
  vcpu        = 2
  type        = "hvf" # Hardware acceleration for macOS
  running     = true

  # Boot and Architecture configuration
  os = {
    type         = "hvm"
    type_arch    = "aarch64"
    type_machine = "virt"

    # UEFI Loader is required for ARM64 boot on Apple Silicon
    loader      = "/opt/homebrew/Cellar/qemu/10.2.1/share/qemu/edk2-aarch64-code.fd"
    loader_type = "pflash"

    # NVRAM setup to persist EFI variables
    nv_ram = {
      nv_ram   = "${abspath(path.module)}/ubuntu-vm_VARS.fd"
      template = "/opt/homebrew/Cellar/qemu/10.2.1/share/qemu/edk2-arm-vars.fd"
    }

    features = {
      acpi = true
    }

    cpu = {
      mode = "host-passthrough"
    }
  }

  # Attached devices: Disks, Network, and Consoles
  devices = {
    disks = [
      # Main System Disk
      {
        source = {
          volume = {
            pool   = libvirt_volume.ubuntu_disk.pool
            volume = libvirt_volume.ubuntu_disk.name
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
      # Cloud-init Seed Disk
      {
        source = {
          volume = {
            pool   = libvirt_volume.ubuntu_seed_volume.pool
            volume = libvirt_volume.ubuntu_seed_volume.name
          }
        }
        target = {
          dev = "vdb"
          bus = "virtio"
        }
      }
    ]

    # User-mode networking (SLIRP)
    interfaces = [
      {
        type = "user"
        model = {
          type = "virtio"
        }
      }
    ]

    # VNC Graphics Configuration
    graphics = [
      {
        vnc = {
          auto_port = true
          listen    = "127.0.0.1"
        }
      }
    ]

    video = [
      {
        model = {
          type = "virtio"
        }
      }
    ]

    # Serial and Console access for debugging
    serials = [
      {
        type = "pty"
        target = {
          type  = "system-serial"
          port  = "0"
          model = {
            name = "pl011"
          }
        }
      }
    ]

    consoles = [
      {
        type = "pty"
        target = {
          type = "serial"
          port = "0"
        }
      }
    ]

    inputs = [
      { type = "tablet", bus = "virtio" },
      { type = "keyboard", bus = "virtio" }
    ]
  }
}

# Post-provisioning: Automatic SSH Port Forwarding
# Maps host port 2222 to guest port 22 via QEMU monitor command (HMP)
resource "null_resource" "auto_ssh_port" {
  # Ensure the VM is created before attempting port mapping
  depends_on = [libvirt_domain.ubuntu_vm]

  # Force execution on every 'apply' to maintain the mapping
  triggers = {
    always_run = timestamp()
  }

  provisioner "local-exec" {
    command = <<-EOT
      echo "Waiting for VM to initialize..."
      sleep 10
      virsh -c qemu:///session qemu-monitor-command ubuntu-vm --hmp "hostfwd_add hostnet0 tcp::2222-:22"
      echo "SSH access available at: ssh ubuntu@localhost -p 2222"
    EOT
  }
}
