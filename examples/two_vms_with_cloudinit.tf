# Example: Two Virtual Machines with Cloud-Init Configuration
#
# This example demonstrates provisioning two Ubuntu VMs with:
# - Automatic IP assignment via DHCP on the default network
# - Root password configured via cloud-init
# - SSH server enabled and accessible
# - Using Ubuntu 22.04 LTS cloud image

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

# Download Ubuntu 22.04 (Jammy) cloud image
# Ubuntu cloud images have excellent cloud-init support
resource "libvirt_volume" "ubuntu_base" {
  name   = "ubuntu-jammy-base.qcow2"
  pool   = "default"
  format = "qcow2"

  create = {
    content = {
      # Ubuntu 22.04 LTS (Jammy Jellyfish) cloud image
      url = "https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img"
    }
  }
}

# Create boot disk for VM1 (uses base image as backing store)
resource "libvirt_volume" "vm1_disk" {
  name   = "vm1-disk.qcow2"
  pool   = "default"
  format = "qcow2"

  # Start with 2GB, will grow as needed
  capacity = 2147483648 # 2GB in bytes

  backing_store = {
    path   = libvirt_volume.ubuntu_base.path
    format = "qcow2"
  }
}

# Download Alpine Linux 3.22 cloud image
resource "libvirt_volume" "alpine_base" {
  name   = "alpine-3.22-base.qcow2"
  pool   = "default"
  format = "qcow2"

  create = {
    content = {
      # Alpine Linux 3.22 generic cloud image (BIOS, cloudinit)
      url = "https://dl-cdn.alpinelinux.org/alpine/v3.22/releases/cloud/generic_alpine-3.22.2-x86_64-bios-cloudinit-r0.qcow2"
    }
  }
}

# Create boot disk for VM2 (uses Alpine base image as backing store)
resource "libvirt_volume" "vm2_disk" {
  name   = "vm2-disk.qcow2"
  pool   = "default"
  format = "qcow2"

  capacity = 2147483648 # 2GB in bytes

  backing_store = {
    path   = libvirt_volume.alpine_base.path
    format = "qcow2"
  }
}

# Cloud-init configuration for VM1
resource "libvirt_cloudinit_disk" "vm1_init" {
  name = "vm1-cloudinit"

  # User-data: Configure root password, enable SSH, install packages
  user_data = <<-EOF
    #cloud-config
    # Set root password to "password" (change this!)
    chpasswd:
      list: |
        root:password
      expire: false

    # Enable SSH password authentication
    ssh_pwauth: true

    # Install and enable SSH server
    packages:
      - openssh-server

    # Optional: Add SSH public key for key-based auth (more secure)
    # ssh_authorized_keys:
    #   - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC0/Ho... your-key-here

    # Set timezone
    timezone: UTC

    # Final message on console
    final_message: "VM1 is ready! SSH: ssh root@<IP>"
  EOF

  # Meta-data: Instance identification
  meta_data = <<-EOF
    instance-id: vm1-001
    local-hostname: ubuntu-vm1
  EOF

  # Network config: Use DHCP (default behavior)
  network_config = <<-EOF
    version: 2
    ethernets:
      eth0:
        dhcp4: true
  EOF
}

# Cloud-init configuration for VM2
resource "libvirt_cloudinit_disk" "vm2_init" {
  name = "vm2-cloudinit"

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

    final_message: "VM2 is ready! SSH: ssh root@<IP>"
  EOF

  meta_data = <<-EOF
    instance-id: vm2-001
    local-hostname: ubuntu-vm2
  EOF

  network_config = <<-EOF
    version: 2
    ethernets:
      eth0:
        dhcp4: true
  EOF
}

# Upload cloud-init ISO for VM1 to a volume
resource "libvirt_volume" "vm1_cloudinit" {
  name = "vm1-cloudinit.iso"
  pool = "default"
  # Format will be auto-detected as "iso"

  create = {
    content = {
      url = libvirt_cloudinit_disk.vm1_init.path
    }
  }
}

# Upload cloud-init ISO for VM2 to a volume
resource "libvirt_volume" "vm2_cloudinit" {
  name = "vm2-cloudinit.iso"
  pool = "default"
  # Format will be auto-detected as "iso"

  create = {
    content = {
      url = libvirt_cloudinit_disk.vm2_init.path
    }
  }
}

# Virtual Machine 1
resource "libvirt_domain" "vm1" {
  name   = "ubuntu-vm1"
  memory = 1048576 # 1 GB in KiB (1024 * 1024)
  vcpu   = 1

  # Boot configuration
  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  # Attached disks
  devices = {
    disks = [
      # Main system disk
      {
        source = {
          pool   = libvirt_volume.vm1_disk.pool
          volume = libvirt_volume.vm1_disk.name
        }
        target = "vda"
        bus    = "virtio"
      },
      # Cloud-init config disk (will be detected automatically)
      {
        device = "cdrom"
        source = {
          pool   = libvirt_volume.vm1_cloudinit.pool
          volume = libvirt_volume.vm1_cloudinit.name
        }
        target = "sda"
        bus    = "sata"
      }
    ]

    # Network interface on default network (DHCP)
    interfaces = [
      {
        type  = "network"
        model = "virtio"
        source = {
          network = "default"
        }
      }
    ]

    # Graphics console (VNC)
    graphics = {
      vnc = {
        autoport = "yes"
        listen   = "127.0.0.1"
      }
    }
  }

  # Start the VM automatically
  running = true
}

# Virtual Machine 2
resource "libvirt_domain" "vm2" {
  name   = "ubuntu-vm2"
  memory = 1048576 # 1 GB in KiB (1024 * 1024)
  vcpu   = 1

  os = {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    disks = [
      {
        source = {
          pool   = libvirt_volume.vm2_disk.pool
          volume = libvirt_volume.vm2_disk.name
        }
        target = "vda"
        bus    = "virtio"
      },
      {
        device = "cdrom"
        source = {
          pool   = libvirt_volume.vm2_cloudinit.pool
          volume = libvirt_volume.vm2_cloudinit.name
        }
        target = "sda"
        bus    = "sata"
      }
    ]

    interfaces = [
      {
        type  = "network"
        model = "virtio"
        source = {
          network = "default"
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

# Outputs to help locate the VMs
output "vm1_id" {
  value       = libvirt_domain.vm1.id
  description = "Domain ID for VM1"
}

output "vm2_id" {
  value       = libvirt_domain.vm2.id
  description = "Domain ID for VM2"
}

output "instructions" {
  value = <<-EOF

    Virtual machines have been created!

    To find IP addresses assigned by DHCP:
      sudo virsh domifaddr ubuntu-vm1
      sudo virsh domifaddr ubuntu-vm2

    Or check the DHCP leases:
      sudo virsh net-dhcp-leases default

    To connect via SSH (once you know the IP):
      ssh root@<IP-ADDRESS>
      Password: password

    To view VM console:
      sudo virsh console ubuntu-vm1
      sudo virsh console ubuntu-vm2

    To connect via VNC:
      1. Find VNC port: sudo virsh domdisplay ubuntu-vm1
      2. Connect with VNC client to that address

    Note: It may take 30-60 seconds after boot for cloud-init to complete
          and the SSH server to be available.
  EOF
}
