# Basic volume
resource "libvirt_volume" "example" {
  name     = "example.qcow2"
  pool     = "default"
  capacity = 10737418240 # 10 GB
  format   = "qcow2"
}

# Volume with backing store
resource "libvirt_volume" "base" {
  name     = "base.qcow2"
  pool     = "default"
  capacity = 10737418240
  format   = "qcow2"
}

resource "libvirt_volume" "overlay" {
  name     = "overlay.qcow2"
  pool     = "default"
  capacity = 10737418240

  backing_store = {
    path   = libvirt_volume.base.path
    format = "qcow2"
  }
}

# Volume from HTTP URL upload
resource "libvirt_volume" "ubuntu_base" {
  name   = "ubuntu-22.04.qcow2"
  pool   = "default"
  format = "qcow2"

  create = {
    content = {
      url = "https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img"
    }
  }
  # capacity is automatically computed from Content-Length header
}

# Volume from local file upload
resource "libvirt_volume" "from_local" {
  name   = "custom-image.qcow2"
  pool   = "default"
  format = "qcow2"

  create = {
    content = {
      url = "/path/to/local/image.qcow2"
      # or: url = "file:///path/to/local/image.qcow2"
    }
  }
  # capacity is automatically computed from file size
}
