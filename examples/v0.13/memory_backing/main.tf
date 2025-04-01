# Configure the provider
provider "libvirt" {
  uri = "qemu:///system"
}

# Simple memfd example
resource "libvirt_domain" "memfd_example" {
  name   = "memfd-example"
  memory = 2048
  vcpu   = 2

  memory_backing {
    source_type = "memfd"
  }
}

# Complex memory backing example
resource "libvirt_domain" "memory_backing_example" {
  name   = "memory-backing-example"
  memory = 4096
  vcpu   = 4

  memory_backing {
    source_type = "memfd"
    access_mode = "shared"
    allocation_mode = "immediate"
    discard = true
    locked = true
    
    hugepages {
      size = 2048
      nodeset = "0-1"
    }
  }
}