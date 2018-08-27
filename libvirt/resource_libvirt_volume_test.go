package libvirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	libvirt "github.com/libvirt/libvirt-go"
)

func testAccCheckLibvirtVolumeDestroy(s *terraform.State) error {
	virConn := testAccProvider.Meta().(*Client).libvirt

	for _, rs := range s.RootModule().Resources {
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

func testAccCheckLibvirtVolumeExists(n string, volume *libvirt.StorageVol) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
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

func TestAccLibvirtVolume_Basic(t *testing.T) {
	var volume libvirt.StorageVol
	randomResourceName := acctest.RandString(10)
	randomVolumeName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_volume" "%s" {
					name = "%s"
					size =  1073741824
				}`, randomResourceName, randomVolumeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume"+randomResourceName, &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomResourceName, "name", randomVolumeName),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomResourceName, "size", "1073741824"),
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
	randomResourceName := acctest.RandString(10)
	randomVolumeName := acctest.RandString(10)
	testAccCheckLibvirtVolumeConfigBasic := fmt.Sprintf(`
	resource "libvirt_volume" "%s" {
		name = "%s"
		size =  1073741824
	}`, randomResourceName, randomVolumeName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLibvirtVolumeConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomResourceName, &volume),
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
	randomResourceName := acctest.RandString(10)
	randomVolumeName := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "libvirt_volume" "%s" {
						name = "%s"
						size =  1073741824
					}

					resource "libvirt_volume" "%s-2" {
						name = "%s"
						size =  1073741824
					}
					`, randomResourceName, randomVolumeName, randomResourceName, randomVolumeName),

				ExpectError: regexp.MustCompile("storage volume '" + randomVolumeName + "' already exists"),
			},
		},
	})
}

func TestAccLibvirtVolume_DownloadFromSource(t *testing.T) {
	var volume libvirt.StorageVol
	randomResourceName := acctest.RandString(10)
	randomVolumeName := acctest.RandString(10)

	fws := fileWebServer{}
	if err := fws.Start(); err != nil {
		t.Fatal(err)
	}
	defer fws.Stop()

	content := []byte("this is a qcow image... well, it is not")
	url, _, err := fws.AddFile(content)
	if err != nil {
		t.Fatal(err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_volume" "%s" {
					name   = "%s"
					source = "%s"
				}`, randomResourceName, randomVolumeName, url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeName, &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomResourceName, "name", randomResourceName),
				),
			},
		},
	})
}

func TestAccLibvirtVolume_Format(t *testing.T) {
	var volume libvirt.StorageVol
	randomResourceName := acctest.RandString(10)
	randomVolumeName := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_volume" "%s" {
					name   = "%s"
					format = "raw"
					size   =  1073741824
				}`, randomResourceName, randomVolumeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomResourceName, &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomResourceName, "name", randomVolumeName),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomResourceName, "size", "1073741824"),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomResourceName, "format", "raw"),
				),
			},
		},
	})
}

func TestAccLibvirtVolume_Import(t *testing.T) {
	var volume libvirt.StorageVol
	randomResourceName := acctest.RandString(10)
	randomVolumeName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(`
					resource "libvirt_volume" "%s" {
							name   = "%s"
							format = "raw"
							size   =  1073741824
					}`, randomResourceName, randomVolumeName),
			},
			resource.TestStep{
				ResourceName: randomResourceName,
				ImportState:  true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomResourceName, &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomResourceName, "name", randomVolumeName),
					resource.TestCheckResourceAttr(
						"libvirt_volume."+randomResourceName, "size", "1073741824"),
				),
			},
		},
	})
}
