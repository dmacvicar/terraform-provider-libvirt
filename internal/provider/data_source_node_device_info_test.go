package provider

import (
	"context"
	"fmt"
	"sort"
	"testing"

	golibvirt "github.com/digitalocean/go-libvirt"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNodeDeviceInfoDataSource_pci(t *testing.T) {
	deviceName := testAccPickNodeDeviceNameByCapability(t, "pci")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeDeviceInfoDataSourceConfigPCI(deviceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify basic fields
					resource.TestCheckResourceAttrSet("data.libvirt_node_device_info.test", "id"),
					resource.TestCheckResourceAttrSet("data.libvirt_node_device_info.test", "name"),
					resource.TestCheckResourceAttrSet("data.libvirt_node_device_info.test", "path"),
					// Verify capability type
					resource.TestCheckResourceAttr("data.libvirt_node_device_info.test", "capability.type", "pci"),
					// Verify PCI fields are set
					resource.TestCheckResourceAttrSet("data.libvirt_node_device_info.test", "capability.domain"),
					resource.TestCheckResourceAttrSet("data.libvirt_node_device_info.test", "capability.bus"),
					resource.TestCheckResourceAttrSet("data.libvirt_node_device_info.test", "capability.slot"),
					resource.TestCheckResourceAttrSet("data.libvirt_node_device_info.test", "capability.function"),
				),
			},
		},
	})
}

func testAccPickNodeDeviceNameByCapability(t *testing.T, capability string) string {
	ctx := context.Background()
	client, err := libvirt.NewClient(ctx, testAccLibvirtURI())
	if err != nil {
		t.Fatalf("failed to create libvirt client: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	cap := golibvirt.OptString{capability}
	count, err := client.Libvirt().NodeNumOfDevices(cap, 0)
	if err != nil {
		t.Fatalf("failed to list %s devices: %v", capability, err)
	}
	if count == 0 {
		t.Skipf("no %s devices available in libvirt", capability)
	}

	deviceNames, err := client.Libvirt().NodeListDevices(cap, count, 0)
	if err != nil {
		t.Fatalf("failed to list %s devices: %v", capability, err)
	}
	if len(deviceNames) == 0 {
		t.Skipf("no %s devices available in libvirt", capability)
	}
	sort.Strings(deviceNames)
	return deviceNames[0]
}

func testAccNodeDeviceInfoDataSourceConfigPCI(name string) string {
	return fmt.Sprintf(`
# Then get details about the first PCI device
data "libvirt_node_device_info" "test" {
  name = %q
}
`, name)
}
