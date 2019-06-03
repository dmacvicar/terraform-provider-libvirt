# instance the provider
provider "libvirt" {
  uri = "qemu:///system"
}

locals {
  ceph_user = "libvirt"
  ceph_uuid = "a453e843-abdd-4f56-b27f-a3eac672c8b4"
  ceph_pool = "rbd"

  ceph_mons = [
    "rbd://10.0.0.1:6789",
    "rbd://10.0.0.2:6789",
    "rbd://10.0.0.3:6789",
  ]

  ceph_config = {
    user = "${local.ceph_user}"
    uuid = "${local.ceph_uuid}"
    pool = "${local.ceph_pool}"
    mons = "${local.ceph_mons}"
  }
}

# Same pool as in the ubuntu example. But it won't be used for the actual vm disk
resource "libvirt_pool" "ubuntu" {
  name = "ubuntu"
  type = "dir"
  path = "/tmp/terraform-provider-libvirt-pool-ubuntu"
}

# We use a volume that already exists on the ceph cluster as a base for this vm
resource "libvirt_volume" "ubuntu-server" {
  name             = "ubuntu-server"
  pool             = "rbd_image"
  base_volume_name = "ubuntu-16.04-server-cloudimg"
  format           = "raw"

  ## To actually work with a base image from ceph we need the clone option
  #clone            = true
}

data "template_file" "user_data" {
  template = "${file("${path.module}/cloud_init.cfg")}"
}

data "template_file" "network_config" {
  template = "${file("${path.module}/network_config.cfg")}"
}

# for more info about paramater check this out
# https://github.com/dmacvicar/terraform-provider-libvirt/blob/master/website/docs/r/cloudinit.html.markdown
# Use CloudInit to add our ssh-key to the instance
# you can add also meta_data field
resource "libvirt_cloudinit_disk" "commoninit" {
  name           = "commoninit.iso"
  user_data      = "${data.template_file.user_data.rendered}"
  network_config = "${data.template_file.network_config.rendered}"
  pool           = "${libvirt_pool.ubuntu.name}"
}

# Create the machine
resource "libvirt_domain" "domain-ubuntu" {
  name   = "ubuntu-terraform"
  memory = "512"
  vcpu   = 1

  cloudinit = "${libvirt_cloudinit_disk.commoninit.id}"

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
    type        = "tcp"
    target_type = "virtio"
    target_port = "1"
  }

  disk {
    volume_id = "${libvirt_volume.ubuntu-server.id}"
    ceph      = "${local.ceph_config}"
  }

  graphics {
    type        = "spice"
    listen_type = "address"
    autoport    = true
  }
}

# IPs: use wait_for_lease true or after creation use terraform refresh and terraform show for the ips of domain

