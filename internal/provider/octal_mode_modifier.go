package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// octalModePlanModifier handles libvirt's octal mode normalization (e.g., 770 -> 0770).
type octalModePlanModifier struct{}

func (m octalModePlanModifier) Description(ctx context.Context) string {
	return "Handles libvirt's octal mode normalization (e.g., 770 -> 0770)"
}

func (m octalModePlanModifier) MarkdownDescription(ctx context.Context) string {
	return "Handles libvirt's octal mode normalization (e.g., `770` -> `0770`)"
}

func (m octalModePlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the state is null (new resource), use the plan value
	if req.StateValue.IsNull() {
		return
	}

	// If the plan is null or unknown, use it as-is
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	planValue := req.PlanValue.ValueString()

	stateValue := req.StateValue.ValueString()
	canonicalState := canonicalOctalMode(stateValue)

	// If plan and state are the same, no change needed
	if planValue == stateValue {
		return
	}

	// Normalize before comparison
	if canonicalOctalMode(planValue) == canonicalState {
		// Keep the state value to avoid unnecessary updates
		resp.PlanValue = types.StringValue(stateValue)
	}
}

func canonicalOctalMode(value string) string {
	trimmed := strings.TrimLeft(value, "0")
	if trimmed == "" {
		return "0"
	}
	return "0" + trimmed
}

// OctalModePlanModifier returns a plan modifier that handles octal mode normalization.
func OctalModePlanModifier() planmodifier.String {
	return octalModePlanModifier{}
}
