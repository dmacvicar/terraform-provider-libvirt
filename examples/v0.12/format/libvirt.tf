provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_network" "tf" {
  name      = "tf"
  domain    = "tf.local"
  mode      = "nat"
  addresses = ["10.0.100.0/24"]
}

# raw image from file
resource "libvirt_volume" "debian8-raw" {
  name   = "debian8-raw"
  format = "raw"
  source = "http://localhost:8000/debian8.img"
}

# qcow2 image from file
resource "libvirt_volume" "debian8-qcow2" {
  name   = "debian8-qcow2"
  source = "http://localhost:8000/debian8.qcow2"
}

# volume with raw backing storage
resource "libvirt_volume" "vol-debian8-raw" {
  name           = "vol-debian8-raw"
  base_volume_id = libvirt_volume.debian8-raw.id
}

# volume with qcow2 backing storage
resource "libvirt_volume" "vol-debian8-qcow2" {
  name           = "vol-debian8-qcow2"
  base_volume_id = libvirt_volume.debian8-qcow2.id
}

# domain using raw-backed volume
resource "libvirt_domain" "domain-debian8-raw" {
  name   = "domain-debian8-raw"
  memory = "256"
  vcpu   = 1

  network_interface {
    network_name = "tf"
  }

  disk {
    volume_id = libvirt_volume.vol-debian8-raw.id
  }

  graphics {
    type        = "spice"
    listen_type = "address"
    autoport    = true
  }
}

# domain using qcow2-backed volume
resource "libvirt_domain" "domain-debian8-qcow2" {
  name   = "domain-debian8-qcow2"
  memory = "256"
  vcpu   = 1

  network_interface {
    network_name = "tf"
  }

  disk {
    volume_id = libvirt_volume.vol-debian8-qcow2.id
  }

  graphics {
    type        = "spice"
    listen_type = "address"
    autoport    = true
  }
}

terraform {
  required_version = ">= 0.12"
}
