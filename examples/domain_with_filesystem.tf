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
  memory_unit   = "MiB"
  vcpu   = 1

  os = {
    type         = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  devices = {
    # Share host directories with the guest using virtio-9p
    filesystems = [
      {
        source = {
          mount = {
            dir = "/home/user/shared"
          }
        }
        target = {
          dir = "shared"
        }
        access_mode = "mapped"
        read_only   = true                  # Mount as read-only
      },
      {
        source = {
          mount = {
            dir = "/var/data"
          }
        }
        target = {
          dir = "data"
        }
        access_mode = "passthrough" # Better performance, less secure
        read_only   = false
      }
    ]
  }
}

# In the guest, mount with:
# mount -t 9p -o trans=virtio shared /mnt/shared
# mount -t 9p -o trans=virtio data /mnt/data
