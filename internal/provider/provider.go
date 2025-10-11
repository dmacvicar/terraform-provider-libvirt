package provider

import (
	"context"
	"os"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the provider.Provider interface
var _ provider.Provider = &LibvirtProvider{}

// LibvirtProvider defines the provider implementation
type LibvirtProvider struct {
	version string
}

// LibvirtProviderModel describes the provider data model
type LibvirtProviderModel struct {
	URI types.String `tfsdk:"uri"`
}

// New creates a new provider instance
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &LibvirtProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name
func (p *LibvirtProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "libvirt"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data
func (p *LibvirtProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for libvirt. Manage virtual machines, networks, and storage using libvirt.",
		MarkdownDescription: `
The Libvirt provider is used to interact with libvirt to manage virtual machines,
networks, storage pools, and other resources.

This provider follows the [libvirt XML schemas](https://libvirt.org/format.html) closely,
providing fine-grained control over all libvirt features.
`,
		Attributes: map[string]schema.Attribute{
			"uri": schema.StringAttribute{
				Description: "Libvirt connection URI. Defaults to qemu:///system if not specified. " +
					"See https://libvirt.org/uri.html for URI format documentation.",
				MarkdownDescription: "Libvirt connection URI. Defaults to `qemu:///system` if not specified. " +
					"See [libvirt URI documentation](https://libvirt.org/uri.html) for details.",
				Optional: true,
			},
		},
	}
}

// Configure prepares the provider for use
func (p *LibvirtProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Libvirt provider")

	var config LibvirtProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine URI with precedence: config > env var > default
	uri := "qemu:///system"

	// Check environment variable first
	if envURI := os.Getenv("LIBVIRT_DEFAULT_URI"); envURI != "" {
		uri = envURI
	}

	// Config overrides environment variable
	if !config.URI.IsNull() && !config.URI.IsUnknown() {
		uri = config.URI.ValueString()
	}

	tflog.Debug(ctx, "Connecting to libvirt", map[string]any{
		"uri": uri,
	})

	// Create libvirt client connection
	client, err := libvirt.NewClient(ctx, uri)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Connect to Libvirt",
			"An error occurred while connecting to libvirt.\n\n"+
				"URI: "+uri+"\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Verify the connection works
	if err := client.Ping(ctx); err != nil {
		_ = client.Close()
		resp.Diagnostics.AddError(
			"Libvirt Connection Failed",
			"Connected to libvirt but the connection test failed.\n\n"+
				"URI: "+uri+"\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Make the client available to resources and data sources
	resp.DataSourceData = client
	resp.ResourceData = client
}

// Resources returns the list of resources supported by this provider
func (p *LibvirtProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDomainResource,
		NewPoolResource,
		NewVolumeResource,
		NewNetworkResource,
	}
}

// DataSources returns the list of data sources supported by this provider
func (p *LibvirtProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// TODO: Add data sources as they are implemented
	}
}
