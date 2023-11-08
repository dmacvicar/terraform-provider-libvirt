package libvirt

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccLibvirtNodeDeviceInfoDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceNodeDeviceInfoPCI,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.libvirt_node_device_info.device", "name", regexp.MustCompile(`^pci_0000_00_00_0$`)),
				),
			}, {
				Config: testAccDataSourceNodeDeviceInfoSystem,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.libvirt_node_device_info.device", "name", regexp.MustCompile(`^computer$`)),
				),
			},
		},
	})
}

const testAccDataSourceNodeDeviceInfoPCI = `
data "libvirt_node_device_info" "device" {
	name = "pci_0000_00_00_0"
}`

const testAccDataSourceNodeDeviceInfoSystem = `
data "libvirt_node_device_info" "device" {
	name = "computer"
}`
