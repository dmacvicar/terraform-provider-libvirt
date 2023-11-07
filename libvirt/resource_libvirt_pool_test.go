package libvirt

import (
	"fmt"
	"regexp"
	"testing"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
							testImportStateCheckResourceAttr("libvirt_pool."+randomPoolResource, "path", poolPath),
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

func TestAccLibvirtPool_Basic(t *testing.T) {
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

					if err := client.libvirt.StoragePoolDestroy(pool); err != nil {
						t.Errorf(err.Error())
					}

					if err := client.libvirt.StoragePoolDelete(pool, libvirt.StoragePoolDeleteNormal); err != nil {
						t.Errorf(err.Error())
					}

					if err := client.libvirt.StoragePoolUndefine(pool); err != nil {
						t.Errorf(err.Error())
					}
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
				ExpectError: regexp.MustCompile(`"path" attribute is requires for storage pools of type "dir"`),
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

