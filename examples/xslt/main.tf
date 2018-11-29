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
    xslt = "${file("nicmodel.xsl")}"
  }
}
