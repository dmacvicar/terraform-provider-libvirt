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
				Config: testAccDataSourceNodeDevicesPci,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.libvirt_node_devices.node", "devices[0]", regexp.MustCompile(`^pci_\d{4}_\d{2}_\d{2}_\d{1}`)),
				),
			},
			{
				Config: testAccDataSourceNodeDevicesSystem,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.libvirt_node_devices.node", "devices[0]", regexp.MustCompile(`^computer$`)),
				),
			},
		},
	})
}

const testAccDataSourceNodeDevicesPci = `
data "libvirt_node_devices" "node" {
	capability = "pci"
}`

const testAccDataSourceNodeDevicesSystem = `
data "libvirt_node_devices" "node" {
	capability = "system"
}`
