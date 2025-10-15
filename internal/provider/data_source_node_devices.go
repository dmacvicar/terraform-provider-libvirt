package provider

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	golibvirt "github.com/digitalocean/go-libvirt"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &NodeDevicesDataSource{}

func NewNodeDevicesDataSource() datasource.DataSource {
	return &NodeDevicesDataSource{}
}

type NodeDevicesDataSource struct {
	client *libvirt.Client
}

type NodeDevicesDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Capability types.String `tfsdk:"capability"`
	Devices    types.Set    `tfsdk:"devices"`
}

func (d *NodeDevicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_devices"
}

func (d *NodeDevicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Enumerates devices available on the libvirt host node.\n\n" +
			"This data source lists devices by capability type, useful for discovering " +
			"PCI devices for passthrough, USB devices, network interfaces, storage devices, and more.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Internal identifier for this data source.",
			},
			"capability": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Filter devices by capability type. Valid values include:\n\n" +
					"  - `system` - Host system information\n" +
					"  - `pci` - PCI devices\n" +
					"  - `usb_device` - USB devices\n" +
					"  - `usb` - USB host controllers\n" +
					"  - `net` - Network interfaces\n" +
					"  - `scsi_host` - SCSI host adapters\n" +
					"  - `scsi` - SCSI devices\n" +
					"  - `storage` - Storage devices\n" +
					"  - `drm` - DRM devices\n" +
					"  - `mdev` - Mediated devices\n" +
					"  - `ccw` - s390 CCW devices\n" +
					"  - `css` - s390 CSS devices\n" +
					"  - `ap_queue` - s390 AP queue devices\n" +
					"  - `ap_card` - s390 AP card devices\n" +
					"  - `ap_matrix` - s390 AP matrix devices\n" +
					"  - `ccw_group` - s390 CCW group devices\n\n" +
					"If not specified, all devices are returned.",
			},
			"devices": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of device names. Device names are in libvirt's internal format (e.g., `pci_0000_00_1f_2`, `net_eth0_00_11_22_33_44_55`).",
			},
		},
	}
}

func (d *NodeDevicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*libvirt.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *libvirt.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *NodeDevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NodeDevicesDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare capability filter (optional)
	var cap golibvirt.OptString
	if !data.Capability.IsNull() && !data.Capability.IsUnknown() {
		cap = append(cap, data.Capability.ValueString())
	}

	// Get number of devices
	numDevices, err := d.client.Libvirt().NodeNumOfDevices(cap, 0)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get device count",
			fmt.Sprintf("Unable to retrieve number of devices: %s", err),
		)
		return
	}

	// List devices
	deviceNames, err := d.client.Libvirt().NodeListDevices(cap, numDevices, 0)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list devices",
			fmt.Sprintf("Unable to list devices: %s", err),
		)
		return
	}

	// Sort devices for consistent ordering
	sort.Strings(deviceNames)

	// Convert to Terraform set
	deviceElements := make([]types.String, len(deviceNames))
	for i, name := range deviceNames {
		deviceElements[i] = types.StringValue(name)
	}

	devicesSet, diags := types.SetValueFrom(ctx, types.StringType, deviceElements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Devices = devicesSet

	// Generate ID
	capFilter := "all"
	if !data.Capability.IsNull() {
		capFilter = data.Capability.ValueString()
	}
	idStr := fmt.Sprintf("%s-%d", capFilter, len(deviceNames))
	data.ID = types.StringValue(strconv.Itoa(hashString(idStr)))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
