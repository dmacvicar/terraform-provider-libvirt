package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	libvirtclient "github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	resource.AddTestSweepers("libvirt_volume", &resource.Sweeper{
		Name:         "libvirt_volume",
		Dependencies: []string{"libvirt_domain"},
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

			// For each pool, list volumes and delete test volumes
			for _, pool := range pools {
				vols, _, err := client.Libvirt().StoragePoolListAllVolumes(pool, 1, 0)
				if err != nil {
					continue // Skip pools we can't read
				}

				for _, vol := range vols {
					if strings.HasPrefix(vol.Name, "test-") || strings.HasPrefix(vol.Name, "test_") {
						if err := client.Libvirt().StorageVolDelete(vol, 0); err != nil {
							fmt.Printf("Warning: failed to delete volume %s: %v\n", vol.Name, err)
						}
					}
				}
			}

			return nil
		},
	})
}

func TestAccVolumeResource_basic(t *testing.T) {
	poolPath := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVolumeResourceConfigBasic("test-volume", poolPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_volume.test", "name", "test-volume.qcow2"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "pool", "test-pool-volume"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "capacity", "1073741824"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "format", "qcow2"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "id"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "key"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "path"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "allocation"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccVolumeResourceConfigBasic(name, poolPath string) string {
	return fmt.Sprintf(`

resource "libvirt_pool" "test" {
  name = "test-pool-volume"
  type = "dir"
  target = {
    path = %[2]q
  }
}

resource "libvirt_volume" "test" {
  name     = "%[1]s.qcow2"
  pool     = libvirt_pool.test.name
  capacity = 1073741824
  format   = "qcow2"
}
`, name, poolPath)
}

func TestAccVolumeResource_backingStore(t *testing.T) {
	poolPath := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceConfigBackingStore("test-volume-cow", poolPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_volume.base", "name", "test-volume-cow-base.qcow2"),
					resource.TestCheckResourceAttr("libvirt_volume.cow", "name", "test-volume-cow.qcow2"),
					resource.TestCheckResourceAttrSet("libvirt_volume.cow", "backing_store.path"),
				),
			},
		},
	})
}

func testAccVolumeResourceConfigBackingStore(name, poolPath string) string {
	return fmt.Sprintf(`

resource "libvirt_pool" "test" {
  name = "test-pool-backing"
  type = "dir"
  target = {
    path = %[2]q
  }
}

resource "libvirt_volume" "base" {
  name     = "%[1]s-base.qcow2"
  pool     = libvirt_pool.test.name
  capacity = 1073741824
  format   = "qcow2"
}

resource "libvirt_volume" "cow" {
  name     = "%[1]s.qcow2"
  pool     = libvirt_pool.test.name
  capacity = 1073741824
  format   = "qcow2"

  backing_store = {
    path   = libvirt_volume.base.path
    format = "qcow2"
  }
}
`, name, poolPath)
}

func TestAccVolumeResource_withDomain(t *testing.T) {
	poolPath := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceConfigWithDomain("test-integration", poolPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_pool.test", "name", "test-pool-integration"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "name", "test-integration.qcow2"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-integration"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.disks.0.target.dev", "vda"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.disks.0.source.pool", "test-pool-integration"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.disks.0.source.volume", "test-integration.qcow2"),
				),
			},
		},
	})
}

func testAccVolumeResourceConfigWithDomain(name, poolPath string) string {
	return fmt.Sprintf(`

resource "libvirt_pool" "test" {
  name = "test-pool-integration"
  type = "dir"
  target = {
    path = %[2]q
  }
}

resource "libvirt_volume" "test" {
  name     = "%[1]s.qcow2"
  pool     = libvirt_pool.test.name
  capacity = 1073741824
  format   = "qcow2"
}

resource "libvirt_domain" "test" {
  name   = "test-domain-integration"
  memory = 512
  vcpu   = 1
  type   = "kvm"

  os = {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }

  devices = {
    disks = [
      {
        source = {
          pool   = libvirt_pool.test.name
          volume = libvirt_volume.test.name
        }
        target = {
          dev = "vda"
          bus = "virtio"
        }
      }
    ]
  }
}
`, name, poolPath)
}

func TestAccVolumeResource_uploadFromFile(t *testing.T) {
	poolPath := t.TempDir()

	// Create a test file to upload
	sourceDir := t.TempDir()
	sourceFilePath := sourceDir + "/source.img"

	// Write test content to the source file
	testContent := make([]byte, 1024*1024) // 1MB test file
	for i := range testContent {
		testContent[i] = byte(i % 256)
	}
	if err := os.WriteFile(sourceFilePath, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceConfigUploadFromFile("test-volume-upload", poolPath, sourceFilePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_volume.test", "name", "test-volume-upload.img"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "pool", "test-pool-upload"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "format", "raw"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "capacity", "1048576"), // 1MB
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "id"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "key"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "path"),
				),
			},
		},
	})
}

func testAccVolumeResourceConfigUploadFromFile(name, poolPath, sourceFile string) string {
	return fmt.Sprintf(`
resource "libvirt_pool" "test" {
  name = "test-pool-upload"
  type = "dir"
  target = {
    path = %[2]q
  }
}

resource "libvirt_volume" "test" {
  name   = "%[1]s.img"
  pool   = libvirt_pool.test.name
  format = "raw"

  create = {
    content = {
      url = %[3]q
    }
  }
}
`, name, poolPath, sourceFile)
}
