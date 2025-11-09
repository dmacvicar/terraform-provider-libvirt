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

// DomainResourceModel describes the resource data model
type DomainResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	UUID           types.String `tfsdk:"uuid"`
	HWUUID         types.String `tfsdk:"hwuuid"`
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
	OnPoweroff     types.String `tfsdk:"on_poweroff"`
	OnReboot       types.String `tfsdk:"on_reboot"`
	OnCrash        types.String `tfsdk:"on_crash"`
	IOThreads      types.Int64  `tfsdk:"iothreads"`
	Running        types.Bool   `tfsdk:"running"`
	Autostart      types.Bool   `tfsdk:"autostart"`
	Metadata       types.String `tfsdk:"metadata"`

	OS       types.Object `tfsdk:"os"`
	Features types.Object `tfsdk:"features"`
	CPU      types.Object `tfsdk:"cpu"`
	PM       types.Object `tfsdk:"pm"`
	Create   types.Object `tfsdk:"create"`
	Devices  types.Object `tfsdk:"devices"`
	Destroy  types.Object `tfsdk:"destroy"`
	Clock    types.Object `tfsdk:"clock"`
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
	NVRAM          types.Object `tfsdk:"nvram"`
}

// DomainNVRAMModel describes NVRAM configuration
type DomainNVRAMModel struct {
	Path           types.String `tfsdk:"path"`
	Source         types.Object `tfsdk:"source"`
	Template       types.String `tfsdk:"template"`
	Format         types.String `tfsdk:"format"`
	TemplateFormat types.String `tfsdk:"template_format"`
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
	Timers     types.List   `tfsdk:"timer"`
}

// DomainTimerModel describes a clock timer
type DomainTimerModel struct {
	Name       types.String `tfsdk:"name"`
	Track      types.String `tfsdk:"track"`
	TickPolicy types.String `tfsdk:"tickpolicy"`
	Frequency  types.Int64  `tfsdk:"frequency"`
	Mode       types.String `tfsdk:"mode"`
	Present    types.String `tfsdk:"present"`
	CatchUp    types.Object `tfsdk:"catchup"`
}

// DomainTimerCatchUpModel describes timer catchup configuration
type DomainTimerCatchUpModel struct {
	Threshold types.Int64 `tfsdk:"threshold"`
	Slew      types.Int64 `tfsdk:"slew"`
	Limit     types.Int64 `tfsdk:"limit"`
}

// DomainPMModel describes power management configuration
type DomainPMModel struct {
	SuspendToMem  types.String `tfsdk:"suspend_to_mem"`
	SuspendToDisk types.String `tfsdk:"suspend_to_disk"`
}

// DomainDiskModel describes a disk device
type DomainDiskModel struct {
	Device types.String           `tfsdk:"device"`
	Source *DomainDiskSourceModel `tfsdk:"source"`
	Target *DomainDiskTargetModel `tfsdk:"target"`
	WWN    types.String           `tfsdk:"wwn"`
}

// DomainDiskSourceModel describes the disk source
type DomainDiskSourceModel struct {
	Pool   types.String `tfsdk:"pool"`
	Volume types.String `tfsdk:"volume"`
	File   types.String `tfsdk:"file"`
	Block  types.String `tfsdk:"block"`
}

// DomainDiskTargetModel describes the disk target mapping
type DomainDiskTargetModel struct {
	Dev types.String `tfsdk:"dev"`
	Bus types.String `tfsdk:"bus"`
}

// DomainInterfaceModel describes a network interface
type DomainInterfaceModel struct {
	Type      types.String                `tfsdk:"type"`
	MAC       types.String                `tfsdk:"mac"`
	Model     types.String                `tfsdk:"model"`
	Source    *DomainInterfaceSourceModel `tfsdk:"source"`
	WaitForIP types.Object                `tfsdk:"wait_for_ip"`
}

// DomainInterfaceSourceModel describes the interface source
type DomainInterfaceSourceModel struct {
	Network   types.String `tfsdk:"network"`
	PortGroup types.String `tfsdk:"portgroup"`
	Bridge    types.String `tfsdk:"bridge"`
	Dev       types.String `tfsdk:"dev"`
	Mode      types.String `tfsdk:"mode"`
}

// DomainInterfaceWaitForIPModel describes wait_for_ip configuration
type DomainInterfaceWaitForIPModel struct {
	Timeout types.Int64  `tfsdk:"timeout"`
	Source  types.String `tfsdk:"source"`
}

// DomainGraphicsModel describes a graphics device
type DomainGraphicsModel struct {
	VNC   types.Object `tfsdk:"vnc"`
	Spice types.Object `tfsdk:"spice"`
}

// DomainGraphicsVNCModel describes VNC graphics configuration
type DomainGraphicsVNCModel struct {
	Socket    types.String `tfsdk:"socket"`
	Port      types.Int64  `tfsdk:"port"`
	AutoPort  types.String `tfsdk:"autoport"`
	WebSocket types.Int64  `tfsdk:"websocket"`
	Listen    types.String `tfsdk:"listen"`
}

// DomainGraphicsSpiceModel describes Spice graphics configuration
type DomainGraphicsSpiceModel struct {
	Port     types.Int64  `tfsdk:"port"`
	TLSPort  types.Int64  `tfsdk:"tlsport"`
	AutoPort types.String `tfsdk:"autoport"`
	Listen   types.String `tfsdk:"listen"`
}

// DomainDevicesModel describes all devices in the domain
type DomainDevicesModel struct {
	Disks       types.List   `tfsdk:"disks"`
	Interfaces  types.List   `tfsdk:"interfaces"`
	Graphics    types.Object `tfsdk:"graphics"`
	Filesystems types.List   `tfsdk:"filesystems"`
	Video       types.Object `tfsdk:"video"`
	Emulator    types.String `tfsdk:"emulator"`
	Consoles    types.List   `tfsdk:"consoles"`
	Serials     types.List   `tfsdk:"serials"`
	RNGs        types.List   `tfsdk:"rngs"`
	TPMs        types.List   `tfsdk:"tpms"`
	Inputs      types.List   `tfsdk:"inputs"`
}

// DomainFilesystemModel describes a filesystem device
type DomainFilesystemModel struct {
	AccessMode types.String `tfsdk:"accessmode"`
	Source     types.String `tfsdk:"source"`
	Target     types.String `tfsdk:"target"`
	ReadOnly   types.Bool   `tfsdk:"readonly"`
}

// DomainVideoModel describes a video device
type DomainVideoModel struct {
	Type types.String `tfsdk:"type"`
}

// DomainConsoleModel describes a console device
type DomainConsoleModel struct {
	Type       types.String `tfsdk:"type"`
	SourcePath types.String `tfsdk:"source_path"`
	TargetType types.String `tfsdk:"target_type"`
	TargetPort types.Int64  `tfsdk:"target_port"`
}

// DomainSerialModel describes a serial device
type DomainSerialModel struct {
	Type       types.String `tfsdk:"type"`
	SourcePath types.String `tfsdk:"source_path"`
	TargetType types.String `tfsdk:"target_type"`
	TargetPort types.Int64  `tfsdk:"target_port"`
}

// DomainRNGModel describes an RNG device
type DomainRNGModel struct {
	Model  types.String `tfsdk:"model"`
	Device types.String `tfsdk:"device"`
}

// DomainTPMModel describes a TPM device
type DomainTPMModel struct {
	Model                   types.String `tfsdk:"model"`
	BackendType             types.String `tfsdk:"backend_type"`
	BackendDevicePath       types.String `tfsdk:"backend_device_path"`
	BackendEncryptionSecret types.String `tfsdk:"backend_encryption_secret"`
	BackendVersion          types.String `tfsdk:"backend_version"`
	BackendPersistentState  types.Bool   `tfsdk:"backend_persistent_state"`
}

// DomainInputModel describes an input device
type DomainInputModel struct {
	Type   types.String `tfsdk:"type"`
	Bus    types.String `tfsdk:"bus"`
	Model  types.String `tfsdk:"model"`
	Driver types.Object `tfsdk:"driver"`
	Source types.Object `tfsdk:"source"`
}

// DomainInputDriverModel describes input device driver options
type DomainInputDriverModel struct {
	IOMMU     types.String `tfsdk:"iommu"`
	ATS       types.String `tfsdk:"ats"`
	Packed    types.String `tfsdk:"packed"`
	PagePerVQ types.String `tfsdk:"page_per_vq"`
}

// DomainInputSourceModel describes input device source
type DomainInputSourceModel struct {
	Passthrough types.Object `tfsdk:"passthrough"`
	EVDev       types.Object `tfsdk:"evdev"`
}

// DomainInputSourcePassthroughModel describes passthrough input source
type DomainInputSourcePassthroughModel struct {
	EVDev types.String `tfsdk:"evdev"`
}

// DomainInputSourceEVDevModel describes evdev input source
type DomainInputSourceEVDevModel struct {
	Dev        types.String `tfsdk:"dev"`
	Grab       types.String `tfsdk:"grab"`
	GrabToggle types.String `tfsdk:"grab_toggle"`
	Repeat     types.String `tfsdk:"repeat"`
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
				Description: "Memory unit (KiB, MiB, GiB, TiB).",
				Optional:    true,
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
				Description: "Domain type (e.g., 'kvm', 'qemu').",
				Required:    true,
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
			"running": schema.BoolAttribute{
				Description: "Whether the domain should be running. If true, the domain will be started after creation. If false or unset, the domain will only be defined but not started.",
				Optional:    true,
			},
			"autostart": schema.BoolAttribute{
				Description: "Whether the domain should be started automatically when the host boots.",
				Optional:    true,
			},
			"metadata": schema.StringAttribute{
				Description: "Custom metadata XML for the domain. Must be valid XML using custom namespaces. Applications must use custom namespaces on their XML nodes/trees, with only one top-level element per namespace. Example: '<app1:foo xmlns:app1=\"http://app1.org/app1/\">content</app1:foo>'",
				MarkdownDescription: `Custom metadata XML for the domain.

This allows applications to store custom information within the domain's XML configuration.

**Requirements:**
- Must be valid XML
- Must use custom namespaces (e.g., ` + "`xmlns:app1=\"http://app1.org/app1/\"`" + `)
- Only one top-level element per namespace

**Example:**
` + "```xml" + `
<app1:foo xmlns:app1="http://app1.org/app1/">
  <app1:bar>some content</app1:bar>
</app1:foo>
` + "```" + `

See [libvirt metadata documentation](https://libvirt.org/formatdomain.html#metadata) for more details.`,
				Optional: true,
			},
			"features": schema.SingleNestedAttribute{
				Description: "Hypervisor features to enable.",
				Optional:    true,
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
			"cpu": schema.SingleNestedAttribute{
				Description: "CPU configuration for the domain.",
				Optional:    true,
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
			"pm": schema.SingleNestedAttribute{
				Description: "Power management configuration for the domain.",
				Optional:    true,
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
			"create": schema.SingleNestedAttribute{
				Description: "Domain start flags. Only used when running=true. Controls how the domain is started.",
				MarkdownDescription: `
Domain start flags corresponding to virDomainCreateFlags. Only used when running=true.

See [libvirt domain documentation](https://libvirt.org/html/libvirt-libvirt-domain.html#virDomainCreateFlags).
`,
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"paused": schema.BoolAttribute{
						Description: "Launch domain in paused state (VIR_DOMAIN_START_PAUSED).",
						Optional:    true,
					},
					"autodestroy": schema.BoolAttribute{
						Description: "Automatically destroy domain when connection closes (VIR_DOMAIN_START_AUTODESTROY).",
						Optional:    true,
					},
					"bypass_cache": schema.BoolAttribute{
						Description: "Avoid filesystem cache pollution (VIR_DOMAIN_START_BYPASS_CACHE).",
						Optional:    true,
					},
					"force_boot": schema.BoolAttribute{
						Description: "Boot domain, discarding any managed save state (VIR_DOMAIN_START_FORCE_BOOT).",
						Optional:    true,
					},
					"validate": schema.BoolAttribute{
						Description: "Validate XML document against schema (VIR_DOMAIN_START_VALIDATE).",
						Optional:    true,
					},
					"reset_nvram": schema.BoolAttribute{
						Description: "Re-initialize NVRAM from template (VIR_DOMAIN_START_RESET_NVRAM).",
						Optional:    true,
					},
				},
			},
			"os": schema.SingleNestedAttribute{
				Description: "Operating system configuration for the domain.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "OS type (e.g., 'hvm' for fully virtualized, 'linux' for paravirtualized).",
						Required:    true,
					},
					"arch": schema.StringAttribute{
						Description: "CPU architecture (e.g., 'x86_64', 'aarch64').",
						Optional:    true,
					},
					"machine": schema.StringAttribute{
						Description: "Machine type (e.g., 'pc', 'q35'). Note: This value represents what you want, but libvirt may internally expand it to a versioned type.",
						Optional:    true,
					},
					"firmware": schema.StringAttribute{
						Description: "Firmware type (e.g., 'efi', 'bios').",
						Optional:    true,
					},
					"boot_devices": schema.ListAttribute{
						Description: "Ordered list of boot devices (e.g., 'hd', 'network', 'cdrom'). If not specified, libvirt may add default boot devices.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					"kernel": schema.StringAttribute{
						Description: "Path to kernel image for direct kernel boot.",
						Optional:    true,
					},
					"initrd": schema.StringAttribute{
						Description: "Path to initrd image for direct kernel boot.",
						Optional:    true,
					},
					"kernel_args": schema.StringAttribute{
						Description: "Kernel command line arguments.",
						Optional:    true,
					},
					"loader_path": schema.StringAttribute{
						Description: "Path to UEFI firmware loader.",
						Optional:    true,
					},
					"loader_readonly": schema.BoolAttribute{
						Description: "Whether the UEFI firmware is read-only.",
						Optional:    true,
					},
					"loader_type": schema.StringAttribute{
						Description: "Loader type ('rom' or 'pflash').",
						Optional:    true,
					},
					"nvram": schema.SingleNestedAttribute{
						Description: "NVRAM configuration for UEFI.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"path": schema.StringAttribute{
								Description: "Path to the NVRAM file for the domain. Mutually exclusive with source.",
								Optional:    true,
							},
							"source": schema.SingleNestedAttribute{
								Description: "NVRAM source configuration for volume-based NVRAM. Mutually exclusive with path.",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"pool": schema.StringAttribute{
										Description: "Storage pool name for volume-based NVRAM. Use with 'volume'.",
										Optional:    true,
									},
									"volume": schema.StringAttribute{
										Description: "Volume name in the storage pool. Use with 'pool'.",
										Optional:    true,
									},
									"file": schema.StringAttribute{
										Description: "Path to NVRAM file. Mutually exclusive with pool/volume and block.",
										Optional:    true,
									},
									"block": schema.StringAttribute{
										Description: "Block device path. Mutually exclusive with pool/volume and file.",
										Optional:    true,
									},
								},
							},
							"template": schema.StringAttribute{
								Description: "Path to NVRAM template file for UEFI variable store. This template is copied to create the domain's NVRAM.",
								Optional:    true,
							},
							"format": schema.StringAttribute{
								Description: "Format of the NVRAM file (e.g., 'raw', 'qcow2').",
								Optional:    true,
							},
							"template_format": schema.StringAttribute{
								Description: "Format of the template file (e.g., 'raw', 'qcow2').",
								Optional:    true,
							},
						},
					},
				},
			},
			"devices": schema.SingleNestedAttribute{
				Description: "Devices attached to the domain (disks, network interfaces, etc.).",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"disks": schema.ListNestedAttribute{
						Description: "Disk devices attached to the domain.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"device": schema.StringAttribute{
									Description: "Device type (disk, cdrom, floppy, lun).",
									Optional:    true,
								},
								"source": schema.SingleNestedAttribute{
									Description: "Disk source configuration. Specify one of: pool+volume for libvirt volumes, file for file paths, or block for block devices.",
									Optional:    true,
									Attributes: map[string]schema.Attribute{
										"pool": schema.StringAttribute{
											Description: "Storage pool name for volume-based disks. Use with 'volume'.",
											Optional:    true,
										},
										"volume": schema.StringAttribute{
											Description: "Volume name in the storage pool. Use with 'pool'.",
											Optional:    true,
										},
										"file": schema.StringAttribute{
											Description: "Path to disk image file. Mutually exclusive with pool/volume and block.",
											Optional:    true,
										},
										"block": schema.StringAttribute{
											Description: "Block device path (e.g., /dev/sdb). Mutually exclusive with pool/volume and file.",
											Optional:    true,
										},
									},
								},
								"target": schema.SingleNestedAttribute{
									Description: "Guest device target mapping.",
									Required:    true,
									Attributes: map[string]schema.Attribute{
										"dev": schema.StringAttribute{
											Description: "Target device name (e.g., vda, sda, hda).",
											Required:    true,
										},
										"bus": schema.StringAttribute{
											Description: "Bus type (virtio, scsi, ide, sata, usb).",
											Optional:    true,
										},
									},
								},
								"wwn": schema.StringAttribute{
									Description: "World Wide Name identifier for the disk (typically for SCSI disks). If not specified for SCSI disks, one will be generated. Format: 16 hex digits.",
									Optional:    true,
									Computed:    true,
								},
							},
						},
					},
					"interfaces": schema.ListNestedAttribute{
						Description: "Network interfaces attached to the domain.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Description: "Interface type (network, bridge, user, direct, etc.).",
									Required:    true,
								},
								"mac": schema.StringAttribute{
									Description: "MAC address for the interface.",
									Optional:    true,
								},
								"model": schema.StringAttribute{
									Description: "Device model (virtio, e1000, rtl8139, etc.).",
									Optional:    true,
								},
								"source": schema.SingleNestedAttribute{
									Description: "Interface source configuration.",
									Optional:    true,
									Attributes: map[string]schema.Attribute{
										"network": schema.StringAttribute{
											Description: "Network name (for type=network).",
											Optional:    true,
										},
										"portgroup": schema.StringAttribute{
											Description: "Port group name (for type=network).",
											Optional:    true,
										},
										"bridge": schema.StringAttribute{
											Description: "Bridge name (for type=bridge).",
											Optional:    true,
										},
										"dev": schema.StringAttribute{
											Description: "Device name (for type=user or type=direct).",
											Optional:    true,
										},
										"mode": schema.StringAttribute{
											Description: "Direct mode (for type=direct). Options: bridge, vepa, private, passthrough.",
											Optional:    true,
										},
									},
								},
								"wait_for_ip": schema.SingleNestedAttribute{
									Description: "Wait for IP address during domain creation. If specified, Terraform will poll for an IP address before considering creation complete. If timeout is reached without obtaining an IP, the domain will be destroyed and creation will fail.",
									Optional:    true,
									Attributes: map[string]schema.Attribute{
										"timeout": schema.Int64Attribute{
											Description: "Maximum time to wait for IP address in seconds. Default: 300 (5 minutes).",
											Optional:    true,
										},
										"source": schema.StringAttribute{
											Description: "Source to query for IP addresses: 'lease' (DHCP), 'agent' (qemu-guest-agent), or 'any' (try both). Default: 'any'.",
											Optional:    true,
										},
									},
								},
							},
						},
					},
					"graphics": schema.SingleNestedAttribute{
						Description: "Graphics device for the domain (VNC or Spice). Only one type can be specified.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"vnc": schema.SingleNestedAttribute{
								Description: "VNC graphics configuration. Mutually exclusive with spice.",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"socket": schema.StringAttribute{
										Description: "UNIX socket path for VNC server. Optional.",
										Optional:    true,
									},
									"port": schema.Int64Attribute{
										Description: "TCP port for VNC server. Use -1 for auto. Optional.",
										Optional:    true,
									},
									"autoport": schema.StringAttribute{
										Description: "Auto-allocate port (yes/no). Optional.",
										Optional:    true,
									},
									"websocket": schema.Int64Attribute{
										Description: "WebSocket port for VNC. Optional.",
										Optional:    true,
									},
									"listen": schema.StringAttribute{
										Description: "Listen address for VNC server. Optional.",
										Optional:    true,
									},
								},
							},
							"spice": schema.SingleNestedAttribute{
								Description: "Spice graphics configuration. Mutually exclusive with vnc.",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"port": schema.Int64Attribute{
										Description: "TCP port for Spice server. Use -1 for auto. Optional.",
										Optional:    true,
									},
									"tlsport": schema.Int64Attribute{
										Description: "TLS port for Spice server. Optional.",
										Optional:    true,
									},
									"autoport": schema.StringAttribute{
										Description: "Auto-allocate port (yes/no). Optional.",
										Optional:    true,
									},
									"listen": schema.StringAttribute{
										Description: "Listen address for Spice server. Optional.",
										Optional:    true,
									},
								},
							},
						},
					},
					"filesystems": schema.ListNestedAttribute{
						Description: "Filesystem devices for sharing host directories with the guest (virtio-9p).",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"accessmode": schema.StringAttribute{
									Description: "Access mode (mapped, passthrough, squash). Defaults to mapped.",
									Optional:    true,
								},
								"source": schema.StringAttribute{
									Description: "Host directory path to share.",
									Required:    true,
								},
								"target": schema.StringAttribute{
									Description: "Mount tag visible in the guest (used to mount the filesystem).",
									Required:    true,
								},
								"readonly": schema.BoolAttribute{
									Description: "Whether the filesystem should be mounted read-only in the guest. Defaults to true.",
									Optional:    true,
								},
							},
						},
					},
					"video": schema.SingleNestedAttribute{
						Description: "Video device for the domain.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Description: "Video device model type (e.g., cirrus, vga, qxl, virtio, vbox, vmvga, gop). Optional.",
								Optional:    true,
							},
						},
					},
					"emulator": schema.StringAttribute{
						Description: "Path to the emulator binary (e.g., /usr/bin/qemu-system-x86_64). Optional, libvirt chooses default if not specified.",
						Optional:    true,
					},
					"consoles": schema.ListNestedAttribute{
						Description: "Console devices for the domain.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Description: "Console source type (pty, file, unix, tcp, etc.). Optional, defaults to pty.",
									Optional:    true,
								},
								"source_path": schema.StringAttribute{
									Description: "Source path for file or unix socket types. Optional.",
									Optional:    true,
								},
								"target_type": schema.StringAttribute{
									Description: "Target type (serial, virtio, xen, etc.). Optional.",
									Optional:    true,
								},
								"target_port": schema.Int64Attribute{
									Description: "Target port number. Optional.",
									Optional:    true,
								},
							},
						},
					},
					"serials": schema.ListNestedAttribute{
						Description: "Serial devices for the domain.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Description: "Serial source type (pty, file, unix, tcp, etc.). Optional, defaults to pty.",
									Optional:    true,
								},
								"source_path": schema.StringAttribute{
									Description: "Source path for file or unix socket types. Optional.",
									Optional:    true,
								},
								"target_type": schema.StringAttribute{
									Description: "Target type (isa-serial, usb-serial, pci-serial, etc.). Optional.",
									Optional:    true,
								},
								"target_port": schema.Int64Attribute{
									Description: "Target port number. Optional.",
									Optional:    true,
								},
							},
						},
					},
					"rngs": schema.ListNestedAttribute{
						Description: "Random number generator devices for the domain.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"model": schema.StringAttribute{
									Description: "RNG device model (virtio, virtio-transitional, virtio-non-transitional). Defaults to virtio.",
									Optional:    true,
								},
								"device": schema.StringAttribute{
									Description: "Backend random device path (e.g., /dev/random, /dev/urandom, /dev/hwrng). Defaults to /dev/urandom.",
									Optional:    true,
								},
							},
						},
					},
					"tpms": schema.ListNestedAttribute{
						Description: "TPM devices for the domain. Only one TPM device is supported per domain.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"model": schema.StringAttribute{
									Description: "TPM device model (e.g., 'tpm-tis', 'tpm-crb', 'tpm-spapr').",
									Optional:    true,
								},
								"backend_type": schema.StringAttribute{
									Description: "TPM backend type ('passthrough', 'emulator'). Defaults to 'emulator'.",
									Optional:    true,
								},
								"backend_device_path": schema.StringAttribute{
									Description: "Device path for passthrough backend (e.g., '/dev/tpm0'). Only used with backend_type='passthrough'.",
									Optional:    true,
								},
								"backend_encryption_secret": schema.StringAttribute{
									Description: "UUID of secret for encrypted state persistence. Only used with backend_type='emulator'.",
									Optional:    true,
								},
								"backend_version": schema.StringAttribute{
									Description: "TPM backend version (e.g., '2.0'). Only used with backend_type='emulator'.",
									Optional:    true,
								},
								"backend_persistent_state": schema.BoolAttribute{
									Description: "Whether TPM state should be persistent across VM restarts. Only used with backend_type='emulator'.",
									Optional:    true,
								},
							},
						},
					},
					"inputs": schema.ListNestedAttribute{
						Description: "Input devices for the domain (keyboard, mouse, tablet, etc.).",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Description: "Input device type ('tablet', 'mouse', 'keyboard', 'passthrough', 'evdev').",
									Required:    true,
								},
								"bus": schema.StringAttribute{
									Description: "Input device bus ('ps2', 'usb', 'xen', 'virtio'). Optional for tablet/mouse/keyboard, required for passthrough/evdev.",
									Optional:    true,
								},
								"model": schema.StringAttribute{
									Description: "Input device model ('virtio', 'virtio-transitional', 'virtio-non-transitional').",
									Optional:    true,
								},
								"driver": schema.SingleNestedAttribute{
									Description: "Virtio driver options for the input device.",
									Optional:    true,
									Attributes: map[string]schema.Attribute{
										"iommu": schema.StringAttribute{
											Description: "Enable IOMMU for the device ('on', 'off').",
											Optional:    true,
										},
										"ats": schema.StringAttribute{
											Description: "Enable ATS (Address Translation Services) ('on', 'off').",
											Optional:    true,
										},
										"packed": schema.StringAttribute{
											Description: "Enable packed virtqueue layout ('on', 'off').",
											Optional:    true,
										},
										"page_per_vq": schema.StringAttribute{
											Description: "Page per virtqueue setting ('on', 'off').",
											Optional:    true,
										},
									},
								},
								"source": schema.SingleNestedAttribute{
									Description: "Input source configuration. Required for passthrough and evdev types.",
									Optional:    true,
									Attributes: map[string]schema.Attribute{
										"passthrough": schema.SingleNestedAttribute{
											Description: "Passthrough input source configuration.",
											Optional:    true,
											Attributes: map[string]schema.Attribute{
												"evdev": schema.StringAttribute{
													Description: "Event device path for passthrough (e.g., '/dev/input/event0').",
													Required:    true,
												},
											},
										},
										"evdev": schema.SingleNestedAttribute{
											Description: "EVDev input source configuration.",
											Optional:    true,
											Attributes: map[string]schema.Attribute{
												"dev": schema.StringAttribute{
													Description: "Event device path (e.g., '/dev/input/event0').",
													Required:    true,
												},
												"grab": schema.StringAttribute{
													Description: "Grab mode ('all', or unspecified).",
													Optional:    true,
												},
												"grab_toggle": schema.StringAttribute{
													Description: "Key combination for grab toggle ('ctrl-ctrl', 'alt-alt', 'shift-shift', 'meta-meta', 'scrolllock', 'ctrl-scrolllock').",
													Optional:    true,
												},
												"repeat": schema.StringAttribute{
													Description: "Enable key repeat ('on', 'off').",
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
			},
			"destroy": schema.SingleNestedAttribute{
				Description: "Domain shutdown behavior. Controls how the domain is stopped when running changes from true to false or when the resource is destroyed.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"graceful": schema.BoolAttribute{
						Description: "Attempt graceful shutdown before forcing. Defaults to true.",
						Optional:    true,
					},
					"timeout": schema.Int64Attribute{
						Description: "Timeout in seconds to wait for graceful shutdown before forcing. Defaults to 300.",
						Optional:    true,
					},
				},
			},
			"clock": schema.SingleNestedAttribute{
				Description: "Clock configuration for the domain.",
				Optional:    true,
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
					"timer": schema.ListNestedAttribute{
						Description: "Timer devices for the guest clock.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "Timer name (platform, pit, rtc, hpet, tsc, kvmclock, hypervclock, armvtimer).",
									Required:    true,
								},
								"track": schema.StringAttribute{
									Description: "Track source (guest, wall).",
									Optional:    true,
								},
								"tickpolicy": schema.StringAttribute{
									Description: "Tick policy (delay, catchup, merge, discard).",
									Optional:    true,
								},
								"frequency": schema.Int64Attribute{
									Description: "Timer frequency in Hz.",
									Optional:    true,
								},
								"mode": schema.StringAttribute{
									Description: "Timer mode (auto, native, emulate, paravirt, smpsafe).",
									Optional:    true,
								},
								"present": schema.StringAttribute{
									Description: "Whether timer is present (yes, no).",
									Optional:    true,
								},
								"catchup": schema.SingleNestedAttribute{
									Description: "Timer catchup configuration.",
									Optional:    true,
									Attributes: map[string]schema.Attribute{
										"threshold": schema.Int64Attribute{
											Description: "Threshold in nanoseconds.",
											Optional:    true,
										},
										"slew": schema.Int64Attribute{
											Description: "Slew in nanoseconds.",
											Optional:    true,
										},
										"limit": schema.Int64Attribute{
											Description: "Limit in nanoseconds.",
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
	domainXML, err := domainModelToXML(ctx, r.client, &plan)
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
		// Cleanup: undefine the domain we just defined
		if undefErr := r.client.Libvirt().DomainUndefine(domain); undefErr != nil {
			tflog.Warn(ctx, "Failed to undefine domain during cleanup", map[string]any{
				"error": undefErr.Error(),
			})
		}
		resp.Diagnostics.AddError(
			"Failed to Read Domain",
			"Domain was created but failed to read back its configuration: "+err.Error(),
		)
		return
	}

	// Parse the returned XML
	parsedDomain, err := libvirt.UnmarshalDomainXML(xmlDesc)
	if err != nil {
		// Cleanup: undefine the domain we just defined
		if undefErr := r.client.Libvirt().DomainUndefine(domain); undefErr != nil {
			tflog.Warn(ctx, "Failed to undefine domain during cleanup", map[string]any{
				"error": undefErr.Error(),
			})
		}
		resp.Diagnostics.AddError(
			"Failed to Parse Domain XML",
			"Failed to parse domain XML from libvirt: "+err.Error(),
		)
		return
	}

	// Update state with computed values
	resp.Diagnostics.Append(xmlToDomainModel(ctx, parsedDomain, &plan)...)
	if resp.Diagnostics.HasError() {
		// Cleanup: undefine the domain we just defined
		if undefErr := r.client.Libvirt().DomainUndefine(domain); undefErr != nil {
			tflog.Warn(ctx, "Failed to undefine domain during cleanup", map[string]any{
				"error": undefErr.Error(),
			})
		}
		return
	}

	// Start the domain if running=true
	if !plan.Running.IsNull() && plan.Running.ValueBool() {
		var flags uint32

		// Build flags from create block
		if !plan.Create.IsNull() && !plan.Create.IsUnknown() {
			var createModel DomainCreateModel
			resp.Diagnostics.Append(plan.Create.As(ctx, &createModel, basetypes.ObjectAsOptions{})...)
			if resp.Diagnostics.HasError() {
				return
			}

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
		}

		// Start the domain
		_, err = r.client.Libvirt().DomainCreateWithFlags(domain, flags)
		if err != nil {
			// Cleanup: undefine the domain we just defined
			if undefErr := r.client.Libvirt().DomainUndefine(domain); undefErr != nil {
				tflog.Warn(ctx, "Failed to undefine domain during cleanup", map[string]any{
					"error": undefErr.Error(),
				})
			}
			resp.Diagnostics.AddError(
				"Failed to Start Domain",
				"Domain was defined but failed to start: "+err.Error(),
			)
			return
		}

		// Wait for IP if configured on any interface
		if !plan.Devices.IsNull() && !plan.Devices.IsUnknown() {
			var devicesModel DomainDevicesModel
			resp.Diagnostics.Append(plan.Devices.As(ctx, &devicesModel, basetypes.ObjectAsOptions{})...)
			if resp.Diagnostics.HasError() {
				return
			}

			// Check if any interface has wait_for_ip configured
			if !devicesModel.Interfaces.IsNull() && !devicesModel.Interfaces.IsUnknown() {
				var interfaces []DomainInterfaceModel
				resp.Diagnostics.Append(devicesModel.Interfaces.ElementsAs(ctx, &interfaces, false)...)
				if resp.Diagnostics.HasError() {
					return
				}

				// Collect all interfaces that need to wait for IP
				for _, iface := range interfaces {
					if !iface.WaitForIP.IsNull() && !iface.WaitForIP.IsUnknown() {
						var waitForIPModel DomainInterfaceWaitForIPModel
						resp.Diagnostics.Append(iface.WaitForIP.As(ctx, &waitForIPModel, basetypes.ObjectAsOptions{})...)
						if resp.Diagnostics.HasError() {
							return
						}

						timeout := int64(300) // Default 5 minutes
						if !waitForIPModel.Timeout.IsNull() && !waitForIPModel.Timeout.IsUnknown() {
							timeout = waitForIPModel.Timeout.ValueInt64()
						}

						source := "any"
						if !waitForIPModel.Source.IsNull() && !waitForIPModel.Source.IsUnknown() {
							source = waitForIPModel.Source.ValueString()
						}

						// Get MAC address if specified (to identify this specific interface)
						mac := ""
						if !iface.MAC.IsNull() && !iface.MAC.IsUnknown() {
							mac = iface.MAC.ValueString()
						}

						// Wait for IP
						tflog.Info(ctx, "Waiting for interface IP address", map[string]any{
							"mac":     mac,
							"timeout": timeout,
							"source":  source,
						})

						if err := waitForInterfaceIP(ctx, r.client, domain, mac, timeout, source); err != nil {
							// Cleanup: destroy and undefine the domain
							if destroyErr := r.client.Libvirt().DomainDestroy(domain); destroyErr != nil {
								tflog.Warn(ctx, "Failed to destroy domain during cleanup", map[string]any{
									"error": destroyErr.Error(),
								})
							}
							if undefErr := r.client.Libvirt().DomainUndefine(domain); undefErr != nil {
								tflog.Warn(ctx, "Failed to undefine domain during cleanup", map[string]any{
									"error": undefErr.Error(),
								})
							}
							macInfo := ""
							if mac != "" {
								macInfo = fmt.Sprintf(" (MAC: %s)", mac)
							}
							resp.Diagnostics.AddError(
								"Failed to Wait for IP Address",
								fmt.Sprintf("Domain was created and started but failed to obtain an IP address%s: %s", macInfo, err),
							)
							return
						}
					}
				}
			}
		}
	}

	// Set autostart if specified
	if !plan.Autostart.IsNull() {
		autostart := int32(0)
		if plan.Autostart.ValueBool() {
			autostart = 1
		}
		err = r.client.Libvirt().DomainSetAutostart(domain, autostart)
		if err != nil {
			// Cleanup: destroy if created, then undefine
			if !plan.Running.IsNull() && plan.Running.ValueBool() {
				if destroyErr := r.client.Libvirt().DomainDestroy(domain); destroyErr != nil {
					tflog.Warn(ctx, "Failed to destroy domain during cleanup", map[string]any{
						"error": destroyErr.Error(),
					})
				}
			}
			if undefErr := r.client.Libvirt().DomainUndefine(domain); undefErr != nil {
				tflog.Warn(ctx, "Failed to undefine domain during cleanup", map[string]any{
					"error": undefErr.Error(),
				})
			}
			resp.Diagnostics.AddError(
				"Failed to Set Autostart",
				"Domain was created but failed to set autostart: "+err.Error(),
			)
			return
		}
	}

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
	resp.Diagnostics.Append(xmlToDomainModel(ctx, parsedDomain, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get autostart status - only if user originally specified it
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

	// Save state
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

	wasRunning := uint32(domainState) == uint32(golibvirt.DomainRunning)

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
		err = waitForDomainState(r.client, oldDomain, uint32(golibvirt.DomainShutoff), 5*time.Second)
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
	domainXML, err := domainModelToXML(ctx, r.client, &plan)
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

	// Handle running state transitions based on plan.Running
	shouldBeRunning := !plan.Running.IsNull() && plan.Running.ValueBool()

	if shouldBeRunning {
		updatedDomain, err := r.client.LookupDomainByUUID(state.UUID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Lookup Updated Domain",
				"Domain was updated but failed to look it up for restart: "+err.Error(),
			)
			return
		}

		// Build flags from create block
		var flags uint32
		if !plan.Create.IsNull() && !plan.Create.IsUnknown() {
			var createModel DomainCreateModel
			resp.Diagnostics.Append(plan.Create.As(ctx, &createModel, basetypes.ObjectAsOptions{})...)
			if resp.Diagnostics.HasError() {
				return
			}

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
		}

		_, err = r.client.Libvirt().DomainCreateWithFlags(updatedDomain, flags)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Failed to Start Domain",
				"Domain was updated but failed to start: "+err.Error(),
			)
			// Continue anyway - domain is defined correctly
		}
	}

	// Read back the domain to get updated state
	finalDomain, err := r.client.LookupDomainByUUID(state.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Updated Domain",
			"Domain was updated but failed to read it back: "+err.Error(),
		)
		return
	}

	xmlDesc, err := r.client.Libvirt().DomainGetXMLDesc(finalDomain, 0)
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

	resp.Diagnostics.Append(xmlToDomainModel(ctx, parsedDomain, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update autostart if it changed
	if !plan.Autostart.IsNull() {
		autostart := int32(0)
		if plan.Autostart.ValueBool() {
			autostart = 1
		}
		err = r.client.Libvirt().DomainSetAutostart(finalDomain, autostart)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Set Autostart",
				"Domain was updated but failed to set autostart: "+err.Error(),
			)
			return
		}
	}

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
	if uint32(domainState) == uint32(golibvirt.DomainRunning) {
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
