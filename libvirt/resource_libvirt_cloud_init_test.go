package libvirt

import (
	"context"
	"fmt"
	"log"
	"testing"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccLibvirtCloudInit_CreateCloudInitDiskAndUpdate(t *testing.T) {
	var volume libvirt.StorageVol
	randomResourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolPath := "/tmp/terraform-provider-libvirt-pool-" + randomPoolName
	// this structs are contents values we expect.
	expectedContents := Expected{UserData: "#cloud-config", NetworkConfig: "network:", MetaData: "instance-id: bamboo"}
	expectedContents2 := Expected{UserData: "#cloud-config2", NetworkConfig: "network2:", MetaData: "instance-id: bamboo2"}
	expectedContentsEmpty := Expected{UserData: "#cloud-config2", NetworkConfig: "", MetaData: ""}
	randomIsoName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + ".iso"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "libvirt_pool" "%s" {
								name = "%s"
								type = "dir"
								path = "%s"
                            }
					resource "libvirt_cloudinit_disk" "%s" {
								name           = "%s"
								user_data      = "#cloud-config"
								meta_data      = "instance-id: bamboo"
								network_config = "network:"
                                pool           = "${libvirt_pool.%s.name}"
							}`, randomPoolName, randomPoolName, randomPoolPath, randomResourceName, randomIsoName, randomPoolName),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"libvirt_cloudinit_disk."+randomResourceName, "name", randomIsoName),
					testAccCheckCloudInitVolumeExists("libvirt_cloudinit_disk."+randomResourceName, &volume),
					expectedContents.testAccCheckCloudInitDiskFilesContent("libvirt_cloudinit_disk."+randomResourceName),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "libvirt_pool" "%s" {
								name = "%s"
								type = "dir"
								path = "%s"
                            }
					resource "libvirt_cloudinit_disk" "%s" {
								name           = "%s"
								user_data      = "#cloud-config2"
								meta_data = "instance-id: bamboo2"
								network_config = "network2:"
                                pool           = "${libvirt_pool.%s.name}"
							}`, randomPoolName, randomPoolName, randomPoolPath, randomResourceName, randomIsoName, randomPoolName),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"libvirt_cloudinit_disk."+randomResourceName, "name", randomIsoName),
					testAccCheckCloudInitVolumeExists("libvirt_cloudinit_disk."+randomResourceName, &volume),
					expectedContents2.testAccCheckCloudInitDiskFilesContent("libvirt_cloudinit_disk."+randomResourceName),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "libvirt_pool" "%s" {
								name = "%s"
								type = "dir"
								path = "%s"
                            }
					resource "libvirt_cloudinit_disk" "%s" {
								name           = "%s"
								user_data      = "#cloud-config2"
                                pool           = "${libvirt_pool.%s.name}"
							}`, randomPoolName, randomPoolName, randomPoolPath, randomResourceName, randomIsoName, randomPoolName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"libvirt_cloudinit_disk."+randomResourceName, "name", randomIsoName),
					testAccCheckCloudInitVolumeExists("libvirt_cloudinit_disk."+randomResourceName, &volume),
					expectedContentsEmpty.testAccCheckCloudInitDiskFilesContent("libvirt_cloudinit_disk."+randomResourceName),
				),
			},
			// when we apply 2 times with same conf, we should not have a diff. See bug:
			// https://github.com/dmacvicar/terraform-provider-libvirt/issues/313
			{
				Config: fmt.Sprintf(`
                        resource "libvirt_pool" "%s" {
								    name = "%s"
                                    type = "dir"
                                    path = "%s"
                                }
						resource "libvirt_cloudinit_disk" "%s" {
									name           = "%s"
									user_data      = "#cloud-config4"
                                    pool           = "${libvirt_pool.%s.name}"
								}`, randomPoolName, randomPoolName, randomPoolPath, randomResourceName, randomIsoName, randomPoolName),
				ExpectNonEmptyPlan: true,
				PlanOnly:           true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"libvirt_cloudinit_disk."+randomResourceName, "name", randomIsoName),
					testAccCheckCloudInitVolumeExists("libvirt_cloudinit_disk."+randomResourceName, &volume),
					expectedContentsEmpty.testAccCheckCloudInitDiskFilesContent("libvirt_cloudinit_disk."+randomResourceName),
				),
			},
		},
	})
}

// The destroy function should always handle the case where the resource might already be destroyed
// (manually, for example). If the resource is already destroyed, this should not return an error.
// This allows Terraform users to manually delete resources without breaking Terraform.
// This test should fail without a proper "Exists" implementation.
func TestAccLibvirtCloudInit_ManuallyDestroyed(t *testing.T) {
	var volume libvirt.StorageVol
	randomResourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolPath := "/tmp/terraform-provider-libvirt-pool-" + randomPoolName

	testAccCheckLibvirtCloudInitConfigBasic := fmt.Sprintf(`
        resource "libvirt_pool" "%s" {
            name = "%s"
            type = "dir"
            path = "%s"
        }
        resource "libvirt_cloudinit_disk" "%s" {
            name           = "%s"
            pool           = "${libvirt_pool.%s.name}"
			user_data      = "#cloud-config\nssh_authorized_keys: []\n"
		}`, randomPoolName, randomPoolName, randomPoolPath, randomResourceName, randomResourceName, randomPoolName)

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
					if volume.Key == "" {
						t.Fatalf("Key is blank")
					}
					if err := volumeDelete(context.Background(), client, volume.Key); err != nil {
						t.Fatal(err)
					}
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

		retrievedVol, err := virConn.StorageVolLookupByKey(cikey)
		if err != nil {
			return err
		}

		if retrievedVol.Key == "" {
			return fmt.Errorf("UUID is blank")
		}

		if retrievedVol.Key != cikey {
			log.Printf("[DEBUG]: retrievedVol.Key is: %s \ncloudinit key is %s", retrievedVol.Key, cikey)
			return fmt.Errorf("Resource ID and cloudinit volume key does not match")
		}

		*volume = retrievedVol

		return nil
	}
}

// this is helper method for test expected values.
type Expected struct {
	UserData, NetworkConfig, MetaData string
}

func (expected *Expected) testAccCheckCloudInitDiskFilesContent(volumeName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt

		rs, err := getResourceFromTerraformState(volumeName, state)
		if err != nil {
			return err
		}

		cloudInitDiskDef, err := newCloudInitDefFromRemoteISO(context.Background(), virConn, rs.Primary.ID)
		if err != nil {
			return err
		}

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
