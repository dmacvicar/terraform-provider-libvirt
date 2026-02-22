package provider

import (
	"context"
	"testing"

	golibvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDomainDestroyFlagsFromDestroy(t *testing.T) {
	t.Parallel()

	destroyAttrTypes := map[string]attr.Type{
		"graceful": types.BoolType,
	}

	tests := []struct {
		name           string
		destroy        types.Object
		expectGraceful bool
	}{
		{
			name:           "null uses default flags",
			destroy:        types.ObjectNull(destroyAttrTypes),
			expectGraceful: false,
		},
		{
			name: "graceful true sets graceful flag",
			destroy: func() types.Object {
				obj, diags := types.ObjectValue(
					destroyAttrTypes,
					map[string]attr.Value{
						"graceful": types.BoolValue(true),
					},
				)
				if diags.HasError() {
					t.Fatalf("failed to create destroy object: %v", diags)
				}
				return obj
			}(),
			expectGraceful: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			flags, diags := domainDestroyFlagsFromDestroy(context.Background(), tc.destroy)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			gotGraceful := flags&golibvirt.DomainDestroyGraceful != 0
			if gotGraceful != tc.expectGraceful {
				t.Fatalf("unexpected graceful flag: got=%t want=%t (flags=%v)", gotGraceful, tc.expectGraceful, flags)
			}
		})
	}
}
