package libvirt

import (
	"fmt"
	"net"
	"strconv"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
)

// a libvirt network DNS host template datasource
//
// Datasource example:
//
// data "libvirt_network_dns_host_template" "k8smasters" {
//   count = "${var.master_count}"
//   ip = "${var.master_ips[count.index]}"
//   hostname = "master-${count.index}"
// }
//
// resource "libvirt_network" "k8snet" {
//   ...
//   dns = [{
//     hosts = [ "${flatten(data.libvirt_network_dns_host_template.k8smasters.*.rendered)}" ]
//   }]
//   ...
// }
//
func datasourceLibvirtNetworkDNSHostTemplate() *schema.Resource {
	return &schema.Resource{
		Read: resourceLibvirtNetworkDNSHostRead,
		Schema: map[string]*schema.Schema{
			"ip": {
				Type:     schema.TypeString,
				Required: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rendered": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
		},
	}
}

func resourceLibvirtNetworkDNSHostRead(d *schema.ResourceData, meta interface{}) error {
	dnsHost := map[string]interface{}{}
	if address, ok := d.GetOk("ip"); ok {
		ip := net.ParseIP(address.(string))
		if ip == nil {
			return fmt.Errorf("Could not parse address '%s'", address)
		}
		dnsHost["ip"] = ip.String()
	}
	if hostname, ok := d.GetOk("hostname"); ok {
		dnsHost["hostname"] = hostname.(string)
	}
	d.Set("rendered", dnsHost)
	d.SetId(strconv.Itoa(hashcode.String(fmt.Sprintf("%v", dnsHost))))

	return nil
}
