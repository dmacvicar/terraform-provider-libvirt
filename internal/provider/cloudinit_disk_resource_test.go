package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCloudInitDiskResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCloudInitDiskResourceConfigBasic("test-cloudinit"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_cloudinit_disk.test", "name", "test-cloudinit"),
					resource.TestCheckResourceAttrSet("libvirt_cloudinit_disk.test", "id"),
					resource.TestCheckResourceAttrSet("libvirt_cloudinit_disk.test", "path"),
					resource.TestCheckResourceAttrSet("libvirt_cloudinit_disk.test", "size"),
					testAccCheckCloudInitDiskExists("libvirt_cloudinit_disk.test"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccCloudInitDiskResource_withNetworkConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudInitDiskResourceConfigWithNetwork("test-cloudinit-net"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_cloudinit_disk.test", "name", "test-cloudinit-net"),
					resource.TestCheckResourceAttrSet("libvirt_cloudinit_disk.test", "network_config"),
					resource.TestCheckResourceAttrSet("libvirt_cloudinit_disk.test", "path"),
					testAccCheckCloudInitDiskExists("libvirt_cloudinit_disk.test"),
				),
			},
		},
	})
}

func TestAccCloudInitDiskResource_withVolume(t *testing.T) {
	poolPath := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudInitDiskResourceConfigWithVolume("test-cloudinit-volume", poolPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_cloudinit_disk.test", "name", "test-cloudinit-volume"),
					resource.TestCheckResourceAttr("libvirt_volume.cloudinit", "name", "test-cloudinit-volume.iso"),
					resource.TestCheckResourceAttrSet("libvirt_volume.cloudinit", "id"),
					testAccCheckCloudInitDiskExists("libvirt_cloudinit_disk.test"),
				),
			},
		},
	})
}

// testAccCheckCloudInitDiskExists verifies that the ISO file exists on disk
func testAccCheckCloudInitDiskExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s", resourceName)
		}

		path := rs.Primary.Attributes["path"]
		if path == "" {
			return fmt.Errorf("no path is set for %s", resourceName)
		}

		// Check if the ISO file exists
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("ISO file does not exist at %s: %w", path, err)
		}

		// Verify it's in the expected temp directory
		expectedDir := filepath.Join(os.TempDir(), "terraform-provider-libvirt-cloudinit")
		if !strings.HasPrefix(path, expectedDir) {
			return fmt.Errorf("ISO file is not in expected directory: got %s, expected prefix %s", path, expectedDir)
		}

		return nil
	}
}

func testAccCloudInitDiskResourceConfigBasic(name string) string {
	return fmt.Sprintf(`
resource "libvirt_cloudinit_disk" "test" {
  name      = %[1]q
  user_data = <<-EOF
    #cloud-config
    users:
      - name: root
        ssh_authorized_keys:
          - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC0/Ho1w+1D4vJccMzEQBzREzCY4NjkrJYh8+9rQJgDrYPrWLe1PJYvDG6r1uDlLrZJhwwq1PcJQw test@example.com
  EOF

  meta_data = <<-EOF
    instance-id: %[1]s-001
    local-hostname: %[1]s
  EOF
}
`, name)
}

func testAccCloudInitDiskResourceConfigWithNetwork(name string) string {
	return fmt.Sprintf(`
resource "libvirt_cloudinit_disk" "test" {
  name      = %[1]q
  user_data = <<-EOF
    #cloud-config
    users:
      - name: root
        ssh_authorized_keys:
          - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC0/Ho1w+1D4vJccMzEQBzREzCY4NjkrJYh8+9rQJgDrYPrWLe1PJYvDG6r1uDlLrZJhwwq1PcJQw test@example.com
  EOF

  meta_data = <<-EOF
    instance-id: %[1]s-001
    local-hostname: %[1]s
  EOF

  network_config = <<-EOF
    version: 2
    ethernets:
      eth0:
        dhcp4: true
  EOF
}
`, name)
}

func testAccCloudInitDiskResourceConfigWithVolume(name, poolPath string) string {
	return fmt.Sprintf(`
resource "libvirt_pool" "test" {
  name = "test-pool-cloudinit"
  type = "dir"
  target = {
    path = %[2]q
  }
}

resource "libvirt_cloudinit_disk" "test" {
  name      = %[1]q
  user_data = <<-EOF
    #cloud-config
    password: password
    chpasswd:
      expire: false
    ssh_pwauth: true
  EOF

  meta_data = <<-EOF
    instance-id: %[1]s-001
    local-hostname: %[1]s
  EOF
}

resource "libvirt_volume" "cloudinit" {
  name   = "%[1]s.iso"
  pool   = libvirt_pool.test.name
  # Let libvirt auto-detect format (it will detect as "iso")

  create = {
    content = {
      url = libvirt_cloudinit_disk.test.path
    }
  }
}
`, name, poolPath)
}
