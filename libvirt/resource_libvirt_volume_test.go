package libvirt

import (
	"encoding/xml"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	libvirt "github.com/libvirt/libvirt-go"
)

func testAccCheckLibvirtVolumeExists(name string, volume *libvirt.StorageVol) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt

		rs, err := getResourceFromTerraformState(name, state)
		if err != nil {
			return err
		}

		retrievedVol, err := getVolumeFromTerraformState(name, state, *virConn)
		if err != nil {
			return err
		}

		realID, err := retrievedVol.GetKey()
		if err != nil {
			return err
		}

		if realID != rs.Primary.ID {
			return fmt.Errorf("Resource ID and volume key does not match")
		}

		*volume = *retrievedVol

		return nil
	}
}

func testAccCheckLibvirtVolumeDoesNotExists(n string, volume *libvirt.StorageVol) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt

		key, err := volume.GetKey()
		if err != nil {
			return fmt.Errorf("Can't retrieve volume key: %s", err)
		}

		vol, err := virConn.LookupStorageVolByKey(key)
		if err == nil {
			vol.Free()
			return fmt.Errorf("Volume '%s' still exists", key)
		}

		return nil
	}
}

func testAccCheckLibvirtVolumeIsBackingStore(name string, volume *libvirt.StorageVol) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt

		vol, err := getVolumeFromTerraformState(name, state, *virConn)
		if err != nil {
			return err
		}

		volXMLDesc, err := vol.GetXMLDesc(0)
		if err != nil {
			return fmt.Errorf("Error retrieving libvirt volume XML description: %s", err)
		}

		volumeDef := newDefVolume()
		err = xml.Unmarshal([]byte(volXMLDesc), &volumeDef)
		if err != nil {
			return fmt.Errorf("Error reading libvirt volume XML description: %s", err)
		}
		if volumeDef.BackingStore == nil {
			return fmt.Errorf("FAIL: the volume was supposed to be a backingstore, but it is not")
		}
		value := volumeDef.BackingStore.Path
		if value == "" {
			return fmt.Errorf("FAIL: the volume was supposed to be a backingstore, but it is not")
		}

		return nil
	}
}

func TestAccLibvirtVolume_Basic(t *testing.T) {
	var volume libvirt.StorageVol
	randomVolumeResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomVolumeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolPath := "/tmp/terraform-provider-libvirt-pool-" + randomPoolName
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
                resource "libvirt_pool" "%s" {
                    name = "%s"
                    type = "dir"
                    path = "%s"
                }

				resource "libvirt_volume" "%s" {
					name = "%s"
					size =  1073741824
                    pool = "${libvirt_pool.%s.name}"
				}`, randomPoolName, randomPoolName, randomPoolPath, randomVolumeResource, randomVolumeName, randomPoolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeResource, &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeResource, "name", randomVolumeName),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeResource, "size", "1073741824"),
				),
			},
		},
	})
}

func TestAccLibvirtVolume_BackingStoreTestByID(t *testing.T) {
	var volume libvirt.StorageVol
	var volume2 libvirt.StorageVol
	random := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolPath := "/tmp/terraform-provider-libvirt-pool-" + random
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
                resource "libvirt_pool" "%s" {
                    name = "%s"
                    type = "dir"
                    path = "%s"
                }
				resource "libvirt_volume" "backing-%s" {
					name = "backing-%s"
					size =  1073741824
                    pool = "${libvirt_pool.%s.name}"
				}
				resource "libvirt_volume" "%s" {
					name = "%s"
					base_volume_id = "${libvirt_volume.backing-%s.id}"
                    pool = "${libvirt_pool.%s.name}"
			        }
				`, random, random, randomPoolPath, random, random, random, random, random, random, random),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume.backing-"+random, &volume),
					testAccCheckLibvirtVolumeIsBackingStore("libvirt_volume."+random, &volume2),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+random, "size", "1073741824"),
				),
			},
		},
	})
}

func TestAccLibvirtVolume_BackingStoreTestByName(t *testing.T) {
	var volume libvirt.StorageVol
	var volume2 libvirt.StorageVol
	random := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolPath := "/tmp/terraform-provider-libvirt-pool-" + random
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
                resource "libvirt_pool" "%s" {
                    name = "%s"
                    type = "dir"
                    path = "%s"
                }
				resource "libvirt_volume" "backing-%s" {
					name = "backing-%s"
					size =  1073741824
                    pool = "${libvirt_pool.%s.name}"
				}
				resource "libvirt_volume" "%s" {
					name = "%s"
                    base_volume_name = "${libvirt_volume.backing-%s.name}"
                    pool = "${libvirt_pool.%s.name}"
			        }
				`, random, random, randomPoolPath, random, random, random, random, random, random, random),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume.backing-"+random, &volume),
					testAccCheckLibvirtVolumeIsBackingStore("libvirt_volume."+random, &volume2),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+random, "size", "1073741824"),
				),
			},
		},
	})
}

// The destroy function should always handle the case where the resource might already be destroyed
// (manually, for example). If the resource is already destroyed, this should not return an error.
// This allows Terraform users to manually delete resources without breaking Terraform.
// This test should fail without a proper "Exists" implementation
func TestAccLibvirtVolume_ManuallyDestroyed(t *testing.T) {
	var volume libvirt.StorageVol
	randomVolumeResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomVolumeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolPath := "/tmp/terraform-provider-libvirt-pool-" + randomPoolName
	testAccCheckLibvirtVolumeConfigBasic := fmt.Sprintf(`
    resource "libvirt_pool" "%s" {
        name = "%s"
        type = "dir"
        path = "%s"
    }
	resource "libvirt_volume" "%s" {
		name = "%s"
		size =  1073741824
        pool = "${libvirt_pool.%s.name}"
	}`, randomPoolName, randomPoolName, randomPoolPath, randomVolumeResource, randomVolumeName, randomPoolName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLibvirtVolumeConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeResource, &volume),
				),
			},
			{
				Config:  testAccCheckLibvirtVolumeConfigBasic,
				Destroy: true,
				PreConfig: func() {
					client := testAccProvider.Meta().(*Client)
					id, err := volume.GetKey()
					if err != nil {
						panic(err)
					}
					volumeDelete(client, id)
				},
			},
		},
	})
}

func TestAccLibvirtVolume_RepeatedName(t *testing.T) {
	var volume libvirt.StorageVol
	randomVolumeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomVolumeResource2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomVolumeResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolPath := "/tmp/terraform-provider-libvirt-pool-" + randomPoolName
	config := fmt.Sprintf(`
    resource "libvirt_pool" "%s" {
        name = "%s"
        type = "dir"
        path = "%s"
    }

	resource "libvirt_volume" "%s" {
		name = "%s"
		size =  1073741824
        pool = "${libvirt_pool.%s.name}"
	}

	resource "libvirt_volume" "%s" {
		name = "%s"
		size =  1073741824
        pool = "${libvirt_pool.%s.name}"
	}
	`, randomPoolName, randomPoolName, randomPoolPath,
		randomVolumeResource, randomVolumeName, randomPoolName,
		randomVolumeResource2, randomVolumeName, randomPoolName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeResource, &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeResource, "name", randomVolumeName),
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeResource2, &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeResource2, "name", randomVolumeName),
				),
			},
		},
	})
}

func TestAccLibvirtVolume_DownloadFromSource(t *testing.T) {
	var volume libvirt.StorageVol
	randomVolumeResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomVolumeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolPath := "/tmp/terraform-provider-libvirt-pool-" + randomPoolName

	fws := fileWebServer{}
	if err := fws.Start(); err != nil {
		t.Fatal(err)
	}
	defer fws.Stop()

	content := []byte("a fake image")
	url, _, err := fws.AddContent(content)
	if err != nil {
		t.Fatal(err)
	}

	config := fmt.Sprintf(`
    resource "libvirt_pool" "%s" {
        name = "%s"
        type = "dir"
        path = "%s"
    }

	resource "libvirt_volume" "%s" {
		name   = "%s"
		source = "%s"
        pool = "${libvirt_pool.%s.name}"
	}`, randomPoolName, randomPoolName, randomPoolPath, randomVolumeResource, randomVolumeName, url, randomPoolName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeResource, &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeResource, "name", randomVolumeName),
				),
			},
		},
	})
}

func TestAccLibvirtVolume_DownloadFromSourceFormat(t *testing.T) {
	var volumeRaw libvirt.StorageVol
	var volumeQCOW2 libvirt.StorageVol
	randomVolumeNameRaw := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomVolumeNameQCOW := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomVolumeResourceRaw := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomVolumeResourceQCOW := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolPath := "/tmp/terraform-provider-libvirt-pool-" + randomPoolName
	qcow2Path, err := filepath.Abs("testdata/test.qcow2")
	if err != nil {
		t.Fatal(err)
	}

	rawPath, err := filepath.Abs("testdata/initrd.img")
	if err != nil {
		t.Fatal(err)
	}

	config := fmt.Sprintf(`
    resource "libvirt_pool" "%s" {
        name = "%s"
        type = "dir"
        path = "%s"
    }
	resource "libvirt_volume" "%s" {
		name   = "%s"
		source = "%s"
        pool = "${libvirt_pool.%s.name}"
	}
  resource "libvirt_volume" "%s" {
		name   = "%s"
		source = "%s"
        pool = "${libvirt_pool.%s.name}"
	}`, randomPoolName, randomPoolName, randomPoolPath, randomVolumeResourceRaw, randomVolumeNameRaw, fmt.Sprintf("file://%s", rawPath), randomPoolName, randomVolumeResourceQCOW, randomVolumeNameQCOW, fmt.Sprintf("file://%s", qcow2Path), randomPoolName)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeResourceRaw, &volumeRaw),
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeResourceQCOW, &volumeQCOW2),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeResourceRaw, "name", randomVolumeNameRaw),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeResourceRaw, "format", "raw"),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeResourceQCOW, "name", randomVolumeNameQCOW),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeResourceQCOW, "format", "qcow2"),
				),
			},
		},
	})
}

func TestAccLibvirtVolume_Format(t *testing.T) {
	var volume libvirt.StorageVol
	randomVolumeResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomVolumeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolPath := "/tmp/terraform-provider-libvirt-pool-" + randomPoolName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
                resource "libvirt_pool" "%s" {
                    name = "%s"
                    type = "dir"
                    path = "%s"
                }

				resource "libvirt_volume" "%s" {
					name   = "%s"
					format = "raw"
					size   =  1073741824
                    pool = "${libvirt_pool.%s.name}"
				}`, randomPoolName, randomPoolName, randomPoolPath, randomVolumeResource, randomVolumeName, randomPoolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeResource, &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeResource, "name", randomVolumeName),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeResource, "size", "1073741824"),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeResource, "format", "raw"),
				),
			},
		},
	})
}

func TestAccLibvirtVolume_Import(t *testing.T) {
	var volume libvirt.StorageVol
	randomVolumeResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomVolumeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolPath := "/tmp/terraform-provider-libvirt-pool-" + randomPoolName
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
                    resource "libvirt_pool" "%s" {
                            name = "%s"
                            type = "dir"
                            path = "%s"
                    }

					resource "libvirt_volume" "%s" {
							name   = "%s"
							format = "raw"
							size   =  1073741824
                            pool = "${libvirt_pool.%s.name}"
					}`, randomPoolName, randomPoolName, randomPoolPath, randomVolumeResource, randomVolumeName, randomPoolName),
			},
			{
				ResourceName: "libvirt_volume." + randomVolumeResource,
				ImportState:  true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeResource, &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeResource, "name", randomVolumeName),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeResource, "size", "1073741824"),
				),
			},
		},
	})
}

func testAccCheckLibvirtVolumeDestroy(state *terraform.State) error {
	virConn := testAccProvider.Meta().(*Client).libvirt
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "libvirt_volume" {
			continue
		}
		_, err := virConn.LookupStorageVolByKey(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf(
				"Error waiting for volume (%s) to be destroyed: %s",
				rs.Primary.ID, err)
		}
	}
	return nil
}
