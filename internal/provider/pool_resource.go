package provider

import (
	"context"
	"fmt"

	golibvirt "github.com/digitalocean/go-libvirt"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

// PoolResourceModel describes the resource data model
type PoolResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	UUID       types.String `tfsdk:"uuid"`
	Capacity   types.Int64  `tfsdk:"capacity"`
	Allocation types.Int64  `tfsdk:"allocation"`
	Available  types.Int64  `tfsdk:"available"`
	Target     types.Object `tfsdk:"target"`
	Source     types.Object `tfsdk:"source"`
}

// PoolTargetModel describes the target block
type PoolTargetModel struct {
	Path        types.String `tfsdk:"path"`
	Permissions types.Object `tfsdk:"permissions"`
}

// PoolPermissionsModel describes permissions for the pool directory
type PoolPermissionsModel struct {
	Owner types.String `tfsdk:"owner"`
	Group types.String `tfsdk:"group"`
	Mode  types.String `tfsdk:"mode"`
	Label types.String `tfsdk:"label"`
}

// PoolSourceModel describes the source block
type PoolSourceModel struct {
	Name   types.String `tfsdk:"name"`
	Device types.List   `tfsdk:"device"`
}

// PoolSourceDeviceModel describes a source device
type PoolSourceDeviceModel struct {
	Path types.String `tfsdk:"path"`
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
	resp.Schema = schema.Schema{
		Description: "Manages a libvirt storage pool. Storage pools provide a unified way to manage storage for virtual machines.",
		MarkdownDescription: `
Manages a libvirt storage pool.

Storage pools provide a common interface for managing storage that can be used by virtual machines.
This resource supports directory-based and LVM-based storage pools.

See the [libvirt storage pool documentation](https://libvirt.org/formatstorage.html) for more details.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Pool UUID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Unique name of the storage pool",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Type of storage pool. Supported values: dir (directory-based), logical (LVM)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"uuid": schema.StringAttribute{
				Description: "UUID of the storage pool",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"capacity": schema.Int64Attribute{
				Description: "Total capacity of the storage pool in bytes",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"allocation": schema.Int64Attribute{
				Description: "Currently allocated space in bytes",
				Computed:    true,
			},
			"available": schema.Int64Attribute{
				Description: "Available space in bytes",
				Computed:    true,
			},
			"target": schema.SingleNestedAttribute{
				Description: "Target path for the storage pool",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"path": schema.StringAttribute{
						Description: "Path where the storage pool is located on the host",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"permissions": schema.SingleNestedAttribute{
						Description: "Permissions for the pool directory",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"owner": schema.StringAttribute{
								Description: "Numeric user ID for the pool directory owner",
								Optional:    true,
							},
							"group": schema.StringAttribute{
								Description: "Numeric group ID for the pool directory group",
								Optional:    true,
							},
							"mode": schema.StringAttribute{
								Description: "Octal permission mode for the pool directory (e.g., '0755')",
								Optional:    true,
							},
							"label": schema.StringAttribute{
								Description: "SELinux label for the pool directory",
								Optional:    true,
							},
						},
					},
				},
			},
			"source": schema.SingleNestedAttribute{
				Description: "Source configuration for the storage pool (required for logical pools)",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Description: "Name of the source (e.g., volume group name for logical pools)",
						Optional:    true,
					},
					"device": schema.ListNestedAttribute{
						Description: "List of devices to use for the storage pool (e.g., physical volumes for logical pools)",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"path": schema.StringAttribute{
									Description: "Path to the device",
									Required:    true,
								},
							},
						},
					},
				},
			},
		},
	}
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

	// Validate pool type
	if poolType != "dir" && poolType != "logical" {
		resp.Diagnostics.AddError(
			"Invalid Pool Type",
			fmt.Sprintf("Pool type must be 'dir' or 'logical', got: %s", poolType),
		)
		return
	}

	// Convert model to libvirt XML
	poolDef := &libvirtxml.StoragePool{
		Type: poolType,
		Name: poolName,
	}

	// Set target path
	if !model.Target.IsNull() {
		var target PoolTargetModel
		diags := model.Target.As(ctx, &target, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		poolDef.Target = &libvirtxml.StoragePoolTarget{
			Path: target.Path.ValueString(),
		}

		// Set target permissions if specified
		if !target.Permissions.IsNull() && !target.Permissions.IsUnknown() {
			var permissions PoolPermissionsModel
			diags := target.Permissions.As(ctx, &permissions, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			poolDef.Target.Permissions = &libvirtxml.StoragePoolTargetPermissions{}

			if !permissions.Owner.IsNull() && !permissions.Owner.IsUnknown() {
				poolDef.Target.Permissions.Owner = permissions.Owner.ValueString()
			}
			if !permissions.Group.IsNull() && !permissions.Group.IsUnknown() {
				poolDef.Target.Permissions.Group = permissions.Group.ValueString()
			}
			if !permissions.Mode.IsNull() && !permissions.Mode.IsUnknown() {
				poolDef.Target.Permissions.Mode = permissions.Mode.ValueString()
			}
			if !permissions.Label.IsNull() && !permissions.Label.IsUnknown() {
				poolDef.Target.Permissions.Label = permissions.Label.ValueString()
			}
		}
	}

	// Set source (for logical pools)
	var skipBuild bool
	if !model.Source.IsNull() {
		var source PoolSourceModel
		diags := model.Source.As(ctx, &source, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		poolDef.Source = &libvirtxml.StoragePoolSource{}

		if !source.Name.IsNull() {
			poolDef.Source.Name = source.Name.ValueString()
		}

		if !source.Device.IsNull() {
			var devices []PoolSourceDeviceModel
			diags := source.Device.ElementsAs(ctx, &devices, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			var libvirtDevices []libvirtxml.StoragePoolSourceDevice
			for _, dev := range devices {
				libvirtDevices = append(libvirtDevices, libvirtxml.StoragePoolSourceDevice{
					Path: dev.Path.ValueString(),
				})
			}
			poolDef.Source.Device = libvirtDevices
		}

		// If no devices specified for logical pool, skip build step
		if len(poolDef.Source.Device) == 0 && poolType == "logical" {
			skipBuild = true
		}
	} else if poolType == "logical" {
		// Logical pool without source means existing VG
		skipBuild = true
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

	// Read back the full state
	resp.Diagnostics.Append(r.readPool(ctx, &model, pool)...)
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

	// Read the pool state
	resp.Diagnostics.Append(r.readPool(ctx, &model, pool)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// readPool reads pool state from libvirt and populates the model
func (r *PoolResource) readPool(ctx context.Context, model *PoolResourceModel, pool golibvirt.StoragePool) diag.Diagnostics {
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

	// Get pool info for capacity/allocation/available
	state, capacity, allocation, available, err := r.client.Libvirt().StoragePoolGetInfo(pool)
	if err != nil {
		diags.AddError(
			"Failed to Get Pool Info",
			fmt.Sprintf("Could not retrieve storage pool info: %s", err),
		)
		return diags
	}
	_ = state // Unused for now

	// Update model
	model.UUID = types.StringValue(poolDef.UUID)
	model.Name = types.StringValue(poolDef.Name)
	model.Type = types.StringValue(poolDef.Type)
	model.Capacity = types.Int64Value(int64(capacity))
	model.Allocation = types.Int64Value(int64(allocation))
	model.Available = types.Int64Value(int64(available))

	// Set target
	if poolDef.Target != nil {
		permissionsObjectType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"owner": types.StringType,
				"group": types.StringType,
				"mode":  types.StringType,
				"label": types.StringType,
			},
		}

		targetModel := PoolTargetModel{
			Path:        types.StringValue(poolDef.Target.Path),
			Permissions: types.ObjectNull(permissionsObjectType.AttrTypes),
		}

		// Set permissions if present and user specified them
		if !model.Target.IsNull() && !model.Target.IsUnknown() {
			var origTarget PoolTargetModel
			d := model.Target.As(ctx, &origTarget, basetypes.ObjectAsOptions{})
			diags.Append(d...)
			if !diags.HasError() && !origTarget.Permissions.IsNull() && !origTarget.Permissions.IsUnknown() && poolDef.Target.Permissions != nil {
				permissionsModel := PoolPermissionsModel{
					Owner: types.StringNull(),
					Group: types.StringNull(),
					Mode:  types.StringNull(),
					Label: types.StringNull(),
				}

				if poolDef.Target.Permissions.Owner != "" {
					permissionsModel.Owner = types.StringValue(poolDef.Target.Permissions.Owner)
				}
				if poolDef.Target.Permissions.Group != "" {
					permissionsModel.Group = types.StringValue(poolDef.Target.Permissions.Group)
				}
				if poolDef.Target.Permissions.Mode != "" {
					permissionsModel.Mode = types.StringValue(poolDef.Target.Permissions.Mode)
				}
				if poolDef.Target.Permissions.Label != "" {
					permissionsModel.Label = types.StringValue(poolDef.Target.Permissions.Label)
				}

				permissionsObj, d := types.ObjectValueFrom(ctx, permissionsObjectType.AttrTypes, permissionsModel)
				diags.Append(d...)
				if !diags.HasError() {
					targetModel.Permissions = permissionsObj
				}
			}
		}

		targetObj, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"path":        types.StringType,
			"permissions": permissionsObjectType,
		}, targetModel)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		model.Target = targetObj
	}

	// Set source (if present and user-specified)
	deviceObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"path": types.StringType,
		},
	}
	sourceAttrTypes := map[string]attr.Type{
		"name":   types.StringType,
		"device": types.ListType{ElemType: deviceObjectType},
	}

	if poolDef.Source != nil && !model.Source.IsNull() && !model.Source.IsUnknown() {

		sourceModel := PoolSourceModel{
			Name:   types.StringNull(),
			Device: types.ListNull(deviceObjectType),
		}

		if poolDef.Source.Name != "" {
			sourceModel.Name = types.StringValue(poolDef.Source.Name)
		}

		if len(poolDef.Source.Device) > 0 {
			var devices []PoolSourceDeviceModel
			for _, dev := range poolDef.Source.Device {
				if dev.Path == "" {
					continue
				}
				devices = append(devices, PoolSourceDeviceModel{
					Path: types.StringValue(dev.Path),
				})
			}
			if len(devices) > 0 {
				deviceList, d := types.ListValueFrom(ctx, deviceObjectType, devices)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
				sourceModel.Device = deviceList
			}
		}

		sourceObj, d := types.ObjectValueFrom(ctx, sourceAttrTypes, sourceModel)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		model.Source = sourceObj
	} else if !model.Source.IsNull() && !model.Source.IsUnknown() {
		model.Source = types.ObjectNull(sourceAttrTypes)
	}

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
