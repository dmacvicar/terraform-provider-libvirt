package libvirt

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	libvirt "github.com/libvirt/libvirt-go"
	//	"github.com/libvirt/libvirt-go-xml"
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
		// if volume is a backing store it has some xml attributes set
		// if there is no backingstore this will panic
		value := volumeDef.BackingStore.Path
		if value == "" {
			return fmt.Errorf("FAIL: the volume was supposed to be a backingstore, but it is not")
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

func TestAccLibvirtVolume_BackingStore(t *testing.T) {
	var volume libvirt.StorageVol
	var volume2 libvirt.StorageVol
	const testAccCheckLibvirtVolumeConfigBasic = `
	resource "libvirt_volume" "terraform-acceptance-test-3" {
		name = "terraform-test3"
		size =  1073741824
	}
	resource "libvirt_volume" "backing-store" {
		name = "backing-store"
		base_volume_id = "${libvirt_volume.terraform-acceptance-test-3.id}"
        }
	`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLibvirtVolumeConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume.terraform-acceptance-test-3", &volume),
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
	const config = `
	resource "libvirt_volume" "terraform-acceptance-test-1" {
		name = "terraform-test"
		size =  1073741824
	}

	resource "libvirt_volume" "terraform-acceptance-test-2" {
		name = "terraform-test"
		size =  1073741824
	}
	`

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
				ImportState:  true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume.terraform-acceptance-test-4", &volume),
					resource.TestCheckResourceAttr(
						"libvirt_volume.terraform-acceptance-test-4", "name", "terraform-test"),
					resource.TestCheckResourceAttr(
						"libvirt_volume.terraform-acceptance-test-4", "size", "1073741824"),
				),
			},
		},
	})
}
