package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNodeDevicesDataSource_all(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeDevicesDataSourceConfigAll(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID is set
					resource.TestCheckResourceAttrSet("data.libvirt_node_devices.test", "id"),
					// Verify devices set is populated (should have at least one device)
					resource.TestCheckResourceAttrSet("data.libvirt_node_devices.test", "devices.#"),
				),
			},
		},
	})
}

func TestAccNodeDevicesDataSource_filtered(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeDevicesDataSourceConfigFiltered("pci"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID is set
					resource.TestCheckResourceAttrSet("data.libvirt_node_devices.test", "id"),
					// Verify capability filter is applied
					resource.TestCheckResourceAttr("data.libvirt_node_devices.test", "capability", "pci"),
					// Verify devices set is populated
					resource.TestCheckResourceAttrSet("data.libvirt_node_devices.test", "devices.#"),
				),
			},
		},
	})
}

func testAccNodeDevicesDataSourceConfigAll() string {
	return `
data "libvirt_node_devices" "test" {
}
`
}

func testAccNodeDevicesDataSourceConfigFiltered(capability string) string {
	return `
data "libvirt_node_devices" "test" {
  capability = "` + capability + `"
}
`
}
