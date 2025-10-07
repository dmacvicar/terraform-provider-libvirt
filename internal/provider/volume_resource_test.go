package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVolumeResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVolumeResourceConfigBasic("test-volume"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_volume.test", "name", "test-volume.qcow2"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "pool", "test-pool-volume"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "capacity", "1073741824"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "format", "qcow2"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "id"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "key"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "path"),
					resource.TestCheckResourceAttrSet("libvirt_volume.test", "allocation"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccVolumeResourceConfigBasic(name string) string {
	return fmt.Sprintf(`
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_pool" "test" {
  name = "test-pool-volume"
  type = "dir"
  target = {
    path = "/tmp/terraform-provider-libvirt-pool-volume"
  }
}

resource "libvirt_volume" "test" {
  name     = "%[1]s.qcow2"
  pool     = libvirt_pool.test.name
  capacity = 1073741824
  format   = "qcow2"
}
`, name)
}

func TestAccVolumeResource_backingStore(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceConfigBackingStore("test-volume-cow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_volume.base", "name", "test-volume-cow-base.qcow2"),
					resource.TestCheckResourceAttr("libvirt_volume.cow", "name", "test-volume-cow.qcow2"),
					resource.TestCheckResourceAttrSet("libvirt_volume.cow", "backing_store.path"),
				),
			},
		},
	})
}

func testAccVolumeResourceConfigBackingStore(name string) string {
	return fmt.Sprintf(`
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_pool" "test" {
  name = "test-pool-backing"
  type = "dir"
  target = {
    path = "/tmp/terraform-provider-libvirt-pool-backing"
  }
}

resource "libvirt_volume" "base" {
  name     = "%[1]s-base.qcow2"
  pool     = libvirt_pool.test.name
  capacity = 1073741824
  format   = "qcow2"
}

resource "libvirt_volume" "cow" {
  name     = "%[1]s.qcow2"
  pool     = libvirt_pool.test.name
  capacity = 1073741824
  format   = "qcow2"

  backing_store = {
    path   = libvirt_volume.base.path
    format = "qcow2"
  }
}
`, name)
}

func TestAccVolumeResource_withDomain(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceConfigWithDomain("test-integration"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_pool.test", "name", "test-pool-integration"),
					resource.TestCheckResourceAttr("libvirt_volume.test", "name", "test-integration.qcow2"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-integration"),
					resource.TestCheckResourceAttr("libvirt_domain.test", "devices.disks.0.target", "vda"),
					resource.TestCheckResourceAttrSet("libvirt_domain.test", "devices.disks.0.volume_id"),
				),
			},
		},
	})
}

func testAccVolumeResourceConfigWithDomain(name string) string {
	return fmt.Sprintf(`
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_pool" "test" {
  name = "test-pool-integration"
  type = "dir"
  target = {
    path = "/tmp/terraform-provider-libvirt-pool-integration"
  }
}

resource "libvirt_volume" "test" {
  name     = "%[1]s.qcow2"
  pool     = libvirt_pool.test.name
  capacity = 1073741824
  format   = "qcow2"
}

resource "libvirt_domain" "test" {
  name   = "test-domain-integration"
  memory = 512
  vcpu   = 1

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }

  devices = {
    disks = [
      {
        volume_id = libvirt_volume.test.id
        target    = "vda"
        bus       = "virtio"
      }
    ]
  }
}
`, name)
}
