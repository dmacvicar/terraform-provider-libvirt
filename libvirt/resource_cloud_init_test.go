package libvirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	libvirt "github.com/libvirt/libvirt-go"
)

func TestCreateCloudIsoViaPlugin(t *testing.T) {
	var volume libvirt.StorageVol
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_cloudinit" "test" {
					name           = "test.iso"
					local_hostname = "tango1"
					pool           = "default"
					user_data      = "#cloud-config\nssh_authorized_keys: []\n"
				}`),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"libvirt_cloudinit.test", "name", "test.iso"),
					resource.TestCheckResourceAttr(
						"libvirt_cloudinit.test", "local_hostname", "tango1"),
					testAccCheckCloudInitVolumeExists("libvirt_cloudinit.test", &volume),
				),
			},
			// 2nd tests Invalid  userdata
			{
				Config: fmt.Sprintf(`
				resource "libvirt_cloudinit" "test" {
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
