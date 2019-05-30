provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_volume" "os_image" {
  name   = "os_image"
  source = "http://download.opensuse.org/repositories/Cloud:/Images:/Leap_42.3/images/openSUSE-Leap-42.3-OpenStack.x86_64.qcow2"
}

resource "libvirt_volume" "volume" {
  name           = "volume-${count.index}"
  base_volume_id = libvirt_volume.os_image.id
  count          = 4
}

resource "libvirt_domain" "domain" {
  name = "domain-${count.index}"

  disk {
    volume_id = element(libvirt_volume.volume.*.id, count.index)
  }

  network_interface {
    network_name   = "default"
    wait_for_lease = true
  }

  count = 4
}

output "ips" {
  value = libvirt_domain.domain.*.network_interface.0.addresses
}

terraform {
  required_version = ">= 0.12"
}
