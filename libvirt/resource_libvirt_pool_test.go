package libvirt

import (
	"context"
	"encoding/xml"
	"fmt"
	"regexp"
	"testing"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"

	testhelper "github.com/dmacvicar/terraform-provider-libvirt/libvirt/helper/test"
)

func testAccCheckLibvirtPoolExists(name string, pool *libvirt.StoragePool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt

		rs, err := getResourceFromTerraformState(name, state)
		if err != nil {
			return fmt.Errorf("Failed to get resource: %w", err)
		}

		retrievedPool, err := getPoolFromTerraformState(name, state, virConn)
		if err != nil {
			return fmt.Errorf("Failed to get pool: %w", err)
		}

		if uuidString(retrievedPool.UUID) == "" {
			return fmt.Errorf("UUID is blank")
		}

		if uuidString(retrievedPool.UUID) != rs.Primary.ID {
			return fmt.Errorf("Resource ID and pool ID does not match")
		}

		*pool = *retrievedPool

		return nil
	}
}

func TestAccLibvirtPool_Import(t *testing.T) {
	var pool libvirt.StoragePool
	randomPoolResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	poolPath := "/tmp/terraform-provider-libvirt-pool-" + randomPoolName
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
                    resource "libvirt_pool" "%s" {
                            name = "%s"
                            type = "dir"
                            path = "%s"
					}`, randomPoolResource, randomPoolName, poolPath),
				Check:   testAccCheckLibvirtPoolExists("libvirt_pool."+randomPoolResource, &pool),
				Destroy: false,
			},
			{
				ResourceName: "libvirt_pool." + randomPoolResource,
				ImportState:  true,
				ImportStateCheck: func(instanceState []*terraform.InstanceState) error {
					// check all instance state imported with same assert
					for i, f := range instanceState {
						if err := composeTestImportStateCheckFunc(
							testImportStateCheckResourceAttr("libvirt_pool."+randomPoolResource, "name", randomPoolName),
							testImportStateCheckResourceAttr("libvirt_pool."+randomPoolResource, "type", "dir"),
							testImportStateCheckResourceAttr("libvirt_pool."+randomPoolResource, "target.0.path", poolPath),
						)(f); err != nil {
							return fmt.Errorf("Check InstanceState n°%d / %d error: %w", i+1, len(instanceState), err)
						}
					}

					return nil
				},
			},
		},
	})
}

// ImportStateCheckFunc one import instance state check function
// differ from github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource.ImportStateCheckFunc
// which is multiple import Instance State check function.
type ImportStateCheckFunc func(is *terraform.InstanceState) error

// composeTestImportStateCheckFunc compose multiple InstanceState check.
func composeTestImportStateCheckFunc(fs ...ImportStateCheckFunc) ImportStateCheckFunc {
	return func(is *terraform.InstanceState) error {
		for i, f := range fs {
			if err := f(is); err != nil {
				return fmt.Errorf("Check %d/%d error: %w", i+1, len(fs), err)
			}
		}

		return nil
	}
}

// testImportStateCheckResourceAttr assert if a terraform.InstanceState as attribute name[key] with value.
func testImportStateCheckResourceAttr(name string, key string, value string) ImportStateCheckFunc {
	return func(instanceState *terraform.InstanceState) error {
		if v, ok := instanceState.Attributes[key]; !ok || v != value {
			if !ok {
				return fmt.Errorf("%s: Attribute '%s' not found", name, key)
			}

			return fmt.Errorf(
				"%s: Attribute '%s' expected %#v, got %#v",
				name,
				key,
				value,
				v)
		}
		return nil
	}
}

func TestAccLibvirtPool_DirBasic(t *testing.T) {
	var pool libvirt.StoragePool
	randomPoolResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	poolPath := "/tmp/cluster-api-provider-libvirt-pool-" + randomPoolName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_pool" "%s" {
					name = "%s"
					type = "dir"
                                        target {
                                          path = "%s"
                                        }
				}`, randomPoolResource, randomPoolName, poolPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtPoolExists("libvirt_pool."+randomPoolResource, &pool),
					resource.TestCheckResourceAttr(
						"libvirt_pool."+randomPoolResource, "name", randomPoolName),
					resource.TestCheckResourceAttr(
						"libvirt_pool."+randomPoolResource, "target.0.path", poolPath),
				),
			},
		},
	})
}

func TestAccLibvirtPool_DirBasicDeprecated(t *testing.T) {
	var pool libvirt.StoragePool
	randomPoolResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	poolPath := "/tmp/cluster-api-provider-libvirt-pool-" + randomPoolName
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_pool" "%s" {
					name = "%s"
					type = "dir"
                    path = "%s"
				}`, randomPoolResource, randomPoolName, poolPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtPoolExists("libvirt_pool."+randomPoolResource, &pool),
					resource.TestCheckResourceAttr(
						"libvirt_pool."+randomPoolResource, "name", randomPoolName),
					resource.TestCheckResourceAttr(
						"libvirt_pool."+randomPoolResource, "path", poolPath),
				),
			},
		},
	})
}

func testAccPreCheckSupportsLogicalPool(t *testing.T) {
	type storagePoolCaps struct {
		Pools []struct {
			Type      string `xml:"type,attr"`
			Supported string `xml:"supported,attr"`
		} `xml:"pool"`
		XMLName xml.Name `xml:"storagepoolCapabilities"`
	}

	client := testAccProvider.Meta().(*Client)

	respStr, err := client.libvirt.ConnectGetStoragePoolCapabilities(0)
	if err != nil {
		t.Fatalf("Error getting storage pool capabilities: %s", err)
	}

	var caps storagePoolCaps
	err = xml.Unmarshal([]byte(respStr), &caps)
	if err != nil {
		t.Fatalf("Error unmarshalling storage pool capabilities: %s", err)
	}

	for _, pool := range caps.Pools {
		if pool.Type == "logical" && pool.Supported != "yes" {
			t.Skip("Storage pool capabilities does not support logical pools")
		}
	}
}

func TestAccLibvirtPool_LVMBasic(t *testing.T) {

	var pool libvirt.StoragePool
	randomPoolResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	// we need the plugin configured before we can test for support for lvm pools.
	diag := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if diag.HasError() {
		t.Fatal("error configuring provider")
	}

	testAccPreCheckSupportsLogicalPool(t)

	blockDev, err := testhelper.CreateTempLoopDevice(t, randomPoolName)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := blockDev.Cleanup(); err != nil {
			t.Errorf("error cleaning up loop device %s: %s", blockDev, err)
		}
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_pool" "%s" {
					name = "%s"
					type = "logical"
                                        source {
                                          device {
                                            path = "%s"
                                          }
                                        }
				}`, randomPoolResource, randomPoolName, blockDev.LoopDevice),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtPoolExists("libvirt_pool."+randomPoolResource, &pool),
					resource.TestCheckResourceAttr(
						"libvirt_pool."+randomPoolResource, "name", randomPoolName),
					resource.TestCheckResourceAttr(
						"libvirt_pool."+randomPoolResource, "source.0.device.0.path", blockDev.LoopDevice),
				),
			},
		},
	})
}

// The destroy function should always handle the case where the resource might already be destroyed
// (manually, for example). If the resource is already destroyed, this should not return an error.
// This allows Terraform users to manually delete resources without breaking Terraform.
// This test should fail without a proper "Exists" implementation.
func TestAccLibvirtPool_ManuallyDestroyed(t *testing.T) {
	var pool libvirt.StoragePool
	randomPoolResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	poolPath := t.TempDir()
	testAccCheckLibvirtPoolConfigBasic := fmt.Sprintf(`
	resource "libvirt_pool" "%s" {
					name = "%s"
					type = "dir"
                    path = "%s"
				}`, randomPoolResource, randomPoolName, poolPath)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLibvirtPoolConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtPoolExists("libvirt_pool."+randomPoolResource, &pool),
				),
			},
			{
				Config:  testAccCheckLibvirtPoolConfigBasic,
				Destroy: true,
				PreConfig: func() {
					// delete the pool out of band (from terraform)
					client := testAccProvider.Meta().(*Client)

					err := client.libvirt.StoragePoolDestroy(pool)
					require.NoError(t, err)

					err = client.libvirt.StoragePoolDelete(pool, libvirt.StoragePoolDeleteNormal)
					require.NoError(t, err)

					err = client.libvirt.StoragePoolUndefine(pool)
					require.NoError(t, err)
				},
			},
		},
	})
}

func TestAccLibvirtPool_UniqueName(t *testing.T) {
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolResource2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	poolPath := t.TempDir()
	poolPath2 := t.TempDir()
	config := fmt.Sprintf(`
	resource "libvirt_pool" "%s" {
		name = "%s"
        type = "dir"
        path = "%s"
	}

	resource "libvirt_pool" "%s" {
		name = "%s"
        type = "dir"
        path = "%s"
	}
	`, randomPoolResource, randomPoolName, poolPath, randomPoolResource2, randomPoolName, poolPath2)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`storage pool '` + randomPoolName + `' (exists already|already exists)`),
			},
		},
	})
}

func TestAccLibvirtPool_NoDirPath(t *testing.T) {
	randomPoolResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_pool" "%s" {
					name = "%s"
					type = "dir"
				}`, randomPoolResource, randomPoolName),
				ExpectError: regexp.MustCompile(`missing storage pool target path`),
			},
		},
	})
}

func testAccCheckLibvirtPoolDestroy(state *terraform.State) error {
	virConn := testAccProvider.Meta().(*Client).libvirt
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "libvirt_pool" {
			continue
		}
		_, err := virConn.StoragePoolLookupByUUID(parseUUID(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf(
				"Error waiting for pool (%s) to be destroyed: %w",
				rs.Primary.ID, err)
		}
	}
	return nil
}
