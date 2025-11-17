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

resource "libvirt_domain" "example_with_metadata" {
  name        = "example-with-metadata"
  title       = "Example Domain"
  description = "This is an example domain with title and description metadata"
  memory      = 512
  memory_unit        = "MiB"
  vcpu        = 1

  os {
    type    = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }
}
