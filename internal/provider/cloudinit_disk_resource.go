package provider

import (
	"bytes"
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
	"github.com/kdomanski/iso9660"
)

// Ensure the implementation satisfies the resource.Resource interface
var _ resource.Resource = &CloudInitDiskResource{}

// CloudInitDiskResource defines the resource implementation
type CloudInitDiskResource struct {
	// No client needed - this is a local file operation
}

// CloudInitDiskResourceModel describes the resource data model
type CloudInitDiskResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	UserData      types.String `tfsdk:"user_data"`
	MetaData      types.String `tfsdk:"meta_data"`
	NetworkConfig types.String `tfsdk:"network_config"`
	Path          types.String `tfsdk:"path"`
	Size          types.Int64  `tfsdk:"size"`
}

// NewCloudInitDiskResource creates a new cloud-init disk resource
func NewCloudInitDiskResource() resource.Resource {
	return &CloudInitDiskResource{}
}

// Metadata returns the resource type name
func (r *CloudInitDiskResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudinit_disk"
}

// Schema defines the resource schema
func (r *CloudInitDiskResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Generates a cloud-init ISO disk image that can be attached to a domain.",
		MarkdownDescription: `
Generates a cloud-init configuration disk as an ISO image with the "cidata" volume label.
This ISO can be uploaded to a libvirt volume and attached to a domain for cloud-init configuration.

Cloud-init will automatically detect and process the configuration from this disk when the VM boots.

## Example Usage

` + "```hcl" + `
resource "libvirt_cloudinit_disk" "init" {
  name      = "vm-init"
  user_data = file("user-data.yaml")
  meta_data = yamlencode({
    instance-id    = "vm-01"
    local-hostname = "webserver"
  })
}

resource "libvirt_volume" "cloudinit" {
  name   = "vm-cloudinit"
  pool   = "default"
  format = "raw"

  create = {
    content = {
      url = libvirt_cloudinit_disk.init.path
    }
  }
}
` + "```" + `

See the [cloud-init documentation](https://cloudinit.readthedocs.io/) for configuration format details.
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
				Description: "Name for this cloud-init disk resource",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_data": schema.StringAttribute{
				Description: "Cloud-init user-data content (usually YAML)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"meta_data": schema.StringAttribute{
				Description: "Cloud-init meta-data content (usually YAML)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_config": schema.StringAttribute{
				Description: "Cloud-init network configuration (optional, usually YAML)",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"path": schema.StringAttribute{
				Description: "Full path to the generated ISO file",
				Computed:    true,
			},
			"size": schema.Int64Attribute{
				Description: "Size of the ISO file in bytes",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure configures the resource (no-op for cloud-init disk)
func (r *CloudInitDiskResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// No configuration needed - this resource doesn't use libvirt connection
}

// Create creates a new cloud-init disk
func (r *CloudInitDiskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model CloudInitDiskResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := model.Name.ValueString()
	userData := model.UserData.ValueString()
	metaData := model.MetaData.ValueString()

	tflog.Debug(ctx, "Creating cloud-init disk", map[string]any{
		"name": name,
	})

	// Create checksum for ID
	h := sha256.New()
	h.Write([]byte(userData))
	h.Write([]byte(metaData))
	if !model.NetworkConfig.IsNull() {
		h.Write([]byte(model.NetworkConfig.ValueString()))
	}
	checksum := fmt.Sprintf("%x", h.Sum(nil))
	model.ID = types.StringValue(checksum[:16]) // Use first 16 chars of checksum

	// Create temp directory for cloud-init ISOs if it doesn't exist
	tmpDir := filepath.Join(os.TempDir(), "terraform-provider-libvirt-cloudinit")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create Temp Directory",
			fmt.Sprintf("Could not create directory for cloud-init ISOs: %s", err),
		)
		return
	}

	// Generate ISO path using checksum for uniqueness
	isoPath := filepath.Join(tmpDir, fmt.Sprintf("cloudinit-%s.iso", checksum[:16]))

	// Check if ISO already exists (for idempotency)
	if _, err := os.Stat(isoPath); err == nil {
		tflog.Info(ctx, "Cloud-init ISO already exists, reusing", map[string]any{
			"path": isoPath,
		})
	} else {
		// Create new ISO
		tflog.Debug(ctx, "Generating cloud-init ISO", map[string]any{
			"path": isoPath,
		})

		if err := r.generateCloudInitISO(ctx, isoPath, userData, metaData, model.NetworkConfig); err != nil {
			resp.Diagnostics.AddError(
				"Failed to Generate ISO",
				fmt.Sprintf("Could not generate cloud-init ISO: %s", err),
			)
			return
		}

		tflog.Info(ctx, "Generated cloud-init ISO", map[string]any{
			"path": isoPath,
		})
	}

	// Get file size
	fileInfo, err := os.Stat(isoPath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Stat ISO",
			fmt.Sprintf("Could not stat generated ISO: %s", err),
		)
		return
	}

	model.Path = types.StringValue(isoPath)
	model.Size = types.Int64Value(fileInfo.Size())

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read reads the cloud-init disk state
func (r *CloudInitDiskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model CloudInitDiskResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	isoPath := model.Path.ValueString()

	// Check if ISO file still exists
	fileInfo, err := os.Stat(isoPath)
	if err != nil {
		if os.IsNotExist(err) {
			tflog.Warn(ctx, "Cloud-init ISO no longer exists, removing from state", map[string]any{
				"path": isoPath,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Stat ISO",
			fmt.Sprintf("Could not stat cloud-init ISO: %s", err),
		)
		return
	}

	// Update size in case it changed
	model.Size = types.Int64Value(fileInfo.Size())

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update updates the cloud-init disk (should trigger replacement)
func (r *CloudInitDiskResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Cloud-init disk cannot be updated. All changes require replacement.",
	)
}

// Delete deletes the cloud-init disk
func (r *CloudInitDiskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model CloudInitDiskResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	isoPath := model.Path.ValueString()

	tflog.Debug(ctx, "Deleting cloud-init ISO", map[string]any{
		"path": isoPath,
	})

	// Remove the ISO file
	if err := os.Remove(isoPath); err != nil && !os.IsNotExist(err) {
		resp.Diagnostics.AddError(
			"Failed to Delete ISO",
			fmt.Sprintf("Could not delete cloud-init ISO: %s", err),
		)
		return
	}

	tflog.Info(ctx, "Deleted cloud-init ISO", map[string]any{
		"path": isoPath,
	})
}

// generateCloudInitISO generates a cloud-init ISO with the given content
func (r *CloudInitDiskResource) generateCloudInitISO(ctx context.Context, isoPath, userData, metaData string, networkConfig types.String) error {
	// Create a new ISO writer
	writer, err := iso9660.NewWriter()
	if err != nil {
		return fmt.Errorf("failed to create ISO writer: %w", err)
	}
	defer func() {
		if err := writer.Cleanup(); err != nil {
			tflog.Warn(ctx, "Failed to cleanup ISO writer", map[string]any{
				"error": err.Error(),
			})
		}
	}()

	// Add user-data file
	if err := writer.AddFile(bytes.NewReader([]byte(userData)), "user-data"); err != nil {
		return fmt.Errorf("failed to add user-data: %w", err)
	}

	// Add meta-data file
	if err := writer.AddFile(bytes.NewReader([]byte(metaData)), "meta-data"); err != nil {
		return fmt.Errorf("failed to add meta-data: %w", err)
	}

	// Add network-config if provided
	if !networkConfig.IsNull() && !networkConfig.IsUnknown() {
		if err := writer.AddFile(bytes.NewReader([]byte(networkConfig.ValueString())), "network-config"); err != nil {
			return fmt.Errorf("failed to add network-config: %w", err)
		}
	}

	// Write ISO to file with "cidata" volume label (required for cloud-init detection)
	outputFile, err := os.OpenFile(isoPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to create ISO file: %w", err)
	}
	defer func() {
		if err := outputFile.Close(); err != nil {
			tflog.Warn(ctx, "Failed to close ISO file", map[string]any{
				"error": err.Error(),
			})
		}
	}()

	if err := writer.WriteTo(outputFile, "cidata"); err != nil {
		return fmt.Errorf("failed to write ISO: %w", err)
	}

	return nil
}
