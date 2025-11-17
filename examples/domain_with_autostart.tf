terraform {
  required_providers {
    libvirt = {
      source = "dmacvicar/libvirt"
    }
  }
}

provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_domain" "example_with_autostart" {
  name      = "example-with-autostart"
  memory    = 512
  memory_unit      = "MiB"
  vcpu      = 1
  autostart = true # Domain will automatically start on host boot

  os {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }
}
