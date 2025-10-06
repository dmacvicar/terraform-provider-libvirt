package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPoolResource_dir(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPoolResourceConfigDir("test-pool-dir"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_pool.test", "name", "test-pool-dir"),
					resource.TestCheckResourceAttr("libvirt_pool.test", "type", "dir"),
					resource.TestCheckResourceAttr("libvirt_pool.test", "target.path", "/tmp/terraform-provider-libvirt-pool-dir"),
					resource.TestCheckResourceAttrSet("libvirt_pool.test", "uuid"),
					resource.TestCheckResourceAttrSet("libvirt_pool.test", "id"),
					resource.TestCheckResourceAttrSet("libvirt_pool.test", "capacity"),
					resource.TestCheckResourceAttrSet("libvirt_pool.test", "allocation"),
					resource.TestCheckResourceAttrSet("libvirt_pool.test", "available"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccPoolResourceConfigDir(name string) string {
	return fmt.Sprintf(`
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_pool" "test" {
  name = %[1]q
  type = "dir"
  target = {
    path = "/tmp/terraform-provider-libvirt-pool-dir"
  }
}
`, name)
}
