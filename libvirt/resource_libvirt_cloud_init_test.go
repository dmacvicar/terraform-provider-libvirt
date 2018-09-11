package libvirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	libvirt "github.com/libvirt/libvirt-go"
)

func TestAccLibvirtCloudInit_CreateCloudInitDisk(t *testing.T) {
	var volume libvirt.StorageVol
	randomResourceName := acctest.RandString(10)
	randomIsoName := acctest.RandString(10) + ".iso"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "libvirt_cloudinit" "%s" {
								name           = "%s"
								user_data          = <<EOF
														#cloud-config
														# vim: syntax=yaml
															write_files:
															-   encoding: b64
			    												content: CiMgVGhpcyBmaWxlIGNvbnRyb2xzIHRoZSBzdGF0ZSBvZiBTRUxpbnV4...
			    												owner: root:root
			    						    				path: /tmp/cloudinit_disk.test
			    						    				permissions: '0644'
															-   content: |
			        										# cloudinit_disk_test
													 EOF
								meta_data = <<EOF
														instance-id: foo-bar
														EOF
								network_config = <<EOF
														network:
			  											version: 2
			  											ethernets:
			    											eno1:
			      										dhcp4: true
																EOF}`, randomResourceName, randomIsoName),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"libvirt_cloudinit."+randomResourceName, "name", randomIsoName),
					testAccCheckCloudInitVolumeExists("libvirt_cloudinit."+randomResourceName, &volume),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "libvirt_cloudinit" "%s" {
								name           = "%s"
								user_data          = <<EOF
														#cloud-config
														# vim: syntax=yaml
															write_files:
															-   encoding: b64
			    												content: CiMgVGhpcyBmaWxlIGNvbnRyb2xzIHRoZSBzdGF0ZSBvZiBTRUxpbnV4...
			    												owner: root:root
			    						    				path: /tmp/cloudinit_disk.test
			    						    				permissions: '0644'
															-   content: |
			        										# cloudinit_disk_test
													 EOF
								meta_data = <<EOF
														instance-id: foo-bar
														EOF
								network_config = <<EOF
														network:
			  											version: 2
			  											ethernets:
			    											eno1:
			      										dhcp4: true
																EOF}`, randomResourceName, randomIsoName),
			},
		},
	})
}

// The destroy function should always handle the case where the resource might already be destroyed
// (manually, for example). If the resource is already destroyed, this should not return an error.
// This allows Terraform users to manually delete resources without breaking Terraform.
// This test should fail without a proper "Exists" implementation
func TestAccLibvirtCloudInit_ManuallyDestroyed(t *testing.T) {
	var volume libvirt.StorageVol
	randomResourceName := acctest.RandString(10)

	testAccCheckLibvirtCloudInitConfigBasic := fmt.Sprintf(`
    	resource "libvirt_cloudinit" "%s" {
  	  name           = "%s"
			pool           = "default"
			user_data      = "#cloud-config\nssh_authorized_keys: []\n"
		}`, randomResourceName, randomResourceName)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLibvirtCloudInitConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudInitVolumeExists("libvirt_cloudinit."+randomResourceName, &volume),
				),
			},
			{
				Config:  testAccCheckLibvirtCloudInitConfigBasic,
				Destroy: true,
				PreConfig: func() {
					client := testAccProvider.Meta().(*Client)
					id, err := volume.GetKey()
					if err != nil {
						panic(err)
					}
					removeVolume(client, id)
				},
			},
		},
	})
}

func testAccCheckCloudInitVolumeExists(n string, volume *libvirt.StorageVol) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No libvirt volume key ID is set")
		}

		cikey, err := getCloudInitVolumeKeyFromTerraformID(rs.Primary.ID)
		retrievedVol, err := virConn.LookupStorageVolByKey(cikey)
		if err != nil {
			return err
		}
		realID, err := retrievedVol.GetKey()
		if err != nil {
			return err
		}

		if realID != cikey {
			fmt.Printf("realID is: %s \ncloudinit key is %s", realID, cikey)
			return fmt.Errorf("Resource ID and cloudinit volume key does not match")
		}

		*volume = *retrievedVol

		return nil
	}
}
