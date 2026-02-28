package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	libvirtclient "github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

func TestAccPoolResource_dir_trailingSlash(t *testing.T) {
	poolPath := t.TempDir() + "/"
	expectedPath := poolPath

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Reproduces inconsistent result when target.path has a trailing slash.
			{
				Config: testAccPoolResourceConfigDir("test-pool-dir-trailing-slash", poolPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_pool.test", "name", "test-pool-dir-trailing-slash"),
					resource.TestCheckResourceAttr("libvirt_pool.test", "type", "dir"),
					resource.TestCheckResourceAttr("libvirt_pool.test", "target.path", expectedPath),
					resource.TestCheckResourceAttrSet("libvirt_pool.test", "uuid"),
					resource.TestCheckResourceAttrSet("libvirt_pool.test", "id"),
					resource.TestCheckResourceAttrSet("libvirt_pool.test", "capacity"),
					resource.TestCheckResourceAttrSet("libvirt_pool.test", "allocation"),
					resource.TestCheckResourceAttrSet("libvirt_pool.test", "available"),
				),
			},
		},
	})
}

func TestAccPoolResource_dir_preservesBackingPathOnDestroy(t *testing.T) {
	poolPath := t.TempDir()
	sentinelPath := filepath.Join(poolPath, "keep-me.txt")
	sentinelContent := []byte("keep")
	if err := os.WriteFile(sentinelPath, sentinelContent, 0o644); err != nil {
		t.Fatalf("failed to create sentinel file: %v", err)
	}

	resourceName := "libvirt_pool.test"
	poolName := "test-pool-preserve-backend"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroyAndSentinelPreserved(poolName, sentinelPath),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolResourceConfigDirWithDestroyDelete(poolName, poolPath, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", poolName),
					resource.TestCheckResourceAttr(resourceName, "type", "dir"),
					resource.TestCheckResourceAttr(resourceName, "target.path", poolPath),
				),
			},
		},
	})
}

func testAccCheckPoolDestroyAndSentinelPreserved(poolName, sentinelPath string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()
		client, err := libvirtclient.NewClient(ctx, testAccLibvirtURI())
		if err != nil {
			return fmt.Errorf("failed to create libvirt client for destroy check: %w", err)
		}
		defer func() { _ = client.Close() }()

		if _, err := client.Libvirt().StoragePoolLookupByName(poolName); err == nil {
			return fmt.Errorf("storage pool %q still exists after destroy", poolName)
		}

		if _, err := os.Stat(sentinelPath); err != nil {
			return fmt.Errorf("expected sentinel file to remain at %q: %w", sentinelPath, err)
		}

		return nil
	}
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

func testAccPoolResourceConfigDirWithDestroyDelete(name, path string, deleteStorage bool) string {
	return fmt.Sprintf(`
resource "libvirt_pool" "test" {
  name = %[1]q
  type = "dir"
  target = {
    path = %[2]q
  }
  destroy = {
    delete = %[3]t
  }
}
`, name, path, deleteStorage)
}
