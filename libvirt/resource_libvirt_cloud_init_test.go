package libvirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	libvirt "github.com/libvirt/libvirt-go"
)

func TestAccLibvirtCloudInit_CreateCloudInitDiskAndUpdate(t *testing.T) {
	var volume libvirt.StorageVol
	randomResourceName := acctest.RandString(10)
	// this structs are contents values we expect.
	expectedContents := Expected{UserData: "#cloud-config", NetworkConfig: "network:", MetaData: "instance-id: bamboo"}
	expectedContents2 := Expected{UserData: "#cloud-config2", NetworkConfig: "network2:", MetaData: "instance-id: bamboo2"}
	expectedContentsEmpty := Expected{UserData: "#cloud-config2", NetworkConfig: "", MetaData: ""}
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
					resource "libvirt_cloudinit_disk" "%s" {
								name           = "%s"
								user_data      = "#cloud-config"
								meta_data      = "instance-id: bamboo"
								network_config = "network:"
							}`, randomResourceName, randomIsoName),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"libvirt_cloudinit_disk."+randomResourceName, "name", randomIsoName),
					testAccCheckCloudInitVolumeExists("libvirt_cloudinit_disk."+randomResourceName, &volume),
					expectedContents.testAccCheckCloudInitDiskFilesContent("libvirt_cloudinit_disk."+randomResourceName, &volume),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "libvirt_cloudinit_disk" "%s" {
								name             = "%s"
								data_source_type = "openstack"
								user_data        = "#cloud-config"
								meta_data        = "instance-id: bamboo"
								network_config   = "network:"
							}`, randomResourceName, randomIsoName),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"libvirt_cloudinit_disk."+randomResourceName, "name", randomIsoName),
					testAccCheckCloudInitVolumeExists("libvirt_cloudinit_disk."+randomResourceName, &volume),
					expectedContents.testAccCheckCloudInitDiskFilesContent("libvirt_cloudinit_disk."+randomResourceName, &volume),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "libvirt_cloudinit_disk" "%s" {
								name             = "%s"
								data_source_type = "ec2"
								user_data        = "#cloud-config"
								meta_data        = "instance-id: bamboo"
								network_config   = "network:"
							}`, randomResourceName, randomIsoName),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"libvirt_cloudinit_disk."+randomResourceName, "name", randomIsoName),
					testAccCheckCloudInitVolumeExists("libvirt_cloudinit_disk."+randomResourceName, &volume),
					expectedContents.testAccCheckCloudInitDiskFilesContent("libvirt_cloudinit_disk."+randomResourceName, &volume),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "libvirt_cloudinit_disk" "%s" {
								name           = "%s"
								user_data      = "#cloud-config2"
								meta_data      = "instance-id: bamboo2"
								network_config = "network2:"
							}`, randomResourceName, randomIsoName),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"libvirt_cloudinit_disk."+randomResourceName, "name", randomIsoName),
					testAccCheckCloudInitVolumeExists("libvirt_cloudinit_disk."+randomResourceName, &volume),
					expectedContents2.testAccCheckCloudInitDiskFilesContent("libvirt_cloudinit_disk."+randomResourceName, &volume),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "libvirt_cloudinit_disk" "%s" {
								name           = "%s"
								user_data      = "#cloud-config2"
							}`, randomResourceName, randomIsoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"libvirt_cloudinit_disk."+randomResourceName, "name", randomIsoName),
					testAccCheckCloudInitVolumeExists("libvirt_cloudinit_disk."+randomResourceName, &volume),
					expectedContentsEmpty.testAccCheckCloudInitDiskFilesContent("libvirt_cloudinit_disk."+randomResourceName, &volume),
				),
			},
			// when we apply 2 times with same conf, we should not have a diff. See bug:
			// https://github.com/dmacvicar/terraform-provider-libvirt/issues/313
			{
				Config: fmt.Sprintf(`
						resource "libvirt_cloudinit_disk" "%s" {
									name           = "%s"
									user_data      = "#cloud-config4"
								}`, randomResourceName, randomIsoName),
				ExpectNonEmptyPlan: true,
				PlanOnly:           true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"libvirt_cloudinit_disk."+randomResourceName, "name", randomIsoName),
					testAccCheckCloudInitVolumeExists("libvirt_cloudinit_disk."+randomResourceName, &volume),
					expectedContentsEmpty.testAccCheckCloudInitDiskFilesContent("libvirt_cloudinit_disk."+randomResourceName, &volume),
				),
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
    	resource "libvirt_cloudinit_disk" "%s" {
  	  name           = "%s"
			pool           = "default"
			user_data      = "#cloud-config\nssh_authorized_keys: []\n"
		}`, randomResourceName, randomResourceName)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLibvirtCloudInitConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudInitVolumeExists("libvirt_cloudinit_disk."+randomResourceName, &volume),
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

func testAccCheckCloudInitVolumeExists(volumeName string, volume *libvirt.StorageVol) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt

		rs, err := getResourceFromTerraformState(volumeName, state)
		if err != nil {
			return err
		}
		cikey, err := getCloudInitVolumeKeyFromTerraformID(rs.Primary.ID)
		if err != nil {
			return err
		}
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

// this is helper method for test expected values
type Expected struct {
	UserData, NetworkConfig, MetaData string
}

func (expected *Expected) testAccCheckCloudInitDiskFilesContent(volumeName string, volume *libvirt.StorageVol) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt

		rs, err := getResourceFromTerraformState(volumeName, state)
		if err != nil {
			return err
		}

		cloudInitDiskDef, err := newCloudInitDefFromRemoteISO(virConn, rs.Primary.ID)

		if cloudInitDiskDef.MetaData != expected.MetaData {
			return fmt.Errorf("metadata '%s' content differs from expected Metadata %s", cloudInitDiskDef.MetaData, expected.MetaData)
		}
		if cloudInitDiskDef.UserData != expected.UserData {
			return fmt.Errorf("userdata '%s' content differs from expected UserData  %s", cloudInitDiskDef.UserData, expected.UserData)
		}
		if cloudInitDiskDef.NetworkConfig != expected.NetworkConfig {
			return fmt.Errorf("networkconfig '%s' content differs from expected NetworkConfigData %s", cloudInitDiskDef.NetworkConfig, expected.NetworkConfig)
		}
		return nil
	}
}
