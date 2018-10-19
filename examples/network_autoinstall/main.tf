provider "libvirt" {
  uri = "qemu:///system"
}
resource "libvirt_volume" "leap15-net" {
  name = "leap15-net"
  pool = "default"
  size = 10000000000
}
resource "libvirt_volume" "kernel" {
  source = "http://download.opensuse.org/distribution/leap/15.0/repo/oss/boot/x86_64/loader/linux"
  name = "kernel"
  pool = "default"
}
resource "libvirt_volume" "initrd" {
  source = "http://download.opensuse.org/distribution/leap/15.0/repo/oss/boot/x86_64/loader/initrd"
  name = "initrd"
  pool = "default"
}
resource "libvirt_domain" "domain-suse-qcow2" {
  name = "leap15"
  memory = "1024"
  vcpu = 1
  network_interface {
  network_name = "default"
  }
  kernel = "${libvirt_volume.kernel.id}"
  initrd = "${libvirt_volume.initrd.id}"
   boot_device {
    dev = ["network", "hd"]
  }
  cmdline {
      "_" = "noapic"
      install = "http://download.opensuse.org/distribution/leap/15.0/repo/oss/"
      AutoYaST = "https://github.com/dmacvicar/terraform-provider-libvirt/tree/master/examples/network_installation/leap15.xml"
  }
  disk {
      volume_id = "${libvirt_volume.leap15-net.id}"
  }
graphics {
    type = "vnc"
    listen_type = "address"
  }
}
