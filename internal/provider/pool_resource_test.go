package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	golibvirt "github.com/digitalocean/go-libvirt"
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

func TestAccPoolResource_dir_deletesBackingPathOnDestroy(t *testing.T) {
	parentDir := t.TempDir()
	poolPath := filepath.Join(parentDir, "pool-delete-on-destroy")
	if err := os.Mkdir(poolPath, 0o755); err != nil {
		t.Fatalf("failed to create pool path: %v", err)
	}

	resourceName := "libvirt_pool.test"
	poolName := "test-pool-delete-backend"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroyAndPathDeleted(poolName, poolPath),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolResourceConfigDirWithDestroyDelete(poolName, poolPath, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", poolName),
					resource.TestCheckResourceAttr(resourceName, "type", "dir"),
					resource.TestCheckResourceAttr(resourceName, "target.path", poolPath),
				),
			},
		},
	})
}

func TestAccPoolResource_dir_createStartAndAutostartOverrides(t *testing.T) {
	poolPath := t.TempDir()
	resourceName := "libvirt_pool.test"
	poolName := "test-pool-create-overrides"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPoolResourceConfigDirWithCreate(poolName, poolPath, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", poolName),
					resource.TestCheckResourceAttr(resourceName, "type", "dir"),
					resource.TestCheckResourceAttr(resourceName, "target.path", poolPath),
					testAccCheckPoolStateAndAutostart(poolName, golibvirt.StoragePoolInactive, 0),
				),
			},
		},
	})
}

func TestAccPoolResource_dir_createBuildOverride(t *testing.T) {
	poolRoot := t.TempDir()
	pathNoBuild := filepath.Join(poolRoot, "pool-build-false")
	pathWithBuild := filepath.Join(poolRoot, "pool-build-true")
	poolNoBuild := "test-pool-build-false"
	poolWithBuild := "test-pool-build-true"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPoolResourceConfigBuildComparison(poolNoBuild, pathNoBuild, poolWithBuild, pathWithBuild),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_pool.no_build", "name", poolNoBuild),
					resource.TestCheckResourceAttr("libvirt_pool.no_build", "target.path", pathNoBuild),
					testAccCheckPathDoesNotExist(pathNoBuild),
					testAccCheckPoolStateAndAutostart(poolNoBuild, golibvirt.StoragePoolInactive, 0),
					resource.TestCheckResourceAttr("libvirt_pool.with_build", "name", poolWithBuild),
					resource.TestCheckResourceAttr("libvirt_pool.with_build", "target.path", pathWithBuild),
					testAccCheckPathExists(pathWithBuild),
					testAccCheckPoolStateAndAutostart(poolWithBuild, golibvirt.StoragePoolInactive, 0),
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

func testAccCheckPoolDestroyAndPathDeleted(poolName, poolPath string) resource.TestCheckFunc {
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

		if _, err := os.Stat(poolPath); !os.IsNotExist(err) {
			if err != nil {
				return fmt.Errorf("expected backing path %q to be deleted: %w", poolPath, err)
			}
			return fmt.Errorf("expected backing path %q to be deleted, but it still exists", poolPath)
		}

		return nil
	}
}

func testAccCheckPathDoesNotExist(path string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			if err != nil {
				return fmt.Errorf("expected path %q to not exist: %w", path, err)
			}
			return fmt.Errorf("expected path %q to not exist, but it exists", path)
		}
		return nil
	}
}

func testAccCheckPathExists(path string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("expected path %q to exist: %w", path, err)
		}
		return nil
	}
}

func testAccCheckPoolStateAndAutostart(poolName string, expectedState golibvirt.StoragePoolState, expectedAutostart int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()
		client, err := libvirtclient.NewClient(ctx, testAccLibvirtURI())
		if err != nil {
			return fmt.Errorf("failed to create libvirt client for pool state check: %w", err)
		}
		defer func() { _ = client.Close() }()

		pool, err := client.Libvirt().StoragePoolLookupByName(poolName)
		if err != nil {
			return fmt.Errorf("failed to lookup pool %q: %w", poolName, err)
		}

		state, _, _, _, err := client.Libvirt().StoragePoolGetInfo(pool)
		if err != nil {
			return fmt.Errorf("failed to get pool info for %q: %w", poolName, err)
		}
		if state != uint8(expectedState) {
			return fmt.Errorf("unexpected pool state for %q: got %d, want %d", poolName, state, uint8(expectedState))
		}

		autostart, err := client.Libvirt().StoragePoolGetAutostart(pool)
		if err != nil {
			return fmt.Errorf("failed to get pool autostart for %q: %w", poolName, err)
		}
		if autostart != expectedAutostart {
			return fmt.Errorf("unexpected pool autostart for %q: got %d, want %d", poolName, autostart, expectedAutostart)
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

func testAccPoolResourceConfigBuildComparison(nameNoBuild, pathNoBuild, nameWithBuild, pathWithBuild string) string {
	return fmt.Sprintf(`
resource "libvirt_pool" "no_build" {
  name = %[1]q
  type = "dir"
  target = {
    path = %[2]q
  }
  create = {
    build     = false
    start     = false
    autostart = false
  }
  destroy = {
    delete = false
  }
}

resource "libvirt_pool" "with_build" {
  name = %[3]q
  type = "dir"
  target = {
    path = %[4]q
  }
  create = {
    build     = true
    start     = false
    autostart = false
  }
  destroy = {
    delete = false
  }
}
`, nameNoBuild, pathNoBuild, nameWithBuild, pathWithBuild)
}

func testAccPoolResourceConfigDirWithCreate(name, path string, start, autostart bool) string {
	return fmt.Sprintf(`
resource "libvirt_pool" "test" {
  name = %[1]q
  type = "dir"
  target = {
    path = %[2]q
  }
  create = {
    start     = %[3]t
    autostart = %[4]t
  }
}
`, name, path, start, autostart)
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
