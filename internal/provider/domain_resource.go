package provider

import (
	"context"
	"fmt"
	"time"

	golibvirt "github.com/digitalocean/go-libvirt"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/generated"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource              = &DomainResource{}
	_ resource.ResourceWithConfigure = &DomainResource{}
)

// NewDomainResource creates a new domain resource
func NewDomainResource() resource.Resource {
	return &DomainResource{}
}

// DomainResource defines the resource implementation
type DomainResource struct {
	client *libvirt.Client
}

// DomainResourceModel embeds the generated domain model and adds provider-specific fields.
type DomainResourceModel struct {
	generated.DomainModel

	Running   types.Bool   `tfsdk:"running"`
	Autostart types.Bool   `tfsdk:"autostart"`
	Create    types.Object `tfsdk:"create"`
	Destroy   types.Object `tfsdk:"destroy"`
}

// DomainInterfaceWaitForIPModel describes wait_for_ip overrides.
type DomainInterfaceWaitForIPModel struct {
	Timeout types.Int64  `tfsdk:"timeout"`
	Source  types.String `tfsdk:"source"`
}

type interfaceWaitForIPConfig struct {
	Index     int
	MAC       string
	Timeout   int64
	Source    string
	Attribute attr.Value
}

// DomainCreateModel describes domain start flags
type DomainCreateModel struct {
	Paused      types.Bool `tfsdk:"paused"`
	Autodestroy types.Bool `tfsdk:"autodestroy"`
	BypassCache types.Bool `tfsdk:"bypass_cache"`
	ForceBoot   types.Bool `tfsdk:"force_boot"`
	Validate    types.Bool `tfsdk:"validate"`
	ResetNVRAM  types.Bool `tfsdk:"reset_nvram"`
}

// DomainDestroyModel describes domain shutdown behavior
type DomainDestroyModel struct {
	Graceful types.Bool  `tfsdk:"graceful"`
	Timeout  types.Int64 `tfsdk:"timeout"`
}

type domainPlanData struct {
	SanitizedModel generated.DomainModel
	WaitConfigs    []interfaceWaitForIPConfig
	WaitAttributes []attr.Value
}

func domainInterfaceWaitForIPSchemaAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Wait for IP address during domain creation. If specified, Terraform will wait until the interface receives an IP.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"timeout": schema.Int64Attribute{
				Description: "Maximum time to wait for IP address in seconds. Default: 300.",
				Optional:    true,
			},
			"source": schema.StringAttribute{
				Description: "Source to query for IP addresses: 'lease', 'agent', or 'any'. Default: 'any'.",
				Optional:    true,
			},
		},
	}
}

func domainDevicesSchemaAttributeWithWaitForIP() schema.SingleNestedAttribute {
	baseAttr := mustSingleNestedAttribute(generated.DomainDeviceListSchemaAttribute(), "DomainDeviceList")

	interfacesAttr, ok := baseAttr.Attributes["interfaces"].(schema.ListNestedAttribute)
	if !ok {
		return baseAttr
	}

	interfaceAttrs := interfacesAttr.NestedObject.Attributes
	interfaceAttrs["wait_for_ip"] = domainInterfaceWaitForIPSchemaAttribute()
	interfacesAttr.NestedObject.Attributes = interfaceAttrs
	baseAttr.Attributes["interfaces"] = interfacesAttr

	return baseAttr
}

func domainInterfaceWaitForIPAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"timeout": types.Int64Type,
		"source":  types.StringType,
	}
}

func copyAttrType(t attr.Type) attr.Type {
	switch typed := t.(type) {
	case types.ObjectType:
		return types.ObjectType{AttrTypes: copyAttrTypesMap(typed.AttrTypes)}
	case types.ListType:
		return types.ListType{ElemType: copyAttrType(typed.ElemType)}
	case types.SetType:
		return types.SetType{ElemType: copyAttrType(typed.ElemType)}
	default:
		return t
	}
}

func copyAttrTypesMap(src map[string]attr.Type) map[string]attr.Type {
	result := make(map[string]attr.Type, len(src))
	for k, v := range src {
		result[k] = copyAttrType(v)
	}
	return result
}

func domainInterfaceAttributeTypesWithWaitForIP() map[string]attr.Type {
	base := copyAttrTypesMap(generated.DomainInterfaceAttributeTypes())
	base["wait_for_ip"] = types.ObjectType{
		AttrTypes: domainInterfaceWaitForIPAttrTypes(),
	}
	return base
}

func domainDeviceListAttributeTypesWithWaitForIP() map[string]attr.Type {
	base := copyAttrTypesMap(generated.DomainDeviceListAttributeTypes())
	if _, ok := base["interfaces"].(types.ListType); ok {
		base["interfaces"] = types.ListType{
			ElemType: types.ObjectType{AttrTypes: domainInterfaceAttributeTypesWithWaitForIP()},
		}
	}
	return base
}

func prepareDomainPlan(ctx context.Context, model *DomainResourceModel) (domainPlanData, diag.Diagnostics) {
	result := domainPlanData{
		SanitizedModel: model.DomainModel,
	}

	cleanDevices, configs, waitAttrs, diags := stripWaitForIP(ctx, model.Devices)
	if diags.HasError() {
		return result, diags
	}

	result.SanitizedModel.Devices = cleanDevices
	result.WaitConfigs = configs
	result.WaitAttributes = waitAttrs
	return result, nil
}

func stripWaitForIP(ctx context.Context, devices types.Object) (types.Object, []interfaceWaitForIPConfig, []attr.Value, diag.Diagnostics) {
	if devices.IsNull() || devices.IsUnknown() {
		return devices, nil, nil, nil
	}

	attrs := devices.Attributes()
	cleanAttrs := make(map[string]attr.Value, len(attrs))
	for k, v := range attrs {
		cleanAttrs[k] = v
	}

	rawInterfaces, ok := attrs["interfaces"]
	if !ok || rawInterfaces.IsNull() || rawInterfaces.IsUnknown() {
		cleanAttrs["interfaces"] = types.ListNull(types.ObjectType{AttrTypes: generated.DomainInterfaceAttributeTypes()})
		cleanObj, diags := types.ObjectValue(generated.DomainDeviceListAttributeTypes(), cleanAttrs)
		return cleanObj, nil, nil, diags
	}

	listVal, ok := rawInterfaces.(basetypes.ListValue)
	if !ok {
		return types.ObjectNull(generated.DomainDeviceListAttributeTypes()), nil, nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid interfaces value", "Expected interfaces to be a list."),
		}
	}

	elements := listVal.Elements()
	cleanInterfaces := make([]attr.Value, len(elements))
	waitAttrs := make([]attr.Value, len(elements))
	var waitConfigs []interfaceWaitForIPConfig

	for i, element := range elements {
		ifaceObj, ok := element.(basetypes.ObjectValue)
		if !ok {
			return types.ObjectNull(generated.DomainDeviceListAttributeTypes()), nil, nil, diag.Diagnostics{
				diag.NewErrorDiagnostic("Invalid interface value", "Expected interface entry to be an object."),
			}
		}

		ifaceAttrs := ifaceObj.Attributes()
		waitVal, hasWait := ifaceAttrs["wait_for_ip"]
		if !hasWait {
			waitVal = types.ObjectNull(domainInterfaceWaitForIPAttrTypes())
		}
		waitAttrs[i] = waitVal

		if hasWait && !waitVal.IsNull() && !waitVal.IsUnknown() {
			var waitModel DomainInterfaceWaitForIPModel
			if waitObj, ok := waitVal.(basetypes.ObjectValue); ok {
				diags := waitObj.As(ctx, &waitModel, basetypes.ObjectAsOptions{})
				if diags.HasError() {
					return types.ObjectNull(generated.DomainDeviceListAttributeTypes()), nil, nil, diags
				}
			}

			timeout := int64(300)
			if !waitModel.Timeout.IsNull() && !waitModel.Timeout.IsUnknown() {
				timeout = waitModel.Timeout.ValueInt64()
			}

			source := "any"
			if !waitModel.Source.IsNull() && !waitModel.Source.IsUnknown() {
				source = waitModel.Source.ValueString()
			}

			macValue, ok := ifaceAttrs["mac"]
			var mac string
			if ok {
				var macModel generated.DomainInterfaceMACModel
				if macObj, ok := macValue.(basetypes.ObjectValue); ok && !macObj.IsNull() && !macObj.IsUnknown() {
					diags := macObj.As(ctx, &macModel, basetypes.ObjectAsOptions{})
					if diags.HasError() {
						return types.ObjectNull(generated.DomainDeviceListAttributeTypes()), nil, nil, diags
					}
				}
				if !macModel.Address.IsNull() && !macModel.Address.IsUnknown() {
					mac = macModel.Address.ValueString()
				}
			}

			waitConfigs = append(waitConfigs, interfaceWaitForIPConfig{
				Index:     i,
				MAC:       mac,
				Timeout:   timeout,
				Source:    source,
				Attribute: waitVal,
			})
		}

		newIfaceAttrs := make(map[string]attr.Value, len(ifaceAttrs))
		for k, v := range ifaceAttrs {
			if k == "wait_for_ip" {
				continue
			}
			newIfaceAttrs[k] = v
		}

		cleanIface, diags := types.ObjectValue(generated.DomainInterfaceAttributeTypes(), newIfaceAttrs)
		if diags.HasError() {
			return types.ObjectNull(generated.DomainDeviceListAttributeTypes()), nil, nil, diags
		}
		cleanInterfaces[i] = cleanIface
	}

	listClean, diags := types.ListValue(types.ObjectType{AttrTypes: generated.DomainInterfaceAttributeTypes()}, cleanInterfaces)
	if diags.HasError() {
		return types.ObjectNull(generated.DomainDeviceListAttributeTypes()), nil, nil, diags
	}

	cleanAttrs["interfaces"] = listClean
	cleanObj, diags := types.ObjectValue(generated.DomainDeviceListAttributeTypes(), cleanAttrs)
	return cleanObj, waitConfigs, waitAttrs, diags
}

func applyWaitForIPValues(ctx context.Context, devices types.Object, waitValues []attr.Value) (types.Object, diag.Diagnostics) {
	if devices.IsNull() || devices.IsUnknown() {
		if len(waitValues) == 0 {
			if devices.IsUnknown() {
				return types.ObjectUnknown(domainDeviceListAttributeTypesWithWaitForIP()), nil
			}
			return types.ObjectNull(domainDeviceListAttributeTypesWithWaitForIP()), nil
		}
		return types.ObjectNull(domainDeviceListAttributeTypesWithWaitForIP()), nil
	}

	attrs := devices.Attributes()
	newAttrs := make(map[string]attr.Value, len(attrs))
	for k, v := range attrs {
		newAttrs[k] = v
	}

	rawInterfaces, ok := attrs["interfaces"]
	if !ok || rawInterfaces.IsNull() || rawInterfaces.IsUnknown() {
		newAttrs["interfaces"] = types.ListNull(types.ObjectType{AttrTypes: domainInterfaceAttributeTypesWithWaitForIP()})
		cleanObj, diags := types.ObjectValue(domainDeviceListAttributeTypesWithWaitForIP(), newAttrs)
		return cleanObj, diags
	}

	listVal, ok := rawInterfaces.(basetypes.ListValue)
	if !ok {
		return types.ObjectNull(domainDeviceListAttributeTypesWithWaitForIP()), diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid interfaces value", "Expected interfaces to be a list."),
		}
	}

	elements := listVal.Elements()
	newInterfaces := make([]attr.Value, len(elements))
	for i, element := range elements {
		ifaceObj, ok := element.(basetypes.ObjectValue)
		if !ok {
			return types.ObjectNull(domainDeviceListAttributeTypesWithWaitForIP()), diag.Diagnostics{
				diag.NewErrorDiagnostic("Invalid interface value", "Expected interface entry to be an object."),
			}
		}

		// Convert interface attributes to ensure proper typing
		// We need to reconstruct the object with the new type that includes wait_for_ip
		ifaceAttrs := ifaceObj.Attributes()
		newIfaceAttrs := make(map[string]attr.Value, len(generated.DomainInterfaceAttributeTypes())+1)

		// Copy all base attributes - they should all be present from FromXML
		for attrName := range generated.DomainInterfaceAttributeTypes() {
			if val, exists := ifaceAttrs[attrName]; exists {
				newIfaceAttrs[attrName] = val
			}
		}

		var waitVal attr.Value
		if i < len(waitValues) && waitValues[i] != nil {
			waitVal = waitValues[i]
		} else {
			waitVal = types.ObjectNull(domainInterfaceWaitForIPAttrTypes())
		}
		newIfaceAttrs["wait_for_ip"] = waitVal

		newIface, diags := types.ObjectValue(domainInterfaceAttributeTypesWithWaitForIP(), newIfaceAttrs)
		if diags.HasError() {
			return types.ObjectNull(domainDeviceListAttributeTypesWithWaitForIP()), diags
		}
		newInterfaces[i] = newIface
	}

	listWithWait, diags := types.ListValue(types.ObjectType{AttrTypes: domainInterfaceAttributeTypesWithWaitForIP()}, newInterfaces)
	if diags.HasError() {
		return types.ObjectNull(domainDeviceListAttributeTypesWithWaitForIP()), diags
	}
	newAttrs["interfaces"] = listWithWait

	cleanObj, diags := types.ObjectValue(domainDeviceListAttributeTypesWithWaitForIP(), newAttrs)
	return cleanObj, diags
}

func domainStartFlagsFromCreate(ctx context.Context, createVal types.Object) (uint32, diag.Diagnostics) {
	if createVal.IsNull() || createVal.IsUnknown() {
		return 0, nil
	}

	var createModel DomainCreateModel
	diags := createVal.As(ctx, &createModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return 0, diags
	}

	var flags uint32
	if !createModel.Paused.IsNull() && createModel.Paused.ValueBool() {
		flags |= uint32(golibvirt.DomainStartPaused)
	}
	if !createModel.Autodestroy.IsNull() && createModel.Autodestroy.ValueBool() {
		flags |= uint32(golibvirt.DomainStartAutodestroy)
	}
	if !createModel.BypassCache.IsNull() && createModel.BypassCache.ValueBool() {
		flags |= uint32(golibvirt.DomainStartBypassCache)
	}
	if !createModel.ForceBoot.IsNull() && createModel.ForceBoot.ValueBool() {
		flags |= uint32(golibvirt.DomainStartForceBoot)
	}
	if !createModel.Validate.IsNull() && createModel.Validate.ValueBool() {
		flags |= uint32(golibvirt.DomainStartValidate)
	}
	if !createModel.ResetNVRAM.IsNull() && createModel.ResetNVRAM.ValueBool() {
		flags |= uint32(golibvirt.DomainStartResetNvram)
	}

	return flags, nil
}

// Metadata returns the resource type name
func (r *DomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

// Schema defines the schema for the resource
func (r *DomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	overrides := map[string]schema.Attribute{
		"devices": domainDevicesSchemaAttributeWithWaitForIP(),
		"running": schema.BoolAttribute{
			Description: "Whether the domain should be started after creation.",
			Optional:    true,
		},
		"autostart": schema.BoolAttribute{
			Description: "Whether the domain should be started automatically when the host boots.",
			Optional:    true,
		},
		"create": schema.SingleNestedAttribute{
			Description: "Start behavior flags passed to libvirt when running is true.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"paused":       schema.BoolAttribute{Optional: true},
				"autodestroy":  schema.BoolAttribute{Optional: true},
				"bypass_cache": schema.BoolAttribute{Optional: true},
				"force_boot":   schema.BoolAttribute{Optional: true},
				"validate":     schema.BoolAttribute{Optional: true},
				"reset_nvram":  schema.BoolAttribute{Optional: true},
			},
		},
		"destroy": schema.SingleNestedAttribute{
			Description: "Destroy behavior when Terraform removes the domain.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"graceful": schema.BoolAttribute{Optional: true},
				"timeout":  schema.Int64Attribute{Optional: true},
			},
		},
	}

	schemaDef := generated.DomainSchema(overrides)
	schemaDef.Description = "Manages a libvirt domain (virtual machine)."
	schemaDef.MarkdownDescription = `
Manages a libvirt domain (virtual machine).

This resource follows the [libvirt domain XML schema](https://libvirt.org/formatdomain.html) closely,
providing fine-grained control over VM configuration.
`
	resp.Schema = schemaDef
}

// Configure adds the provider configured client to the resource
func (r *DomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*libvirt.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *libvirt.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates a new domain
func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planData, diags := prepareDomainPlan(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainXML, err := generated.DomainToXML(ctx, &planData.SanitizedModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Domain Configuration",
			"Failed to convert domain configuration to XML: "+err.Error(),
		)
		return
	}

	xmlString, err := libvirt.MarshalDomainXML(domainXML)
	if err != nil {
		resp.Diagnostics.AddError(
			"XML Marshaling Failed",
			"Failed to marshal domain XML: "+err.Error(),
		)
		return
	}

	domain, err := r.client.Libvirt().DomainDefineXML(xmlString)
	if err != nil {
		resp.Diagnostics.AddError(
			"Domain Creation Failed",
			"Failed to define domain in libvirt: "+err.Error(),
		)
		return
	}

	cleanupOnError := func() {
		if !plan.Running.IsNull() && plan.Running.ValueBool() {
			if destroyErr := r.client.Libvirt().DomainDestroy(domain); destroyErr != nil {
				tflog.Warn(ctx, "Failed to destroy domain during cleanup", map[string]any{"error": destroyErr.Error()})
			}
		}
		if undefErr := r.client.Libvirt().DomainUndefine(domain); undefErr != nil {
			tflog.Warn(ctx, "Failed to undefine domain during cleanup", map[string]any{"error": undefErr.Error()})
		}
	}

	if !plan.Running.IsNull() && plan.Running.ValueBool() {
		flags, startDiags := domainStartFlagsFromCreate(ctx, plan.Create)
		resp.Diagnostics.Append(startDiags...)
		if resp.Diagnostics.HasError() {
			cleanupOnError()
			return
		}

		if _, err := r.client.Libvirt().DomainCreateWithFlags(domain, flags); err != nil {
			cleanupOnError()
			resp.Diagnostics.AddError(
				"Failed to Start Domain",
				"Domain was defined but failed to start: "+err.Error(),
			)
			return
		}

		for _, waitCfg := range planData.WaitConfigs {
			if err := waitForInterfaceIP(ctx, r.client, domain, waitCfg.MAC, waitCfg.Timeout, waitCfg.Source); err != nil {
				cleanupOnError()
				info := ""
				if waitCfg.MAC != "" {
					info = fmt.Sprintf(" (MAC: %s)", waitCfg.MAC)
				}
				resp.Diagnostics.AddError(
					"Failed to Wait for IP Address",
					fmt.Sprintf("Domain was created and started but failed to obtain an IP address%s: %s", info, err),
				)
				return
			}
		}
	}

	if !plan.Autostart.IsNull() && !plan.Autostart.IsUnknown() {
		autostart := int32(0)
		if plan.Autostart.ValueBool() {
			autostart = 1
		}
		if err := r.client.Libvirt().DomainSetAutostart(domain, autostart); err != nil {
			cleanupOnError()
			resp.Diagnostics.AddError(
				"Failed to Set Autostart",
				"Domain was created but failed to set autostart: "+err.Error(),
			)
			return
		}
	}

	xmlDesc, err := r.client.Libvirt().DomainGetXMLDesc(domain, 0)
	if err != nil {
		cleanupOnError()
		resp.Diagnostics.AddError(
			"Failed to Read Domain",
			"Failed to get domain XML: "+err.Error(),
		)
		return
	}

	parsedDomain, err := libvirt.UnmarshalDomainXML(xmlDesc)
	if err != nil {
		cleanupOnError()
		resp.Diagnostics.AddError(
			"Failed to Parse Domain XML",
			"Failed to parse domain XML from libvirt: "+err.Error(),
		)
		return
	}

	stateModel, err := generated.DomainFromXML(ctx, parsedDomain, &planData.SanitizedModel)
	if err != nil {
		cleanupOnError()
		resp.Diagnostics.AddError(
			"Failed to Convert Domain",
			"Failed to convert domain XML to state: "+err.Error(),
		)
		return
	}

	state := DomainResourceModel{
		DomainModel: *stateModel,
		Running:     plan.Running,
		Autostart:   plan.Autostart,
		Create:      plan.Create,
		Destroy:     plan.Destroy,
	}

	state.Devices, diags = applyWaitForIPValues(ctx, state.Devices, planData.WaitAttributes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		cleanupOnError()
		return
	}

	if !plan.Metadata.IsNull() && !plan.Metadata.IsUnknown() {
		state.Metadata = plan.Metadata
	}

	if state.ID.IsNull() && !plan.ID.IsNull() && !plan.ID.IsUnknown() {
		state.ID = plan.ID
	}

	if !state.Autostart.IsNull() && !state.Autostart.IsUnknown() {
		autostart, err := r.client.Libvirt().DomainGetAutostart(domain)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Get Autostart Status",
				"Failed to read domain autostart setting: "+err.Error(),
			)
			return
		}
		state.Autostart = types.BoolValue(autostart == 1)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read reads the domain state
func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	originalMetadata := state.Metadata
	originalID := state.ID

	if state.UUID.IsNull() || state.UUID.IsUnknown() {
		resp.State.RemoveResource(ctx)
		return
	}

	planData, diags := prepareDomainPlan(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.LookupDomainByUUID(state.UUID.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	xmlDesc, err := r.client.Libvirt().DomainGetXMLDesc(domain, 0)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Domain",
			"Failed to get domain XML: "+err.Error(),
		)
		return
	}

	parsedDomain, err := libvirt.UnmarshalDomainXML(xmlDesc)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Parse Domain XML",
			"Failed to parse domain XML from libvirt: "+err.Error(),
		)
		return
	}

	stateModel, err := generated.DomainFromXML(ctx, parsedDomain, &planData.SanitizedModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Convert Domain",
			"Failed to convert domain XML to state: "+err.Error(),
		)
		return
	}

	state.DomainModel = *stateModel

	state.Devices, diags = applyWaitForIPValues(ctx, state.Devices, planData.WaitAttributes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !originalMetadata.IsNull() && !originalMetadata.IsUnknown() {
		state.Metadata = originalMetadata
	}

	if state.ID.IsNull() && !originalID.IsNull() && !originalID.IsUnknown() {
		state.ID = originalID
	}

	if !state.Autostart.IsNull() && !state.Autostart.IsUnknown() {
		autostart, err := r.client.Libvirt().DomainGetAutostart(domain)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Get Autostart Status",
				"Failed to read domain autostart setting: "+err.Error(),
			)
			return
		}
		state.Autostart = types.BoolValue(autostart == 1)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// waitForDomainState waits for a domain to reach the specified state with a timeout
func waitForDomainState(client *libvirt.Client, domain golibvirt.Domain, targetState uint32, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		state, _, err := client.Libvirt().DomainGetState(domain, 0)
		if err != nil {
			return fmt.Errorf("failed to get domain state: %w", err)
		}
		if uint32(state) == targetState {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("timeout waiting for domain to reach state %d", targetState)
}

// Update updates the domain
func (r *DomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DomainResourceModel
	var state DomainResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.UUID.IsNull() || state.UUID.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing Domain UUID",
			"Existing domain is missing UUID in state",
		)
		return
	}

	planData, diags := prepareDomainPlan(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.UUID.IsNull() && !state.UUID.IsUnknown() {
		planData.SanitizedModel.UUID = state.UUID
	}

	existingDomain, err := r.client.LookupDomainByUUID(state.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Domain Lookup Failed",
			"Failed to look up existing domain: "+err.Error(),
		)
		return
	}

	domainState, _, err := r.client.Libvirt().DomainGetState(existingDomain, 0)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Get Domain State",
			"Failed to check if domain is running: "+err.Error(),
		)
		return
	}

	if uint32(domainState) == uint32(golibvirt.DomainRunning) {
		if err := r.client.Libvirt().DomainShutdown(existingDomain); err != nil {
			resp.Diagnostics.AddError(
				"Failed to Shutdown Domain",
				"Domain must be stopped before updating. Failed to shutdown: "+err.Error(),
			)
			return
		}

		if err := waitForDomainState(r.client, existingDomain, uint32(golibvirt.DomainShutoff), 5*time.Second); err != nil {
			if err := r.client.Libvirt().DomainDestroy(existingDomain); err != nil {
				resp.Diagnostics.AddError(
					"Failed to Stop Domain",
					"Domain must be stopped before updating. Failed to force stop: "+err.Error(),
				)
				return
			}
		}
	}

	domainXML, err := generated.DomainToXML(ctx, &planData.SanitizedModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Domain Configuration",
			"Failed to convert domain configuration to XML: "+err.Error(),
		)
		return
	}

	xmlString, err := libvirt.MarshalDomainXML(domainXML)
	if err != nil {
		resp.Diagnostics.AddError(
			"XML Marshaling Failed",
			"Failed to marshal domain XML: "+err.Error(),
		)
		return
	}

	if err := r.client.Libvirt().DomainUndefine(existingDomain); err != nil {
		resp.Diagnostics.AddError(
			"Domain Undefine Failed",
			"Failed to undefine existing domain: "+err.Error(),
		)
		return
	}

	newDomain, err := r.client.Libvirt().DomainDefineXML(xmlString)
	if err != nil {
		resp.Diagnostics.AddError(
			"Domain Update Failed",
			"Failed to define updated domain in libvirt: "+err.Error(),
		)
		return
	}

	if !plan.Autostart.IsNull() && !plan.Autostart.IsUnknown() {
		autostart := int32(0)
		if plan.Autostart.ValueBool() {
			autostart = 1
		}
		if err := r.client.Libvirt().DomainSetAutostart(newDomain, autostart); err != nil {
			resp.Diagnostics.AddError(
				"Failed to Set Autostart",
				"Domain was updated but failed to set autostart: "+err.Error(),
			)
			return
		}
	}

	shouldBeRunning := !plan.Running.IsNull() && plan.Running.ValueBool()
	if shouldBeRunning {
		flags, startDiags := domainStartFlagsFromCreate(ctx, plan.Create)
		resp.Diagnostics.Append(startDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if _, err := r.client.Libvirt().DomainCreateWithFlags(newDomain, flags); err != nil {
			resp.Diagnostics.AddWarning(
				"Failed to Start Domain",
				"Domain was updated but failed to start: "+err.Error(),
			)
		} else {
			for _, waitCfg := range planData.WaitConfigs {
				if err := waitForInterfaceIP(ctx, r.client, newDomain, waitCfg.MAC, waitCfg.Timeout, waitCfg.Source); err != nil {
					resp.Diagnostics.AddError(
						"Failed to Wait for IP Address",
						fmt.Sprintf("Domain was updated but failed to obtain an IP address (MAC: %s): %s", waitCfg.MAC, err),
					)
					return
				}
			}
		}
	}

	xmlDesc, err := r.client.Libvirt().DomainGetXMLDesc(newDomain, 0)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Domain",
			"Failed to get domain XML: "+err.Error(),
		)
		return
	}

	parsedDomain, err := libvirt.UnmarshalDomainXML(xmlDesc)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Parse Domain XML",
			"Failed to parse domain XML from libvirt: "+err.Error(),
		)
		return
	}

	stateModel, err := generated.DomainFromXML(ctx, parsedDomain, &planData.SanitizedModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Convert Domain",
			"Failed to convert domain XML to state: "+err.Error(),
		)
		return
	}

	newState := DomainResourceModel{
		DomainModel: *stateModel,
		Running:     plan.Running,
		Autostart:   plan.Autostart,
		Create:      plan.Create,
		Destroy:     plan.Destroy,
	}

	newState.Devices, diags = applyWaitForIPValues(ctx, newState.Devices, planData.WaitAttributes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Metadata.IsNull() && !plan.Metadata.IsUnknown() {
		newState.Metadata = plan.Metadata
	}

	if newState.ID.IsNull() && !state.ID.IsNull() && !state.ID.IsUnknown() {
		newState.ID = state.ID
	}

	if !newState.Autostart.IsNull() && !newState.Autostart.IsUnknown() {
		autostart, err := r.client.Libvirt().DomainGetAutostart(newDomain)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Get Autostart Status",
				"Failed to read domain autostart setting: "+err.Error(),
			)
			return
		}
		newState.Autostart = types.BoolValue(autostart == 1)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Delete deletes the domain
func (r *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	destroyState := DomainDestroyModel{
    Graceful: types.BoolValue(false),
    Timeout:  types.Int64Value(0),
	}
	resp.Diagnostics.Append(state.Destroy.As(ctx, &destroyState, basetypes.ObjectAsOptions{})...)
	if destroyState.Timeout.ValueInt64() == 0 {
		destroyState.Timeout = types.Int64Value(30)
	}

	// Look up the domain
	domain, err := r.client.LookupDomainByUUID(state.UUID.ValueString())
	if err != nil {
		// Domain already gone - that's OK
		return
	}

	// Destroy (stop) the domain if it's running
	domainState, _, err := r.client.Libvirt().DomainGetState(domain, 0)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Get Domain State",
			"Failed to check domain state: "+err.Error(),
		)
		return
	}

	// DomainState values: 0=nostate, 1=running, 2=blocked, 3=paused, 4=shutdown, 5=shutoff, 6=crashed, 7=pmsuspended
	if uint32(domainState) == uint32(golibvirt.DomainRunning) {
		if destroyState.Graceful.ValueBool() {
		  err = r.client.Libvirt().DomainShutdown(domain)
			if err != nil {
				resp.Diagnostics.AddError(
					"Failed to Shutdown Domain",
					"Failed to stop running domain: "+err.Error(),
				)
				return
			}

			err = waitForDomainState(
				r.client,
				domain,
				uint32(golibvirt.DomainShutoff),
				time.Duration(destroyState.Timeout.ValueInt64())*time.Second)
			if err != nil {
				resp.Diagnostics.AddError(
					"Timeout waiting for Domain to shutdown: ",
					"Timeout exceeded waiting for Domain to shutdown: "+domain.Name,
				)
				return
			}
		} else {
			err = r.client.Libvirt().DomainDestroy(domain)
			if err != nil {
				resp.Diagnostics.AddError(
					"Failed to Destroy Domain",
					"Failed to stop running domain: "+err.Error(),
				)
				return
			}
		}
	}

	// Undefine the domain
	// Use DomainUndefineNvram flag to also remove NVRAM files when present (UEFI)
	// Use DomainUndefineTpm flag to also remove TPM state when present.
	// Libvirt added this flag in 8.9.0, so gate it to keep Ubuntu 22.04 (libvirt 8.0.0)
	// working even though it's outdated but still popular.
	const libvirtTPMUndefineMinVersion = 8009000
	flags := golibvirt.DomainUndefineNvram
	if r.client.LibVersion() >= libvirtTPMUndefineMinVersion {
		flags |= golibvirt.DomainUndefineTpm
	}
	err = r.client.Libvirt().DomainUndefineFlags(domain, flags)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Undefine Domain",
			"Failed to undefine domain: "+err.Error(),
		)
		return
	}
}

// waitForInterfaceIP polls for IP addresses on a domain's interfaces
// If mac is specified, waits for that specific interface to get an IP
// If mac is empty, waits for any interface to get an IP
// Returns error if timeout is reached without obtaining an IP
func waitForInterfaceIP(ctx context.Context, client *libvirt.Client, domain golibvirt.Domain, mac string, timeout int64, sourceStr string) error {
	if timeout == 0 {
		timeout = 300 // Default 5 minutes
	}
	if sourceStr == "" {
		sourceStr = "any"
	}

	// Determine source(s) to query
	var sources []golibvirt.DomainInterfaceAddressesSource
	switch sourceStr {
	case "lease":
		sources = []golibvirt.DomainInterfaceAddressesSource{golibvirt.DomainInterfaceAddressesSrcLease}
	case "agent":
		sources = []golibvirt.DomainInterfaceAddressesSource{golibvirt.DomainInterfaceAddressesSrcAgent}
	case "any":
		sources = []golibvirt.DomainInterfaceAddressesSource{
			golibvirt.DomainInterfaceAddressesSrcLease,
			golibvirt.DomainInterfaceAddressesSrcAgent,
		}
	default:
		return fmt.Errorf("invalid source: %s (must be 'lease', 'agent', or 'any')", sourceStr)
	}

	// Poll for IP with timeout
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	pollInterval := 5 * time.Second

	for {
		// Try each source until we get an IP
		for _, source := range sources {
			ifaces, err := client.Libvirt().DomainInterfaceAddresses(domain, uint32(source), 0)
			if err == nil && len(ifaces) > 0 {
				// Check if we're looking for a specific MAC or any interface
				if mac != "" {
					// Look for specific interface by MAC
					for _, iface := range ifaces {
						// Check if this interface matches the MAC we're looking for
						if len(iface.Hwaddr) > 0 && iface.Hwaddr[0] == mac {
							if len(iface.Addrs) > 0 {
								// Found the interface and it has an IP
								return nil
							}
							// Found the interface but no IP yet, keep polling
							break
						}
					}
				} else {
					// Wait for any interface to get an IP
					for _, iface := range ifaces {
						if len(iface.Addrs) > 0 {
							// Got at least one IP address
							return nil
						}
					}
				}
			}
		}

		// Check timeout
		if time.Now().After(deadline) {
			if mac != "" {
				return fmt.Errorf("timeout waiting for IP address on interface %s after %d seconds", mac, timeout)
			}
			return fmt.Errorf("timeout waiting for IP address after %d seconds", timeout)
		}

		// Wait before next poll
		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled while waiting for IP")
		case <-time.After(pollInterval):
			// Continue polling
		}
	}
}
