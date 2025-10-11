package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	libvirtclient "github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func init() {
	resource.AddTestSweepers("libvirt_network", &resource.Sweeper{
		Name: "libvirt_network",
		F: func(uri string) error {
			ctx := context.Background()
			client, err := libvirtclient.NewClient(ctx, uri)
			if err != nil {
				return fmt.Errorf("failed to create libvirt client: %w", err)
			}
			defer func() { _ = client.Close() }()

			// List all networks
			networks, _, err := client.Libvirt().ConnectListAllNetworks(1, 0)
			if err != nil {
				return fmt.Errorf("failed to list networks: %w", err)
			}

			// Delete test networks (prefix: test-)
			for _, network := range networks {
				if strings.HasPrefix(network.Name, "test-") || strings.HasPrefix(network.Name, "test_") {
					// Try to destroy if active
					_ = client.Libvirt().NetworkDestroy(network)
					// Undefine the network
					if err := client.Libvirt().NetworkUndefine(network); err != nil {
						fmt.Printf("Warning: failed to undefine network %s: %v\n", network.Name, err)
					}
				}
			}

			return nil
		},
	})
}

func testAccCheckNetworkDestroy(s *terraform.State) error {
	ctx := context.Background()
	client, err := libvirtclient.NewClient(ctx, testAccLibvirtURI())
	if err != nil {
		return fmt.Errorf("failed to create libvirt client: %w", err)
	}
	defer func() { _ = client.Close() }()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "libvirt_network" {
			continue
		}

		uuid := rs.Primary.Attributes["uuid"]
		if uuid == "" {
			continue
		}

		// Try to find the network - it should not exist
		_, err := client.LookupNetworkByUUID(uuid)
		if err == nil {
			return fmt.Errorf("network %s still exists after destroy", uuid)
		}
	}

	return nil
}

func TestAccNetworkResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNetworkDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNetworkResourceConfigBasic("test-network-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network.test", "name", "test-network-basic"),
					resource.TestCheckResourceAttr("libvirt_network.test", "mode", "nat"),
					resource.TestCheckResourceAttr("libvirt_network.test", "addresses.#", "1"),
					resource.TestCheckResourceAttr("libvirt_network.test", "addresses.0", "10.17.3.0/24"),
					resource.TestCheckResourceAttrSet("libvirt_network.test", "uuid"),
					resource.TestCheckResourceAttrSet("libvirt_network.test", "id"),
					resource.TestCheckResourceAttrSet("libvirt_network.test", "bridge"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccNetworkResource_isolated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfigIsolated("test-network-isolated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("libvirt_network.test", "name", "test-network-isolated"),
					resource.TestCheckResourceAttr("libvirt_network.test", "mode", "none"),
					resource.TestCheckResourceAttr("libvirt_network.test", "addresses.#", "1"),
				),
			},
		},
	})
}

func testAccNetworkResourceConfigBasic(name string) string {
	return fmt.Sprintf(`
resource "libvirt_network" "test" {
  name      = %[1]q
  mode      = "nat"
  addresses = ["10.17.3.0/24"]
  autostart = false
}
`, name)
}

func testAccNetworkResourceConfigIsolated(name string) string {
	return fmt.Sprintf(`
resource "libvirt_network" "test" {
  name      = %[1]q
  mode      = "none"
  addresses = ["192.168.100.0/24"]
}
`, name)
}
