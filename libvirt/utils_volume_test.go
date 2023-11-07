package libvirt

import (
	"fmt"
	"os"
	"testing"
	"time"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestTimeFromEpoch(t *testing.T) {
	if ts := timeFromEpoch(""); ts.UnixNano() > 0 {
		t.Fatalf("expected timestamp '0.0', got %v.%v", ts.Unix(), ts.Nanosecond())
	}

	if ts := timeFromEpoch("abc"); ts.UnixNano() > 0 {
		t.Fatalf("expected timestamp '0.0', got %v.%v", ts.Unix(), ts.Nanosecond())
	}

	if ts := timeFromEpoch("123"); ts.UnixNano() != time.Unix(123, 0).UnixNano() {
		t.Fatalf("expected timestamp '123.0', got %v.%v", ts.Unix(), ts.Nanosecond())
	}

	if ts := timeFromEpoch("123.456"); ts.UnixNano() != time.Unix(123, 456).UnixNano() {
		t.Fatalf("expected timestamp '123.456', got %v.%v", ts.Unix(), ts.Nanosecond())
	}
}

func TestAccUtilsVolume_UploadVolumeCopier(t *testing.T) {

	var volume libvirt.StorageVol
	randomVolumeResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomVolumeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolPath := t.TempDir()

	tmpfile, err := os.CreateTemp(t.TempDir(), "test-image-")
	if err != nil {
		t.Fatal(err)
	}

	// simulate uploading a 1G file usig a sparse file
	if err:= tmpfile.Truncate(1024*1024*1024); err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	url := fmt.Sprintf("file://%s", tmpfile.Name())

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
