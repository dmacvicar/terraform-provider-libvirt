# Terraform ignition configuration 
# All configuration options are detailed at
# https://www.terraform.io/docs/providers/ignition/index.html

data "ignition_config" "startup" {
  users = [
    data.ignition_user.core.rendered,
  ]

  files = [
    element(data.ignition_file.hostname.*.rendered, count.index),
  ]

  ## Relevant for the QEMU Guest Agent example
  #systemd = [
  #  "${data.ignition_systemd_unit.mount-images.rendered}",
  #  "${data.ignition_systemd_unit.qemu-agent.rendered}"
  #]
  count = var.hosts
}

# Replace the default hostname with our generated one
data "ignition_file" "hostname" {
  path       = "/etc/hostname"
  mode       = 420 # decimal 0644

  content {
    content = format(var.hostname_format, count.index + 1)
  }

  count = var.hosts
}

# Example configuration for the basic `core` user
data "ignition_user" "core" {
  name = "core"

  #Example password: foobar
  password_hash = "$5$XMoeOXG6$8WZoUCLhh8L/KYhsJN2pIRb3asZ2Xos3rJla.FA1TI7"
  # Preferably use the ssh key auth instead
  #ssh_authorized_keys = "${list()}"
}

## Relevant for the QEMU Guest Agent example
#data "ignition_systemd_unit" "mount-images" {
#  name = "var-mnt-images.mount"
#  enabled = true
#  content = "${file("${path.module}/qemu-agent/docker-images.mount")}"
#}
## Relevant for the QEMU Guest Agent example
#data "ignition_systemd_unit" "qemu-agent" {
#  name = "qemu-agent.service"
#  enabled = true
#  content = "${file("${path.module}/qemu-agent/qemu-agent.service")}"
#}
