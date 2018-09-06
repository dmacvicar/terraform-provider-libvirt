package libvirt

import (
	"encoding/xml"
	"fmt"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	libvirt "github.com/libvirt/libvirt-go"
)

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

func testAccCheckLibvirtVolumeExists(name string, volume *libvirt.StorageVol) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No libvirt volume key ID is set")
		}

		retrievedVol, err := virConn.LookupStorageVolByKey(rs.Primary.ID)
		if err != nil {
			return err
		}
		fmt.Printf("The ID is %s", rs.Primary.ID)

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
	return func(s *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt

		resource, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if resource.Primary.ID == "" {
			return fmt.Errorf("No libvirt volume key ID is set")
		}

		vol, err := virConn.LookupStorageVolByKey(resource.Primary.ID)
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
	randomVolumeResource := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_volume" "%s" {
					name = "terraform-test"
					size =  1073741824
				}`, randomVolumeResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeResource, &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeResource, "name", "terraform-test"),
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
	randomVolumeResource := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_volume" "%s" {
					name = "terraform-test3"
					size =  1073741824
				}
				resource "libvirt_volume" "backing-store" {
					name = "backing-store"
					base_volume_id = "${libvirt_volume.%s.id}"
			        }
				`, randomVolumeResource, randomVolumeResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeResource, &volume),
					testAccCheckLibvirtVolumeIsBackingStore("libvirt_volume.backing-store", &volume2),
				),
			},
		},
	})
}

func TestAccLibvirtVolume_BackingStoreTestByName(t *testing.T) {
	var volume libvirt.StorageVol
	var volume2 libvirt.StorageVol
	randomVolumeResource := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "libvirt_volume" "%s" {
						name = "terraform-test3"
						size =  1073741824
					}
					resource "libvirt_volume" "backing-store" {
						name = "backing-store"
						base_volume_name = "${libvirt_volume.%s.name}"
				  }	`, randomVolumeResource, randomVolumeResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeResource, &volume),
					testAccCheckLibvirtVolumeIsBackingStore("libvirt_volume.backing-store", &volume2),
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
	randomVolumeResource := acctest.RandString(10)
	testAccCheckLibvirtVolumeConfigBasic := fmt.Sprintf(`
	resource "libvirt_volume" "%s" {
		name = "terraform-test"
		size =  1073741824
	}`, randomVolumeResource)

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
					removeVolume(client, id)
				},
			},
		},
	})
}

func TestAccLibvirtVolume_UniqueName(t *testing.T) {
	randomVolumeName := acctest.RandString(10)
	randomVolumeName2 := acctest.RandString(10)
	config := fmt.Sprintf(`
	resource "libvirt_volume" "%s" {
		name = "terraform-test"
		size =  1073741824
	}

	resource "libvirt_volume" "%s" {
		name = "terraform-test"
		size =  1073741824
	}
	`, randomVolumeName, randomVolumeName2)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`storage volume 'terraform-test' already exists`),
			},
		},
	})
}

func TestAccLibvirtVolume_DownloadFromSource(t *testing.T) {
	var volume libvirt.StorageVol
	randomVolumeName := acctest.RandString(10)

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
	resource "libvirt_volume" "%s" {
		name   = "terraform-test"
		source = "%s"
	}`, randomVolumeName, url)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeName, &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeName, "name", "terraform-test"),
				),
			},
		},
	})
}

func TestAccLibvirtVolume_DownloadFromSourceFormat(t *testing.T) {
	var volumeRaw libvirt.StorageVol
	var volumeQCOW2 libvirt.StorageVol
	randomVolumeNameRaw := acctest.RandString(10)
	randomVolumeNameQCOW := acctest.RandString(10)
	qcow2Path, err := filepath.Abs("testdata/test.qcow2")
	if err != nil {
		t.Fatal(err)
	}

	rawPath, err := filepath.Abs("testdata/initrd.img")
	if err != nil {
		t.Fatal(err)
	}

	config := fmt.Sprintf(`
	resource "libvirt_volume" "%s" {
		name   = "terraform-test-raw"
		source = "%s"
	}
  resource "libvirt_volume" "%s" {
		name   = "terraform-test-qcow2"
		source = "%s"
	}`, randomVolumeNameRaw, fmt.Sprintf("file://%s", rawPath), randomVolumeNameQCOW, fmt.Sprintf("file://%s", qcow2Path))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeNameRaw, &volumeRaw),
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeNameQCOW, &volumeQCOW2),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeNameRaw, "name", "terraform-test-raw"),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeNameRaw, "format", "raw"),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeNameQCOW, "name", "terraform-test-qcow2"),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeNameQCOW, "format", "qcow2"),
				),
			},
		},
	})
}

func TestAccLibvirtVolume_Format(t *testing.T) {
	var volume libvirt.StorageVol
	randomVolumeName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_volume" "%s" {
					name   = "terraform-test"
					format = "raw"
					size   =  1073741824
				}`, randomVolumeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeName, &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeName, "name", "terraform-test"),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeName, "size", "1073741824"),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeName, "format", "raw"),
				),
			},
		},
	})
}

func TestAccLibvirtVolume_Import(t *testing.T) {
	var volume libvirt.StorageVol
	randomVolumeName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(`
					resource "libvirt_volume" "%s" {
							name   = "terraform-test"
							format = "raw"
							size   =  1073741824
					}`, randomVolumeName),
			},
			resource.TestStep{
				ResourceName: "libvirt_volume." + randomVolumeName,
				ImportState:  true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeName, &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeName, "name", "terraform-test"),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomVolumeName, "size", "1073741824"),
				),
			},
		},
	})
}
