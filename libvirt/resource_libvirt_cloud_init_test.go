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

func TestAccLibvirtCloudInit_CreateCloudIsoViaPlugin(t *testing.T) {
	var volume libvirt.StorageVol
	randomResourceName := acctest.RandString(10)
	randomIsoName := acctest.RandString(10) + ".iso"
	randomLocalHostname := acctest.RandString(5) + ".iso"

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
					local_hostname = "%s"
					pool           = "default"
					user_data      = "#cloud-config\nssh_authorized_keys: []\n"
				}`, randomResourceName, randomIsoName, randomLocalHostname),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"libvirt_cloudinit."+randomResourceName, "name", randomIsoName),
					resource.TestCheckResourceAttr(
						"libvirt_cloudinit"+randomResourceName, "local_hostname", randomLocalHostname),
					testAccCheckCloudInitVolumeExists("libvirt_cloudinit."+randomResourceName, &volume),
				),
			},
			// 2nd tests Invalid  userdata
			{
				Config: fmt.Sprintf(`
				resource "libvirt_cloudinit" "testfail" {
					name           = "commoninit2.iso"
					local_hostname = "samba2"
					pool           = "default"
					user_data      = "invalidgino"
				}`),
				ExpectError: regexp.MustCompile("Error merging UserData with UserDataRaw: yaml: unmarshal errors"),
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
	randomIsoName := acctest.RandString(9) + ".iso"
	randomLocalHostname := acctest.RandString(5) + ".iso"

	testAccCheckLibvirtCloudInitConfigBasic := fmt.Sprintf(`
    	resource "libvirt_cloudinit" "%s" {
  	  name           = "%s"
			local_hostname = "%s"
			pool           = "default"
			user_data      = "#cloud-config\nssh_authorized_keys: []\n"
		}`, randomResourceName, randomIsoName, randomLocalHostname)

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
