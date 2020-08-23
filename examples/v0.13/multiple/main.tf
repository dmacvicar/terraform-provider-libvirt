terraform {
 required_version = ">= 0.13"
  required_providers {
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "0.6.2"
    }
  }
}

provider "libvirt" {
  uri = "qemu:///system"
}

provider "libvirt" {
  alias = "remotehost"
  uri   = "qemu+ssh://root@remotehost-p1.qa.suse.de/system"
}

resource "libvirt_volume" "local-qcow2" {
  name   = "local-qcow2"
  pool   = "default"
  format = "qcow2"
  size   = 100000
}

resource "libvirt_volume" "remotehost-qcow2" {
  provider = libvirt.remotehost
  name     = "remotehost-qcow2"
  pool     = "default"
  format   = "qcow2"
  size     = 100000
}

resource "libvirt_domain" "local-domain" {
  name   = "local"
  memory = "2048"
  vcpu   = 2

  disk {
    volume_id = libvirt_volume.local-qcow2.id
  }
}

resource "libvirt_domain" "remotehost-domain" {
  provider = libvirt.remotehost
  name     = "remotehost"
  memory   = "2048"
  vcpu     = 2

  disk {
    volume_id = libvirt_volume.remotehost-qcow2.id
  }
}
