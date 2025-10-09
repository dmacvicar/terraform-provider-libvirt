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

resource "libvirt_domain" "example_with_filesystem" {
  name   = "example-with-filesystem"
  memory = 512
  unit   = "MiB"
  vcpu   = 1

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }

  devices = {
    # Share host directories with the guest using virtio-9p
    filesystems = [
      {
        source     = "/home/user/shared"  # Host directory
        target     = "shared"              # Mount tag (use to mount in guest)
        accessmode = "mapped"              # Security mode
        readonly   = true                  # Mount as read-only
      },
      {
        source     = "/var/data"
        target     = "data"
        accessmode = "passthrough" # Better performance, less secure
        readonly   = false
      }
    ]
  }
}

# In the guest, mount with:
# mount -t 9p -o trans=virtio shared /mnt/shared
# mount -t 9p -o trans=virtio data /mnt/data
