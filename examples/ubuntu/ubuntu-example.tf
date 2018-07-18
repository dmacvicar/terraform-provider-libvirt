# instance the provider
provider "libvirt" {
  uri = "qemu:///system"
}

# We fetch the latest ubuntu release image from their mirrors
resource "libvirt_volume" "ubuntu-qcow2" {
  name = "ubuntu-qcow2"
  pool = "default"
  source = "https://cloud-images.ubuntu.com/releases/xenial/release/ubuntu-16.04-server-cloudimg-amd64-disk1.img"
  format = "qcow2"
}

# Create a network for our VMs
resource "libvirt_network" "vm_network" {
   name = "vm_network"
   addresses = ["10.0.1.0/24"]
}

# Use CloudInit to add our ssh-key to the instance
resource "libvirt_cloudinit" "commoninit" {
          name           = "commoninit.iso"
          ssh_authorized_key = "<ssh-key-here>"
        }


# Create the machine
resource "libvirt_domain" "domain-ubuntu" {
  name = "ubuntu-terraform"
  memory = "512"
  vcpu = 1

  cloudinit = "${libvirt_cloudinit.commoninit.id}"

  network_interface {
    network_name = "vm_network"
  }

  # IMPORTANT
  # Ubuntu can hang if an isa-serial is not present at boot time.
  # If you find your CPU 100% and never is available this is why
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
       volume_id = "${libvirt_volume.ubuntu-qcow2.id}"
  }
  graphics {
    type = "spice"
    listen_type = "address"
    autoport = "yes"
  }
}

# Print the Boxes IP
# Note: you can use `virsh domifaddr <vm_name> <interface>` to get the ip later
output "ip" {
  value = "${libvirt_domain.domain-ubuntu.network_interface.0.addresses.0}"
}
