package provider

import (
	"context"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/generated"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func preserveVolumeTargetPermissionsMode(ctx context.Context, model *generated.StorageVolumeModel, plan *generated.StorageVolumeModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if model == nil || plan == nil {
		return diags
	}
	if model.Target.IsNull() || model.Target.IsUnknown() || plan.Target.IsNull() || plan.Target.IsUnknown() {
		return diags
	}

	var target generated.StorageVolumeTargetModel
	if d := model.Target.As(ctx, &target, basetypes.ObjectAsOptions{}); d.HasError() {
		diags.AddError("Failed to read volume target", d.Errors()[0].Summary())
		return diags
	}

	var planTarget generated.StorageVolumeTargetModel
	if d := plan.Target.As(ctx, &planTarget, basetypes.ObjectAsOptions{}); d.HasError() {
		diags.AddError("Failed to read planned volume target", d.Errors()[0].Summary())
		return diags
	}

	if target.Permissions.IsNull() || target.Permissions.IsUnknown() || planTarget.Permissions.IsNull() || planTarget.Permissions.IsUnknown() {
		return diags
	}

	var perms generated.StorageVolumeTargetPermissionsModel
	if d := target.Permissions.As(ctx, &perms, basetypes.ObjectAsOptions{}); d.HasError() {
		diags.AddError("Failed to read volume target permissions", d.Errors()[0].Summary())
		return diags
	}

	var planPerms generated.StorageVolumeTargetPermissionsModel
	if d := planTarget.Permissions.As(ctx, &planPerms, basetypes.ObjectAsOptions{}); d.HasError() {
		diags.AddError("Failed to read planned volume target permissions", d.Errors()[0].Summary())
		return diags
	}

	if planPerms.Mode.IsNull() || planPerms.Mode.IsUnknown() || perms.Mode.IsNull() || perms.Mode.IsUnknown() {
		return diags
	}

	if canonicalOctalMode(planPerms.Mode.ValueString()) != canonicalOctalMode(perms.Mode.ValueString()) {
		return diags
	}

	perms.Mode = planPerms.Mode
	permsObj, d := types.ObjectValueFrom(ctx, generated.StorageVolumeTargetPermissionsAttributeTypes(), &perms)
	if d.HasError() {
		diags.AddError("Failed to update volume target permissions", d.Errors()[0].Summary())
		return diags
	}

	target.Permissions = permsObj
	targetObj, d := types.ObjectValueFrom(ctx, generated.StorageVolumeTargetAttributeTypes(), &target)
	if d.HasError() {
		diags.AddError("Failed to update volume target", d.Errors()[0].Summary())
		return diags
	}

	model.Target = targetObj
	return diags
}

func preservePoolTargetPermissionsMode(ctx context.Context, model *generated.StoragePoolModel, plan *generated.StoragePoolModel) diag.Diagnostics {
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

	if target.Permissions.IsNull() || target.Permissions.IsUnknown() || planTarget.Permissions.IsNull() || planTarget.Permissions.IsUnknown() {
		return diags
	}

	var perms generated.StoragePoolTargetPermissionsModel
	if d := target.Permissions.As(ctx, &perms, basetypes.ObjectAsOptions{}); d.HasError() {
		diags.AddError("Failed to read pool target permissions", d.Errors()[0].Summary())
		return diags
	}

	var planPerms generated.StoragePoolTargetPermissionsModel
	if d := planTarget.Permissions.As(ctx, &planPerms, basetypes.ObjectAsOptions{}); d.HasError() {
		diags.AddError("Failed to read planned pool target permissions", d.Errors()[0].Summary())
		return diags
	}

	if planPerms.Mode.IsNull() || planPerms.Mode.IsUnknown() || perms.Mode.IsNull() || perms.Mode.IsUnknown() {
		return diags
	}

	if canonicalOctalMode(planPerms.Mode.ValueString()) != canonicalOctalMode(perms.Mode.ValueString()) {
		return diags
	}

	perms.Mode = planPerms.Mode
	permsObj, d := types.ObjectValueFrom(ctx, generated.StoragePoolTargetPermissionsAttributeTypes(), &perms)
	if d.HasError() {
		diags.AddError("Failed to update pool target permissions", d.Errors()[0].Summary())
		return diags
	}

	target.Permissions = permsObj
	targetObj, d := types.ObjectValueFrom(ctx, generated.StoragePoolTargetAttributeTypes(), &target)
	if d.HasError() {
		diags.AddError("Failed to update pool target", d.Errors()[0].Summary())
		return diags
	}

	model.Target = targetObj
	return diags
}
