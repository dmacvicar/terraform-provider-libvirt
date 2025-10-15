package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNodeDeviceInfoDataSource_pci(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeDeviceInfoDataSourceConfigPCI(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify basic fields
					resource.TestCheckResourceAttrSet("data.libvirt_node_device_info.test", "id"),
					resource.TestCheckResourceAttrSet("data.libvirt_node_device_info.test", "name"),
					resource.TestCheckResourceAttrSet("data.libvirt_node_device_info.test", "path"),
					// Verify capability type
					resource.TestCheckResourceAttr("data.libvirt_node_device_info.test", "capability.type", "pci"),
					// Verify PCI fields are set
					resource.TestCheckResourceAttrSet("data.libvirt_node_device_info.test", "capability.domain"),
					resource.TestCheckResourceAttrSet("data.libvirt_node_device_info.test", "capability.bus"),
					resource.TestCheckResourceAttrSet("data.libvirt_node_device_info.test", "capability.slot"),
					resource.TestCheckResourceAttrSet("data.libvirt_node_device_info.test", "capability.function"),
				),
			},
		},
	})
}

func testAccNodeDeviceInfoDataSourceConfigPCI() string {
	return `
# First get a list of PCI devices
data "libvirt_node_devices" "pci" {
  capability = "pci"
}

# Then get details about the first PCI device
data "libvirt_node_device_info" "test" {
  name = tolist(data.libvirt_node_devices.pci.devices)[0]
}
`
}
