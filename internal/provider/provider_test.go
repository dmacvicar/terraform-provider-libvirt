package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestMain enables test sweepers and other test setup
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"libvirt": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccLibvirtURI returns the libvirt URI to use for acceptance tests.
// It can be configured via the LIBVIRT_TEST_URI environment variable.
// Defaults to qemu:///system if not set.
func testAccLibvirtURI() string {
	uri := os.Getenv("LIBVIRT_TEST_URI")
	if uri == "" {
		uri = "qemu:///system"
	}
	return uri
}

func testAccPreCheck(t *testing.T) {
	// You can add checks here to verify libvirt is available
	// For now, we'll let the provider connection handle this
}
