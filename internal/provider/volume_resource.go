package provider

import (
	"context"
	"fmt"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

	tflog.Debug(ctx, "Creating storage volume", map[string]any{
		"name": model.Name.ValueString(),
		"pool": model.Pool.ValueString(),
	})

	// TODO: Implement create logic
	resp.Diagnostics.AddError(
		"Not Implemented",
		"Volume resource create is not yet implemented",
	)
}

// Read reads the storage volume state
func (r *VolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model VolumeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: Implement read logic
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

	tflog.Debug(ctx, "Deleting storage volume", map[string]any{
		"name": model.Name.ValueString(),
	})

	// TODO: Implement delete logic
}

// ImportState imports an existing storage volume
func (r *VolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
