package libvirt

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccLibvirtNodeDevicesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceNodeDevices_pci,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.libvirt_node_devices.node", "devices[0]", regexp.MustCompile(`^pci_\d{4}_\d{2}_\d{2}_\d{1}`)),
				),
			},
			{
				Config: testAccDataSourceNodeDevices_system,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.libvirt_node_devices.node", "devices[0]", regexp.MustCompile(`^computer$`)),
				),
			},
		},
	})
}

const testAccDataSourceNodeDevices_pci = `
data "libvirt_node_devices" "node" {
	capability = "pci"
}`

const testAccDataSourceNodeDevices_system = `
data "libvirt_node_devices" "node" {
	capability = "system"
}`
