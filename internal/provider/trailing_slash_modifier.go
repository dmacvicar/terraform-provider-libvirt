package provider

import (
	"context"
	"strings"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/generated"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

	canonicalPlan := canonicalPath(planValue)

	// If the state is null (new resource), nothing else to compare
	if req.StateValue.IsNull() {
		return
	}

	stateValue := req.StateValue.ValueString()

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

func preservePoolTargetPath(ctx context.Context, model *generated.StoragePoolModel, plan *generated.StoragePoolModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if model == nil || plan == nil {
		return diags
	}
	if model.Target.IsNull() || model.Target.IsUnknown() || plan.Target.IsNull() || plan.Target.IsUnknown() {
		return diags
	}

	var target generated.StoragePoolTargetModel
	if d := model.Target.As(ctx, &target, basetypes.ObjectAsOptions{}); d.HasError() {
		diags.AddError("Failed to read pool target", d.Errors()[0].Summary())
		return diags
	}

	var planTarget generated.StoragePoolTargetModel
	if d := plan.Target.As(ctx, &planTarget, basetypes.ObjectAsOptions{}); d.HasError() {
		diags.AddError("Failed to read planned pool target", d.Errors()[0].Summary())
		return diags
	}

	if target.Path.IsNull() || target.Path.IsUnknown() || planTarget.Path.IsNull() || planTarget.Path.IsUnknown() {
		return diags
	}

	if canonicalPath(planTarget.Path.ValueString()) != canonicalPath(target.Path.ValueString()) {
		return diags
	}

	target.Path = planTarget.Path
	targetObj, d := types.ObjectValueFrom(ctx, generated.StoragePoolTargetAttributeTypes(), &target)
	if d.HasError() {
		diags.AddError("Failed to update pool target", d.Errors()[0].Summary())
		return diags
	}

	model.Target = targetObj
	return diags
}
