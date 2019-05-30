provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_domain" "xslt-demo-domain" {
  name   = "xslt-demo-domain"
  memory = "512"

  network_interface {
    network_name = "default"
  }

  xml {
    xslt = file("nicmodel.xsl")
  }
}

terraform {
  required_version = ">= 0.12"
}
