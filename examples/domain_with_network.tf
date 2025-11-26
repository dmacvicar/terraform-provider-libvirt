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

resource "libvirt_domain" "example" {
  name   = "example-vm"
  memory = 2048
  memory_unit   = "MiB"
  vcpu   = 2
  type   = "kvm"

  os = {
    type         = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    interfaces = [
      # Network interface connected to default network
      {
        model = {
          type = "virtio"
        }
        source = {
          network = {
            network = "default"
          }
        }
      },
      # Additional network interface on a specific network
      {
        model = {
          type = "virtio"
        }
        source = {
          network = {
            network    = "custom-network"
            port_group = "web-servers"
          }
        }
      },
      # Bridge interface
      {
        source = {
          bridge = {
            bridge = "br0"
          }
        }
      }
    ]
  }
}
