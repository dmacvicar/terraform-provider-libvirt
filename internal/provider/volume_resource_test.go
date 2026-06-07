package provider

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	libvirtclient "github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
					resource.TestCheckResourceAttr("libvirt_volume.test", "target.format.type", "qcow2"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "id"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "key"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "target.path"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "allocation"),
				),
			},
			{
				Config:   testAccVolumeResourceConfigBasic("test-volume", poolPath),
				PlanOnly: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccVolumeResource_explicitSparseAllocation(t *testing.T) {
	poolPath := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceConfigAllocation("test-volume-sparse", poolPath, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_volume.test", "capacity", "1048576"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "allocation", "0"),
				),
			},
			{
				Config:   testAccVolumeResourceConfigAllocation("test-volume-sparse", poolPath, 0),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVolumeResource_explicitFullAllocation(t *testing.T) {
	poolPath := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceConfigAllocation("test-volume-full", poolPath, 1024*1024),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_volume.test", "capacity", "1048576"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "allocation", "1048576"),
				),
			},
			{
				Config:   testAccVolumeResourceConfigAllocation("test-volume-full", poolPath, 1024*1024),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVolumeResource_importByKey(t *testing.T) {
	poolPath := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceConfigBasic("test-volume-import", poolPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "key"),
				),
			},
			{
				ResourceName: "libvirt_volume.test",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					volume, ok := state.RootModule().Resources["libvirt_volume.test"]
					if !ok {
						return "", fmt.Errorf("libvirt_volume.test not found in state")
					}

					key := volume.Primary.Attributes["key"]
					if key == "" {
						return "", fmt.Errorf("libvirt_volume.test key not found in state")
					}

					return key, nil
				},
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allocation_unit",
					"capacity_unit",
					"physical_unit",
					"target.permissions",
					"target.timestamps",
					"type",
				},
			},
		},
	})
}

func TestAccVolumeResource_permissionsMode(t *testing.T) {
	poolPath := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceConfigPermissionsMode("test-volume-perms", poolPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_volume.test", "name", "test-volume-perms.qcow2"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "pool", "test-pool-volume-perms"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "target.path"),
				),
			},
			{
				Config:   testAccVolumeResourceConfigPermissionsMode("test-volume-perms", poolPath),
				PlanOnly: true,
			},
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
  target = {
    format = {
      type = "qcow2"
    }
  }
}
`, name, poolPath)
}

func testAccVolumeResourceConfigPermissionsMode(name, poolPath string) string {
	return fmt.Sprintf(`

resource "libvirt_pool" "test" {
  name = "test-pool-volume-perms"
  type = "dir"
  target = {
    path = %[2]q
  }
}

resource "libvirt_volume" "test" {
  name     = "%[1]s.qcow2"
  pool     = libvirt_pool.test.name
  capacity = 1073741824
  target = {
    permissions = {
      mode = "770"
    }
    format = {
      type = "qcow2"
    }
  }
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
  target = {
    format = {
      type = "qcow2"
    }
  }
}

resource "libvirt_volume" "cow" {
  name     = "%[1]s.qcow2"
  pool     = libvirt_pool.test.name
  capacity = 1073741824
  target = {
    format = {
      type = "qcow2"
    }
  }

  backing_store = {
    path = libvirt_volume.base.path
    format = {
      type = "qcow2"
    }
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
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.disks.0.source.volume.pool", "test-pool-integration"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.disks.0.source.volume.volume", "test-integration.qcow2"),
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
  target = {
    format = {
      type = "qcow2"
    }
  }
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
          volume = {
            pool   = libvirt_pool.test.name
            volume = libvirt_volume.test.name
          }
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

// TestAccVolumeResource_capacityUnit reproduces issue #1253: using capacity_unit
// (e.g. "GiB") causes "Provider produced inconsistent result after apply" because
// libvirt normalises the value to bytes on readback and the provider was always
// reading the raw bytes value instead of preserving the plan's unit-converted value.
func TestAccVolumeResource_capacityUnit(t *testing.T) {
	poolPath := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create a 1 GiB volume using capacity_unit.
				// Before the fix this errors: "Provider produced inconsistent result
				// after apply: .capacity: was cty.NumberIntVal(1), but now
				// cty.NumberIntVal(1073741824)".
				Config: testAccVolumeResourceConfigCapacityUnit("test-volume-cap-unit", poolPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_volume.test", "capacity", "1"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "capacity_unit", "GiB"),
				),
			},
			{
				// Ensure there is no perpetual diff after the first apply.
				Config:   testAccVolumeResourceConfigCapacityUnit("test-volume-cap-unit", poolPath),
				PlanOnly: true,
			},
		},
	})
}

func testAccVolumeResourceConfigCapacityUnit(name, poolPath string) string {
	return fmt.Sprintf(`
resource "libvirt_pool" "test" {
  name = "test-pool-cap-unit"
  type = "dir"
  target = {
    path = %[2]q
  }
}

resource "libvirt_volume" "test" {
  name          = "%[1]s.qcow2"
  pool          = libvirt_pool.test.name
  capacity      = 1
  capacity_unit = "GiB"
  target = {
    format = {
      type = "qcow2"
    }
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
					resource.TestCheckResourceAttr("libvirt_volume.test", "target.format.type", "raw"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "capacity", "1048576"), // 1MB
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "id"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "key"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "target.path"),
				),
			},
		},
	})
}

func TestAccVolumeResource_createContentChangeRequiresReplace(t *testing.T) {
	poolPath := t.TempDir()

	sourceDir := t.TempDir()
	sourceFilePath1 := sourceDir + "/source1.img"
	sourceFilePath2 := sourceDir + "/source2.img"

	if err := os.WriteFile(sourceFilePath1, bytes.Repeat([]byte("a"), 1024*1024), 0644); err != nil {
		t.Fatalf("Failed to create source1 file: %v", err)
	}
	if err := os.WriteFile(sourceFilePath2, bytes.Repeat([]byte("b"), 1024*1024), 0644); err != nil {
		t.Fatalf("Failed to create source2 file: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceConfigUploadFromFile("test-volume-replace", poolPath, sourceFilePath1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_volume.test", "name", "test-volume-replace.img"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "id"),
				),
			},
			{
				Config: testAccVolumeResourceConfigUploadFromFile("test-volume-replace", poolPath, sourceFilePath2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_volume.test", "name", "test-volume-replace.img"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "id"),
				),
			},
		},
	})
}

func TestAccVolumeResource_uploadFromHTTPWithContentLength(t *testing.T) {
	poolPath := t.TempDir()
	testContent := bytes.Repeat([]byte("a"), 1024*1024)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/image.raw" {
			http.NotFound(w, r)
			return
		}

		http.ServeContent(w, r, "image.raw", time.Time{}, bytes.NewReader(testContent))
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceConfigUploadFromURL("test-volume-upload-http", poolPath, server.URL+"/image.raw", nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_volume.test", "name", "test-volume-upload-http.img"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "pool", "test-pool-upload-url"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "target.format.type", "raw"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "capacity", "1048576"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "id"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "key"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "target.path"),
				),
			},
		},
	})
}

func TestAccVolumeResource_uploadIntoLargerCapacity(t *testing.T) {
	poolPath := t.TempDir()
	testContent := bytes.Repeat([]byte("d"), 1024*1024)
	requestedCapacity := int64(2 * 1024 * 1024)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/image.raw" {
			http.NotFound(w, r)
			return
		}

		http.ServeContent(w, r, "image.raw", time.Time{}, bytes.NewReader(testContent))
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceConfigUploadFromURL(
					"test-volume-upload-larger",
					poolPath,
					server.URL+"/image.raw",
					&requestedCapacity,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_volume.test", "capacity", "2097152"),
					testAccCheckVolumeFileSize("libvirt_volume.test", requestedCapacity),
				),
			},
		},
	})
}

func TestAccVolumeResource_uploadFromHTTPWithoutContentLengthUsesCapacity(t *testing.T) {
	poolPath := t.TempDir()
	testContent := bytes.Repeat([]byte("b"), 1024*1024)
	expectedCapacity := int64(len(testContent))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/image.raw" {
			http.NotFound(w, r)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer does not support flushing")
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)

		for offset := 0; offset < len(testContent); offset += 4096 {
			end := offset + 4096
			if end > len(testContent) {
				end = len(testContent)
			}
			if _, err := w.Write(testContent[offset:end]); err != nil {
				return
			}
			flusher.Flush()
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceConfigUploadFromURL("test-volume-upload-http-no-length", poolPath, server.URL+"/image.raw", &expectedCapacity),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_volume.test", "name", "test-volume-upload-http-no-length.img"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "pool", "test-pool-upload-url"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "target.format.type", "raw"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "capacity", "1048576"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "id"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "key"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "target.path"),
				),
			},
		},
	})
}

func TestAccVolumeResource_uploadFromHTTPWithoutContentLengthRequiresCapacity(t *testing.T) {
	poolPath := t.TempDir()
	testContent := bytes.Repeat([]byte("c"), 1024*1024)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/image.raw" {
			http.NotFound(w, r)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer does not support flushing")
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)

		for offset := 0; offset < len(testContent); offset += 4096 {
			end := offset + 4096
			if end > len(testContent) {
				end = len(testContent)
			}
			if _, err := w.Write(testContent[offset:end]); err != nil {
				return
			}
			flusher.Flush()
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccVolumeResourceConfigUploadFromURL("test-volume-upload-http-no-length-error", poolPath, server.URL+"/image.raw", nil),
				ExpectError: regexp.MustCompile(`(?s)Missing Capacity.*upload source does not provide\s+Content-Length`),
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
  target = {
    format = {
      type = "raw"
    }
  }

  create = {
    content = {
      url = %[3]q
    }
  }
}
`, name, poolPath, sourceFile)
}

func testAccVolumeResourceConfigAllocation(name, poolPath string, allocation int64) string {
	return fmt.Sprintf(`
resource "libvirt_pool" "test" {
  name = "test-pool-allocation-%[1]s"
  type = "dir"
  target = {
    path = %[2]q
  }
}

resource "libvirt_volume" "test" {
  name       = "%[1]s.raw"
  pool       = libvirt_pool.test.name
  capacity   = 1048576
  allocation = %[3]d
  target = {
    format = {
      type = "raw"
    }
  }
}
`, name, poolPath, allocation)
}

func testAccCheckVolumeFileSize(resourceName string, expected int64) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		volume, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("%s not found in state", resourceName)
		}

		path := volume.Primary.Attributes["path"]
		if path == "" {
			return fmt.Errorf("%s path not found in state", resourceName)
		}

		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("stat volume %s: %w", path, err)
		}
		if info.Size() != expected {
			return fmt.Errorf("expected volume file size %d, got %d", expected, info.Size())
		}
		return nil
	}
}

func testAccVolumeResourceConfigUploadFromURL(name, poolPath, sourceURL string, capacity *int64) string {
	capacityConfig := ""
	if capacity != nil {
		capacityConfig = fmt.Sprintf("  capacity = %d\n", *capacity)
	}

	return fmt.Sprintf(`
resource "libvirt_pool" "test" {
  name = "test-pool-upload-url"
  type = "dir"
  target = {
    path = %[2]q
  }
}

resource "libvirt_volume" "test" {
  name = "%[1]s.img"
  pool = libvirt_pool.test.name
%[4]s  target = {
    format = {
      type = "raw"
    }
  }

  create = {
    content = {
      url = %[3]q
    }
  }
}
`, name, poolPath, sourceURL, capacityConfig)
}
