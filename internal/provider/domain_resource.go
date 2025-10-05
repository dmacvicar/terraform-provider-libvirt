package provider

import (
	"context"
	"fmt"
	"time"

	golibvirt "github.com/digitalocean/go-libvirt"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// DomainResourceModel describes the resource data model
type DomainResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	UUID        types.String `tfsdk:"uuid"`
	HWUUID      types.String `tfsdk:"hwuuid"`
	Type           types.String `tfsdk:"type"`
	Title          types.String `tfsdk:"title"`
	Description    types.String `tfsdk:"description"`
	Bootloader     types.String `tfsdk:"bootloader"`
	BootloaderArgs types.String `tfsdk:"bootloader_args"`
	Memory         types.Int64  `tfsdk:"memory"`
	Unit           types.String `tfsdk:"unit"`
	CurrentMemory  types.Int64  `tfsdk:"current_memory"`
	MaxMemory      types.Int64  `tfsdk:"max_memory"`
	MaxMemorySlots types.Int64  `tfsdk:"max_memory_slots"`
	VCPU           types.Int64  `tfsdk:"vcpu"`
	OnPoweroff types.String `tfsdk:"on_poweroff"`
	OnReboot   types.String `tfsdk:"on_reboot"`
	OnCrash    types.String `tfsdk:"on_crash"`
	IOThreads  types.Int64  `tfsdk:"iothreads"`

	// Blocks
	OS       *DomainOSModel       `tfsdk:"os"`
	Features *DomainFeaturesModel `tfsdk:"features"`
	CPU      *DomainCPUModel      `tfsdk:"cpu"`
	Clock    *DomainClockModel    `tfsdk:"clock"`
	PM       *DomainPMModel       `tfsdk:"pm"`
	Disks    []DomainDiskModel    `tfsdk:"disk"`

	// TODO: Add more fields as we implement them:
	// - iothreads
	// - current_memory, max_memory
	// - features
	// - cpu
	// - clock
	// - devices
	// - etc.
}

// DomainOSModel describes the OS configuration
type DomainOSModel struct {
	Type           types.String `tfsdk:"type"`
	Arch           types.String `tfsdk:"arch"`
	Machine        types.String `tfsdk:"machine"`
	Firmware       types.String `tfsdk:"firmware"`
	BootDevices    types.List   `tfsdk:"boot_devices"`
	Kernel         types.String `tfsdk:"kernel"`
	Initrd         types.String `tfsdk:"initrd"`
	KernelArgs     types.String `tfsdk:"kernel_args"`
	LoaderPath     types.String `tfsdk:"loader_path"`
	LoaderReadOnly types.Bool   `tfsdk:"loader_readonly"`
	LoaderType     types.String `tfsdk:"loader_type"`
	NVRAMPath      types.String `tfsdk:"nvram_path"`
}

// DomainCPUModel describes CPU configuration
type DomainCPUModel struct {
	Mode               types.String `tfsdk:"mode"`
	Match              types.String `tfsdk:"match"`
	Check              types.String `tfsdk:"check"`
	Migratable         types.String `tfsdk:"migratable"`
	DeprecatedFeatures types.String `tfsdk:"deprecated_features"`
	Model              types.String `tfsdk:"model"`
	Vendor             types.String `tfsdk:"vendor"`
}

// DomainClockModel describes clock configuration
type DomainClockModel struct {
	Offset     types.String `tfsdk:"offset"`
	Basis      types.String `tfsdk:"basis"`
	Adjustment types.String `tfsdk:"adjustment"`
	TimeZone   types.String `tfsdk:"timezone"`
}

// DomainPMModel describes power management configuration
type DomainPMModel struct {
	SuspendToMem  types.String `tfsdk:"suspend_to_mem"`
	SuspendToDisk types.String `tfsdk:"suspend_to_disk"`
}

// DomainDiskModel describes a disk device
type DomainDiskModel struct {
	Device types.String `tfsdk:"device"`
	Source types.String `tfsdk:"source"`
	Target types.String `tfsdk:"target"`
	Bus    types.String `tfsdk:"bus"`
}

// DomainFeaturesModel describes VM features
type DomainFeaturesModel struct {
	PAE          types.Bool   `tfsdk:"pae"`
	ACPI         types.Bool   `tfsdk:"acpi"`
	APIC         types.Bool   `tfsdk:"apic"`
	Viridian     types.Bool   `tfsdk:"viridian"`
	PrivNet      types.Bool   `tfsdk:"privnet"`
	HAP          types.String `tfsdk:"hap"`
	PMU          types.String `tfsdk:"pmu"`
	VMPort       types.String `tfsdk:"vmport"`
	PVSpinlock   types.String `tfsdk:"pvspinlock"`
	VMCoreInfo   types.String `tfsdk:"vmcoreinfo"`
	HTM          types.String `tfsdk:"htm"`
	NestedHV     types.String `tfsdk:"nested_hv"`
	CCFAssist    types.String `tfsdk:"ccf_assist"`
	RAS          types.String `tfsdk:"ras"`
	PS2          types.String `tfsdk:"ps2"`
	IOAPICDriver types.String `tfsdk:"ioapic_driver"`
	GICVersion   types.String `tfsdk:"gic_version"`
}

// Metadata returns the resource type name
func (r *DomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

// Schema defines the schema for the resource
func (r *DomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a libvirt domain (virtual machine).",
		MarkdownDescription: `
Manages a libvirt domain (virtual machine).

This resource follows the [libvirt domain XML schema](https://libvirt.org/formatdomain.html) closely,
providing fine-grained control over VM configuration.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Domain identifier (UUID)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Domain name. Must be unique on the host.",
				Required:    true,
			},
			"uuid": schema.StringAttribute{
				Description: "Domain UUID. If not specified, one will be generated.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hwuuid": schema.StringAttribute{
				Description: "Hardware UUID for the domain.",
				Optional:    true,
			},
			"memory": schema.Int64Attribute{
				Description: "Maximum memory allocation in the specified unit. Default unit is KiB.",
				Required:    true,
			},
			"unit": schema.StringAttribute{
				Description: "Memory unit (KiB, MiB, GiB, TiB). Defaults to KiB if not specified.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"current_memory": schema.Int64Attribute{
				Description: "Actual memory allocation at boot time. If not set, defaults to memory value.",
				Optional:    true,
			},
			"max_memory": schema.Int64Attribute{
				Description: "Maximum memory for hotplug. Must be >= memory.",
				Optional:    true,
			},
			"max_memory_slots": schema.Int64Attribute{
				Description: "Number of slots for memory hotplug. Required when max_memory is set.",
				Optional:    true,
			},
			"vcpu": schema.Int64Attribute{
				Description: "Number of virtual CPUs.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Domain type (e.g., 'kvm', 'qemu'). Defaults to 'kvm'.",
				Optional:    true,
				Computed:    true,
			},
			"title": schema.StringAttribute{
				Description: "Short description title for the domain.",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of the domain.",
				Optional:    true,
			},
			"bootloader": schema.StringAttribute{
				Description: "Bootloader path for paravirtualized guests (Xen).",
				Optional:    true,
			},
			"bootloader_args": schema.StringAttribute{
				Description: "Arguments to pass to bootloader.",
				Optional:    true,
			},
			"on_poweroff": schema.StringAttribute{
				Description: "Action to take when guest requests poweroff (destroy, restart, preserve, rename-restart).",
				Optional:    true,
			},
			"on_reboot": schema.StringAttribute{
				Description: "Action to take when guest requests reboot (destroy, restart, preserve, rename-restart).",
				Optional:    true,
			},
			"on_crash": schema.StringAttribute{
				Description: "Action to take when guest crashes (destroy, restart, preserve, rename-restart, coredump-destroy, coredump-restart).",
				Optional:    true,
			},
			"iothreads": schema.Int64Attribute{
				Description: "Number of I/O threads for virtio disks.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"os": schema.SingleNestedBlock{
				Description: "Operating system configuration for the domain.",
				MarkdownDescription: `
Operating system configuration. See [libvirt OS element documentation](https://libvirt.org/formatdomain.html#operating-system-booting).
`,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "OS type (e.g., 'hvm' for fully virtualized, 'linux' for paravirtualized). Required.",
						Required:    true,
					},
					"arch": schema.StringAttribute{
						Description: "CPU architecture (e.g., 'x86_64', 'aarch64'). Optional.",
						Optional:    true,
					},
					"machine": schema.StringAttribute{
						Description: "Machine type (e.g., 'pc', 'q35'). Optional. " +
							"Note: This value represents what you want, but libvirt may internally expand it to a versioned type.",
						Optional: true,
					},
					"firmware": schema.StringAttribute{
						Description: "Firmware type (e.g., 'efi', 'bios'). Optional.",
						Optional:    true,
					},
					"boot_devices": schema.ListAttribute{
						Description: "Ordered list of boot devices (e.g., 'hd', 'network', 'cdrom'). Optional. " +
							"If not specified, libvirt may add default boot devices.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					"kernel": schema.StringAttribute{
						Description: "Path to kernel image for direct kernel boot. Optional.",
						Optional:    true,
					},
					"initrd": schema.StringAttribute{
						Description: "Path to initrd image for direct kernel boot. Optional.",
						Optional:    true,
					},
					"kernel_args": schema.StringAttribute{
						Description: "Kernel command line arguments. Optional.",
						Optional:    true,
					},
					"loader_path": schema.StringAttribute{
						Description: "Path to UEFI firmware loader. Optional.",
						Optional:    true,
					},
					"loader_readonly": schema.BoolAttribute{
						Description: "Whether the UEFI firmware is read-only. Optional.",
						Optional:    true,
					},
					"loader_type": schema.StringAttribute{
						Description: "Loader type ('rom' or 'pflash'). Optional.",
						Optional:    true,
					},
					"nvram_path": schema.StringAttribute{
						Description: "Path to NVRAM template for UEFI. Optional.",
						Optional:    true,
					},
				},
			},
			"features": schema.SingleNestedBlock{
				Description: "Hypervisor features to enable.",
				Attributes: map[string]schema.Attribute{
					"pae": schema.BoolAttribute{
						Description: "Physical Address Extension mode.",
						Optional:    true,
					},
					"acpi": schema.BoolAttribute{
						Description: "ACPI support.",
						Optional:    true,
					},
					"apic": schema.BoolAttribute{
						Description: "APIC support.",
						Optional:    true,
					},
					"viridian": schema.BoolAttribute{
						Description: "Viridian enlightenments for Windows guests.",
						Optional:    true,
					},
					"privnet": schema.BoolAttribute{
						Description: "Private network namespace.",
						Optional:    true,
					},
					"hap": schema.StringAttribute{
						Description: "Hardware Assisted Paging (on, off).",
						Optional:    true,
					},
					"pmu": schema.StringAttribute{
						Description: "Performance Monitoring Unit (on, off).",
						Optional:    true,
					},
					"vmport": schema.StringAttribute{
						Description: "VMware IO port emulation (on, off, auto).",
						Optional:    true,
					},
					"pvspinlock": schema.StringAttribute{
						Description: "Paravirtualized spinlock prevention (on, off).",
						Optional:    true,
					},
					"vmcoreinfo": schema.StringAttribute{
						Description: "VM crash information (on, off).",
						Optional:    true,
					},
					"htm": schema.StringAttribute{
						Description: "Hardware Transactional Memory (on, off).",
						Optional:    true,
					},
					"nested_hv": schema.StringAttribute{
						Description: "Nested HV (on, off).",
						Optional:    true,
					},
					"ccf_assist": schema.StringAttribute{
						Description: "CCF Assist (on, off).",
						Optional:    true,
					},
					"ras": schema.StringAttribute{
						Description: "Reliability, Availability and Serviceability (on, off).",
						Optional:    true,
					},
					"ps2": schema.StringAttribute{
						Description: "PS/2 controller (on, off).",
						Optional:    true,
					},
					"ioapic_driver": schema.StringAttribute{
						Description: "IOAPIC driver (kvm, qemu).",
						Optional:    true,
					},
					"gic_version": schema.StringAttribute{
						Description: "GIC version for ARM guests.",
						Optional:    true,
					},
				},
			},
			"cpu": schema.SingleNestedBlock{
				Description: "CPU configuration for the domain.",
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						Description: "CPU mode (host-model, host-passthrough, custom).",
						Optional:    true,
					},
					"match": schema.StringAttribute{
						Description: "CPU match policy (exact, minimum, strict).",
						Optional:    true,
					},
					"check": schema.StringAttribute{
						Description: "CPU check policy (none, partial, full).",
						Optional:    true,
					},
					"migratable": schema.StringAttribute{
						Description: "Whether the CPU is migratable (on, off).",
						Optional:    true,
					},
					"deprecated_features": schema.StringAttribute{
						Description: "How to handle deprecated features (allow, forbid).",
						Optional:    true,
					},
					"model": schema.StringAttribute{
						Description: "CPU model name.",
						Optional:    true,
					},
					"vendor": schema.StringAttribute{
						Description: "CPU vendor.",
						Optional:    true,
					},
				},
			},
			"clock": schema.SingleNestedBlock{
				Description: "Clock configuration for the domain.",
				Attributes: map[string]schema.Attribute{
					"offset": schema.StringAttribute{
						Description: "Clock offset (utc, localtime, timezone, variable).",
						Optional:    true,
					},
					"basis": schema.StringAttribute{
						Description: "Clock basis (utc, localtime).",
						Optional:    true,
					},
					"adjustment": schema.StringAttribute{
						Description: "Clock adjustment in seconds.",
						Optional:    true,
					},
					"timezone": schema.StringAttribute{
						Description: "Timezone name when offset is 'timezone'.",
						Optional:    true,
					},
				},
			},
			"pm": schema.SingleNestedBlock{
				Description: "Power management configuration for the domain.",
				Attributes: map[string]schema.Attribute{
					"suspend_to_mem": schema.StringAttribute{
						Description: "Suspend to memory policy (yes, no).",
						Optional:    true,
					},
					"suspend_to_disk": schema.StringAttribute{
						Description: "Suspend to disk policy (yes, no).",
						Optional:    true,
					},
				},
			},
			"disk": schema.ListNestedBlock{
				Description: "Disk devices attached to the domain.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"device": schema.StringAttribute{
							Description: "Device type (disk, cdrom, floppy, lun).",
							Optional:    true,
						},
						"source": schema.StringAttribute{
							Description: "Path to the disk image file.",
							Optional:    true,
						},
						"target": schema.StringAttribute{
							Description: "Target device name (e.g., vda, sda, hda).",
							Required:    true,
						},
						"bus": schema.StringAttribute{
							Description: "Bus type (virtio, scsi, ide, sata, usb).",
							Optional:    true,
						},
					},
				},
			},
		},
	}
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

	// Convert Terraform model to libvirt XML
	domainXML, err := domainModelToXML(&plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Domain Configuration",
			"Failed to convert domain configuration to XML: "+err.Error(),
		)
		return
	}

	// Marshal to XML string
	xmlString, err := libvirt.MarshalDomainXML(domainXML)
	if err != nil {
		resp.Diagnostics.AddError(
			"XML Marshaling Failed",
			"Failed to marshal domain XML: "+err.Error(),
		)
		return
	}

	// Define the domain in libvirt
	domain, err := r.client.Libvirt().DomainDefineXML(xmlString)
	if err != nil {
		resp.Diagnostics.AddError(
			"Domain Creation Failed",
			"Failed to define domain in libvirt: "+err.Error(),
		)
		return
	}

	// Get the domain XML back to capture UUID and other computed fields
	xmlDesc, err := r.client.Libvirt().DomainGetXMLDesc(domain, 0)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Domain",
			"Domain was created but failed to read back its configuration: "+err.Error(),
		)
		return
	}

	// Parse the returned XML
	parsedDomain, err := libvirt.UnmarshalDomainXML(xmlDesc)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Parse Domain XML",
			"Failed to parse domain XML from libvirt: "+err.Error(),
		)
		return
	}

	// Update state with computed values
	xmlToDomainModel(parsedDomain, &plan)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the domain state
func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Look up domain by UUID
	domain, err := r.client.LookupDomainByUUID(state.UUID.ValueString())
	if err != nil {
		// Domain not found - remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Get the domain XML
	xmlDesc, err := r.client.Libvirt().DomainGetXMLDesc(domain, 0)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Domain",
			"Failed to get domain XML: "+err.Error(),
		)
		return
	}

	// Parse the XML
	parsedDomain, err := libvirt.UnmarshalDomainXML(xmlDesc)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Parse Domain XML",
			"Failed to parse domain XML from libvirt: "+err.Error(),
		)
		return
	}

	// Update state
	xmlToDomainModel(parsedDomain, &state)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// waitForDomainState waits for a domain to reach the specified state with a timeout
func waitForDomainState(client *libvirt.Client, domain golibvirt.Domain, targetState int32, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		state, _, err := client.Libvirt().DomainGetState(domain, 0)
		if err != nil {
			return fmt.Errorf("failed to get domain state: %w", err)
		}
		if state == targetState {
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

	// Look up the existing domain
	oldDomain, err := r.client.LookupDomainByUUID(state.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Domain Lookup Failed",
			"Failed to look up existing domain: "+err.Error(),
		)
		return
	}

	// Check if domain is running - we need to shut it down for updates
	domainState, _, err := r.client.Libvirt().DomainGetState(oldDomain, 0)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Get Domain State",
			"Failed to check if domain is running: "+err.Error(),
		)
		return
	}

	wasRunning := domainState == 1 // 1 = VIR_DOMAIN_RUNNING

	// If domain is running, shut it down first
	if wasRunning {
		// Try graceful shutdown first
		err = r.client.Libvirt().DomainShutdown(oldDomain)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Shutdown Domain",
				"Domain must be stopped before updating. Failed to shutdown: "+err.Error(),
			)
			return
		}

		// Wait for shutdown with 5 second timeout
		err = waitForDomainState(r.client, oldDomain, 5, 5*time.Second) // 5 = VIR_DOMAIN_SHUTOFF
		if err != nil {
			// Graceful shutdown failed, force it
			err = r.client.Libvirt().DomainDestroy(oldDomain)
			if err != nil {
				resp.Diagnostics.AddError(
					"Failed to Stop Domain",
					"Domain must be stopped before updating. Failed to force stop: "+err.Error(),
				)
				return
			}
		}
	}

	// Convert Terraform model to libvirt XML
	domainXML, err := domainModelToXML(&plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Domain Configuration",
			"Failed to convert domain configuration to XML: "+err.Error(),
		)
		return
	}

	// Preserve UUID from state
	domainXML.UUID = state.UUID.ValueString()

	// Marshal to XML string
	xmlString, err := libvirt.MarshalDomainXML(domainXML)
	if err != nil {
		resp.Diagnostics.AddError(
			"XML Marshaling Failed",
			"Failed to marshal domain XML: "+err.Error(),
		)
		return
	}

	// Undefine the old domain
	err = r.client.Libvirt().DomainUndefine(oldDomain)
	if err != nil {
		resp.Diagnostics.AddError(
			"Domain Undefine Failed",
			"Failed to undefine existing domain: "+err.Error(),
		)
		return
	}

	// Define the updated domain
	_, err = r.client.Libvirt().DomainDefineXML(xmlString)
	if err != nil {
		resp.Diagnostics.AddError(
			"Domain Update Failed",
			"Failed to define updated domain in libvirt: "+err.Error(),
		)
		return
	}

	// If domain was running before, start it again
	if wasRunning {
		updatedDomain, err := r.client.LookupDomainByUUID(state.UUID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Lookup Updated Domain",
				"Domain was updated but failed to look it up for restart: "+err.Error(),
			)
			return
		}

		err = r.client.Libvirt().DomainCreate(updatedDomain)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Failed to Restart Domain",
				"Domain was updated but failed to restart automatically: "+err.Error(),
			)
			// Continue anyway - domain is defined correctly
		}
	}

	// Read back the domain to get updated state
	updatedDomain, err := r.client.LookupDomainByUUID(state.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Updated Domain",
			"Domain was updated but failed to read it back: "+err.Error(),
		)
		return
	}

	xmlDesc, err := r.client.Libvirt().DomainGetXMLDesc(updatedDomain, 0)
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

	xmlToDomainModel(parsedDomain, &plan)

	// Save updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the domain
func (r *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
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
	if domainState == 1 { // running
		err = r.client.Libvirt().DomainDestroy(domain)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Destroy Domain",
				"Failed to stop running domain: "+err.Error(),
			)
			return
		}
	}

	// Undefine the domain
	err = r.client.Libvirt().DomainUndefine(domain)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Undefine Domain",
			"Failed to undefine domain: "+err.Error(),
		)
		return
	}
}
