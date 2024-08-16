package libvirt

import (
	"fmt"
	"testing"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccLibvirtCombustion_Basic(t *testing.T) {
	var volume libvirt.StorageVol
	randomCombustionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolPath := "/tmp/terraform-provider-libvirt-pool-" + randomPoolName
	config := fmt.Sprintf(`
    resource "libvirt_pool" "%s" {
        name = "%s"
        type = "dir"
        path = "%s"
    }

	resource "libvirt_combustion" "combustion" {
		name    = "%s"
		content = "#!/bin/bash
# combustion: network
echo 'root:$6$3aQC9rrDLHiTf1yR$NoKe9tko0kFIpu0rQ2y/FOO' | chpasswd -e
"
        pool    = "${libvirt_pool.%s.name}"
	}
	`, randomPoolName, randomPoolName, randomPoolPath, randomCombustionName, randomPoolName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtCombustionDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIgnitionVolumeExists("libvirt_combustion.combustion", &volume),
					resource.TestCheckResourceAttr(
						"libvirt_combustion.combustion", "name", randomCombustionName),
					resource.TestCheckResourceAttr(
						"libvirt_combustion.combustion", "pool", randomPoolName),
				),
			},
		},
	})
}

func testAccCheckLibvirtCombustionDestroy(s *terraform.State) error {
	virtConn := testAccProvider.Meta().(*Client).libvirt
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "libvirt_combustion" {
			continue
		}
		// Try to find the Ignition Volume
		ignKey, errKey := getIgnitionVolumeKeyFromTerraformID(rs.Primary.ID)
		if errKey != nil {
			return errKey
		}
		_, err := virtConn.StorageVolLookupByKey(ignKey)
		if err == nil {
			return fmt.Errorf(
				"Error waiting for CombustionVolume (%s) to be destroyed: %w",
				ignKey, err)
		}
	}
	return nil
}
