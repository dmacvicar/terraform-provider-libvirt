package libvirt

import (
	"fmt"
	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccLibvirtVolumeDatasource(t *testing.T) {
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
					size = 1073741824
					pool = "%s"
				}

				data "libvirt_volume" "%s" {
					name = "%s"
					pool = "%s"
				}
				"`, randomPoolName,
					randomPoolName,
					randomPoolPath,
					randomVolumeResource,
					randomVolumeName,
					randomPoolName,
					randomVolumeResource,
					randomVolumeName,
					randomPoolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("data.libvirt_volume."+randomVolumeResource, &volume),
					resource.TestCheckResourceAttr("data.libvirt_volume."+randomVolumeResource, "name", randomVolumeName),
					resource.TestCheckResourceAttr("data.libvirt_volume."+randomVolumeResource, "pool", randomPoolName),
				),
			},
		},
	})
}
