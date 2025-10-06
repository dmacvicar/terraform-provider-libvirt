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
	Path types.String `tfsdk:"path"`
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

	tflog.Debug(ctx, "Creating storage pool", map[string]any{
		"name": model.Name.ValueString(),
		"type": model.Type.ValueString(),
	})

	// TODO: Implement create logic
	resp.Diagnostics.AddError(
		"Not Implemented",
		"Pool resource create is not yet implemented",
	)
}

// Read reads the storage pool state
func (r *PoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model PoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: Implement read logic
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

	tflog.Debug(ctx, "Deleting storage pool", map[string]any{
		"name": model.Name.ValueString(),
	})

	// TODO: Implement delete logic
}

// ImportState imports an existing storage pool
func (r *PoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
