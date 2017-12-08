package libvirt

import (
	"fmt"
	"testing"

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

	const testAccCheckLibvirtVolumeConfigBasic = `
	resource "libvirt_volume" "terraform-acceptance-test-1" {
		name = "terraform-test"
		size =  1073741824
	}`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLibvirtVolumeConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume.terraform-acceptance-test-1", &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume.terraform-acceptance-test-1", "name", "terraform-test"),
					resource.TestCheckResourceAttr(
						"libvirt_volume.terraform-acceptance-test-1", "size", "1073741824"),
				),
			},
		},
	})
}

func TestAccLibvirtVolume_DownloadFromSource(t *testing.T) {
	var volume libvirt.StorageVol

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

	const testAccCheckLibvirtVolumeConfigSource = `
	resource "libvirt_volume" "terraform-acceptance-test-2" {
		name   = "terraform-test"
		source = "%s"
	}`
	config := fmt.Sprintf(testAccCheckLibvirtVolumeConfigSource, url)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume.terraform-acceptance-test-2", &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume.terraform-acceptance-test-2", "name", "terraform-test"),
				),
			},
		},
	})
}

func TestAccLibvirtVolume_Format(t *testing.T) {
	var volume libvirt.StorageVol

	const testAccCheckLibvirtVolumeConfigFormat = `
	resource "libvirt_volume" "terraform-acceptance-test-3" {
		name   = "terraform-test"
		format = "raw"
		size   =  1073741824
	}`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLibvirtVolumeConfigFormat,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume.terraform-acceptance-test-3", &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume.terraform-acceptance-test-3", "name", "terraform-test"),
					resource.TestCheckResourceAttr(
						"libvirt_volume.terraform-acceptance-test-3", "size", "1073741824"),
					resource.TestCheckResourceAttr(
						"libvirt_volume.terraform-acceptance-test-3", "format", "raw"),
				),
			},
		},
	})
}

func TestAccLibvirtVolume_Import(t *testing.T) {
	var volume libvirt.StorageVol

	const testAccCheckLibvirtVolumeConfigImport = `
	resource "libvirt_volume" "terraform-acceptance-test-4" {
		name   = "terraform-test"
		format = "raw"
		size   =  1073741824
	}`

	resourceName := "libvirt_volume.terraform-acceptance-test-4"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckLibvirtVolumeConfigImport,
			},
			resource.TestStep{
				ResourceName: resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			}
			resource.TestStep{
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume.terraform-acceptance-test-3", &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume.terraform-acceptance-test-3", "name", "terraform-test"),
					resource.TestCheckResourceAttr(
						"libvirt_volume.terraform-acceptance-test-3", "size", "1073741824"),
					resource.TestCheckResourceAttr(
						"libvirt_volume.terraform-acceptance-test-3", "format", "raw"),
				),
			}
		},
	})
}