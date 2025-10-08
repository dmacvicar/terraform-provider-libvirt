package provider

import (
	"context"
	"fmt"
	"net"
	"strings"

	libvirtclient "github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"libvirt.org/go/libvirtxml"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &NetworkResource{}
var _ resource.ResourceWithImportState = &NetworkResource{}

func NewNetworkResource() resource.Resource {
	return &NetworkResource{}
}

// NetworkResource defines the resource implementation.
type NetworkResource struct {
	client *libvirtclient.Client
}

// NetworkResourceModel describes the resource data model.
type NetworkResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	UUID      types.String `tfsdk:"uuid"`
	Mode      types.String `tfsdk:"mode"`
	Bridge    types.String `tfsdk:"bridge"`
	Addresses types.List   `tfsdk:"addresses"`
	Autostart types.Bool   `tfsdk:"autostart"`
}

func (r *NetworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *NetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a libvirt virtual network.",
		Description:         "Manages a libvirt virtual network.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Network identifier (UUID)",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Network name. Must be unique on the host.",
				Required:    true,
			},
			"uuid": schema.StringAttribute{
				Description: "Network UUID. If not specified, one will be generated.",
				Optional:    true,
				Computed:    true,
			},
			"mode": schema.StringAttribute{
				Description: "Network forwarding mode: 'nat' (default), 'none' (isolated), 'route', 'open', 'bridge'.",
				Optional:    true,
			},
			"bridge": schema.StringAttribute{
				Description: "Bridge name. If not specified, libvirt will auto-generate one.",
				Optional:    true,
				Computed:    true,
			},
			"addresses": schema.ListAttribute{
				Description: "List of IPv4/IPv6 CIDR addresses for the network (e.g., '10.17.3.0/24', 'fd00::/64').",
				Optional:    true,
				ElementType: types.StringType,
			},
			"autostart": schema.BoolAttribute{
				Description: "Whether the network should be started automatically when the host boots.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *NetworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*libvirtclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *libvirt.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model NetworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build libvirt network XML
	network := libvirtxml.Network{
		Name: model.Name.ValueString(),
	}

	// Set UUID if provided
	if !model.UUID.IsNull() && !model.UUID.IsUnknown() {
		network.UUID = model.UUID.ValueString()
	}

	// Set forward mode
	mode := "nat" // default
	if !model.Mode.IsNull() && !model.Mode.IsUnknown() {
		mode = model.Mode.ValueString()
	}

	if mode != "none" {
		network.Forward = &libvirtxml.NetworkForward{
			Mode: mode,
		}
	}

	// Set bridge name if provided
	if !model.Bridge.IsNull() && !model.Bridge.IsUnknown() {
		network.Bridge = &libvirtxml.NetworkBridge{
			Name: model.Bridge.ValueString(),
		}
	}

	// Process addresses (IP ranges)
	if !model.Addresses.IsNull() && !model.Addresses.IsUnknown() {
		var addresses []string
		diags := model.Addresses.ElementsAs(ctx, &addresses, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, addr := range addresses {
			ip, ipNet, err := net.ParseCIDR(addr)
			if err != nil {
				resp.Diagnostics.AddError(
					"Invalid Address",
					fmt.Sprintf("Failed to parse address %q: %s", addr, err),
				)
				return
			}

			networkIP := libvirtxml.NetworkIP{
				Address: ip.String(),
			}

			// Determine if IPv4 or IPv6
			if ip.To4() != nil {
				networkIP.Family = "ipv4"
				ones, _ := ipNet.Mask.Size()
				networkIP.Prefix = uint(ones)
			} else {
				networkIP.Family = "ipv6"
				ones, _ := ipNet.Mask.Size()
				networkIP.Prefix = uint(ones)
			}

			network.IPs = append(network.IPs, networkIP)
		}
	}

	// Marshal to XML
	xmlDoc, err := network.Marshal()
	if err != nil {
		resp.Diagnostics.AddError(
			"XML Marshal Failed",
			fmt.Sprintf("Failed to marshal network XML: %s", err),
		)
		return
	}

	// Define the network in libvirt
	net, err := r.client.Libvirt().NetworkDefineXML(xmlDoc)
	if err != nil {
		resp.Diagnostics.AddError(
			"Network Creation Failed",
			fmt.Sprintf("Failed to define network in libvirt: %s", err),
		)
		return
	}

	// Set autostart if requested
	autostart := int32(0)
	if !model.Autostart.IsNull() && !model.Autostart.IsUnknown() {
		if model.Autostart.ValueBool() {
			autostart = 1
		}
	}

	if err := r.client.Libvirt().NetworkSetAutostart(net, autostart); err != nil {
		// Non-fatal, just warn
		resp.Diagnostics.AddWarning(
			"Autostart Configuration Failed",
			fmt.Sprintf("Network created but failed to set autostart: %s", err),
		)
	}

	// Start the network
	if err := r.client.Libvirt().NetworkCreate(net); err != nil {
		resp.Diagnostics.AddError(
			"Network Start Failed",
			fmt.Sprintf("Network defined but failed to start: %s", err),
		)
		return
	}

	// Read back the network to get computed values
	uuidStr := libvirtclient.UUIDString(net.UUID)
	model.ID = types.StringValue(uuidStr)
	model.UUID = types.StringValue(uuidStr)

	// Read full network details
	if err := r.readNetwork(ctx, &model, uuidStr); err != nil {
		resp.Diagnostics.AddError(
			"Network Read Failed",
			fmt.Sprintf("Network created but failed to read back: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model NetworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := model.ID.ValueString()
	if uuid == "" {
		uuid = model.UUID.ValueString()
	}

	if err := r.readNetwork(ctx, &model, uuid); err != nil {
		if strings.Contains(err.Error(), "Network not found") {
			// Network was deleted outside of Terraform
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Network Read Failed",
			fmt.Sprintf("Failed to read network: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model NetworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state to check what changed
	var state NetworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := state.ID.ValueString()

	// Lookup the network
	net, err := r.client.LookupNetworkByUUID(uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Network Lookup Failed",
			fmt.Sprintf("Failed to find network: %s", err),
		)
		return
	}

	// Handle autostart changes
	if !model.Autostart.Equal(state.Autostart) {
		autostart := int32(0)
		if !model.Autostart.IsNull() && !model.Autostart.IsUnknown() {
			if model.Autostart.ValueBool() {
				autostart = 1
			}
		}

		if err := r.client.Libvirt().NetworkSetAutostart(net, autostart); err != nil {
			resp.Diagnostics.AddError(
				"Autostart Update Failed",
				fmt.Sprintf("Failed to update autostart: %s", err),
			)
			return
		}
	}

	// Read back to get current state
	if err := r.readNetwork(ctx, &model, uuid); err != nil {
		resp.Diagnostics.AddError(
			"Network Read Failed",
			fmt.Sprintf("Network updated but failed to read back: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model NetworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := model.ID.ValueString()

	// Lookup the network
	net, err := r.client.LookupNetworkByUUID(uuid)
	if err != nil {
		// Already deleted
		if strings.Contains(err.Error(), "Network not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Network Lookup Failed",
			fmt.Sprintf("Failed to find network for deletion: %s", err),
		)
		return
	}

	// Check if network is active
	active, err := r.client.Libvirt().NetworkIsActive(net)
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Network Status Check Failed",
			fmt.Sprintf("Failed to check network status: %s", err),
		)
	}

	// Destroy (stop) the network if it's active
	if active == 1 {
		if err := r.client.Libvirt().NetworkDestroy(net); err != nil {
			resp.Diagnostics.AddError(
				"Network Stop Failed",
				fmt.Sprintf("Failed to stop network: %s", err),
			)
			return
		}
	}

	// Undefine (delete) the network
	if err := r.client.Libvirt().NetworkUndefine(net); err != nil {
		resp.Diagnostics.AddError(
			"Network Delete Failed",
			fmt.Sprintf("Failed to delete network: %s", err),
		)
		return
	}
}

func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readNetwork reads network details from libvirt and populates the model
func (r *NetworkResource) readNetwork(ctx context.Context, model *NetworkResourceModel, uuid string) error {
	// Lookup the network by UUID
	net, err := r.client.LookupNetworkByUUID(uuid)
	if err != nil {
		return fmt.Errorf("network not found: %w", err)
	}

	// Get network XML
	xmlDoc, err := r.client.Libvirt().NetworkGetXMLDesc(net, 0)
	if err != nil {
		return fmt.Errorf("failed to get network XML: %w", err)
	}

	// Unmarshal XML
	var network libvirtxml.Network
	if err := network.Unmarshal(xmlDoc); err != nil {
		return fmt.Errorf("failed to unmarshal network XML: %w", err)
	}

	// Populate model
	model.ID = types.StringValue(network.UUID)
	model.UUID = types.StringValue(network.UUID)
	model.Name = types.StringValue(network.Name)

	// Mode
	if network.Forward != nil && network.Forward.Mode != "" {
		model.Mode = types.StringValue(network.Forward.Mode)
	} else {
		model.Mode = types.StringValue("none")
	}

	// Bridge name
	if network.Bridge != nil && network.Bridge.Name != "" {
		model.Bridge = types.StringValue(network.Bridge.Name)
	} else {
		model.Bridge = types.StringNull()
	}

	// Addresses
	var addresses []string
	for _, ip := range network.IPs {
		if ip.Address != "" && ip.Prefix > 0 {
			addr := fmt.Sprintf("%s/%d", ip.Address, ip.Prefix)
			addresses = append(addresses, addr)
		}
	}
	if len(addresses) > 0 {
		addressList, diag := types.ListValueFrom(ctx, types.StringType, addresses)
		if diag.HasError() {
			return fmt.Errorf("failed to convert addresses: %v", diag.Errors())
		}
		model.Addresses = addressList
	} else {
		model.Addresses = types.ListNull(types.StringType)
	}

	// Autostart
	autostart, err := r.client.Libvirt().NetworkGetAutostart(net)
	if err != nil {
		// Non-fatal, just set to false
		model.Autostart = types.BoolValue(false)
	} else {
		model.Autostart = types.BoolValue(autostart == 1)
	}

	return nil
}
