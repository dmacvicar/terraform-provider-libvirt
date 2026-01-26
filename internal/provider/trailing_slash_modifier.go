package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// trailingSlashPlanModifier handles libvirt trimming trailing slashes from paths.
type trailingSlashPlanModifier struct{}

func (m trailingSlashPlanModifier) Description(ctx context.Context) string {
	return "Handles libvirt trimming trailing slashes from paths"
}

func (m trailingSlashPlanModifier) MarkdownDescription(ctx context.Context) string {
	return "Handles libvirt trimming trailing slashes from paths"
}

func (m trailingSlashPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the plan is null or unknown, use it as-is
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	planValue := req.PlanValue.ValueString()
	stateValue := req.StateValue.ValueString()

	canonicalPlan := canonicalPath(planValue)

	// Normalize on create as well, so state matches libvirt's canonical path.
	if canonicalPlan != planValue {
		resp.PlanValue = types.StringValue(canonicalPlan)
		return
	}

	// If the state is null (new resource), nothing else to compare
	if req.StateValue.IsNull() {
		return
	}

	// If plan and state are the same, no change needed
	if planValue == stateValue {
		return
	}

	if canonicalPlan == canonicalPath(stateValue) {
		resp.PlanValue = types.StringValue(stateValue)
	}
}

func canonicalPath(value string) string {
	if value == "/" {
		return value
	}
	return strings.TrimRight(value, "/")
}

// TrailingSlashPlanModifier returns a plan modifier that handles trailing slash normalization.
func TrailingSlashPlanModifier() planmodifier.String {
	return trailingSlashPlanModifier{}
}
