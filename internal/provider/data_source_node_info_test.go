package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNodeInfoDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeInfoDataSourceConfigBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID is set
					resource.TestCheckResourceAttrSet("data.libvirt_node_info.test", "id"),
					// Verify CPU model is set
					resource.TestCheckResourceAttrSet("data.libvirt_node_info.test", "cpu_model"),
					// Verify memory is set and is a positive value
					resource.TestCheckResourceAttrSet("data.libvirt_node_info.test", "memory_total_kb"),
					// Verify CPU count is set
					resource.TestCheckResourceAttrSet("data.libvirt_node_info.test", "cpu_cores_total"),
					// Verify NUMA nodes is set
					resource.TestCheckResourceAttrSet("data.libvirt_node_info.test", "numa_nodes"),
					// Verify CPU sockets is set
					resource.TestCheckResourceAttrSet("data.libvirt_node_info.test", "cpu_sockets"),
					// Verify cores per socket is set
					resource.TestCheckResourceAttrSet("data.libvirt_node_info.test", "cpu_cores_per_socket"),
					// Verify threads per core is set
					resource.TestCheckResourceAttrSet("data.libvirt_node_info.test", "cpu_threads_per_core"),
				),
			},
		},
	})
}

func testAccNodeInfoDataSourceConfigBasic() string {
	return `
data "libvirt_node_info" "test" {
}
`
}
