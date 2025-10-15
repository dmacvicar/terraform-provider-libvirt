package provider

import (
	"context"
	"fmt"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"libvirt.org/go/libvirtxml"
)

var _ datasource.DataSource = &NodeDeviceInfoDataSource{}

func NewNodeDeviceInfoDataSource() datasource.DataSource {
	return &NodeDeviceInfoDataSource{}
}

type NodeDeviceInfoDataSource struct {
	client *libvirt.Client
}

type NodeDeviceInfoDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Path       types.String `tfsdk:"path"`
	Parent     types.String `tfsdk:"parent"`
	Capability types.Object `tfsdk:"capability"`
}

func (d *NodeDeviceInfoDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_device_info"
}

func (d *NodeDeviceInfoDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches detailed information about a specific libvirt host node device.\n\n" +
			"This data source provides comprehensive details about hardware devices, " +
			"including PCI devices for passthrough, USB devices, network interfaces, and storage devices.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Internal identifier for this data source.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Device name from `libvirt_node_devices` data source (e.g., `pci_0000_00_1f_2`).",
			},
			"path": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Sysfs path to the device.",
			},
			"parent": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Parent device name in the device hierarchy.",
			},
			"capability": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Device capability details. Fields populated depend on the device type.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Capability type (e.g., `pci`, `usb_device`, `net`, `storage`).",
					},
					// PCI device fields
					"domain": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "PCI domain number.",
					},
					"bus": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "PCI/USB bus number.",
					},
					"slot": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "PCI slot number.",
					},
					"function": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "PCI function number.",
					},
					"class": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "PCI class code (e.g., `0x030000` for VGA).",
					},
					"product_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Product ID (vendor-specific identifier).",
					},
					"product_name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Product name or description.",
					},
					"vendor_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Vendor ID.",
					},
					"vendor_name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Vendor name.",
					},
					"iommu_group": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "IOMMU group number (for PCI passthrough).",
					},
					// USB device fields
					"device_number": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "USB device number.",
					},
					// Network interface fields
					"interface": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Network interface name (e.g., `eth0`).",
					},
					"address": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "MAC address or device address.",
					},
					"link_speed": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Network link speed.",
					},
					"link_state": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Network link state (e.g., `up`, `down`).",
					},
					// Storage device fields
					"block": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Block device path (e.g., `/dev/sda`).",
					},
					"drive_type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Drive type (e.g., `disk`, `cdrom`).",
					},
					"model": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Device model name.",
					},
					"serial": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Device serial number.",
					},
					"size": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Storage capacity in bytes.",
					},
					"logical_block_size": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Logical block size in bytes.",
					},
					"num_blocks": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Number of blocks.",
					},
					// SCSI device fields
					"host": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "SCSI host number.",
					},
					"target": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "SCSI target number.",
					},
					"lun": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "SCSI LUN (Logical Unit Number).",
					},
					"scsi_type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "SCSI device type.",
					},
				},
			},
		},
	}
}

func (d *NodeDeviceInfoDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NodeDeviceInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NodeDeviceInfoDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceName := data.Name.ValueString()

	// Get device XML
	xmlDesc, err := d.client.Libvirt().NodeDeviceGetXMLDesc(deviceName, 0)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get device information",
			fmt.Sprintf("Unable to retrieve device %s: %s", deviceName, err),
		)
		return
	}

	// Parse XML using libvirtxml
	var device libvirtxml.NodeDevice
	if err := device.Unmarshal(xmlDesc); err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse device XML",
			fmt.Sprintf("Unable to parse device XML for %s: %s", deviceName, err),
		)
		return
	}

	// Populate basic fields
	data.Path = types.StringValue(device.Path)
	data.Parent = types.StringValue(device.Parent)

	// Convert capability to Terraform object
	capability, diags := convertNodeDeviceCapability(ctx, device.Capability)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Capability = capability

	// Use device name as ID
	data.ID = types.StringValue(deviceName)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func convertNodeDeviceCapability(ctx context.Context, cap libvirtxml.NodeDeviceCapability) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Define attribute types for the capability object
	attrTypes := map[string]attr.Type{
		"type":               types.StringType,
		"domain":             types.Int64Type,
		"bus":                types.Int64Type,
		"slot":               types.Int64Type,
		"function":           types.Int64Type,
		"class":              types.StringType,
		"product_id":         types.StringType,
		"product_name":       types.StringType,
		"vendor_id":          types.StringType,
		"vendor_name":        types.StringType,
		"iommu_group":        types.Int64Type,
		"device_number":      types.Int64Type,
		"interface":          types.StringType,
		"address":            types.StringType,
		"link_speed":         types.StringType,
		"link_state":         types.StringType,
		"block":              types.StringType,
		"drive_type":         types.StringType,
		"model":              types.StringType,
		"serial":             types.StringType,
		"size":               types.Int64Type,
		"logical_block_size": types.Int64Type,
		"num_blocks":         types.Int64Type,
		"host":               types.Int64Type,
		"target":             types.Int64Type,
		"lun":                types.Int64Type,
		"scsi_type":          types.StringType,
	}

	// Initialize all attributes as null
	attrs := make(map[string]attr.Value)
	for key, attrType := range attrTypes {
		switch attrType {
		case types.StringType:
			attrs[key] = types.StringNull()
		case types.Int64Type:
			attrs[key] = types.Int64Null()
		}
	}

	// Populate fields based on capability type
	switch {
	case cap.PCI != nil:
		attrs["type"] = types.StringValue("pci")
		if cap.PCI.Domain != nil {
			attrs["domain"] = types.Int64Value(int64(*cap.PCI.Domain))
		}
		if cap.PCI.Bus != nil {
			attrs["bus"] = types.Int64Value(int64(*cap.PCI.Bus))
		}
		if cap.PCI.Slot != nil {
			attrs["slot"] = types.Int64Value(int64(*cap.PCI.Slot))
		}
		if cap.PCI.Function != nil {
			attrs["function"] = types.Int64Value(int64(*cap.PCI.Function))
		}
		if cap.PCI.Class != "" {
			attrs["class"] = types.StringValue(cap.PCI.Class)
		}
		if cap.PCI.Product.ID != "" {
			attrs["product_id"] = types.StringValue(cap.PCI.Product.ID)
		}
		if cap.PCI.Product.Name != "" {
			attrs["product_name"] = types.StringValue(cap.PCI.Product.Name)
		}
		if cap.PCI.Vendor.ID != "" {
			attrs["vendor_id"] = types.StringValue(cap.PCI.Vendor.ID)
		}
		if cap.PCI.Vendor.Name != "" {
			attrs["vendor_name"] = types.StringValue(cap.PCI.Vendor.Name)
		}
		if cap.PCI.IOMMUGroup != nil {
			attrs["iommu_group"] = types.Int64Value(int64(cap.PCI.IOMMUGroup.Number))
		}

	case cap.USBDevice != nil:
		attrs["type"] = types.StringValue("usb_device")
		attrs["bus"] = types.Int64Value(int64(cap.USBDevice.Bus))
		attrs["device_number"] = types.Int64Value(int64(cap.USBDevice.Device))
		if cap.USBDevice.Product.ID != "" {
			attrs["product_id"] = types.StringValue(cap.USBDevice.Product.ID)
		}
		if cap.USBDevice.Product.Name != "" {
			attrs["product_name"] = types.StringValue(cap.USBDevice.Product.Name)
		}
		if cap.USBDevice.Vendor.ID != "" {
			attrs["vendor_id"] = types.StringValue(cap.USBDevice.Vendor.ID)
		}
		if cap.USBDevice.Vendor.Name != "" {
			attrs["vendor_name"] = types.StringValue(cap.USBDevice.Vendor.Name)
		}

	case cap.Net != nil:
		attrs["type"] = types.StringValue("net")
		if cap.Net.Interface != "" {
			attrs["interface"] = types.StringValue(cap.Net.Interface)
		}
		if cap.Net.Address != "" {
			attrs["address"] = types.StringValue(cap.Net.Address)
		}
		if cap.Net.Link != nil {
			if cap.Net.Link.Speed != "" {
				attrs["link_speed"] = types.StringValue(cap.Net.Link.Speed)
			}
			if cap.Net.Link.State != "" {
				attrs["link_state"] = types.StringValue(cap.Net.Link.State)
			}
		}

	case cap.Storage != nil:
		attrs["type"] = types.StringValue("storage")
		if cap.Storage.Block != "" {
			attrs["block"] = types.StringValue(cap.Storage.Block)
		}
		if cap.Storage.Bus != "" {
			attrs["bus"] = types.StringValue(cap.Storage.Bus)
		}
		if cap.Storage.DriverType != "" {
			attrs["drive_type"] = types.StringValue(cap.Storage.DriverType)
		}
		if cap.Storage.Model != "" {
			attrs["model"] = types.StringValue(cap.Storage.Model)
		}
		if cap.Storage.Vendor != "" {
			attrs["vendor_name"] = types.StringValue(cap.Storage.Vendor)
		}
		if cap.Storage.Serial != "" {
			attrs["serial"] = types.StringValue(cap.Storage.Serial)
		}
		if cap.Storage.Size != nil {
			attrs["size"] = types.Int64Value(int64(*cap.Storage.Size))
		}
		if cap.Storage.LogicalBlockSize != nil {
			attrs["logical_block_size"] = types.Int64Value(int64(*cap.Storage.LogicalBlockSize))
		}
		if cap.Storage.NumBlocks != nil {
			attrs["num_blocks"] = types.Int64Value(int64(*cap.Storage.NumBlocks))
		}

	case cap.SCSI != nil:
		attrs["type"] = types.StringValue("scsi")
		attrs["host"] = types.Int64Value(int64(cap.SCSI.Host))
		attrs["bus"] = types.Int64Value(int64(cap.SCSI.Bus))
		attrs["target"] = types.Int64Value(int64(cap.SCSI.Target))
		attrs["lun"] = types.Int64Value(int64(cap.SCSI.Lun))
		if cap.SCSI.Type != "" {
			attrs["scsi_type"] = types.StringValue(cap.SCSI.Type)
		}

	default:
		// For other capability types, just set the type if we can determine it
		if cap.System != nil {
			attrs["type"] = types.StringValue("system")
		} else if cap.USB != nil {
			attrs["type"] = types.StringValue("usb")
		} else if cap.SCSIHost != nil {
			attrs["type"] = types.StringValue("scsi_host")
			attrs["host"] = types.Int64Value(int64(cap.SCSIHost.Host))
		} else if cap.DRM != nil {
			attrs["type"] = types.StringValue("drm")
		} else {
			attrs["type"] = types.StringValue("unknown")
		}
	}

	obj, objDiags := types.ObjectValue(attrTypes, attrs)
	diags.Append(objDiags...)

	return obj, diags
}
