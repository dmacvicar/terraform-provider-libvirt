terraform {
  required_providers {
    libvirt = {
      source = "dmacvicar/libvirt"
    }
  }
}

# Configure the Libvirt Provider
provider "libvirt" {
  # Connection URI - defaults to qemu:///system if not specified
  # uri = "qemu:///system"

  # For user session:
  # uri = "qemu:///session"

  # For remote connections (not yet implemented):
  # uri = "qemu+ssh://user@remote-host/system"
}
