package provider

import (
	"context"
	"fmt"
	"strings"

	libvirtclient "github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	IPs       types.List   `tfsdk:"ips"`
	Autostart types.Bool   `tfsdk:"autostart"`
}

// NetworkIPModel describes an IP configuration in the network
type NetworkIPModel struct {
	Address   types.String `tfsdk:"address"`
	Netmask   types.String `tfsdk:"netmask"`
	Prefix    types.Int64  `tfsdk:"prefix"`
	Family    types.String `tfsdk:"family"`
	LocalPtr  types.String `tfsdk:"local_ptr"`
	DHCP      types.Object `tfsdk:"dhcp"`
}

// NetworkDHCPModel describes DHCP configuration for an IP
type NetworkDHCPModel struct {
	Ranges types.List `tfsdk:"ranges"`
	Hosts  types.List `tfsdk:"hosts"`
}

// NetworkDHCPRangeModel describes a DHCP range
type NetworkDHCPRangeModel struct {
	Start types.String `tfsdk:"start"`
	End   types.String `tfsdk:"end"`
}

// NetworkDHCPHostModel describes a DHCP host entry
type NetworkDHCPHostModel struct {
	MAC    types.String `tfsdk:"mac"`
	IP     types.String `tfsdk:"ip"`
	Name   types.String `tfsdk:"name"`
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
			"ips": schema.ListNestedAttribute{
				Description: "IP address configurations for the network. Each entry can specify address, netmask/prefix, family, and DHCP settings.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							Description: "IP address (e.g., '10.17.3.1', 'fd00::1'). Optional - if not specified, one will be derived from the network.",
							Optional:    true,
						},
						"netmask": schema.StringAttribute{
							Description: "Network mask for IPv4 (e.g., '255.255.255.0'). Mutually exclusive with prefix.",
							Optional:    true,
						},
						"prefix": schema.Int64Attribute{
							Description: "Network prefix length (e.g., 24 for 255.255.255.0, 64 for IPv6). Mutually exclusive with netmask.",
							Optional:    true,
						},
						"family": schema.StringAttribute{
							Description: "Address family ('ipv4' or 'ipv6'). Optional - will be auto-detected from address if not specified.",
							Optional:    true,
						},
						"local_ptr": schema.StringAttribute{
							Description: "Whether to generate local PTR records ('yes' or 'no').",
							Optional:    true,
						},
						"dhcp": schema.SingleNestedAttribute{
							Description: "DHCP server configuration for this IP range.",
							Optional:    true,
							Attributes: map[string]schema.Attribute{
								"ranges": schema.ListNestedAttribute{
									Description: "DHCP address ranges to hand out.",
									Optional:    true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"start": schema.StringAttribute{
												Description: "Start IP address of the range.",
												Required:    true,
											},
											"end": schema.StringAttribute{
												Description: "End IP address of the range.",
												Required:    true,
											},
										},
									},
								},
								"hosts": schema.ListNestedAttribute{
									Description: "Static DHCP host entries.",
									Optional:    true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"mac": schema.StringAttribute{
												Description: "MAC address of the host.",
												Required:    true,
											},
											"ip": schema.StringAttribute{
												Description: "IP address to assign to this host.",
												Required:    true,
											},
											"name": schema.StringAttribute{
												Description: "Hostname for this host.",
												Optional:    true,
											},
										},
									},
								},
							},
						},
					},
				},
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

	// Process IP configurations
	if !model.IPs.IsNull() && !model.IPs.IsUnknown() {
		var ips []NetworkIPModel
		diags := model.IPs.ElementsAs(ctx, &ips, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, ipModel := range ips {
			networkIP := libvirtxml.NetworkIP{}

			// Set address if provided
			if !ipModel.Address.IsNull() && !ipModel.Address.IsUnknown() {
				networkIP.Address = ipModel.Address.ValueString()
			}

			// Set netmask if provided
			if !ipModel.Netmask.IsNull() && !ipModel.Netmask.IsUnknown() {
				networkIP.Netmask = ipModel.Netmask.ValueString()
			}

			// Set prefix if provided
			if !ipModel.Prefix.IsNull() && !ipModel.Prefix.IsUnknown() {
				networkIP.Prefix = uint(ipModel.Prefix.ValueInt64())
			}

			// Set family if provided
			if !ipModel.Family.IsNull() && !ipModel.Family.IsUnknown() {
				networkIP.Family = ipModel.Family.ValueString()
			}

			// Set local PTR if provided
			if !ipModel.LocalPtr.IsNull() && !ipModel.LocalPtr.IsUnknown() {
				networkIP.LocalPtr = ipModel.LocalPtr.ValueString()
			}

			// Process DHCP configuration if provided
			if !ipModel.DHCP.IsNull() && !ipModel.DHCP.IsUnknown() {
				var dhcpModel NetworkDHCPModel
				diags := ipModel.DHCP.As(ctx, &dhcpModel, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				dhcpConfig := libvirtxml.NetworkDHCP{}

				// Process DHCP ranges
				if !dhcpModel.Ranges.IsNull() && !dhcpModel.Ranges.IsUnknown() {
					var ranges []NetworkDHCPRangeModel
					diags := dhcpModel.Ranges.ElementsAs(ctx, &ranges, false)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}

					for _, rangeModel := range ranges {
						dhcpRange := libvirtxml.NetworkDHCPRange{
							Start: rangeModel.Start.ValueString(),
							End:   rangeModel.End.ValueString(),
						}

						
						dhcpConfig.Ranges = append(dhcpConfig.Ranges, dhcpRange)
					}
				}

				// Process DHCP hosts
				if !dhcpModel.Hosts.IsNull() && !dhcpModel.Hosts.IsUnknown() {
					var hosts []NetworkDHCPHostModel
					diags := dhcpModel.Hosts.ElementsAs(ctx, &hosts, false)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}

					for _, hostModel := range hosts {
						dhcpHost := libvirtxml.NetworkDHCPHost{
							MAC: hostModel.MAC.ValueString(),
							IP:  hostModel.IP.ValueString(),
						}

						if !hostModel.Name.IsNull() && !hostModel.Name.IsUnknown() {
							dhcpHost.Name = hostModel.Name.ValueString()
						}

						dhcpConfig.Hosts = append(dhcpConfig.Hosts, dhcpHost)
					}
				}

				// Only set DHCP if we have some configuration
				if len(dhcpConfig.Ranges) > 0 || len(dhcpConfig.Hosts) > 0 {
					networkIP.DHCP = &dhcpConfig
				}
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

	// Mode - only set if user originally specified it
	if !model.Mode.IsNull() && !model.Mode.IsUnknown() {
		if network.Forward != nil && network.Forward.Mode != "" {
			model.Mode = types.StringValue(network.Forward.Mode)
		} else {
			model.Mode = types.StringValue("none")
		}
	}

	// Bridge name
	if network.Bridge != nil && network.Bridge.Name != "" {
		model.Bridge = types.StringValue(network.Bridge.Name)
	} else {
		model.Bridge = types.StringNull()
	}

	// IPs
	var ips []NetworkIPModel
	for _, ip := range network.IPs {
		ipModel := NetworkIPModel{}

		// Set address
		if ip.Address != "" {
			ipModel.Address = types.StringValue(ip.Address)
		} else {
			ipModel.Address = types.StringNull()
		}

		// Set netmask
		if ip.Netmask != "" {
			ipModel.Netmask = types.StringValue(ip.Netmask)
		} else {
			ipModel.Netmask = types.StringNull()
		}

		// Set prefix
		if ip.Prefix > 0 {
			ipModel.Prefix = types.Int64Value(int64(ip.Prefix))
		} else {
			ipModel.Prefix = types.Int64Null()
		}

		// Set family
		if ip.Family != "" {
			ipModel.Family = types.StringValue(ip.Family)
		} else {
			ipModel.Family = types.StringNull()
		}

		// Set local PTR
		if ip.LocalPtr != "" {
			ipModel.LocalPtr = types.StringValue(ip.LocalPtr)
		} else {
			ipModel.LocalPtr = types.StringNull()
		}

		// Process DHCP configuration if present
		if ip.DHCP != nil {
			dhcpModel := NetworkDHCPModel{}

			// Process DHCP ranges
			if len(ip.DHCP.Ranges) > 0 {
				var ranges []NetworkDHCPRangeModel
				for _, dhcpRange := range ip.DHCP.Ranges {
					rangeModel := NetworkDHCPRangeModel{
						Start: types.StringValue(dhcpRange.Start),
						End:   types.StringValue(dhcpRange.End),
					}
					ranges = append(ranges, rangeModel)
				}

				rangesObjType := types.ObjectType{AttrTypes: getNetworkDHCPRangeAttrTypes()}
				rangesList, diag := types.ListValueFrom(ctx, rangesObjType, ranges)
				if diag.HasError() {
					return fmt.Errorf("failed to convert DHCP ranges: %v", diag.Errors())
				}
				dhcpModel.Ranges = rangesList
			} else {
				rangesObjType := types.ObjectType{AttrTypes: getNetworkDHCPRangeAttrTypes()}
				dhcpModel.Ranges = types.ListNull(rangesObjType)
			}

			// Process DHCP hosts
			if len(ip.DHCP.Hosts) > 0 {
				var hosts []NetworkDHCPHostModel
				for _, dhcpHost := range ip.DHCP.Hosts {
					hostModel := NetworkDHCPHostModel{
						MAC: types.StringValue(dhcpHost.MAC),
						IP:  types.StringValue(dhcpHost.IP),
					}
					if dhcpHost.Name != "" {
						hostModel.Name = types.StringValue(dhcpHost.Name)
					} else {
						hostModel.Name = types.StringNull()
					}
					hosts = append(hosts, hostModel)
				}

				hostsObjType := types.ObjectType{AttrTypes: getNetworkDHCPHostAttrTypes()}
				hostsList, diag := types.ListValueFrom(ctx, hostsObjType, hosts)
				if diag.HasError() {
					return fmt.Errorf("failed to convert DHCP hosts: %v", diag.Errors())
				}
				dhcpModel.Hosts = hostsList
			} else {
				hostsObjType := types.ObjectType{AttrTypes: getNetworkDHCPHostAttrTypes()}
				dhcpModel.Hosts = types.ListNull(hostsObjType)
			}

			
			// Create DHCP object
			dhcpObj, diag := types.ObjectValueFrom(ctx, getNetworkDHCPAttrTypes(), dhcpModel)
			if diag.HasError() {
				return fmt.Errorf("failed to convert DHCP configuration: %v", diag.Errors())
			}
			ipModel.DHCP = dhcpObj
		} else {
			ipModel.DHCP = types.ObjectNull(getNetworkDHCPAttrTypes())
		}

		ips = append(ips, ipModel)
	}

	if len(ips) > 0 {
		ipsObjType := types.ObjectType{AttrTypes: getNetworkIPAttrTypes()}
		ipsList, diag := types.ListValueFrom(ctx, ipsObjType, ips)
		if diag.HasError() {
			return fmt.Errorf("failed to convert IPs: %v", diag.Errors())
		}
		model.IPs = ipsList
	} else {
		ipsObjType := types.ObjectType{AttrTypes: getNetworkIPAttrTypes()}
		model.IPs = types.ListNull(ipsObjType)
	}

	// Autostart - always read the actual value
	autostart, err := r.client.Libvirt().NetworkGetAutostart(net)
	if err != nil {
		// Non-fatal, just set to false
		model.Autostart = types.BoolValue(false)
	} else {
		model.Autostart = types.BoolValue(autostart == 1)
	}

	return nil
}

// Helper functions to get attribute types for nested objects
func getNetworkIPAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"address":    types.StringType,
		"netmask":    types.StringType,
		"prefix":     types.Int64Type,
		"family":     types.StringType,
		"local_ptr":  types.StringType,
		"dhcp":       types.ObjectType{AttrTypes: getNetworkDHCPAttrTypes()},
	}
}

func getNetworkDHCPAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ranges": types.ListType{ElemType: types.ObjectType{AttrTypes: getNetworkDHCPRangeAttrTypes()}},
		"hosts":  types.ListType{ElemType: types.ObjectType{AttrTypes: getNetworkDHCPHostAttrTypes()}},
	}
}

func getNetworkDHCPRangeAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"start": types.StringType,
		"end":   types.StringType,
	}
}

func getNetworkDHCPHostAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mac":  types.StringType,
		"ip":   types.StringType,
		"name": types.StringType,
	}
}

