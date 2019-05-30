provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_volume" "ubuntu-cloud-uefi" {
  name   = "ubuntu-cloud-uefi"
  source = "https://cloud-images.ubuntu.com/xenial/current/xenial-server-cloudimg-amd64-uefi1.img"
}

resource "libvirt_volume" "volume" {
  name           = "vm${count.index}"
  base_volume_id = libvirt_volume.ubuntu-cloud-uefi.id
  count          = 1
}

resource "libvirt_domain" "domain" {
  count  = 2
  name   = "ubuntu-cloud-${count.index}"
  memory = "512"
  vcpu   = 1

  # This file is usually present as part of the ovmf firmware package in many
  # Linux distributions.
  firmware = "/usr/share/OVMF/OVMF_CODE.fd"

  nvram {
    # This is the file which will back the UEFI NVRAM content.
    file = "/var/lib/libvirt/qemu/nvram/vm${count.index}_VARS.fd"

    # This file needs to be provided by the user.
    template = "/srv/provisioning/terraform/debian-stable-uefi_VARS.fd"
  }

  disk {
    volume_id = element(libvirt_volume.volume.*.id, count.index)
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
