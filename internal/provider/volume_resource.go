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

// VolumeResourceModel extends generated model with resource-specific fields
type VolumeResourceModel struct {
	generated.StorageVolumeModel
	ID     types.String `tfsdk:"id"`     // Resource-specific ID
	Pool   types.String `tfsdk:"pool"`   // Provider-specific: which pool to create in
	Path   types.String `tfsdk:"path"`   // Computed: convenience field mirroring target.path
	Create types.Object `tfsdk:"create"` // Provider-specific: upload content on create
}

// VolumeCreateModel describes the create block for volume initialization
type VolumeCreateModel struct {
	Content types.Object `tfsdk:"content"`
}

// VolumeCreateContentModel describes the content block for uploading from URL
type VolumeCreateContentModel struct {
	URL types.String `tfsdk:"url"`
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
	// Get base target schema from generated code
	baseTargetAttr := mustSingleNestedAttribute(generated.StorageVolumeTargetSchemaAttribute(), "StorageVolumeTarget")

	// Override path to be Computed
	targetAttrs := baseTargetAttr.Attributes
	targetAttrs["path"] = schema.StringAttribute{
		Description: "Volume path on the host filesystem",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	// Normalize permissions.mode to avoid diffs (e.g., 770 vs 0770)
	permissionsAttr := mustSingleNestedAttribute(targetAttrs["permissions"], "StorageVolumeTargetPermissions")
	permissionsAttrs := permissionsAttr.Attributes
	modeAttr := mustStringAttribute(permissionsAttrs["mode"], "StorageVolumeTargetPermissions.mode")
	modeAttr.PlanModifiers = append(modeAttr.PlanModifiers, OctalModePlanModifier())
	permissionsAttrs["mode"] = modeAttr
	permissionsAttr.Attributes = permissionsAttrs
	targetAttrs["permissions"] = permissionsAttr

	// Use generated schema with provider-specific overrides
	resp.Schema = generated.StorageVolumeSchema(map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Volume identifier (same as key)",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"pool": schema.StringAttribute{
			Description: "Name of the storage pool where the volume will be created",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"path": schema.StringAttribute{
			Description: "Volume path on the host filesystem (same as target.path)",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"capacity": schema.Int64Attribute{
			Description: "Volume capacity in bytes (required unless using create.content)",
			Optional:    true,
			Computed:    true,
		},
		"target": schema.SingleNestedAttribute{
			Optional:   true,
			Attributes: targetAttrs,
		},
		"create": schema.SingleNestedAttribute{
			Description: "Volume creation options for initializing volume content from external sources",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"content": schema.SingleNestedAttribute{
					Description: "Upload content from a URL or local file",
					Required:    true,
					Attributes: map[string]schema.Attribute{
						"url": schema.StringAttribute{
							Description: "URL to download content from",
							Required:    true,
						},
					},
				},
			},
		},
	})
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
	var volumeCapacity int64
	if uploadStream != nil {
		volumeCapacity = uploadCapacity
	} else if model.Capacity.IsNull() || model.Capacity.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing Capacity",
			"Volume capacity is required when not uploading from a URL",
		)
		return
	} else {
		volumeCapacity = model.Capacity.ValueInt64()
	}

	// Create a model for XML conversion with computed capacity
	xmlModel := model.StorageVolumeModel
	xmlModel.Capacity = types.Int64Value(volumeCapacity)

	// Convert model to libvirt XML using generated conversion
	volumeDef, err := generated.StorageVolumeToXML(ctx, &xmlModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Model to XML Conversion Failed",
			fmt.Sprintf("Failed to convert model to XML: %s", err),
		)
		return
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
			// Upload failed, try to clean up the volume (ignore cleanup errors to preserve original error)
			if delErr := r.client.Libvirt().StorageVolDelete(volume, 0); delErr != nil {
				tflog.Warn(ctx, "Failed to delete volume during cleanup", map[string]any{
					"error": delErr.Error(),
				})
			}
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

	// Save the plan for preserving user intent
	planModel := model.StorageVolumeModel

	// Read back the full state
	resp.Diagnostics.Append(r.readVolume(ctx, &model, volume, &planModel)...)
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

	// Read the volume state (use current state as plan to preserve user intent)
	resp.Diagnostics.Append(r.readVolume(ctx, &model, volume, &model.StorageVolumeModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// readVolume reads volume state from libvirt and populates the model
// plan parameter is used to preserve user intent (only populate fields user specified)
func (r *VolumeResource) readVolume(ctx context.Context, model *VolumeResourceModel, volume golibvirt.StorageVol, plan *generated.StorageVolumeModel) diag.Diagnostics {
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

	// Convert XML to model using generated conversion
	volumeModel, err := generated.StorageVolumeFromXML(ctx, &volumeDef, plan)
	if err != nil {
		diags.AddError(
			"XML to Model Conversion Failed",
			fmt.Sprintf("Failed to convert XML to model: %s", err),
		)
		return diags
	}

	diags.Append(preserveVolumeTargetPermissionsMode(ctx, volumeModel, plan)...)
	if diags.HasError() {
		return diags
	}

	// Update the embedded model
	model.StorageVolumeModel = *volumeModel

	// Populate computed fields that generated conversion might skip
	// target.path is Computed, always populate it
	if volumeDef.Target != nil && volumeDef.Target.Path != "" {
		// Populate top-level path for convenience
		model.Path = types.StringValue(volumeDef.Target.Path)

		// Extract current target from the model (already properly converted)
		if !model.Target.IsNull() {
			var targetModel generated.StorageVolumeTargetModel
			targetDiags := model.Target.As(ctx, &targetModel, basetypes.ObjectAsOptions{})
			diags.Append(targetDiags...)
			if !diags.HasError() {
				// Update just the path field
				targetModel.Path = types.StringValue(volumeDef.Target.Path)

				// Write back
				targetObj, objDiags := types.ObjectValueFrom(ctx, generated.StorageVolumeTargetAttributeTypes(), targetModel)
				diags.Append(objDiags...)
				if !diags.HasError() {
					model.Target = targetObj
				}
			}
		}
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
