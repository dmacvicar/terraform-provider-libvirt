package libvirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccLibvirtNetworkDataSource_DNSHostTemplate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtNetworkDestroy,
		Steps: []resource.TestStep{

			{
				Config: `data "libvirt_network_dns_host_template" "bootstrap" {
  count = 2
  ip = "1.1.1.${count.index}"
  hostname = "myhost${count.index}"
}`,
				Check: resource.ComposeTestCheckFunc(
					checkDNSHostTemplate("data.libvirt_network_dns_host_template.bootstrap.0", "ip", "1.1.1.0"),
					checkDNSHostTemplate("data.libvirt_network_dns_host_template.bootstrap.0", "hostname", "myhost0"),
					checkDNSHostTemplate("data.libvirt_network_dns_host_template.bootstrap.1", "ip", "1.1.1.1"),
					checkDNSHostTemplate("data.libvirt_network_dns_host_template.bootstrap.1", "hostname", "myhost1"),
				),
			},
		},
	})
}

func checkDNSHostTemplate(id, name, value string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		rs, err := getResourceFromTerraformState(id, state)
		if err != nil {
			return err
		}

		v := rs.Primary.Attributes[name]
		if v != value {
			return fmt.Errorf(
				"Value for %s is %s, not %s", name, v, value)
		}

		return nil
	}
}
