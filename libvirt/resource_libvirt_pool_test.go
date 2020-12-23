package libvirt

import (
	"fmt"
	"regexp"
	"testing"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func testAccCheckLibvirtPoolExists(name string, pool *libvirt.StoragePool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt

		rs, err := getResourceFromTerraformState(name, state)
		if err != nil {
			return fmt.Errorf("Failed to get resource: %s", err)
		}

		retrievedPool, err := getPoolFromTerraformState(name, state, virConn)
		if err != nil {
			return fmt.Errorf("Failed to get pool: %s", err)
		}

		realID := retrievedPool.UUID
		if uuidString(realID) == "" {
			return fmt.Errorf("UUID is blank")
		}

		if uuidString(realID) != rs.Primary.ID {
			return fmt.Errorf("Resource ID and pool ID does not match")
		}

		*pool = *retrievedPool

		return nil
	}
}

func testAccCheckLibvirtPoolDoesNotExists(n string, pool *libvirt.StoragePool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt

		id := pool.UUID

		if uuidString(id) == "" {
			return fmt.Errorf("UUID is blank")
		}

		_, err := virConn.StoragePoolLookupByUUID(id)
		if err == nil {
			return fmt.Errorf("Pool '%s' still exists", id)
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
// This test should fail without a proper "Exists" implementation
func TestAccLibvirtPool_ManuallyDestroyed(t *testing.T) {
	var pool libvirt.StoragePool
	randomPoolResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	poolPath := "/tmp/cluster-api-provider-libvirt-pool-" + randomPoolName
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
					client := testAccProvider.Meta().(*Client)
					id := pool.UUID
					if uuidString(id) == "" {
						panic(fmt.Errorf("UUID is blank"))
					}
					deletePool(client, uuidString(id))
				},
			},
		},
	})
}

func TestAccLibvirtPool_UniqueName(t *testing.T) {
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolResource2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	poolPath := "/tmp/cluster-api-provider-libvirt-pool-" + randomPoolName
	poolPath2 := "/tmp/cluster-api-provider-libvirt-pool-" + randomPoolName + "-2"
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
				"Error waiting for pool (%s) to be destroyed: %s",
				rs.Primary.ID, err)
		}
	}
	return nil
}
