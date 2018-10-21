provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_domain" "domain-suse-qcow2" {
  name                = "loop"
  network_autoinstall = true
}
