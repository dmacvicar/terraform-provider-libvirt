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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
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
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Pool         types.String `tfsdk:"pool"`
	Type         types.String `tfsdk:"type"`
	Key          types.String `tfsdk:"key"`
	Capacity     types.Int64  `tfsdk:"capacity"`
	Allocation   types.Int64  `tfsdk:"allocation"`
	Path         types.String `tfsdk:"path"`
	Format       types.String `tfsdk:"format"`
	BackingStore types.Object `tfsdk:"backing_store"`
	Permissions  types.Object `tfsdk:"permissions"`
	Create       types.Object `tfsdk:"create"`
}

// VolumeCreateModel describes the create block for volume initialization
type VolumeCreateModel struct {
	Content types.Object `tfsdk:"content"`
}

// VolumeCreateContentModel describes the content block for uploading from URL
type VolumeCreateContentModel struct {
	URL types.String `tfsdk:"url"`
}

// VolumeBackingStoreModel describes the backing store block
type VolumeBackingStoreModel struct {
	Path   types.String `tfsdk:"path"`
	Format types.String `tfsdk:"format"`
}

// VolumePermissionsModel describes permissions for the volume
type VolumePermissionsModel struct {
	Owner types.String `tfsdk:"owner"`
	Group types.String `tfsdk:"group"`
	Mode  types.String `tfsdk:"mode"`
	Label types.String `tfsdk:"label"`
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
			"type": schema.StringAttribute{
				Description: "Volume type (file, block, dir, network, netdir)",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
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
				Description: "Volume capacity in bytes. Required for empty volumes, computed when using create.content",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
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
			"permissions": schema.SingleNestedAttribute{
				Description: "Permissions for the volume file",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"owner": schema.StringAttribute{
						Description: "Numeric user ID for the volume file owner",
						Optional:    true,
					},
					"group": schema.StringAttribute{
						Description: "Numeric group ID for the volume file group",
						Optional:    true,
					},
					"mode": schema.StringAttribute{
						Description: "Octal permission mode for the volume file (e.g., '0644')",
						Optional:    true,
					},
					"label": schema.StringAttribute{
						Description: "SELinux label for the volume file",
						Optional:    true,
					},
				},
			},
			"create": schema.SingleNestedAttribute{
				Description: "Volume creation options for initializing volume content from external sources",
				Optional:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"content": schema.SingleNestedAttribute{
						Description: "Upload content from a URL or local file",
						Required:    true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								Description: "URL to download content from (supports https://, file://, or absolute paths)",
								Required:    true,
							},
						},
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

	// Check if we're uploading content from a URL
	var uploadStream *URLStream
	var uploadCapacity int64

	if !model.Create.IsNull() && !model.Create.IsUnknown() {
		var createModel VolumeCreateModel
		diags := model.Create.As(ctx, &createModel, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !createModel.Content.IsNull() && !createModel.Content.IsUnknown() {
			var contentModel VolumeCreateContentModel
			diags := createModel.Content.As(ctx, &contentModel, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			uploadURL := contentModel.URL.ValueString()
			tflog.Debug(ctx, "Opening URL stream for volume upload", map[string]any{
				"url": uploadURL,
			})

			// Open the URL stream
			stream, err := OpenURLStream(ctx, uploadURL)
			if err != nil {
				resp.Diagnostics.AddError(
					"Failed to Open URL",
					fmt.Sprintf("Could not open URL for upload: %s", err),
				)
				return
			}
			uploadStream = stream
			uploadCapacity = stream.Size

			tflog.Info(ctx, "URL stream opened", map[string]any{
				"url":  uploadURL,
				"size": uploadCapacity,
			})
		}
	}

	// Determine capacity: from upload stream or from user-provided value
	var capacity int64
	if uploadStream != nil {
		capacity = uploadCapacity
	} else {
		if model.Capacity.IsNull() || model.Capacity.IsUnknown() {
			resp.Diagnostics.AddError(
				"Missing Capacity",
				"Volume capacity is required when not uploading from a URL",
			)
			return
		}
		capacity = model.Capacity.ValueInt64()
	}

	// Build volume definition
	volumeDef := &libvirtxml.StorageVolume{
		Name: volumeName,
		Capacity: &libvirtxml.StorageVolumeSize{
			Value: uint64(capacity),
		},
		Target: &libvirtxml.StorageVolumeTarget{},
	}

	// Set type if specified
	if !model.Type.IsNull() && !model.Type.IsUnknown() {
		volumeDef.Type = model.Type.ValueString()
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

	// Set permissions if specified
	if !model.Permissions.IsNull() && !model.Permissions.IsUnknown() {
		var permissions VolumePermissionsModel
		diags := model.Permissions.As(ctx, &permissions, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		volumeDef.Target.Permissions = &libvirtxml.StorageVolumeTargetPermissions{}

		if !permissions.Owner.IsNull() && !permissions.Owner.IsUnknown() {
			volumeDef.Target.Permissions.Owner = permissions.Owner.ValueString()
		}
		if !permissions.Group.IsNull() && !permissions.Group.IsUnknown() {
			volumeDef.Target.Permissions.Group = permissions.Group.ValueString()
		}
		if !permissions.Mode.IsNull() && !permissions.Mode.IsUnknown() {
			volumeDef.Target.Permissions.Mode = permissions.Mode.ValueString()
		}
		if !permissions.Label.IsNull() && !permissions.Label.IsUnknown() {
			volumeDef.Target.Permissions.Label = permissions.Label.ValueString()
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

	// Upload content if we have a stream
	if uploadStream != nil {
		defer func() {
			if err := uploadStream.Reader.Close(); err != nil {
				tflog.Warn(ctx, "Failed to close upload stream", map[string]any{
					"error": err.Error(),
				})
			}
		}()

		tflog.Info(ctx, "Uploading content to volume", map[string]any{
			"size": uploadCapacity,
		})

		// Upload the content using StorageVolUpload
		// The 0 flag means start at offset 0, uploadCapacity is the length
		err = r.client.Libvirt().StorageVolUpload(volume, uploadStream.Reader, 0, uint64(uploadCapacity), 0)
		if err != nil {
			// Upload failed, try to clean up the volume
			_ = r.client.Libvirt().StorageVolDelete(volume, 0)
			resp.Diagnostics.AddError(
				"Volume Upload Failed",
				fmt.Sprintf("Failed to upload content to volume: %s", err),
			)
			return
		}

		tflog.Info(ctx, "Successfully uploaded content to volume", map[string]any{
			"bytes": uploadCapacity,
		})
	}

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

	if volumeDef.Type != "" {
		model.Type = types.StringValue(volumeDef.Type)
	}

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

	// Set permissions if present and user specified them
	if !model.Permissions.IsNull() && !model.Permissions.IsUnknown() && volumeDef.Target != nil && volumeDef.Target.Permissions != nil {
		permissionsModel := VolumePermissionsModel{
			Owner: types.StringNull(),
			Group: types.StringNull(),
			Mode:  types.StringNull(),
			Label: types.StringNull(),
		}

		if volumeDef.Target.Permissions.Owner != "" {
			permissionsModel.Owner = types.StringValue(volumeDef.Target.Permissions.Owner)
		}
		if volumeDef.Target.Permissions.Group != "" {
			permissionsModel.Group = types.StringValue(volumeDef.Target.Permissions.Group)
		}
		if volumeDef.Target.Permissions.Mode != "" {
			permissionsModel.Mode = types.StringValue(volumeDef.Target.Permissions.Mode)
		}
		if volumeDef.Target.Permissions.Label != "" {
			permissionsModel.Label = types.StringValue(volumeDef.Target.Permissions.Label)
		}

		permissionsObj, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"owner": types.StringType,
			"group": types.StringType,
			"mode":  types.StringType,
			"label": types.StringType,
		}, permissionsModel)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		model.Permissions = permissionsObj
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
