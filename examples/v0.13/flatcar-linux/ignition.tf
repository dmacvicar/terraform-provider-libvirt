data "ignition_config" "ignition" {
  users = [
    data.ignition_user.core.rendered,
  ]

  files = [
    element(data.ignition_file.hostname.*.rendered, count.index)
  ]

  networkd = [
    "${data.ignition_networkd_unit.network-dhcp.rendered}",
  ]

  systemd = [
    "${data.ignition_systemd_unit.etcd-member[count.index].rendered}",
  ]

  count = var.hosts
}

data "ignition_file" "hostname" {
  filesystem = "root"
  path       = "/etc/hostname"
  mode       = 420

  content {
    content = format(var.hostname_format, count.index + 1)
  }

  count = var.hosts
}

data "ignition_user" "core" {
  name = "core"

  ssh_authorized_keys = ["ssh-rsa <your-ssh-pub-key>"]
}

data "ignition_networkd_unit" "network-dhcp" {
  name    = "00-wired.network"
  content = "${file("${path.module}/units/00-wired.network")}"
}

data "ignition_systemd_unit" "etcd-member" {
  name    = "etcd-member.service"
  enabled = true
  dropin {
    content = "${data.template_file.etcd-member[count.index].rendered}"
    name    = "20-etcd-member.conf"
  }
  count = var.hosts
}

resource "random_string" "token" {
  length  = 16
  special = false
}

data "template_file" "etcd-member" {
  template = "${file("${path.module}/units/20-etcd-member.conf")}"
  count    = var.hosts
  vars = {
    node_name     = format(var.hostname_format, count.index + 1)
    private_ip    = format("192.168.122.1%02d", count.index + 1)
    cluster_token = random_string.token.result
  }
}

