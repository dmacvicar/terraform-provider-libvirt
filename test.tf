provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_volume" "coreos-qcow2" {
  name = "coreos-qcow2"
  source = "https://stable.release.core-os.net/amd64-usr/current/coreos_production_qemu_image.img.bz2"
  format = "qcow2"
}
