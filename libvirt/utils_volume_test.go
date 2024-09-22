package libvirt

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestAccUtilsVolume_UploaderDownloader(t *testing.T) {
	var volume libvirt.StorageVol
	randomVolumeResource := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomVolumeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	randomPoolPath := t.TempDir()

	imagePath, err := filepath.Abs("testdata/test.qcow2")
	require.NoError(t, err)

	url := fmt.Sprintf("file://%s", imagePath)

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
					func(state *terraform.State) error {
						virConn := testAccProvider.Meta().(*Client).libvirt

						file, err := os.CreateTemp("", "downloader-")
						require.NoError(t, err)

						defer os.Remove(file.Name())
						defer file.Close()

						downloader := newVolumeDownloader(virConn, &volume)
						err = downloader(file)
						require.NoError(t, err)

						_, err = file.Seek(0, 0)
						require.NoError(t, err)

						h := sha256.New()
						_, err = io.Copy(h, file)
						require.NoError(t, err)

						assert.Equal(t, "0f71acdc66da59b04121b939573bec2e5be78a6cdf829b64142cf0a93a7076f5", fmt.Sprintf("%x", h.Sum(nil)))
						return nil
					},
				),
			},
		},
	})

}
