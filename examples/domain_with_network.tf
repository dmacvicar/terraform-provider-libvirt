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
  unit   = "MiB"
  vcpu   = 2
  type   = "kvm"

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }

  # Network interface connected to default network
  interface {
    type  = "network"
    model = "virtio"
    source {
      network = "default"
    }
  }

  # Additional network interface on a specific network
  interface {
    type  = "network"
    model = "virtio"
    source {
      network   = "custom-network"
      portgroup = "web-servers"
    }
  }

  # Bridge interface
  interface {
    type = "bridge"
    source {
      bridge = "br0"
    }
  }
}
