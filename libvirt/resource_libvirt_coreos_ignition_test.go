package libvirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	libvirt "github.com/libvirt/libvirt-go"
)

func TestAccLibvirtIgnition_Basic(t *testing.T) {
	var volume libvirt.StorageVol
	var config = fmt.Sprintf(`
	    data "ignition_systemd_unit" "acceptance-test-systemd" {
    		name = "example.service"
    		content = "[Service]\nType=oneshot\nExecStart=/usr/bin/echo Hello World\n\n[Install]\nWantedBy=multi-user.target"
	    }

	    data "ignition_config" "acceptance-test-config" {
    		systemd = [
        		"${data.ignition_systemd_unit.acceptance-test-systemd.id}",
    		]
	    }

	    resource "libvirt_ignition" "ignition" {
	    	name = "ignition"
	    	content = "${data.ignition_config.acceptance-test-config.rendered}"
	    }
	`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtIgnitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIgnitionVolumeExists("libvirt_ignition.ignition", &volume),
					resource.TestCheckResourceAttr(
						"libvirt_ignition.ignition", "name", "ignition"),
					resource.TestCheckResourceAttr(
						"libvirt_ignition.ignition", "pool", "default"),
				),
			},
		},
	})
}

func testAccCheckIgnitionVolumeExists(n string, volume *libvirt.StorageVol) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No libvirt ignition key ID is set")
		}

		ignKey, err := getIgnitionVolumeKeyFromTerraformID(rs.Primary.ID)
		if err != nil {
			return err
		}

		retrievedVol, err := virConn.LookupStorageVolByKey(ignKey)
		if err != nil {
			return err
		}
		fmt.Printf("The ID is %s", rs.Primary.ID)

		realId, err := retrievedVol.GetKey()
		if err != nil {
			return err
		}

		if realId != ignKey {
			return fmt.Errorf("Resource ID and volume key does not match")
		}

		*volume = *retrievedVol
		return nil
	}
}

func testAccCheckLibvirtIgnitionDestroy(s *terraform.State) error {
	virtConn := testAccProvider.Meta().(*Client).libvirt

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "libvirt_ignition" {
			continue
		}

		// Try to find the Ignition Volume

		ignKey, errKey := getIgnitionVolumeKeyFromTerraformID(rs.Primary.ID)
		if errKey != nil {
			return errKey
		}

		_, err := virtConn.LookupStorageVolByKey(ignKey)
		if err == nil {
			return fmt.Errorf(
				"Error waiting for IgnitionVolume (%s) to be destroyed: %s",
				ignKey, err)
		}
	}

	return nil
}
