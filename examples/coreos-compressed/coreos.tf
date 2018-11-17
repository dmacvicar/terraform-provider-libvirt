# instance the provider
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_volume" "coreos-qcow2" {
  name = "coreos-qcow2"
  source = "https://github.com/MalloZup/terraform-provider-libvirt/raw/compressed-images/libvirt/testdata/gzip/test.qcow2.gz"
  //source = "https://stable.release.core-os.net/amd64-usr/current/coreos_production_qemu_image.img.bz2"
  //source = "https://cloud-images.ubuntu.com/cosmic/current/cosmic-server-cloudimg-amd64-lxd.tar.xz"
}

# Create the machine
resource "libvirt_domain" "domain-coreos" {
  name   = "coreos-terraform"
  memory = "512"
  vcpu   = 1

  network_interface {
    network_name = "default"
  }

  # IMPORTANT: this is a known bug on cloud images, since they expect a console
  # we need to pass it
  # https://bugs.launchpad.net/cloud-images/+bug/1573095
  console {
    type        = "pty"
    target_port = "0"
    target_type = "serial"
  }

  console {
    type        = "pty"
    target_type = "virtio"
    target_port = "1"
  }

  disk {
    volume_id = "${libvirt_volume.coreos-qcow2.id}"
  }

  graphics {
    type        = "spice"
    listen_type = "address"
    autoport    = true
  }
}

# IPs: use wait_for_lease true or after creation use terraform refresh and terraform show for the ips of domain

