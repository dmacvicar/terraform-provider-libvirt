# Basic volume
resource "libvirt_volume" "example" {
  name     = "example.qcow2"
  pool     = "default"
  capacity = 10737418240 # 10 GB
  target = {
    format = {
      type = "qcow2"
    }
  }
}

# Volume with backing store
resource "libvirt_volume" "base" {
  name     = "base.qcow2"
  pool     = "default"
  capacity = 10737418240
  target = {
    format = {
      type = "qcow2"
    }
  }
}

resource "libvirt_volume" "overlay" {
  name     = "overlay.qcow2"
  pool     = "default"
  capacity = 10737418240

  backing_store = {
    path   = libvirt_volume.base.path
    format = {
      type = "qcow2"
    }
  }
}

# Volume from HTTP URL upload
resource "libvirt_volume" "ubuntu_base" {
  name   = "ubuntu-22.04.qcow2"
  pool   = "default"
  target = {
    format = {
      type = "qcow2"
    }
  }

  create = {
    content = {
      url = "https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img"
    }
  }
  # capacity is automatically computed from Content-Length when available
}

# Volume from local file upload
resource "libvirt_volume" "from_local" {
  name   = "custom-image.qcow2"
  pool   = "default"
  target = {
    format = {
      type = "qcow2"
    }
  }

  create = {
    content = {
      url = "/path/to/local/image.qcow2"
      # or: url = "file:///path/to/local/image.qcow2"
    }
  }
  # capacity is automatically computed from file size
}

# Volume adopting an existing file on the libvirt host
resource "libvirt_volume" "adopted" {
  name = "ubuntu-22.04.qcow2"
  pool = "default"

  source = {
    host_path = "/var/lib/libvirt/images/ubuntu-22.04.qcow2"
  }
  # capacity is automatically computed from the existing file
}

# Volume cloned from an existing volume (full copy)
resource "libvirt_volume" "vm_clone" {
  name     = "vm-disk.qcow2"
  pool     = "default"
  capacity = 21474836480 # 20 GB (may be larger than source)

  clone = {
    volume     = libvirt_volume.base.name
    pool       = libvirt_volume.base.pool
    full_clone = true # full copy (default), set false for reflink/COW
  }
}
