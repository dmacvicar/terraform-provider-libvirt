provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_cloudinit_disk" "commoninit" {
  name      = "commoninit.iso"
  user_data = "#cloud-config"
}

# Up to 4 local file CDROMs can be attached to the domain. When attached,
# cloudinit reserves one of the slots.
resource "libvirt_domain" "example" {
  name = "example"

  cloudinit = libvirt_cloudinit_disk.commoninit.id

  disk {
    file = "${path.module}/image.iso"
  }

  disk {
    file = "${path.module}/image2.iso"
  }

  disk {
    file = "${path.module}/image3.iso"
  }
}

terraform {
  required_version = ">= 0.12"
}
