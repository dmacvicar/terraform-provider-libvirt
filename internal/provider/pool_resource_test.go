package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	libvirtclient "github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	resource.AddTestSweepers("libvirt_pool", &resource.Sweeper{
		Name:         "libvirt_pool",
		Dependencies: []string{"libvirt_volume"},
		F: func(uri string) error {
			ctx := context.Background()
			client, err := libvirtclient.NewClient(ctx, uri)
			if err != nil {
				return fmt.Errorf("failed to create libvirt client: %w", err)
			}
			defer func() { _ = client.Close() }()

			// List all storage pools
			pools, _, err := client.Libvirt().ConnectListAllStoragePools(1, 0)
			if err != nil {
				return fmt.Errorf("failed to list storage pools: %w", err)
			}

			// Delete test pools (prefix: test-)
			for _, pool := range pools {
				if strings.HasPrefix(pool.Name, "test-") || strings.HasPrefix(pool.Name, "test_") {
					// Try to destroy if active
					_ = client.Libvirt().StoragePoolDestroy(pool)
					// Undefine the pool
					if err := client.Libvirt().StoragePoolUndefine(pool); err != nil {
						fmt.Printf("Warning: failed to undefine pool %s: %v\n", pool.Name, err)
					}
				}
			}

			return nil
		},
	})
}

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
