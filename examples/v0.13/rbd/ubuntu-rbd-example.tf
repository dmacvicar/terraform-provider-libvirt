terraform {
 required_version = ">= 0.13"
  required_providers {
    libvirt = {
      source  = "gxben/libvirt"
      version = "0.6.11"
    }
  }
}

# instance the provider
provider "libvirt" {
  uri = "qemu:///system"
}

# We previously added an Ubuntu QCOW2 cloud-image to Ceph, that will be used as a reference template.
# $ wget http://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64-disk-kvm.img
# $ qemu-img convert -p -f qcow2 -O rbd ubuntu-20.04-server-cloudimg-amd64-disk-kvm.img rbd:rbd_pool/ubuntu-cloudimg-20.04

# For convenience, it is easier to configure the RBD pool at libvirt system level rather than Terraform level.
# This can be done once through:
# $ virsh pool-define pool.xml
# $ virsh pool-start rbd_pool
# where pool.xml can be inspired from the following (where ceph.acme.com is the Ceph monitor host):
# <pool type="rbd">
#   <name>rbd_pool</name>
#   <source>
#     <name>rbd_pool</name>
#     <host name='ceph.acme.com' port='3300'/>
#   </source>
# </pool>

# We then ask Terraform to create a new volume on RBD by cloning the RBD reference template and resize it to 50 GB.
resource "libvirt_volume" "os-disk" {
  name             = "os-disk"
  type             = "rbd"
  pool             = "rbd_pool"
  format           = "raw"
  base_volume_pool = "rbd_pool"
  base_volume_name = "ubuntu-cloudimg-20.04"
  size             = "53687091200" # 50 GB
}

# Create the machine
resource "libvirt_domain" "domain-ubuntu" {
  name   = "ubuntu-terraform"
  memory = "1024"
  vcpu   = 1

  disk {
    rbd       = true
    rbd_host  = "ceph.acme.com"
    rbd_port  = "3300"
    rbd_pool  = "rbd_pool"
    rbd_image = "os-disk"
  }
}
