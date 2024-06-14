# instance the provider
provider "libvirt" {
  uri = var.cluscfg.uri[0]
  #uri = "qemu:///system"
}

variable run_kube_master_init {
  default = "- [ /usr/local/bin/kube_master_init.sh ]"
}

resource "libvirt_pool" "clusterpool" {
  name = var.cluscfg.pool_name
  type = "dir"
  path = var.cluscfg.pool_path
}

 # We fetch the latest ubuntu release image from their mirrors
resource "libvirt_volume" "ubuntu-bionic-qcow2" {
  name   = "ubuntu-bionic-qcow2"
  pool   = libvirt_pool.clusterpool.name
  source = "file:///hpnvmezpool/vms/ubuntuimagepool/ubuntu-bionic-qcow2"
  format = "qcow2"
}

resource "libvirt_volume" "cluster-disk" {
  name             = "${format(var.cluscfg.hostname_format, count.index + 1)}_bioinic.qcow2"
  count            = var.cluscfg.clusterhosts
  base_volume_name = libvirt_volume.ubuntu-bionic-qcow2.name
  pool             = libvirt_pool.clusterpool.name
  format           = "qcow2"
  size             = var.cluscfg.volumesize
}

# Create script for initializing kubernetes cluster, this will become a part  of cloud-init
# user data and will be executed via run
data "template_file" "kube_master_init" {
  template = "${file("${path.module}/kube_master_init.tpl")}"
  vars = {
    clusteruser      = var.cluscfg.clusteruser
    pod_network_cidr = var.cluscfg.pod_network_cidr
  }
}

data "template_file" "user_data" {
  template = file("${path.module}/cloud_init.tpl")
  count    = var.cluscfg.clusterhosts

  vars = {
    basehostname = "${format(var.cluscfg.hostname_format, count.index + 1)}"
    hostname = format("%s.%s","${format(var.cluscfg.hostname_format, count.index + 1)}", var.cluscfg.domain_name)
    clusteruser            = var.cluscfg.clusteruser         
    clusteruserfullname    = var.cluscfg.clusteruserfullname   
    passwdhash             = var.cluscfg.passwdhash        
    ssh_authorized_key     = var.cluscfg.ssh_authorized_key
    root_passwd            = var.cluscfg.root_passwd
    write_files_section    = data.template_file.kube_master_init.rendered
    kube_master_init       = count.index == 0 ? var.run_kube_master_init: ""
  }
}

data "template_file" "network_config" {
  template = file("${path.module}/network_config.tpl")
  count    = var.cluscfg.clusterhosts

  vars = {
    # For this lab working assumption is we allocate IPs statically
    # and ip allocation starts at var.cluscfg.start_ip
    ip                    = "${format(var.cluscfg.ip_format, count.index + var.cluscfg.start_ip)}"
    searchdomain          = var.cluscfg.domain_name 
    gateway               = var.cluscfg.gateway
    nameservers           = var.cluscfg.nameservers
    mtu                   = var.cluscfg.mtu
  }
}

resource "libvirt_cloudinit_disk" "commoninit" {
  count          = var.cluscfg.clusterhosts
  name           = "${format(var.cluscfg.hostname_format, count.index + 1)}_commoninit.iso"
  user_data      = data.template_file.user_data[count.index].rendered
  network_config = data.template_file.network_config[count.index].rendered
  pool           = libvirt_pool.clusterpool.name
}

resource "libvirt_network" "clusternet" {
  # the name used by libvirt
  #name = var.cluscfg.network_name
  name = var.cluscfg.network_name

  # mode can be: "nat" (default), "none", "route", "bridge"
  #mode = "bridge"
  mode = "nat"

  #  the domain used by the DNS server in this network
  domain = var.cluscfg.domain_name

  # list addresses allowed for domains connected
  # also derived to define the host addresses
  # also derived to define the addresses served by the DHCP server
  # Turning off dhcp for now
  #addresses = ["192.168.200.0/24"]
  addresses = [var.cluscfg.network_range]

  # (optional) the bridge device defines the name of a bridge device
  # which will be used to construct the virtual network.
  # (only necessary in "bridge" mode)
  #bridge = "nm-bridge"

  # (optional) the MTU for the network. If not supplied, the underlying device's
  # default is used (usually 1500)
  mtu = var.cluscfg.mtu

  # (Optional) DNS configuration
  dns {
    # (Optional, default false)
    # Set to true, if no other option is specified and you still want to 
    # enable dns.
    enabled = true

    # (Optional, default false)
    # true: DNS requests under this domain will only be resolved by the
    # virtual network's own DNS server
    # false: Unresolved requests will be forwarded to the host's
    # upstream DNS server if the virtual network's DNS server does not
    # have an answer.
    local_only = true

    # (Optional) one or more DNS forwarder entries.  One or both of
    # "address" and "domain" must be specified.  The format is:
    forwarders {
        address = var.cluscfg.nameserver_forwarder
        domain = var.cluscfg.domain_name
      } 

    dynamic "hosts" {
      for_each = range(var.cluscfg.clusterhosts)
      content {
        hostname = format(var.cluscfg.hostname_format, hosts.value + 1)
        ip = format(var.cluscfg.ip_format, hosts.value + var.cluscfg.start_ip)
      }
    }

  }
    
  # (Optional) one or more static routes.
  # "cidr" and "gateway" must be specified. The format is:
  routes {
      cidr = var.cluscfg.network_range
      gateway = var.cluscfg.gateway
    }
}

resource "libvirt_domain" "clusternode" {
  count  = var.cluscfg.clusterhosts
  name   = format(var.cluscfg.hostname_format, count.index + 1)
  vcpu   = var.cluscfg.vcpu
  memory = var.cluscfg.memory

  cloudinit = libvirt_cloudinit_disk.commoninit[count.index].id

  disk {
    volume_id = element(libvirt_volume.cluster-disk.*.id, count.index)
  }

  network_interface {
    network_id = libvirt_network.clusternet.id
  }

  # IMPORTANT: known bug on cloud images, since they expect a console
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

  graphics {
    type        = "spice"
    listen_type = "address"
    autoport    = true
  }
}


terraform {
  required_version = ">= 0.12"
}

