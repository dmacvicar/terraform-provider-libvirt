package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// machineTypePlanModifier handles libvirt's machine type expansion
// For example, "q35" becomes "pc-q35-10.1"
type machineTypePlanModifier struct{}

func (m machineTypePlanModifier) Description(ctx context.Context) string {
	return "Handles libvirt's machine type expansion (e.g., q35 -> pc-q35-10.1)"
}

func (m machineTypePlanModifier) MarkdownDescription(ctx context.Context) string {
	return "Handles libvirt's machine type expansion (e.g., `q35` -> `pc-q35-10.1`)"
}

func (m machineTypePlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
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

	// If plan and state are the same, no change needed
	if planValue == stateValue {
		return
	}

	// Check if the state value is an expanded version of the plan value
	// For example, plan="q35" and state="pc-q35-10.1"
	if isExpandedMachineType(planValue, stateValue) {
		// Keep the state value to avoid unnecessary updates
		resp.PlanValue = types.StringValue(stateValue)
	}
}

// isExpandedMachineType checks if stateValue is an expanded version of planValue
func isExpandedMachineType(plan, state string) bool {
	// Common patterns:
	// - "q35" expands to "pc-q35-X.Y"
	// - "pc" expands to "pc-i440fx-X.Y"
	// - Already versioned types should match exactly

	// If state contains the plan value, it's likely an expansion
	if strings.Contains(state, plan) {
		return true
	}

	// Check specific patterns
	switch plan {
	case "q35":
		return strings.HasPrefix(state, "pc-q35-")
	case "pc":
		return strings.HasPrefix(state, "pc-i440fx-") || strings.HasPrefix(state, "pc-")
	}

	return false
}

// MachineTypePlanModifier returns a plan modifier that handles machine type expansion
func MachineTypePlanModifier() planmodifier.String {
	return machineTypePlanModifier{}
}
