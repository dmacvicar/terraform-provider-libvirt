# main module for kvm-guest

resource "libvirt_volume" "os_volume" {
  name = "os_volume_${var.name}.qcow2"
  pool = "ssdpool"

  create = {
    content = {
      url    = var.image
      format = "qcow2"
    }
  }
}

resource "libvirt_volume" "data_volume" {
  count    = var.data_disk ? 1 : 0
  name     = "data_volume_${var.name}.qcow2"
  pool     = var.data_pool
  capacity = 214748364800

  target = {
    format = {
      type = "qcow2"
    }
  }
}

resource "libvirt_cloudinit_disk" "cloudinit_seed" {
  name = "${var.name}-cloudinit"

  user_data = templatefile(var.cloudinit_template_path, {
    server_name = var.name
    vpc_domain  = var.vpc_domain
  })

  network_config = templatefile(var.network_template_path, {})

  meta_data = yamlencode({
    "instance-id"    = var.name
    "local-hostname" = var.name
  })
}

resource "libvirt_volume" "cloudinit_seed_volume" {
  name = "${var.name}-cloudinit_seed.iso"
  pool = "cloudinit"

  create = {
    content = {
      url = libvirt_cloudinit_disk.cloudinit_seed.path
    }
  }
}

resource "libvirt_domain" "kvm_vpc_guest" {
  name         = "${var.name}.${var.vpc_domain}"
  memory       = var.memory
  memory_unit  = "MiB"
  vcpu         = var.vcpus
  type         = "kvm"
  autostart    = true
  running      = true

  os = {
    type         = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  cpu = {
    mode = "host-passthrough"
  }

  devices = {
    disks = concat(
      [
      {
        device = "cdrom"
        source = {
          volume = {
            pool = libvirt_volume.cloudinit_seed_volume.pool
            volume = libvirt_volume.cloudinit_seed_volume.name
          }
        }
        target = {
          dev = "sda"
          bus = "sata"
          }
        },
        {
          source = {
            volume = {
              pool   = libvirt_volume.os_volume.pool
              volume = libvirt_volume.os_volume.name
            }
          }
          driver = { type = "qcow2" }
          target = { dev = "vda", bus = "virtio" }
        }
      ],
      [
        for vol in libvirt_volume.data_volume : {
          source = {
            volume = {
              pool   = vol.pool
              volume = vol.name
            }
          }
          driver = { type = "qcow2" }
          target = { dev = "vdb", bus = "virtio" }
          }
        ]
    )

    interfaces = [
      {
        wait_for_ip = {
          timeout = 300
          source = "any"
        }
        model = { type = "virtio" }
        source = {
          network = {
            network = var.network
          }
        }
      }
    ]
    consoles = [
      {
        source = {
          target = {
            type = "serial"
            port = "0"
          }
        }
      }
    ]
    channels = [
      {
        source = {
          unix = {
            }
        }
        target = {
          virt_io = {
            name = "org.qemu.guest_agent.0"
          }
        }
      }
    ]
  }
}

 data "libvirt_domain_interface_addresses" "kvm_vpc_guest" {
   domain = libvirt_domain.kvm_vpc_guest.name
   source = "any"
   depends_on = [libvirt_domain.kvm_vpc_guest]
 }
