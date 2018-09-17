package libvirt

import (
	"fmt"

	"github.com/hashicorp/terraform/terraform"
)

// This file contain function helpers used for testsuite/testacc

// getResourceFromTerraformState get aresource by name
// from terraform states produced during testacc
// and return the resource
func getResourceFromTerraformState(resourceName string, state *terraform.State) (*terraform.ResourceState, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("Not found: %s", resourceName)
	}

	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("No libvirt resource key ID is set")
	}
	return rs, nil

}
