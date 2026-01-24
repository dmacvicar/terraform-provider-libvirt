package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestOctalModePlanModifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		plan      types.String
		state     types.String
		wantValue types.String
	}{
		{
			name:      "equivalent_with_leading_zero",
			plan:      types.StringValue("770"),
			state:     types.StringValue("0770"),
			wantValue: types.StringValue("0770"),
		},
		{
			name:      "already_equal",
			plan:      types.StringValue("0770"),
			state:     types.StringValue("0770"),
			wantValue: types.StringValue("0770"),
		},
		{
			name:      "different_modes",
			plan:      types.StringValue("750"),
			state:     types.StringValue("0770"),
			wantValue: types.StringValue("750"),
		},
		{
			name:      "plan_unknown",
			plan:      types.StringUnknown(),
			state:     types.StringValue("0770"),
			wantValue: types.StringUnknown(),
		},
		{
			name:      "plan_null",
			plan:      types.StringNull(),
			state:     types.StringValue("0770"),
			wantValue: types.StringNull(),
		},
		{
			name:      "state_null",
			plan:      types.StringValue("770"),
			state:     types.StringNull(),
			wantValue: types.StringValue("770"),
		},
	}

	mod := OctalModePlanModifier()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := planmodifier.StringRequest{
				PlanValue:  tt.plan,
				StateValue: tt.state,
			}
			resp := &planmodifier.StringResponse{
				PlanValue: tt.plan,
			}

			mod.PlanModifyString(context.Background(), req, resp)

			if !resp.PlanValue.Equal(tt.wantValue) {
				t.Fatalf("unexpected plan value: got %v, want %v", resp.PlanValue, tt.wantValue)
			}
		})
	}
}
