package provider

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the resource.Resource interface
var _ resource.Resource = &IgnitionResource{}

// IgnitionResource defines the resource implementation
type IgnitionResource struct {
	// No client needed - this is a local file operation
}

// IgnitionResourceModel describes the resource data model
type IgnitionResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Content types.String `tfsdk:"content"`
	Path    types.String `tfsdk:"path"`
	Size    types.Int64  `tfsdk:"size"`
}

// NewIgnitionResource creates a new ignition resource
func NewIgnitionResource() resource.Resource {
	return &IgnitionResource{}
}

// Metadata returns the resource type name
func (r *IgnitionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ignition"
}

// Schema defines the resource schema
func (r *IgnitionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Generates an Ignition configuration file that can be uploaded to a libvirt volume.",
		MarkdownDescription: `
Generates an Ignition configuration file for CoreOS/Fedora CoreOS systems.

Ignition is a provisioning tool that reads a configuration file and provisions the machine
accordingly on first boot. This resource generates the Ignition file that can be uploaded
to a volume and provided to the virtual machine.

## Example Usage

` + "```hcl" + `
resource "libvirt_ignition" "fcos" {
  name    = "fcos-ignition"
  content = data.ignition_config.fcos.rendered
}

resource "libvirt_volume" "ignition" {
  name   = "fcos-ignition.ign"
  pool   = "default"
  format = "raw"

  create = {
    content = {
      url = libvirt_ignition.fcos.path
    }
  }
}
` + "```" + `

See the [Ignition documentation](https://coreos.github.io/ignition/) for configuration details.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier (checksum of content)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name for this ignition resource",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				Description: "Ignition configuration content (JSON)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"path": schema.StringAttribute{
				Description: "Full path to the generated ignition file",
				Computed:    true,
			},
			"size": schema.Int64Attribute{
				Description: "Size of the file in bytes",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure configures the resource (no-op for ignition)
func (r *IgnitionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// No configuration needed - this resource doesn't use libvirt connection
}

// Create creates a new ignition file
func (r *IgnitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model IgnitionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := model.Name.ValueString()
	content := model.Content.ValueString()

	tflog.Debug(ctx, "Creating ignition file", map[string]any{
		"name": name,
	})

	// Create checksum for ID
	h := sha256.New()
	h.Write([]byte(content))
	checksum := fmt.Sprintf("%x", h.Sum(nil))
	model.ID = types.StringValue(checksum[:16]) // Use first 16 chars of checksum

	// Create temp directory for ignition files if it doesn't exist
	tmpDir := filepath.Join(os.TempDir(), "terraform-provider-libvirt-ignition")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create Temp Directory",
			fmt.Sprintf("Could not create directory for ignition files: %s", err),
		)
		return
	}

	// Generate file path using checksum for uniqueness
	filePath := filepath.Join(tmpDir, fmt.Sprintf("ignition-%s.ign", checksum[:16]))

	// Check if file already exists (for idempotency)
	if _, err := os.Stat(filePath); err == nil {
		tflog.Info(ctx, "Ignition file already exists, reusing", map[string]any{
			"path": filePath,
		})
	} else {
		// Write new file
		tflog.Debug(ctx, "Writing ignition file", map[string]any{
			"path": filePath,
		})

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			resp.Diagnostics.AddError(
				"Failed to Write File",
				fmt.Sprintf("Could not write ignition file: %s", err),
			)
			return
		}

		tflog.Info(ctx, "Written ignition file", map[string]any{
			"path": filePath,
		})
	}

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Stat File",
			fmt.Sprintf("Could not stat ignition file: %s", err),
		)
		return
	}

	model.Path = types.StringValue(filePath)
	model.Size = types.Int64Value(fileInfo.Size())

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read reads the ignition file state
func (r *IgnitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model IgnitionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filePath := model.Path.ValueString()

	// Check if file still exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			tflog.Warn(ctx, "Ignition file no longer exists, removing from state", map[string]any{
				"path": filePath,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Stat File",
			fmt.Sprintf("Could not stat ignition file: %s", err),
		)
		return
	}

	// Update size in case it changed
	model.Size = types.Int64Value(fileInfo.Size())

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update updates the ignition file (should trigger replacement)
func (r *IgnitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Ignition file cannot be updated. All changes require replacement.",
	)
}

// Delete deletes the ignition file
func (r *IgnitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model IgnitionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filePath := model.Path.ValueString()

	tflog.Debug(ctx, "Deleting ignition file", map[string]any{
		"path": filePath,
	})

	// Remove the file
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		resp.Diagnostics.AddError(
			"Failed to Delete File",
			fmt.Sprintf("Could not delete ignition file: %s", err),
		)
		return
	}

	tflog.Info(ctx, "Deleted ignition file", map[string]any{
		"path": filePath,
	})
}
