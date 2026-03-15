package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestDomainResourceSchemaIncludesHypervisorNamespaceFields(t *testing.T) {
	r := &DomainResource{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	for _, attr := range []string{
		"qemu_commandline",
		"qemu_capabilities",
		"qemu_override",
		"qemu_deprecation",
		"lxc_namespace",
		"bhyve_commandline",
		"vmware_data_center_path",
		"xen_commandline",
	} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Fatalf("expected domain schema to expose %q", attr)
		}
	}
}
