package provider

import (
	"context"
	"fmt"

	golibvirt "github.com/digitalocean/go-libvirt"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"libvirt.org/go/libvirtxml"
)

// Ensure the implementation satisfies the resource.Resource interface
var _ resource.Resource = &VolumeResource{}
var _ resource.ResourceWithImportState = &VolumeResource{}

// VolumeResource defines the resource implementation
type VolumeResource struct {
	client *libvirt.Client
}

// VolumeResourceModel describes the resource data model
type VolumeResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Pool       types.String `tfsdk:"pool"`
	Key        types.String `tfsdk:"key"`
	Capacity   types.Int64  `tfsdk:"capacity"`
	Allocation types.Int64  `tfsdk:"allocation"`
	Path       types.String `tfsdk:"path"`
	Format     types.String `tfsdk:"format"`
	BackingStore types.Object `tfsdk:"backing_store"`
}

// VolumeBackingStoreModel describes the backing store block
type VolumeBackingStoreModel struct {
	Path   types.String `tfsdk:"path"`
	Format types.String `tfsdk:"format"`
}

// NewVolumeResource creates a new volume resource
func NewVolumeResource() resource.Resource {
	return &VolumeResource{}
}

// Metadata returns the resource type name
func (r *VolumeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

// Schema defines the resource schema
func (r *VolumeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a libvirt storage volume. Volumes are stored in storage pools and can be attached to domains as disks.",
		MarkdownDescription: `
Manages a libvirt storage volume.

Storage volumes are images (qcow2, raw, etc.) stored in a storage pool that can be attached to virtual machines.

See the [libvirt storage volume documentation](https://libvirt.org/formatstorage.html) for more details.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Volume identifier (same as key)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the storage volume",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pool": schema.StringAttribute{
				Description: "Name of the storage pool where the volume will be created",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				Description: "Unique key of the storage volume",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"capacity": schema.Int64Attribute{
				Description: "Volume capacity in bytes",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"allocation": schema.Int64Attribute{
				Description: "Currently allocated size in bytes",
				Computed:    true,
			},
			"path": schema.StringAttribute{
				Description: "Full path to the volume on the host",
				Computed:    true,
			},
			"format": schema.StringAttribute{
				Description: "Volume format (qcow2, raw, etc.)",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"backing_store": schema.SingleNestedAttribute{
				Description: "Backing store configuration for copy-on-write volumes",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"path": schema.StringAttribute{
						Description: "Path to the backing volume",
						Required:    true,
					},
					"format": schema.StringAttribute{
						Description: "Format of the backing volume",
						Optional:    true,
					},
				},
			},
		},
	}
}

// Configure configures the resource
func (r *VolumeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new storage volume
func (r *VolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model VolumeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	volumeName := model.Name.ValueString()
	poolName := model.Pool.ValueString()

	tflog.Debug(ctx, "Creating storage volume", map[string]any{
		"name": volumeName,
		"pool": poolName,
	})

	// Look up the pool
	pool, err := r.client.Libvirt().StoragePoolLookupByName(poolName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Pool Not Found",
			fmt.Sprintf("Storage pool '%s' not found: %s", poolName, err),
		)
		return
	}

	// Build volume definition
	volumeDef := &libvirtxml.StorageVolume{
		Name: volumeName,
		Capacity: &libvirtxml.StorageVolumeSize{
			Value: uint64(model.Capacity.ValueInt64()),
		},
		Target: &libvirtxml.StorageVolumeTarget{},
	}

	// Set format if specified
	if !model.Format.IsNull() && !model.Format.IsUnknown() {
		volumeDef.Target.Format = &libvirtxml.StorageVolumeTargetFormat{
			Type: model.Format.ValueString(),
		}
	}

	// Set backing store if specified
	if !model.BackingStore.IsNull() {
		var backingStore VolumeBackingStoreModel
		diags := model.BackingStore.As(ctx, &backingStore, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		volumeDef.BackingStore = &libvirtxml.StorageVolumeBackingStore{
			Path: backingStore.Path.ValueString(),
		}

		if !backingStore.Format.IsNull() {
			volumeDef.BackingStore.Format = &libvirtxml.StorageVolumeTargetFormat{
				Type: backingStore.Format.ValueString(),
			}
		}
	}

	// Marshal to XML
	xmlDoc, err := volumeDef.Marshal()
	if err != nil {
		resp.Diagnostics.AddError(
			"XML Marshaling Failed",
			fmt.Sprintf("Failed to marshal storage volume XML: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Generated volume XML", map[string]any{"xml": xmlDoc})

	// Create the volume
	volume, err := r.client.Libvirt().StorageVolCreateXML(pool, xmlDoc, 0)
	if err != nil {
		resp.Diagnostics.AddError(
			"Volume Creation Failed",
			fmt.Sprintf("Failed to create storage volume: %s", err),
		)
		return
	}

	// Set the ID (use key)
	model.ID = types.StringValue(volume.Key)
	model.Key = types.StringValue(volume.Key)

	tflog.Info(ctx, "Created storage volume", map[string]any{
		"key":  volume.Key,
		"name": volumeName,
	})

	// Read back the full state
	resp.Diagnostics.Append(r.readVolume(ctx, &model, volume)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read reads the storage volume state
func (r *VolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model VolumeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Look up the volume by key
	volume, err := r.client.Libvirt().StorageVolLookupByKey(model.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Volume Not Found",
			fmt.Sprintf("Storage volume not found, removing from state: %s", err),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	// Read the volume state
	resp.Diagnostics.Append(r.readVolume(ctx, &model, volume)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// readVolume reads volume state from libvirt and populates the model
func (r *VolumeResource) readVolume(ctx context.Context, model *VolumeResourceModel, volume golibvirt.StorageVol) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get volume XML
	xmlDoc, err := r.client.Libvirt().StorageVolGetXMLDesc(volume, 0)
	if err != nil {
		diags.AddError(
			"Failed to Get Volume XML",
			fmt.Sprintf("Could not retrieve storage volume XML: %s", err),
		)
		return diags
	}

	// Parse XML
	var volumeDef libvirtxml.StorageVolume
	if err := volumeDef.Unmarshal(xmlDoc); err != nil {
		diags.AddError(
			"Failed to Parse Volume XML",
			fmt.Sprintf("Could not parse storage volume XML: %s", err),
		)
		return diags
	}

	// Get volume info for allocation
	volType, capacity, allocation, err := r.client.Libvirt().StorageVolGetInfo(volume)
	if err != nil {
		diags.AddError(
			"Failed to Get Volume Info",
			fmt.Sprintf("Could not retrieve storage volume info: %s", err),
		)
		return diags
	}
	_ = volType // Unused for now

	// Update model
	model.Name = types.StringValue(volumeDef.Name)
	model.Key = types.StringValue(volumeDef.Key)
	model.Capacity = types.Int64Value(int64(capacity))
	model.Allocation = types.Int64Value(int64(allocation))

	if volumeDef.Target != nil {
		model.Path = types.StringValue(volumeDef.Target.Path)

		if volumeDef.Target.Format != nil {
			model.Format = types.StringValue(volumeDef.Target.Format.Type)
		}
	}

	// Set backing store if present
	if volumeDef.BackingStore != nil {
		backingStoreModel := VolumeBackingStoreModel{
			Path: types.StringValue(volumeDef.BackingStore.Path),
		}

		if volumeDef.BackingStore.Format != nil {
			backingStoreModel.Format = types.StringValue(volumeDef.BackingStore.Format.Type)
		} else {
			backingStoreModel.Format = types.StringNull()
		}

		backingStoreObj, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"path":   types.StringType,
			"format": types.StringType,
		}, backingStoreModel)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		model.BackingStore = backingStoreObj
	}

	return diags
}

// Update updates the storage volume (volumes are immutable, so this should not be called)
func (r *VolumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Storage volumes cannot be updated. All changes require replacement.",
	)
}

// Delete deletes the storage volume
func (r *VolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model VolumeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	volumeName := model.Name.ValueString()

	tflog.Debug(ctx, "Deleting storage volume", map[string]any{
		"name": volumeName,
		"key":  model.Key.ValueString(),
	})

	// Look up the volume by key
	volume, err := r.client.Libvirt().StorageVolLookupByKey(model.Key.ValueString())
	if err != nil {
		// Volume doesn't exist, consider it deleted
		tflog.Info(ctx, "Storage volume not found, considering deleted", map[string]any{
			"name": volumeName,
		})
		return
	}

	// Delete the volume
	if err := r.client.Libvirt().StorageVolDelete(volume, 0); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Delete Volume",
			fmt.Sprintf("Could not delete storage volume: %s", err),
		)
		return
	}

	tflog.Info(ctx, "Deleted storage volume", map[string]any{
		"name": volumeName,
	})
}

// ImportState imports an existing storage volume
func (r *VolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
