package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDomainUpdateOptionsFromUpdate(t *testing.T) {
	t.Parallel()

	updateAttrTypes := map[string]attr.Type{
		"shutdown": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"timeout": types.Int64Type,
			},
		},
	}

	tests := []struct {
		name          string
		update        types.Object
		wantTimeout   time.Duration
		wantForceStop bool
	}{
		{
			name:          "null uses default timeout and force stop",
			update:        types.ObjectNull(updateAttrTypes),
			wantTimeout:   30 * time.Second,
			wantForceStop: true,
		},
		{
			name: "configured shutdown timeout overrides default",
			update: func() types.Object {
				obj, diags := types.ObjectValue(
					updateAttrTypes,
					map[string]attr.Value{
						"shutdown": types.ObjectValueMust(
							map[string]attr.Type{
								"timeout": types.Int64Type,
							},
							map[string]attr.Value{
								"timeout": types.Int64Value(60),
							},
						),
					},
				)
				if diags.HasError() {
					t.Fatalf("failed to create update object: %v", diags)
				}
				return obj
			}(),
			wantTimeout:   60 * time.Second,
			wantForceStop: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options, diags := domainUpdateOptionsFromUpdate(context.Background(), tc.update)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			if options.ShutdownTimeout != tc.wantTimeout {
				t.Fatalf("unexpected timeout: got=%s want=%s", options.ShutdownTimeout, tc.wantTimeout)
			}

			if options.ForceOnTimeout != tc.wantForceStop {
				t.Fatalf("unexpected force-on-timeout: got=%t want=%t", options.ForceOnTimeout, tc.wantForceStop)
			}
		})
	}
}

func TestDomainDestroyOptionsFromDestroy(t *testing.T) {
	t.Parallel()

	destroyAttrTypes := map[string]attr.Type{
		"graceful": types.BoolType,
		"shutdown": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"timeout": types.Int64Type,
			},
		},
	}

	destroy, diags := types.ObjectValue(
		destroyAttrTypes,
		map[string]attr.Value{
			"graceful": types.BoolValue(true),
			"shutdown": types.ObjectValueMust(
				map[string]attr.Type{
					"timeout": types.Int64Type,
				},
				map[string]attr.Value{
					"timeout": types.Int64Value(45),
				},
			),
		},
	)
	if diags.HasError() {
		t.Fatalf("failed to create destroy object: %v", diags)
	}

	options, diags := domainDestroyOptionsFromDestroy(context.Background(), destroy)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !options.ShutdownEnabled {
		t.Fatal("expected shutdown to be enabled")
	}

	if options.ShutdownTimeout != 45*time.Second {
		t.Fatalf("unexpected timeout: got=%s want=%s", options.ShutdownTimeout, 45*time.Second)
	}

	if options.ForceOnTimeout {
		t.Fatal("expected destroy shutdown timeout to fail instead of force stopping")
	}
}
