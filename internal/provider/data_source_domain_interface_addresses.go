package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	golibvirt "github.com/digitalocean/go-libvirt"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &DomainInterfaceAddressesDataSource{}

func NewDomainInterfaceAddressesDataSource() datasource.DataSource {
	return &DomainInterfaceAddressesDataSource{}
}

type DomainInterfaceAddressesDataSource struct {
	client *libvirt.Client
}

type DomainInterfaceAddressesDataSourceModel struct {
	ID           types.String            `tfsdk:"id"`
	Domain       types.String            `tfsdk:"domain"`
	Source       types.String            `tfsdk:"source"`
	Interfaces   []InterfaceAddressModel `tfsdk:"interfaces"`
	AgentTimeout types.Int64             `tfsdk:"agent_timeout"`
}

type InterfaceAddressModel struct {
	Name   types.String     `tfsdk:"name"`
	Hwaddr types.String     `tfsdk:"hwaddr"`
	Addrs  []IPAddressModel `tfsdk:"addrs"`
}

type IPAddressModel struct {
	Type   types.String `tfsdk:"type"`
	Addr   types.String `tfsdk:"addr"`
	Prefix types.Int64  `tfsdk:"prefix"`
}

func (d *DomainInterfaceAddressesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_interface_addresses"
}

func (d *DomainInterfaceAddressesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Queries IP addresses for a libvirt domain's network interfaces.\n\n" +
			"This data source uses libvirt's `virDomainInterfaceAddresses` API to retrieve " +
			"IP address information from DHCP leases or the QEMU guest agent.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Internal identifier for this data source (domain UUID).",
			},
			"domain": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "Domain UUID or name to query. Use `libvirt_domain.example.id` or " +
					"`libvirt_domain.example.name` to reference a managed domain.",
			},
			"source": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Source to query for IP addresses:\n" +
					"- `lease` - Query DHCP server leases (fast, no guest agent needed)\n" +
					"- `agent` - Query QEMU guest agent (requires qemu-guest-agent installed in guest)\n" +
					"- `any` - Try both sources (default)\n\n" +
					"If not specified, attempts both sources.",
				Validators: []validator.String{
					stringvalidator.OneOf("lease", "agent", "any"),
				},
			},
			"agent_timeout": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The timeout in seconds to wait for interfaces to become available when using the `agent` source. QEMU guest agent may take some time to start after the domain is booted.",
			},
			"interfaces": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of network interfaces with their IP addresses.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Interface name on the host (e.g., `vnet0`).",
						},
						"hwaddr": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "MAC address of the interface.",
						},
						"addrs": schema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "List of IP addresses assigned to this interface.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Address type: `ipv4` or `ipv6`.",
									},
									"addr": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "IP address.",
									},
									"prefix": schema.Int64Attribute{
										Computed:            true,
										MarkdownDescription: "Network prefix length (e.g., 24 for 255.255.255.0).",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *DomainInterfaceAddressesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// lookupDomain looks up a domain by UUID or name.
// Tries UUID first (since it has a specific format), then falls back to name.
func (d *DomainInterfaceAddressesDataSource) lookupDomain(nameOrUUID string) (golibvirt.Domain, error) {
	// Try UUID first
	domain, err := d.client.LookupDomainByUUID(nameOrUUID)
	if err == nil {
		return domain, nil
	}

	// Fall back to name lookup
	domain, err = d.client.Libvirt().DomainLookupByName(nameOrUUID)
	if err != nil {
		return golibvirt.Domain{}, fmt.Errorf("domain not found by UUID or name '%s': %w", nameOrUUID, err)
	}

	return domain, nil
}

func (d *DomainInterfaceAddressesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DomainInterfaceAddressesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Lookup domain by UUID or name
	domainIdentifier := config.Domain.ValueString()
	domain, err := d.lookupDomain(domainIdentifier)
	if err != nil {
		resp.Diagnostics.AddError(
			"Domain Not Found",
			fmt.Sprintf("Unable to find domain '%s': %s", domainIdentifier, err),
		)
		return
	}

	// // Determine source(s) to query
	sourceFuncs := []getInterfacesFunc{}
	sourceStr := "any"
	if !config.Source.IsNull() && !config.Source.IsUnknown() {
		sourceStr = config.Source.ValueString()
	}

	timeout := time.Duration(0)
	if !config.AgentTimeout.IsNull() && !config.AgentTimeout.IsUnknown() {
		timeout = time.Duration(config.AgentTimeout.ValueInt64()) * time.Second
	}

	agentFn := getQEMUGuestAgentInterfaces(d.client, domain)
	if timeout > 0 {
		agentFn = getInterfacesWithRetry(ctx, agentFn, domain, timeout)
	}

	switch sourceStr {
	case "lease":
		sourceFuncs = append(sourceFuncs, getDHCPServerLeaseInterfaces(d.client, domain))
	case "agent":
		sourceFuncs = append(sourceFuncs, agentFn)
	case "any":
		sourceFuncs = append(sourceFuncs, getDHCPServerLeaseInterfaces(d.client, domain))
		sourceFuncs = append(sourceFuncs, agentFn)
	}

	var ifaces []golibvirt.DomainInterface
	var lastErr error

	for _, sourceFn := range sourceFuncs {
		ifaces, err = sourceFn()
		if err == nil && len(ifaces) > 0 {
			// Found interfaces with at least one result
			break
		}
		if err != nil {
			lastErr = err
		}
	}

	// If we got no interfaces and had errors, report it
	if len(ifaces) == 0 && lastErr != nil {
		resp.Diagnostics.AddWarning(
			"No Interface Addresses Found",
			fmt.Sprintf("Unable to retrieve interface addresses for domain '%s': %s\n\n"+
				"This may be normal if:\n"+
				"- Domain has not obtained an IP address yet\n"+
				"- DHCP is not configured (for source='lease')\n"+
				"- QEMU guest agent is not running (for source='agent')",
				domainIdentifier, lastErr),
		)
	}

	// Convert to model
	// Initialize as empty slice (not nil) so Terraform gets [] instead of null
	interfaces := []InterfaceAddressModel{}
	for _, iface := range ifaces {
		ifaceModel := InterfaceAddressModel{
			Name: types.StringValue(iface.Name),
		}

		// Handle optional hwaddr (OptString is a []string)
		if len(iface.Hwaddr) > 0 {
			ifaceModel.Hwaddr = types.StringValue(iface.Hwaddr[0])
		} else {
			ifaceModel.Hwaddr = types.StringNull()
		}

		// Convert addresses
		// Initialize as empty slice (not nil) so Terraform gets [] instead of null
		addrs := []IPAddressModel{}
		for _, addr := range iface.Addrs {
			addrType := "ipv4"
			if addr.Type == int32(golibvirt.IPAddrTypeIpv6) {
				addrType = "ipv6"
			}

			addrs = append(addrs, IPAddressModel{
				Type:   types.StringValue(addrType),
				Addr:   types.StringValue(addr.Addr),
				Prefix: types.Int64Value(int64(addr.Prefix)),
			})
		}
		ifaceModel.Addrs = addrs

		interfaces = append(interfaces, ifaceModel)
	}

	// Populate result
	uuidStr := libvirt.UUIDString(domain.UUID)
	config.ID = types.StringValue(uuidStr)
	config.Interfaces = interfaces

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// getInterfacesFunc defines a function type that retrieves domain interfaces.
type getInterfacesFunc func() ([]golibvirt.DomainInterface, error)

// getDHCPServerLeaseInterfaces returns a function that retrieves interfaces using DHCP server leases.
func getDHCPServerLeaseInterfaces(client *libvirt.Client, domain golibvirt.Domain) getInterfacesFunc {
	return func() ([]golibvirt.DomainInterface, error) {
		return client.Libvirt().DomainInterfaceAddresses(domain, uint32(golibvirt.DomainInterfaceAddressesSrcLease), 0)
	}
}

// getQEMUGuestAgentInterfaces returns a function that retrieves interfaces using the QEMU guest agent.
func getQEMUGuestAgentInterfaces(client *libvirt.Client, domain golibvirt.Domain) getInterfacesFunc {
	return func() ([]golibvirt.DomainInterface, error) {
		return client.Libvirt().DomainInterfaceAddresses(domain, uint32(golibvirt.DomainInterfaceAddressesSrcAgent), 0)
	}
}

func getInterfacesWithRetry(ctx context.Context, fn getInterfacesFunc, domain golibvirt.Domain, timeout time.Duration) getInterfacesFunc {
	return func() ([]golibvirt.DomainInterface, error) {
		pollInterval := 5 * time.Second
		ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		tflog.Debug(ctxWithTimeout, "Starting to wait and retry for interfaces to become available", map[string]interface{}{
			"domain":   domain.Name,
			"timeout":  timeout.String(),
			"interval": pollInterval.String(),
		})

		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()
		retryCnt := 0
		var lastErr error

		for {
			select {
			case <-ctxWithTimeout.Done():
				return nil, fmt.Errorf("timed out waiting for interfaces to become available: %w", lastErr)
			case <-ticker.C:
				retryCnt++
				ifaces, err := fn()
				if err != nil {
					var libvirtErr golibvirt.Error
					if errors.As(err, &libvirtErr) {
						if libvirtErr.Code == uint32(golibvirt.ErrAgentUnresponsive) {
							tflog.Debug(ctx, "Agent unresponsive", map[string]interface{}{
								"attempt": retryCnt,
								"error":   err.Error(),
							})
						} else {
							tflog.Debug(ctx, "Libvirt error, will retry in 5 seconds", map[string]interface{}{
								"attempt": retryCnt,
								"error":   err.Error(),
								"code":    libvirtErr.Code,
							})
						}
					} else {
						tflog.Debug(ctx, "Interfaces not available yet, will retry in 5 seconds", map[string]interface{}{
							"attempt": retryCnt,
							"error":   err.Error(),
						})
					}

					lastErr = err
					continue
				}

				return ifaces, nil
			}
		}
	}
}
