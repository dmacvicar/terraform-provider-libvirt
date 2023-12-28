package libvirt

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccLibvirtNodeInfoDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceNodeInfo,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.libvirt_node_info.info", "cpus", regexp.MustCompile(`^\d+`)),
				),
			}, {
				Config: testAccDataSourceNodeInfo,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.libvirt_node_info.info", "numa_nodes", regexp.MustCompile(`^\d+`)),
				),
			},
		},
	})
}

const testAccDataSourceNodeInfo = `
data "libvirt_node_info" "info" {

}`
