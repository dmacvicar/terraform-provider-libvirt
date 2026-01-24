package provider

import (
	"context"
	"fmt"

	golibvirt "github.com/digitalocean/go-libvirt"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/generated"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"libvirt.org/go/libvirtxml"
)

// Ensure the implementation satisfies the resource.Resource interface
var _ resource.Resource = &PoolResource{}
var _ resource.ResourceWithImportState = &PoolResource{}

// PoolResource defines the resource implementation
type PoolResource struct {
	client *libvirt.Client
}

// PoolResourceModel extends generated model with resource-specific ID field
type PoolResourceModel struct {
	generated.StoragePoolModel
	ID types.String `tfsdk:"id"` // Resource-specific ID
}

// NewPoolResource creates a new pool resource
func NewPoolResource() resource.Resource {
	return &PoolResource{}
}

// Metadata returns the resource type name
func (r *PoolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pool"
}

// Schema defines the resource schema
func (r *PoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Get base target schema from generated code
	baseTargetAttr := mustSingleNestedAttribute(generated.StoragePoolTargetSchemaAttribute(), "StoragePoolTarget")
	targetAttrs := baseTargetAttr.Attributes

	// Normalize permissions.mode to avoid diffs (e.g., 770 vs 0770)
	permissionsAttr := mustSingleNestedAttribute(targetAttrs["permissions"], "StoragePoolTargetPermissions")
	permissionsAttrs := permissionsAttr.Attributes
	modeAttr := mustStringAttribute(permissionsAttrs["mode"], "StoragePoolTargetPermissions.mode")
	modeAttr.PlanModifiers = append(modeAttr.PlanModifiers, OctalModePlanModifier())
	permissionsAttrs["mode"] = modeAttr
	permissionsAttr.Attributes = permissionsAttrs
	targetAttrs["permissions"] = permissionsAttr

	// Use generated schema with resource-specific overrides
	resp.Schema = generated.StoragePoolSchema(map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Pool UUID (same as uuid)",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"target": schema.SingleNestedAttribute{
			Optional:   true,
			Attributes: targetAttrs,
		},
	})
}

// Configure configures the resource
func (r *PoolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*libvirt.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *libvirt.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates a new storage pool
func (r *PoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model PoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	poolName := model.Name.ValueString()
	poolType := model.Type.ValueString()

	tflog.Debug(ctx, "Creating storage pool", map[string]any{
		"name": poolName,
		"type": poolType,
	})

	// Convert model to libvirt XML using generated conversion
	poolDef, err := generated.StoragePoolToXML(ctx, &model.StoragePoolModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Model to XML Conversion Failed",
			fmt.Sprintf("Failed to convert model to XML: %s", err),
		)
		return
	}

	// Determine if we should skip the build step
	// For logical pools without source devices, we assume the VG already exists
	var skipBuild bool
	if poolType == "logical" {
		if poolDef.Source == nil || len(poolDef.Source.Device) == 0 {
			skipBuild = true
		}
	}

	// Marshal to XML
	xmlDoc, err := poolDef.Marshal()
	if err != nil {
		resp.Diagnostics.AddError(
			"XML Marshaling Failed",
			fmt.Sprintf("Failed to marshal storage pool XML: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Generated pool XML", map[string]any{"xml": xmlDoc})

	// Define the pool
	pool, err := r.client.Libvirt().StoragePoolDefineXML(xmlDoc, 0)
	if err != nil {
		resp.Diagnostics.AddError(
			"Pool Creation Failed",
			fmt.Sprintf("Failed to define storage pool: %s", err),
		)
		return
	}

	// Build the pool (unless we're skipping)
	poolBuilt := false
	if !skipBuild {
		if err := r.client.Libvirt().StoragePoolBuild(pool, 0); err != nil {
			// Cleanup: undefine the pool we just defined
			if undefErr := r.client.Libvirt().StoragePoolUndefine(pool); undefErr != nil {
				tflog.Warn(ctx, "Failed to undefine pool during cleanup", map[string]any{
					"error": undefErr.Error(),
				})
			}
			resp.Diagnostics.AddError(
				"Pool Build Failed",
				fmt.Sprintf("Failed to build storage pool: %s", err),
			)
			return
		}
		poolBuilt = true
	}

	// Set autostart
	if err := r.client.Libvirt().StoragePoolSetAutostart(pool, 1); err != nil {
		// Cleanup: delete if built, then undefine
		if poolBuilt {
			if deleteErr := r.client.Libvirt().StoragePoolDelete(pool, 0); deleteErr != nil {
				tflog.Warn(ctx, "Failed to delete pool during cleanup", map[string]any{
					"error": deleteErr.Error(),
				})
			}
		}
		if undefErr := r.client.Libvirt().StoragePoolUndefine(pool); undefErr != nil {
			tflog.Warn(ctx, "Failed to undefine pool during cleanup", map[string]any{
				"error": undefErr.Error(),
			})
		}
		resp.Diagnostics.AddError(
			"Pool Autostart Failed",
			fmt.Sprintf("Failed to set pool autostart: %s", err),
		)
		return
	}

	// Start the pool
	if err := r.client.Libvirt().StoragePoolCreate(pool, 0); err != nil {
		// Cleanup: delete if built, then undefine
		if poolBuilt {
			if deleteErr := r.client.Libvirt().StoragePoolDelete(pool, 0); deleteErr != nil {
				tflog.Warn(ctx, "Failed to delete pool during cleanup", map[string]any{
					"error": deleteErr.Error(),
				})
			}
		}
		if undefErr := r.client.Libvirt().StoragePoolUndefine(pool); undefErr != nil {
			tflog.Warn(ctx, "Failed to undefine pool during cleanup", map[string]any{
				"error": undefErr.Error(),
			})
		}
		resp.Diagnostics.AddError(
			"Pool Start Failed",
			fmt.Sprintf("Failed to start storage pool: %s", err),
		)
		return
	}

	// Refresh to get current state
	if err := r.client.Libvirt().StoragePoolRefresh(pool, 0); err != nil {
		// Cleanup: destroy, delete if built, then undefine
		if destroyErr := r.client.Libvirt().StoragePoolDestroy(pool); destroyErr != nil {
			tflog.Warn(ctx, "Failed to destroy pool during cleanup", map[string]any{
				"error": destroyErr.Error(),
			})
		}
		if poolBuilt {
			if deleteErr := r.client.Libvirt().StoragePoolDelete(pool, 0); deleteErr != nil {
				tflog.Warn(ctx, "Failed to delete pool during cleanup", map[string]any{
					"error": deleteErr.Error(),
				})
			}
		}
		if undefErr := r.client.Libvirt().StoragePoolUndefine(pool); undefErr != nil {
			tflog.Warn(ctx, "Failed to undefine pool during cleanup", map[string]any{
				"error": undefErr.Error(),
			})
		}
		resp.Diagnostics.AddError(
			"Pool Refresh Failed",
			fmt.Sprintf("Failed to refresh storage pool: %s", err),
		)
		return
	}

	// Set the ID (use UUID)
	uuid := libvirt.UUIDString(pool.UUID)
	model.ID = types.StringValue(uuid)

	tflog.Info(ctx, "Created storage pool", map[string]any{
		"id":   uuid,
		"name": poolName,
	})

	// Save the plan for preserving user intent
	planModel := model.StoragePoolModel

	// Read back the full state
	resp.Diagnostics.Append(r.readPoolWithPlan(ctx, &model, pool, &planModel)...)
	if resp.Diagnostics.HasError() {
		// Cleanup: destroy, delete if built, then undefine
		if destroyErr := r.client.Libvirt().StoragePoolDestroy(pool); destroyErr != nil {
			tflog.Warn(ctx, "Failed to destroy pool during cleanup", map[string]any{
				"error": destroyErr.Error(),
			})
		}
		if poolBuilt {
			if deleteErr := r.client.Libvirt().StoragePoolDelete(pool, 0); deleteErr != nil {
				tflog.Warn(ctx, "Failed to delete pool during cleanup", map[string]any{
					"error": deleteErr.Error(),
				})
			}
		}
		if undefErr := r.client.Libvirt().StoragePoolUndefine(pool); undefErr != nil {
			tflog.Warn(ctx, "Failed to undefine pool during cleanup", map[string]any{
				"error": undefErr.Error(),
			})
		}
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read reads the storage pool state
func (r *PoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model PoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Look up the pool
	pool, err := r.client.LookupPoolByUUID(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Pool Not Found",
			fmt.Sprintf("Storage pool not found, removing from state: %s", err),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	// Read the pool state (use current state as plan to preserve user intent)
	resp.Diagnostics.Append(r.readPoolWithPlan(ctx, &model, pool, &model.StoragePoolModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// readPoolWithPlan reads pool state from libvirt and populates the model
// plan parameter is used to preserve user intent (only populate fields user specified)
func (r *PoolResource) readPoolWithPlan(ctx context.Context, model *PoolResourceModel, pool golibvirt.StoragePool, plan *generated.StoragePoolModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get pool XML
	xmlDoc, err := r.client.Libvirt().StoragePoolGetXMLDesc(pool, 0)
	if err != nil {
		diags.AddError(
			"Failed to Get Pool XML",
			fmt.Sprintf("Could not retrieve storage pool XML: %s", err),
		)
		return diags
	}

	// Parse XML
	var poolDef libvirtxml.StoragePool
	if err := poolDef.Unmarshal(xmlDoc); err != nil {
		diags.AddError(
			"Failed to Parse Pool XML",
			fmt.Sprintf("Could not parse storage pool XML: %s", err),
		)
		return diags
	}

	// Convert XML to model using generated conversion
	poolModel, err := generated.StoragePoolFromXML(ctx, &poolDef, plan)
	if err != nil {
		diags.AddError(
			"XML to Model Conversion Failed",
			fmt.Sprintf("Failed to convert XML to model: %s", err),
		)
		return diags
	}

	diags.Append(preservePoolTargetPermissionsMode(ctx, poolModel, plan)...)
	if diags.HasError() {
		return diags
	}

	// Update the embedded model
	model.StoragePoolModel = *poolModel

	return diags
}

// Update updates the storage pool (pools are immutable, so this should not be called)
func (r *PoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Storage pools cannot be updated. All changes require replacement.",
	)
}

// Delete deletes the storage pool
func (r *PoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model PoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	poolName := model.Name.ValueString()

	tflog.Debug(ctx, "Deleting storage pool", map[string]any{
		"name": poolName,
	})

	// Look up the pool
	pool, err := r.client.LookupPoolByUUID(model.ID.ValueString())
	if err != nil {
		// Pool doesn't exist, consider it deleted
		tflog.Info(ctx, "Storage pool not found, considering deleted", map[string]any{
			"name": poolName,
		})
		return
	}

	// Get pool XML to check if we should delete the underlying storage
	xmlDoc, err := r.client.Libvirt().StoragePoolGetXMLDesc(pool, 0)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Get Pool XML",
			fmt.Sprintf("Could not retrieve storage pool XML: %s", err),
		)
		return
	}

	var poolDef libvirtxml.StoragePool
	if err := poolDef.Unmarshal(xmlDoc); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Parse Pool XML",
			fmt.Sprintf("Could not parse storage pool XML: %s", err),
		)
		return
	}

	// Destroy (stop) the pool if it's active
	if err := r.client.Libvirt().StoragePoolDestroy(pool); err != nil {
		// Pool might already be inactive, that's okay
		tflog.Debug(ctx, "Pool destroy returned error (may already be inactive)", map[string]any{
			"error": err.Error(),
		})
	}

	// Determine if we should delete the underlying storage
	// For dir pools: always delete
	// For logical pools: only delete if we created the VG (i.e., if source devices were specified)
	shouldDelete := poolDef.Type == "dir"
	if poolDef.Type == "logical" && poolDef.Source != nil && len(poolDef.Source.Device) > 0 {
		shouldDelete = true
	}

	if shouldDelete {
		// Delete the pool's storage
		if err := r.client.Libvirt().StoragePoolDelete(pool, 0); err != nil {
			resp.Diagnostics.AddError(
				"Failed to Delete Pool Storage",
				fmt.Sprintf("Could not delete storage pool storage: %s", err),
			)
			return
		}
	}

	// Undefine the pool
	if err := r.client.Libvirt().StoragePoolUndefine(pool); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Undefine Pool",
			fmt.Sprintf("Could not undefine storage pool: %s", err),
		)
		return
	}

	tflog.Info(ctx, "Deleted storage pool", map[string]any{
		"name": poolName,
	})
}

// ImportState imports an existing storage pool
func (r *PoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
