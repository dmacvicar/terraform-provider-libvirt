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
var _ resource.Resource = &CombustionResource{}

// CombustionResource defines the resource implementation
type CombustionResource struct {
	// No client needed - this is a local file operation
}

// CombustionResourceModel describes the resource data model
type CombustionResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Content types.String `tfsdk:"content"`
	Path    types.String `tfsdk:"path"`
	Size    types.Int64  `tfsdk:"size"`
}

// NewCombustionResource creates a new combustion resource
func NewCombustionResource() resource.Resource {
	return &CombustionResource{}
}

// Metadata returns the resource type name
func (r *CombustionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_combustion"
}

// Schema defines the resource schema
func (r *CombustionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Generates a Combustion script file that can be uploaded to a libvirt volume.",
		MarkdownDescription: `
Generates a Combustion script file for openSUSE MicroOS/Elemental systems.

Combustion is a minimal provisioning framework that runs shell scripts on first boot.
This resource generates the script file that can be uploaded to a volume and provided
to the virtual machine.

## Example Usage

` + "```hcl" + `
resource "libvirt_combustion" "microos" {
  name    = "microos-combustion"
  content = <<-EOF
    #!/bin/bash
    # combustion: network
    echo "root:password" | chpasswd
    systemctl enable sshd
  EOF
}

resource "libvirt_volume" "combustion" {
  name   = "microos-combustion.sh"
  pool   = "default"
  format = "raw"

  create = {
    content = {
      url = libvirt_combustion.microos.path
    }
  }
}
` + "```" + `

See the [Combustion documentation](https://github.com/openSUSE/combustion) for script format details.
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
				Description: "Name for this combustion resource",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				Description: "Combustion script content (shell script)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"path": schema.StringAttribute{
				Description: "Full path to the generated combustion script file",
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

// Configure configures the resource (no-op for combustion)
func (r *CombustionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// No configuration needed - this resource doesn't use libvirt connection
}

// Create creates a new combustion script file
func (r *CombustionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model CombustionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := model.Name.ValueString()
	content := model.Content.ValueString()

	tflog.Debug(ctx, "Creating combustion script", map[string]any{
		"name": name,
	})

	// Create checksum for ID
	h := sha256.New()
	h.Write([]byte(content))
	checksum := fmt.Sprintf("%x", h.Sum(nil))
	model.ID = types.StringValue(checksum[:16]) // Use first 16 chars of checksum

	// Create temp directory for combustion scripts if it doesn't exist
	tmpDir := filepath.Join(os.TempDir(), "terraform-provider-libvirt-combustion")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create Temp Directory",
			fmt.Sprintf("Could not create directory for combustion scripts: %s", err),
		)
		return
	}

	// Generate file path using checksum for uniqueness
	filePath := filepath.Join(tmpDir, fmt.Sprintf("script-%s.sh", checksum[:16]))

	// Check if file already exists (for idempotency)
	if _, err := os.Stat(filePath); err == nil {
		tflog.Info(ctx, "Combustion script already exists, reusing", map[string]any{
			"path": filePath,
		})
	} else {
		// Write new file (make it executable)
		tflog.Debug(ctx, "Writing combustion script", map[string]any{
			"path": filePath,
		})

		if err := os.WriteFile(filePath, []byte(content), 0755); err != nil {
			resp.Diagnostics.AddError(
				"Failed to Write File",
				fmt.Sprintf("Could not write combustion script: %s", err),
			)
			return
		}

		tflog.Info(ctx, "Written combustion script", map[string]any{
			"path": filePath,
		})
	}

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Stat File",
			fmt.Sprintf("Could not stat combustion script: %s", err),
		)
		return
	}

	model.Path = types.StringValue(filePath)
	model.Size = types.Int64Value(fileInfo.Size())

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read reads the combustion script state
func (r *CombustionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model CombustionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filePath := model.Path.ValueString()

	// Check if file still exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			tflog.Warn(ctx, "Combustion script no longer exists, removing from state", map[string]any{
				"path": filePath,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Stat File",
			fmt.Sprintf("Could not stat combustion script: %s", err),
		)
		return
	}

	// Update size in case it changed
	model.Size = types.Int64Value(fileInfo.Size())

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update updates the combustion script (should trigger replacement)
func (r *CombustionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Combustion script cannot be updated. All changes require replacement.",
	)
}

// Delete deletes the combustion script
func (r *CombustionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model CombustionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filePath := model.Path.ValueString()

	tflog.Debug(ctx, "Deleting combustion script", map[string]any{
		"path": filePath,
	})

	// Remove the file
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		resp.Diagnostics.AddError(
			"Failed to Delete File",
			fmt.Sprintf("Could not delete combustion script: %s", err),
		)
		return
	}

	tflog.Info(ctx, "Deleted combustion script", map[string]any{
		"path": filePath,
	})
}
