provider "libvirt" {
  uri = "qemu:///system"
}

// blank 10GB image for net install.
resource "libvirt_volume" "debian9-qcow2" {
  name   = "debian9-qcow2"
  pool   = "default"
  format = "qcow2"
  size   = 10000000000
}

// set boot order hd, network
resource "libvirt_domain" "domain-debian9-qcow2" {
  name   = "debian9"
  memory = "1024"
  vcpu   = 1

  network_interface {
    bridge = "br0"
    mac    = "52:54:00:b2:2f:86"
  }

  boot_device {
    dev = ["hd", "network"]
  }

  disk {
    volume_id = libvirt_volume.debian9-qcow2.id
  }

  graphics {
    type        = "vnc"
    listen_type = "address"
  }
}

terraform {
  required_version = ">= 0.12"
}
