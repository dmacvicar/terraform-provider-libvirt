provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_volume" "volume-test" {
                name = "volume-test"
                size =  1073741824
}
