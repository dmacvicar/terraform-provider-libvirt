package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPoolResource_dir(t *testing.T) {
	poolPath := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPoolResourceConfigDir("test-pool-dir", poolPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_pool.test", "name", "test-pool-dir"),
					resource.TestCheckResourceAttr("libvirt_pool.test", "type", "dir"),
					resource.TestCheckResourceAttr("libvirt_pool.test", "target.path", poolPath),
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

func testAccPoolResourceConfigDir(name, path string) string {
	return fmt.Sprintf(`
resource "libvirt_pool" "test" {
  name = %[1]q
  type = "dir"
  target = {
    path = %[2]q
  }
}
`, name, path)
}
