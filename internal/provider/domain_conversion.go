// Package provider implements the Terraform provider for libvirt.
package provider

import (
	"context"
	"fmt"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"libvirt.org/go/libvirtxml"
)

const (
	unitKiB = "KiB"
	unitMiB = "MiB"
	unitGiB = "GiB"
	unitTiB = "TiB"
)

// convertMemory converts memory values between units
func convertMemory(value int64, fromUnit, toUnit string) int64 {
	// First convert to KiB
	var kib int64
	switch fromUnit {
	case unitKiB, "":
		kib = value
	case unitMiB:
		kib = value * 1024
	case unitGiB:
		kib = value * 1024 * 1024
	case unitTiB:
		kib = value * 1024 * 1024 * 1024
	default:
		kib = value // fallback
	}

	// Then convert to target unit
	switch toUnit {
	case unitKiB, "":
		return kib
	case unitMiB:
		return kib / 1024
	case unitGiB:
		return kib / (1024 * 1024)
	case unitTiB:
		return kib / (1024 * 1024 * 1024)
	default:
		return kib // fallback
	}
}

// domainModelToXML converts a DomainResourceModel to libvirtxml.Domain
// TODO: Consider refactoring this function to reduce complexity once we add more fields
func domainModelToXML(ctx context.Context, client *libvirt.Client, model *DomainResourceModel) (*libvirtxml.Domain, error) {
	domain := &libvirtxml.Domain{
		Name: model.Name.ValueString(),
	}

	// Set domain type (default to kvm)
	if model.Type.IsNull() || model.Type.IsUnknown() {
		domain.Type = "kvm"
	} else {
		domain.Type = model.Type.ValueString()
	}

	// Set UUID if provided
	if !model.UUID.IsNull() && !model.UUID.IsUnknown() {
		domain.UUID = model.UUID.ValueString()
	}

	// Set HWUUID if provided
	if !model.HWUUID.IsNull() && !model.HWUUID.IsUnknown() {
		domain.HWUUID = model.HWUUID.ValueString()
	}

	// Set title
	if !model.Title.IsNull() && !model.Title.IsUnknown() {
		domain.Title = model.Title.ValueString()
	}

	// Set description
	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		domain.Description = model.Description.ValueString()
	}

	// Set bootloader (Xen paravirt)
	if !model.Bootloader.IsNull() && !model.Bootloader.IsUnknown() {
		domain.Bootloader = model.Bootloader.ValueString()
	}
	if !model.BootloaderArgs.IsNull() && !model.BootloaderArgs.IsUnknown() {
		domain.BootloaderArgs = model.BootloaderArgs.ValueString()
	}

	// Set memory
	if !model.Memory.IsNull() && !model.Memory.IsUnknown() {
		unit := "KiB"
		if !model.Unit.IsNull() && !model.Unit.IsUnknown() {
			unit = model.Unit.ValueString()
		}
		memValue := uint(model.Memory.ValueInt64())
		domain.Memory = &libvirtxml.DomainMemory{
			Value: memValue,
			Unit:  unit,
		}
	}

	// Set current memory
	if !model.CurrentMemory.IsNull() && !model.CurrentMemory.IsUnknown() {
		unit := "KiB"
		if !model.Unit.IsNull() && !model.Unit.IsUnknown() {
			unit = model.Unit.ValueString()
		}
		currentMemValue := uint(model.CurrentMemory.ValueInt64())
		domain.CurrentMemory = &libvirtxml.DomainCurrentMemory{
			Value: currentMemValue,
			Unit:  unit,
		}
	}

	// Set max memory
	if !model.MaxMemory.IsNull() && !model.MaxMemory.IsUnknown() {
		unit := "KiB"
		if !model.Unit.IsNull() && !model.Unit.IsUnknown() {
			unit = model.Unit.ValueString()
		}
		maxMemValue := uint(model.MaxMemory.ValueInt64())
		maxMem := &libvirtxml.DomainMaxMemory{
			Value: maxMemValue,
			Unit:  unit,
		}
		if !model.MaxMemorySlots.IsNull() && !model.MaxMemorySlots.IsUnknown() {
			maxMem.Slots = uint(model.MaxMemorySlots.ValueInt64())
		}
		domain.MaximumMemory = maxMem
	}

	// Set VCPU
	if !model.VCPU.IsNull() && !model.VCPU.IsUnknown() {
		vcpuValue := uint(model.VCPU.ValueInt64())
		domain.VCPU = &libvirtxml.DomainVCPU{
			Value: vcpuValue,
		}
	}

	// Set lifecycle actions
	if !model.OnPoweroff.IsNull() && !model.OnPoweroff.IsUnknown() {
		domain.OnPoweroff = model.OnPoweroff.ValueString()
	}
	if !model.OnReboot.IsNull() && !model.OnReboot.IsUnknown() {
		domain.OnReboot = model.OnReboot.ValueString()
	}
	if !model.OnCrash.IsNull() && !model.OnCrash.IsUnknown() {
		domain.OnCrash = model.OnCrash.ValueString()
	}

	// Set I/O threads
	if !model.IOThreads.IsNull() && !model.IOThreads.IsUnknown() {
		domain.IOThreads = uint(model.IOThreads.ValueInt64())
	}

	// Set OS configuration
	if model.OS != nil {
		os := &libvirtxml.DomainOS{}

		// OS type is required
		if model.OS.Type.IsNull() || model.OS.Type.IsUnknown() {
			return nil, fmt.Errorf("os.type is required")
		}

		os.Type = &libvirtxml.DomainOSType{
			Type: model.OS.Type.ValueString(),
		}

		// Optional OS type attributes
		if !model.OS.Arch.IsNull() && !model.OS.Arch.IsUnknown() {
			os.Type.Arch = model.OS.Arch.ValueString()
		}
		if !model.OS.Machine.IsNull() && !model.OS.Machine.IsUnknown() {
			os.Type.Machine = model.OS.Machine.ValueString()
		}

		// Boot devices
		if !model.OS.BootDevices.IsNull() && !model.OS.BootDevices.IsUnknown() {
			var bootDevices []types.String
			model.OS.BootDevices.ElementsAs(context.Background(), &bootDevices, false)
			for _, dev := range bootDevices {
				if !dev.IsNull() && !dev.IsUnknown() {
					os.BootDevices = append(os.BootDevices, libvirtxml.DomainBootDevice{
						Dev: dev.ValueString(),
					})
				}
			}
		}

		// Direct kernel boot
		if !model.OS.Kernel.IsNull() && !model.OS.Kernel.IsUnknown() {
			os.Kernel = model.OS.Kernel.ValueString()
		}
		if !model.OS.Initrd.IsNull() && !model.OS.Initrd.IsUnknown() {
			os.Initrd = model.OS.Initrd.ValueString()
		}
		if !model.OS.KernelArgs.IsNull() && !model.OS.KernelArgs.IsUnknown() {
			os.Cmdline = model.OS.KernelArgs.ValueString()
		}

		// UEFI loader
		if !model.OS.LoaderPath.IsNull() && !model.OS.LoaderPath.IsUnknown() {
			loader := &libvirtxml.DomainLoader{
				Path: model.OS.LoaderPath.ValueString(),
			}
			if !model.OS.LoaderReadOnly.IsNull() && !model.OS.LoaderReadOnly.IsUnknown() {
				readOnly := "no"
				if model.OS.LoaderReadOnly.ValueBool() {
					readOnly = "yes"
				}
				loader.Readonly = readOnly
			}
			if !model.OS.LoaderType.IsNull() && !model.OS.LoaderType.IsUnknown() {
				loader.Type = model.OS.LoaderType.ValueString()
			}
			os.Loader = loader
		}

		// NVRAM
		if !model.OS.NVRAMPath.IsNull() && !model.OS.NVRAMPath.IsUnknown() {
			os.NVRam = &libvirtxml.DomainNVRam{
				NVRam: model.OS.NVRAMPath.ValueString(),
			}
		}

		domain.OS = os
	}

	// Set features
	if model.Features != nil {
		features := &libvirtxml.DomainFeatureList{}

		// PAE - presence-based
		if !model.Features.PAE.IsNull() && !model.Features.PAE.IsUnknown() {
			if model.Features.PAE.ValueBool() {
				features.PAE = &libvirtxml.DomainFeature{}
			}
		}

		// ACPI - presence-based
		if !model.Features.ACPI.IsNull() && !model.Features.ACPI.IsUnknown() {
			if model.Features.ACPI.ValueBool() {
				features.ACPI = &libvirtxml.DomainFeature{}
			}
		}

		// APIC - presence-based
		if !model.Features.APIC.IsNull() && !model.Features.APIC.IsUnknown() {
			if model.Features.APIC.ValueBool() {
				features.APIC = &libvirtxml.DomainFeatureAPIC{}
			}
		}

		// Viridian - presence-based
		if !model.Features.Viridian.IsNull() && !model.Features.Viridian.IsUnknown() {
			if model.Features.Viridian.ValueBool() {
				features.Viridian = &libvirtxml.DomainFeature{}
			}
		}

		// PrivNet - presence-based
		if !model.Features.PrivNet.IsNull() && !model.Features.PrivNet.IsUnknown() {
			if model.Features.PrivNet.ValueBool() {
				features.PrivNet = &libvirtxml.DomainFeature{}
			}
		}

		// HAP - state-based
		if !model.Features.HAP.IsNull() && !model.Features.HAP.IsUnknown() {
			features.HAP = &libvirtxml.DomainFeatureState{
				State: model.Features.HAP.ValueString(),
			}
		}

		// PMU - state-based
		if !model.Features.PMU.IsNull() && !model.Features.PMU.IsUnknown() {
			features.PMU = &libvirtxml.DomainFeatureState{
				State: model.Features.PMU.ValueString(),
			}
		}

		// VMPort - state-based
		if !model.Features.VMPort.IsNull() && !model.Features.VMPort.IsUnknown() {
			features.VMPort = &libvirtxml.DomainFeatureState{
				State: model.Features.VMPort.ValueString(),
			}
		}

		// PVSpinlock - state-based
		if !model.Features.PVSpinlock.IsNull() && !model.Features.PVSpinlock.IsUnknown() {
			features.PVSpinlock = &libvirtxml.DomainFeatureState{
				State: model.Features.PVSpinlock.ValueString(),
			}
		}

		// VMCoreInfo - state-based
		if !model.Features.VMCoreInfo.IsNull() && !model.Features.VMCoreInfo.IsUnknown() {
			features.VMCoreInfo = &libvirtxml.DomainFeatureState{
				State: model.Features.VMCoreInfo.ValueString(),
			}
		}

		// HTM - state-based
		if !model.Features.HTM.IsNull() && !model.Features.HTM.IsUnknown() {
			features.HTM = &libvirtxml.DomainFeatureState{
				State: model.Features.HTM.ValueString(),
			}
		}

		// NestedHV - state-based
		if !model.Features.NestedHV.IsNull() && !model.Features.NestedHV.IsUnknown() {
			features.NestedHV = &libvirtxml.DomainFeatureState{
				State: model.Features.NestedHV.ValueString(),
			}
		}

		// CCFAssist - state-based
		if !model.Features.CCFAssist.IsNull() && !model.Features.CCFAssist.IsUnknown() {
			features.CCFAssist = &libvirtxml.DomainFeatureState{
				State: model.Features.CCFAssist.ValueString(),
			}
		}

		// RAS - state-based
		if !model.Features.RAS.IsNull() && !model.Features.RAS.IsUnknown() {
			features.RAS = &libvirtxml.DomainFeatureState{
				State: model.Features.RAS.ValueString(),
			}
		}

		// PS2 - state-based
		if !model.Features.PS2.IsNull() && !model.Features.PS2.IsUnknown() {
			features.PS2 = &libvirtxml.DomainFeatureState{
				State: model.Features.PS2.ValueString(),
			}
		}

		// IOAPIC - driver attribute
		if !model.Features.IOAPICDriver.IsNull() && !model.Features.IOAPICDriver.IsUnknown() {
			features.IOAPIC = &libvirtxml.DomainFeatureIOAPIC{
				Driver: model.Features.IOAPICDriver.ValueString(),
			}
		}

		// GIC - version attribute
		if !model.Features.GICVersion.IsNull() && !model.Features.GICVersion.IsUnknown() {
			features.GIC = &libvirtxml.DomainFeatureGIC{
				Version: model.Features.GICVersion.ValueString(),
			}
		}

		domain.Features = features
	}

	// Set CPU
	if model.CPU != nil {
		cpu := &libvirtxml.DomainCPU{}

		if !model.CPU.Mode.IsNull() && !model.CPU.Mode.IsUnknown() {
			cpu.Mode = model.CPU.Mode.ValueString()
		}

		if !model.CPU.Match.IsNull() && !model.CPU.Match.IsUnknown() {
			cpu.Match = model.CPU.Match.ValueString()
		}

		if !model.CPU.Check.IsNull() && !model.CPU.Check.IsUnknown() {
			cpu.Check = model.CPU.Check.ValueString()
		}

		if !model.CPU.Migratable.IsNull() && !model.CPU.Migratable.IsUnknown() {
			cpu.Migratable = model.CPU.Migratable.ValueString()
		}

		if !model.CPU.DeprecatedFeatures.IsNull() && !model.CPU.DeprecatedFeatures.IsUnknown() {
			cpu.DeprecatedFeatures = model.CPU.DeprecatedFeatures.ValueString()
		}

		if !model.CPU.Model.IsNull() && !model.CPU.Model.IsUnknown() {
			cpu.Model = &libvirtxml.DomainCPUModel{
				Value: model.CPU.Model.ValueString(),
			}
		}

		if !model.CPU.Vendor.IsNull() && !model.CPU.Vendor.IsUnknown() {
			cpu.Vendor = model.CPU.Vendor.ValueString()
		}

		domain.CPU = cpu
	}

	// Set Clock
	if model.Clock != nil {
		clock := &libvirtxml.DomainClock{}

		if !model.Clock.Offset.IsNull() && !model.Clock.Offset.IsUnknown() {
			clock.Offset = model.Clock.Offset.ValueString()
		}

		if !model.Clock.Basis.IsNull() && !model.Clock.Basis.IsUnknown() {
			clock.Basis = model.Clock.Basis.ValueString()
		}

		if !model.Clock.Adjustment.IsNull() && !model.Clock.Adjustment.IsUnknown() {
			clock.Adjustment = model.Clock.Adjustment.ValueString()
		}

		if !model.Clock.TimeZone.IsNull() && !model.Clock.TimeZone.IsUnknown() {
			clock.TimeZone = model.Clock.TimeZone.ValueString()
		}

		// Convert timers
		for _, timerModel := range model.Clock.Timers {
			timer := libvirtxml.DomainTimer{}

			if !timerModel.Name.IsNull() && !timerModel.Name.IsUnknown() {
				timer.Name = timerModel.Name.ValueString()
			}

			if !timerModel.Track.IsNull() && !timerModel.Track.IsUnknown() {
				timer.Track = timerModel.Track.ValueString()
			}

			if !timerModel.TickPolicy.IsNull() && !timerModel.TickPolicy.IsUnknown() {
				timer.TickPolicy = timerModel.TickPolicy.ValueString()
			}

			if !timerModel.Frequency.IsNull() && !timerModel.Frequency.IsUnknown() {
				timer.Frequency = uint64(timerModel.Frequency.ValueInt64())
			}

			if !timerModel.Mode.IsNull() && !timerModel.Mode.IsUnknown() {
				timer.Mode = timerModel.Mode.ValueString()
			}

			if !timerModel.Present.IsNull() && !timerModel.Present.IsUnknown() {
				timer.Present = timerModel.Present.ValueString()
			}

			// Convert catchup
			if timerModel.CatchUp != nil {
				catchup := &libvirtxml.DomainTimerCatchUp{}

				if !timerModel.CatchUp.Threshold.IsNull() && !timerModel.CatchUp.Threshold.IsUnknown() {
					catchup.Threshold = uint(timerModel.CatchUp.Threshold.ValueInt64())
				}

				if !timerModel.CatchUp.Slew.IsNull() && !timerModel.CatchUp.Slew.IsUnknown() {
					catchup.Slew = uint(timerModel.CatchUp.Slew.ValueInt64())
				}

				if !timerModel.CatchUp.Limit.IsNull() && !timerModel.CatchUp.Limit.IsUnknown() {
					catchup.Limit = uint(timerModel.CatchUp.Limit.ValueInt64())
				}

				timer.CatchUp = catchup
			}

			clock.Timer = append(clock.Timer, timer)
		}

		domain.Clock = clock
	}

	// Set PM
	if model.PM != nil {
		pm := &libvirtxml.DomainPM{}

		if !model.PM.SuspendToMem.IsNull() && !model.PM.SuspendToMem.IsUnknown() {
			pm.SuspendToMem = &libvirtxml.DomainPMPolicy{
				Enabled: model.PM.SuspendToMem.ValueString(),
			}
		}

		if !model.PM.SuspendToDisk.IsNull() && !model.PM.SuspendToDisk.IsUnknown() {
			pm.SuspendToDisk = &libvirtxml.DomainPMPolicy{
				Enabled: model.PM.SuspendToDisk.ValueString(),
			}
		}

		domain.PM = pm
	}

	// Set Devices (Disks and Interfaces)
	if !model.Devices.IsNull() && !model.Devices.IsUnknown() {
		var devices DomainDevicesModel
		diags := model.Devices.As(ctx, &devices, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, fmt.Errorf("failed to extract devices: %v", diags.Errors())
		}

		if domain.Devices == nil {
			domain.Devices = &libvirtxml.DomainDeviceList{}
		}

		// Process disks
		if !devices.Disks.IsNull() && !devices.Disks.IsUnknown() {
			var disks []DomainDiskModel
			diags := devices.Disks.ElementsAs(ctx, &disks, false)
			if diags.HasError() {
				return nil, fmt.Errorf("failed to extract disks: %v", diags.Errors())
			}

			for _, diskModel := range disks {
			disk := libvirtxml.DomainDisk{}

			// Set device type (default to disk)
			if !diskModel.Device.IsNull() && !diskModel.Device.IsUnknown() {
				disk.Device = diskModel.Device.ValueString()
			} else {
				disk.Device = "disk"
			}

			// Set source - either from volume_id or direct file path
			var sourcePath string
			if !diskModel.VolumeID.IsNull() && !diskModel.VolumeID.IsUnknown() {
				// Look up volume by key and get its path
				volumeKey := diskModel.VolumeID.ValueString()
				volume, err := client.Libvirt().StorageVolLookupByKey(volumeKey)
				if err != nil {
					return nil, fmt.Errorf("failed to lookup volume %s: %w", volumeKey, err)
				}

				volumeXML, err := client.Libvirt().StorageVolGetXMLDesc(volume, 0)
				if err != nil {
					return nil, fmt.Errorf("failed to get volume XML: %w", err)
				}

				var volumeDef libvirtxml.StorageVolume
				if err := volumeDef.Unmarshal(volumeXML); err != nil {
					return nil, fmt.Errorf("failed to parse volume XML: %w", err)
				}

				if volumeDef.Target != nil {
					sourcePath = volumeDef.Target.Path
				} else {
					return nil, fmt.Errorf("volume %s has no target path", volumeKey)
				}
			} else if !diskModel.Source.IsNull() && !diskModel.Source.IsUnknown() {
				sourcePath = diskModel.Source.ValueString()
			}

			if sourcePath != "" {
				disk.Source = &libvirtxml.DomainDiskSource{
					File: &libvirtxml.DomainDiskSourceFile{
						File: sourcePath,
					},
				}
			}

			// Set target
			if !diskModel.Target.IsNull() && !diskModel.Target.IsUnknown() {
				disk.Target = &libvirtxml.DomainDiskTarget{
					Dev: diskModel.Target.ValueString(),
				}

				// Set bus if specified
				if !diskModel.Bus.IsNull() && !diskModel.Bus.IsUnknown() {
					disk.Target.Bus = diskModel.Bus.ValueString()
				}
			}

				domain.Devices.Disks = append(domain.Devices.Disks, disk)
			}
		}

		// Process interfaces
		if !devices.Interfaces.IsNull() && !devices.Interfaces.IsUnknown() {
			var interfaces []DomainInterfaceModel
			diags := devices.Interfaces.ElementsAs(ctx, &interfaces, false)
			if diags.HasError() {
				return nil, fmt.Errorf("failed to extract interfaces: %v", diags.Errors())
			}

			for _, ifaceModel := range interfaces {
			iface := libvirtxml.DomainInterface{}

			// Set MAC address
			if !ifaceModel.MAC.IsNull() && !ifaceModel.MAC.IsUnknown() {
				iface.MAC = &libvirtxml.DomainInterfaceMAC{
					Address: ifaceModel.MAC.ValueString(),
				}
			}

			// Set model
			if !ifaceModel.Model.IsNull() && !ifaceModel.Model.IsUnknown() {
				iface.Model = &libvirtxml.DomainInterfaceModel{
					Type: ifaceModel.Model.ValueString(),
				}
			}

			// Set source based on type
			if !ifaceModel.Type.IsNull() && !ifaceModel.Type.IsUnknown() && ifaceModel.Source != nil {
				ifaceType := ifaceModel.Type.ValueString()
				source := &libvirtxml.DomainInterfaceSource{}

				switch ifaceType {
				case "network":
					network := &libvirtxml.DomainInterfaceSourceNetwork{}
					if !ifaceModel.Source.Network.IsNull() && !ifaceModel.Source.Network.IsUnknown() {
						network.Network = ifaceModel.Source.Network.ValueString()
					}
					if !ifaceModel.Source.PortGroup.IsNull() && !ifaceModel.Source.PortGroup.IsUnknown() {
						network.PortGroup = ifaceModel.Source.PortGroup.ValueString()
					}
					source.Network = network

				case "bridge":
					bridge := &libvirtxml.DomainInterfaceSourceBridge{}
					if !ifaceModel.Source.Bridge.IsNull() && !ifaceModel.Source.Bridge.IsUnknown() {
						bridge.Bridge = ifaceModel.Source.Bridge.ValueString()
					}
					source.Bridge = bridge

				case "user":
					user := &libvirtxml.DomainInterfaceSourceUser{}
					if !ifaceModel.Source.Dev.IsNull() && !ifaceModel.Source.Dev.IsUnknown() {
						user.Dev = ifaceModel.Source.Dev.ValueString()
					}
					source.User = user
				}

				iface.Source = source
			}

				domain.Devices.Interfaces = append(domain.Devices.Interfaces, iface)
			}
		}

		// Process graphics
		if !devices.Graphics.IsNull() && !devices.Graphics.IsUnknown() {
			var graphicsModel DomainGraphicsModel
			diags := devices.Graphics.As(ctx, &graphicsModel, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				return nil, fmt.Errorf("failed to extract graphics: %v", diags.Errors())
			}

			graphic := libvirtxml.DomainGraphic{}

			// Handle VNC graphics
			if !graphicsModel.VNC.IsNull() && !graphicsModel.VNC.IsUnknown() {
				var vncModel DomainGraphicsVNCModel
				diags := graphicsModel.VNC.As(ctx, &vncModel, basetypes.ObjectAsOptions{})
				if diags.HasError() {
					return nil, fmt.Errorf("failed to extract VNC graphics: %v", diags.Errors())
				}

				vnc := &libvirtxml.DomainGraphicVNC{}

				if !vncModel.Socket.IsNull() && !vncModel.Socket.IsUnknown() {
					vnc.Socket = vncModel.Socket.ValueString()
				}
				if !vncModel.Port.IsNull() && !vncModel.Port.IsUnknown() {
					vnc.Port = int(vncModel.Port.ValueInt64())
				}
				if !vncModel.AutoPort.IsNull() && !vncModel.AutoPort.IsUnknown() {
					vnc.AutoPort = vncModel.AutoPort.ValueString()
				}
				if !vncModel.WebSocket.IsNull() && !vncModel.WebSocket.IsUnknown() {
					vnc.WebSocket = int(vncModel.WebSocket.ValueInt64())
				}
				if !vncModel.Listen.IsNull() && !vncModel.Listen.IsUnknown() {
					vnc.Listen = vncModel.Listen.ValueString()
				}

				graphic.VNC = vnc
			}

			// Handle Spice graphics
			if !graphicsModel.Spice.IsNull() && !graphicsModel.Spice.IsUnknown() {
				var spiceModel DomainGraphicsSpiceModel
				diags := graphicsModel.Spice.As(ctx, &spiceModel, basetypes.ObjectAsOptions{})
				if diags.HasError() {
					return nil, fmt.Errorf("failed to extract Spice graphics: %v", diags.Errors())
				}

				spice := &libvirtxml.DomainGraphicSpice{}

				if !spiceModel.Port.IsNull() && !spiceModel.Port.IsUnknown() {
					spice.Port = int(spiceModel.Port.ValueInt64())
				}
				if !spiceModel.TLSPort.IsNull() && !spiceModel.TLSPort.IsUnknown() {
					spice.TLSPort = int(spiceModel.TLSPort.ValueInt64())
				}
				if !spiceModel.AutoPort.IsNull() && !spiceModel.AutoPort.IsUnknown() {
					spice.AutoPort = spiceModel.AutoPort.ValueString()
				}
				if !spiceModel.Listen.IsNull() && !spiceModel.Listen.IsUnknown() {
					spice.Listen = spiceModel.Listen.ValueString()
				}

				graphic.Spice = spice
			}

			domain.Devices.Graphics = append(domain.Devices.Graphics, graphic)
		}
	}

	return domain, nil
}

// xmlToDomainModel converts libvirtxml.Domain to a DomainResourceModel
func xmlToDomainModel(ctx context.Context, domain *libvirtxml.Domain, model *DomainResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	model.Name = types.StringValue(domain.Name)
	model.Type = types.StringValue(domain.Type)

	if domain.UUID != "" {
		model.UUID = types.StringValue(domain.UUID)
		model.ID = types.StringValue(domain.UUID)
	}

	if domain.HWUUID != "" {
		model.HWUUID = types.StringValue(domain.HWUUID)
	}

	if domain.Title != "" {
		model.Title = types.StringValue(domain.Title)
	}

	if domain.Description != "" {
		model.Description = types.StringValue(domain.Description)
	}

	if domain.Bootloader != "" {
		model.Bootloader = types.StringValue(domain.Bootloader)
	}
	if domain.BootloaderArgs != "" {
		model.BootloaderArgs = types.StringValue(domain.BootloaderArgs)
	}

	if domain.Memory != nil {
		// Libvirt always returns memory in KiB, but we need to convert it back
		// to the unit the user specified in their config to avoid inconsistency errors
		memoryKiB := int64(domain.Memory.Value)

		// If the model already has a unit preference, convert to that
		if !model.Unit.IsNull() && !model.Unit.IsUnknown() {
			targetUnit := model.Unit.ValueString()
			model.Memory = types.Int64Value(convertMemory(memoryKiB, "KiB", targetUnit))
			model.Unit = types.StringValue(targetUnit)
		} else {
			// Otherwise, use what libvirt returned
			model.Memory = types.Int64Value(memoryKiB)
			if domain.Memory.Unit != "" {
				model.Unit = types.StringValue(domain.Memory.Unit)
			} else {
				model.Unit = types.StringValue("KiB")
			}
		}
	}

	// Only set current_memory if user originally specified it, to avoid inconsistency with libvirt defaults
	if !model.CurrentMemory.IsNull() && !model.CurrentMemory.IsUnknown() && domain.CurrentMemory != nil {
		currentMemoryKiB := int64(domain.CurrentMemory.Value)
		if !model.Unit.IsNull() && !model.Unit.IsUnknown() {
			targetUnit := model.Unit.ValueString()
			model.CurrentMemory = types.Int64Value(convertMemory(currentMemoryKiB, "KiB", targetUnit))
		} else {
			model.CurrentMemory = types.Int64Value(currentMemoryKiB)
		}
	}

	if domain.MaximumMemory != nil {
		maxMemoryKiB := int64(domain.MaximumMemory.Value)
		if !model.Unit.IsNull() && !model.Unit.IsUnknown() {
			targetUnit := model.Unit.ValueString()
			model.MaxMemory = types.Int64Value(convertMemory(maxMemoryKiB, "KiB", targetUnit))
		} else {
			model.MaxMemory = types.Int64Value(maxMemoryKiB)
		}
		if domain.MaximumMemory.Slots > 0 {
			model.MaxMemorySlots = types.Int64Value(int64(domain.MaximumMemory.Slots))
		}
	}

	if domain.VCPU != nil {
		model.VCPU = types.Int64Value(int64(domain.VCPU.Value))
	}

	// Only preserve lifecycle actions if user originally specified them, to avoid inconsistency with libvirt defaults
	if !model.OnPoweroff.IsNull() && !model.OnPoweroff.IsUnknown() && domain.OnPoweroff != "" {
		model.OnPoweroff = types.StringValue(domain.OnPoweroff)
	}
	if !model.OnReboot.IsNull() && !model.OnReboot.IsUnknown() && domain.OnReboot != "" {
		model.OnReboot = types.StringValue(domain.OnReboot)
	}
	if !model.OnCrash.IsNull() && !model.OnCrash.IsUnknown() && domain.OnCrash != "" {
		model.OnCrash = types.StringValue(domain.OnCrash)
	}

	if domain.IOThreads > 0 {
		model.IOThreads = types.Int64Value(int64(domain.IOThreads))
	}

	if domain.OS != nil {
		osModel := &DomainOSModel{}

		if domain.OS.Type != nil {
			osModel.Type = types.StringValue(domain.OS.Type.Type)
			if domain.OS.Type.Arch != "" {
				osModel.Arch = types.StringValue(domain.OS.Type.Arch)
			}
			// For machine type, preserve what the user specified in their config
			// even though libvirt may have expanded it (e.g., q35 -> pc-q35-10.1)
			// This prevents unnecessary diffs
			if !model.OS.Machine.IsNull() && !model.OS.Machine.IsUnknown() {
				osModel.Machine = model.OS.Machine
			} else if domain.OS.Type.Machine != "" {
				osModel.Machine = types.StringValue(domain.OS.Type.Machine)
			}
		}

		// For boot devices, preserve what the user specified
		// If they didn't specify any, use what libvirt added
		if !model.OS.BootDevices.IsNull() && !model.OS.BootDevices.IsUnknown() {
			osModel.BootDevices = model.OS.BootDevices
		} else if len(domain.OS.BootDevices) > 0 {
			bootDevices := make([]types.String, len(domain.OS.BootDevices))
			for i, boot := range domain.OS.BootDevices {
				bootDevices[i] = types.StringValue(boot.Dev)
			}
			listValue, _ := types.ListValueFrom(context.Background(), types.StringType, bootDevices)
			osModel.BootDevices = listValue
		}

		if domain.OS.Kernel != "" {
			osModel.Kernel = types.StringValue(domain.OS.Kernel)
		}
		if domain.OS.Initrd != "" {
			osModel.Initrd = types.StringValue(domain.OS.Initrd)
		}
		if domain.OS.Cmdline != "" {
			osModel.KernelArgs = types.StringValue(domain.OS.Cmdline)
		}

		if domain.OS.Loader != nil {
			osModel.LoaderPath = types.StringValue(domain.OS.Loader.Path)
			switch domain.OS.Loader.Readonly {
			case "yes":
				osModel.LoaderReadOnly = types.BoolValue(true)
			case "no":
				osModel.LoaderReadOnly = types.BoolValue(false)
			}
			if domain.OS.Loader.Type != "" {
				osModel.LoaderType = types.StringValue(domain.OS.Loader.Type)
			}
		}

		if domain.OS.NVRam != nil {
			osModel.NVRAMPath = types.StringValue(domain.OS.NVRam.NVRam)
		}

		model.OS = osModel
	}

	// Features
	if domain.Features != nil {
		featuresModel := &DomainFeaturesModel{}

		// PAE - presence = enabled
		if domain.Features.PAE != nil {
			featuresModel.PAE = types.BoolValue(true)
		}

		// ACPI - presence = enabled
		if domain.Features.ACPI != nil {
			featuresModel.ACPI = types.BoolValue(true)
		}

		// APIC - presence = enabled
		if domain.Features.APIC != nil {
			featuresModel.APIC = types.BoolValue(true)
		}

		// Viridian - presence = enabled
		if domain.Features.Viridian != nil {
			featuresModel.Viridian = types.BoolValue(true)
		}

		// PrivNet - presence = enabled
		if domain.Features.PrivNet != nil {
			featuresModel.PrivNet = types.BoolValue(true)
		}

		// HAP - state-based
		if domain.Features.HAP != nil && domain.Features.HAP.State != "" {
			featuresModel.HAP = types.StringValue(domain.Features.HAP.State)
		}

		// PMU - state-based
		if domain.Features.PMU != nil && domain.Features.PMU.State != "" {
			featuresModel.PMU = types.StringValue(domain.Features.PMU.State)
		}

		// VMPort - state-based
		if domain.Features.VMPort != nil && domain.Features.VMPort.State != "" {
			featuresModel.VMPort = types.StringValue(domain.Features.VMPort.State)
		}

		// PVSpinlock - state-based
		if domain.Features.PVSpinlock != nil && domain.Features.PVSpinlock.State != "" {
			featuresModel.PVSpinlock = types.StringValue(domain.Features.PVSpinlock.State)
		}

		// VMCoreInfo - state-based
		if domain.Features.VMCoreInfo != nil && domain.Features.VMCoreInfo.State != "" {
			featuresModel.VMCoreInfo = types.StringValue(domain.Features.VMCoreInfo.State)
		}

		// HTM - state-based
		if domain.Features.HTM != nil && domain.Features.HTM.State != "" {
			featuresModel.HTM = types.StringValue(domain.Features.HTM.State)
		}

		// NestedHV - state-based
		if domain.Features.NestedHV != nil && domain.Features.NestedHV.State != "" {
			featuresModel.NestedHV = types.StringValue(domain.Features.NestedHV.State)
		}

		// CCFAssist - state-based
		if domain.Features.CCFAssist != nil && domain.Features.CCFAssist.State != "" {
			featuresModel.CCFAssist = types.StringValue(domain.Features.CCFAssist.State)
		}

		// RAS - state-based
		if domain.Features.RAS != nil && domain.Features.RAS.State != "" {
			featuresModel.RAS = types.StringValue(domain.Features.RAS.State)
		}

		// PS2 - state-based
		if domain.Features.PS2 != nil && domain.Features.PS2.State != "" {
			featuresModel.PS2 = types.StringValue(domain.Features.PS2.State)
		}

		// IOAPIC - driver attribute
		if domain.Features.IOAPIC != nil && domain.Features.IOAPIC.Driver != "" {
			featuresModel.IOAPICDriver = types.StringValue(domain.Features.IOAPIC.Driver)
		}

		// GIC - version attribute
		if domain.Features.GIC != nil && domain.Features.GIC.Version != "" {
			featuresModel.GICVersion = types.StringValue(domain.Features.GIC.Version)
		}

		model.Features = featuresModel
	}

	// CPU - only preserve if user originally specified it
	if model.CPU != nil && domain.CPU != nil {
		cpuModel := &DomainCPUModel{}

		// Only set mode if user specified it
		if !model.CPU.Mode.IsNull() && !model.CPU.Mode.IsUnknown() && domain.CPU.Mode != "" {
			cpuModel.Mode = types.StringValue(domain.CPU.Mode)
		}

		// Only set match if user specified it
		if !model.CPU.Match.IsNull() && !model.CPU.Match.IsUnknown() && domain.CPU.Match != "" {
			cpuModel.Match = types.StringValue(domain.CPU.Match)
		}

		// Only set check if user specified it
		if !model.CPU.Check.IsNull() && !model.CPU.Check.IsUnknown() && domain.CPU.Check != "" {
			cpuModel.Check = types.StringValue(domain.CPU.Check)
		}

		// Only set migratable if user specified it
		if !model.CPU.Migratable.IsNull() && !model.CPU.Migratable.IsUnknown() && domain.CPU.Migratable != "" {
			cpuModel.Migratable = types.StringValue(domain.CPU.Migratable)
		}

		// Only set deprecated_features if user specified it
		if !model.CPU.DeprecatedFeatures.IsNull() && !model.CPU.DeprecatedFeatures.IsUnknown() && domain.CPU.DeprecatedFeatures != "" {
			cpuModel.DeprecatedFeatures = types.StringValue(domain.CPU.DeprecatedFeatures)
		}

		// Only set model if user specified it
		if !model.CPU.Model.IsNull() && !model.CPU.Model.IsUnknown() && domain.CPU.Model != nil && domain.CPU.Model.Value != "" {
			cpuModel.Model = types.StringValue(domain.CPU.Model.Value)
		}

		// Only set vendor if user specified it
		if !model.CPU.Vendor.IsNull() && !model.CPU.Vendor.IsUnknown() && domain.CPU.Vendor != "" {
			cpuModel.Vendor = types.StringValue(domain.CPU.Vendor)
		}

		model.CPU = cpuModel
	}

	// Clock - only preserve if user originally specified it
	if model.Clock != nil && domain.Clock != nil {
		clockModel := &DomainClockModel{}

		if domain.Clock.Offset != "" {
			clockModel.Offset = types.StringValue(domain.Clock.Offset)
		}

		if domain.Clock.Basis != "" {
			clockModel.Basis = types.StringValue(domain.Clock.Basis)
		}

		if domain.Clock.Adjustment != "" {
			clockModel.Adjustment = types.StringValue(domain.Clock.Adjustment)
		}

		if domain.Clock.TimeZone != "" {
			clockModel.TimeZone = types.StringValue(domain.Clock.TimeZone)
		}

		// Convert timers - only if user specified them
		if len(model.Clock.Timers) > 0 && len(domain.Clock.Timer) > 0 {
			timers := make([]DomainTimerModel, 0, len(domain.Clock.Timer))

			for _, timer := range domain.Clock.Timer {
				timerModel := DomainTimerModel{}

				if timer.Name != "" {
					timerModel.Name = types.StringValue(timer.Name)
				}

				if timer.Track != "" {
					timerModel.Track = types.StringValue(timer.Track)
				}

				if timer.TickPolicy != "" {
					timerModel.TickPolicy = types.StringValue(timer.TickPolicy)
				}

				if timer.Frequency != 0 {
					timerModel.Frequency = types.Int64Value(int64(timer.Frequency))
				}

				if timer.Mode != "" {
					timerModel.Mode = types.StringValue(timer.Mode)
				}

				if timer.Present != "" {
					timerModel.Present = types.StringValue(timer.Present)
				}

				// Convert catchup if present
				if timer.CatchUp != nil {
					catchupModel := &DomainTimerCatchUpModel{}

					if timer.CatchUp.Threshold != 0 {
						catchupModel.Threshold = types.Int64Value(int64(timer.CatchUp.Threshold))
					}

					if timer.CatchUp.Slew != 0 {
						catchupModel.Slew = types.Int64Value(int64(timer.CatchUp.Slew))
					}

					if timer.CatchUp.Limit != 0 {
						catchupModel.Limit = types.Int64Value(int64(timer.CatchUp.Limit))
					}

					timerModel.CatchUp = catchupModel
				}

				timers = append(timers, timerModel)
			}

			clockModel.Timers = timers
		}

		model.Clock = clockModel
	}

	// PM - only preserve if user originally specified it
	if model.PM != nil && domain.PM != nil {
		pmModel := &DomainPMModel{}

		if domain.PM.SuspendToMem != nil && domain.PM.SuspendToMem.Enabled != "" {
			pmModel.SuspendToMem = types.StringValue(domain.PM.SuspendToMem.Enabled)
		}

		if domain.PM.SuspendToDisk != nil && domain.PM.SuspendToDisk.Enabled != "" {
			pmModel.SuspendToDisk = types.StringValue(domain.PM.SuspendToDisk.Enabled)
		}

		model.PM = pmModel
	}

	// Devices - only preserve optional fields the user specified
	if !model.Devices.IsNull() && !model.Devices.IsUnknown() && domain.Devices != nil {
		// Extract the current devices model from state
		var origDevices DomainDevicesModel
		d := model.Devices.As(ctx, &origDevices, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		var disks []DomainDiskModel
		var interfaces []DomainInterfaceModel

		// Process disks - only preserve optional fields the user specified
		if !origDevices.Disks.IsNull() && !origDevices.Disks.IsUnknown() && len(domain.Devices.Disks) > 0 {
			var origDisks []DomainDiskModel
			d := origDevices.Disks.ElementsAs(ctx, &origDisks, false)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			disks = make([]DomainDiskModel, 0, len(origDisks))

			for i := 0; i < len(origDisks) && i < len(domain.Devices.Disks); i++ {
				disk := domain.Devices.Disks[i]
				orig := origDisks[i]
				diskModel := DomainDiskModel{}

				if !orig.Device.IsNull() && !orig.Device.IsUnknown() && disk.Device != "" {
					diskModel.Device = types.StringValue(disk.Device)
				}

				// Preserve source only when the user specified a source path
				if !orig.Source.IsNull() && !orig.Source.IsUnknown() && disk.Source != nil && disk.Source.File != nil && disk.Source.File.File != "" {
					diskModel.Source = types.StringValue(disk.Source.File.File)
				}

				// Preserve volume_id exactly as provided by the user to avoid replacing it with libvirt paths
				if !orig.VolumeID.IsNull() && !orig.VolumeID.IsUnknown() {
					diskModel.VolumeID = orig.VolumeID
				}

				if disk.Target != nil {
					if disk.Target.Dev != "" {
						diskModel.Target = types.StringValue(disk.Target.Dev)
					}
					if !orig.Bus.IsNull() && !orig.Bus.IsUnknown() && disk.Target.Bus != "" {
						diskModel.Bus = types.StringValue(disk.Target.Bus)
					}
				}

				disks = append(disks, diskModel)
			}
		}

		// Process interfaces - only if user specified them
		if !origDevices.Interfaces.IsNull() && !origDevices.Interfaces.IsUnknown() && len(domain.Devices.Interfaces) > 0 {
			var origInterfaces []DomainInterfaceModel
			d := origDevices.Interfaces.ElementsAs(ctx, &origInterfaces, false)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			interfaces = make([]DomainInterfaceModel, 0, len(domain.Devices.Interfaces))

			for i, iface := range domain.Devices.Interfaces {
				// Only process if we have a corresponding model entry
				if i >= len(origInterfaces) {
					break
				}

				ifaceModel := DomainInterfaceModel{}
				origModel := origInterfaces[i]

				// Type - determine from source
				if iface.Source != nil {
					if iface.Source.Network != nil {
						ifaceModel.Type = types.StringValue("network")
					} else if iface.Source.Bridge != nil {
						ifaceModel.Type = types.StringValue("bridge")
					} else if iface.Source.User != nil {
						ifaceModel.Type = types.StringValue("user")
					}
				}

				// MAC - only if user specified it
				if !origModel.MAC.IsNull() && iface.MAC != nil && iface.MAC.Address != "" {
					ifaceModel.MAC = types.StringValue(iface.MAC.Address)
				}

				// Model - only if user specified it
				if !origModel.Model.IsNull() && iface.Model != nil && iface.Model.Type != "" {
					ifaceModel.Model = types.StringValue(iface.Model.Type)
				}

				// Source - only if user specified it
				if origModel.Source != nil && iface.Source != nil {
					sourceModel := &DomainInterfaceSourceModel{}

					if iface.Source.Network != nil {
						if !origModel.Source.Network.IsNull() && iface.Source.Network.Network != "" {
							sourceModel.Network = types.StringValue(iface.Source.Network.Network)
						}
						if !origModel.Source.PortGroup.IsNull() && iface.Source.Network.PortGroup != "" {
							sourceModel.PortGroup = types.StringValue(iface.Source.Network.PortGroup)
						}
					} else if iface.Source.Bridge != nil {
						if !origModel.Source.Bridge.IsNull() && iface.Source.Bridge.Bridge != "" {
							sourceModel.Bridge = types.StringValue(iface.Source.Bridge.Bridge)
						}
					} else if iface.Source.User != nil {
						if !origModel.Source.Dev.IsNull() && iface.Source.User.Dev != "" {
							sourceModel.Dev = types.StringValue(iface.Source.User.Dev)
						}
					}

					ifaceModel.Source = sourceModel
				}

				interfaces = append(interfaces, ifaceModel)
			}
		}

		// Process graphics - only preserve optional fields the user specified
		var graphicsModel *DomainGraphicsModel
		if !origDevices.Graphics.IsNull() && !origDevices.Graphics.IsUnknown() && len(domain.Devices.Graphics) > 0 {
			var origGraphics DomainGraphicsModel
			d := origDevices.Graphics.As(ctx, &origGraphics, basetypes.ObjectAsOptions{})
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			graphic := domain.Devices.Graphics[0] // libvirt only supports one graphics device
			newGraphics := DomainGraphicsModel{}

			// Initialize null objects for unused types
			vncNullObj := types.ObjectNull(map[string]attr.Type{
				"socket":    types.StringType,
				"port":      types.Int64Type,
				"autoport":  types.StringType,
				"websocket": types.Int64Type,
				"listen":    types.StringType,
			})
			spiceNullObj := types.ObjectNull(map[string]attr.Type{
				"port":     types.Int64Type,
				"tlsport":  types.Int64Type,
				"autoport": types.StringType,
				"listen":   types.StringType,
			})

			// Set defaults
			newGraphics.VNC = vncNullObj
			newGraphics.Spice = spiceNullObj

			// Handle VNC
			if !origGraphics.VNC.IsNull() && graphic.VNC != nil {
				var origVNC DomainGraphicsVNCModel
				d := origGraphics.VNC.As(ctx, &origVNC, basetypes.ObjectAsOptions{})
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				vncModel := DomainGraphicsVNCModel{}
				if !origVNC.Socket.IsNull() && graphic.VNC.Socket != "" {
					vncModel.Socket = types.StringValue(graphic.VNC.Socket)
				}
				if !origVNC.Port.IsNull() {
					vncModel.Port = types.Int64Value(int64(graphic.VNC.Port))
				}
				if !origVNC.AutoPort.IsNull() && graphic.VNC.AutoPort != "" {
					vncModel.AutoPort = types.StringValue(graphic.VNC.AutoPort)
				}
				if !origVNC.WebSocket.IsNull() {
					vncModel.WebSocket = types.Int64Value(int64(graphic.VNC.WebSocket))
				}
				if !origVNC.Listen.IsNull() && graphic.VNC.Listen != "" {
					vncModel.Listen = types.StringValue(graphic.VNC.Listen)
				}

				vncObj, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
					"socket":    types.StringType,
					"port":      types.Int64Type,
					"autoport":  types.StringType,
					"websocket": types.Int64Type,
					"listen":    types.StringType,
				}, vncModel)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
				newGraphics.VNC = vncObj
			}

			// Handle Spice
			if !origGraphics.Spice.IsNull() && graphic.Spice != nil {
				var origSpice DomainGraphicsSpiceModel
				d := origGraphics.Spice.As(ctx, &origSpice, basetypes.ObjectAsOptions{})
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				spiceModel := DomainGraphicsSpiceModel{}
				if !origSpice.Port.IsNull() {
					spiceModel.Port = types.Int64Value(int64(graphic.Spice.Port))
				}
				if !origSpice.TLSPort.IsNull() {
					spiceModel.TLSPort = types.Int64Value(int64(graphic.Spice.TLSPort))
				}
				if !origSpice.AutoPort.IsNull() && graphic.Spice.AutoPort != "" {
					spiceModel.AutoPort = types.StringValue(graphic.Spice.AutoPort)
				}
				if !origSpice.Listen.IsNull() && graphic.Spice.Listen != "" {
					spiceModel.Listen = types.StringValue(graphic.Spice.Listen)
				}

				spiceObj, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
					"port":     types.Int64Type,
					"tlsport":  types.Int64Type,
					"autoport": types.StringType,
					"listen":   types.StringType,
				}, spiceModel)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
				newGraphics.Spice = spiceObj
			}

			graphicsModel = &newGraphics
		}

		// Create the devices lists
		disksList, d := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"device":    types.StringType,
				"source":    types.StringType,
				"volume_id": types.StringType,
				"target":    types.StringType,
				"bus":       types.StringType,
			},
		}, disks)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		interfacesList, d := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type":  types.StringType,
				"mac":   types.StringType,
				"model": types.StringType,
				"source": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"network":   types.StringType,
						"portgroup": types.StringType,
						"bridge":    types.StringType,
						"dev":       types.StringType,
					},
				},
			},
		}, interfaces)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		// Create graphics object if present
		var graphicsObj types.Object
		if graphicsModel != nil {
			var err diag.Diagnostics
			graphicsObj, err = types.ObjectValueFrom(ctx, map[string]attr.Type{
				"vnc": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"socket":    types.StringType,
						"port":      types.Int64Type,
						"autoport":  types.StringType,
						"websocket": types.Int64Type,
						"listen":    types.StringType,
					},
				},
				"spice": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"port":     types.Int64Type,
						"tlsport":  types.Int64Type,
						"autoport": types.StringType,
						"listen":   types.StringType,
					},
				},
			}, graphicsModel)
			diags.Append(err...)
			if diags.HasError() {
				return diags
			}
		} else {
			graphicsObj = types.ObjectNull(map[string]attr.Type{
				"vnc": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"socket":    types.StringType,
						"port":      types.Int64Type,
						"autoport":  types.StringType,
						"websocket": types.Int64Type,
						"listen":    types.StringType,
					},
				},
				"spice": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"port":     types.Int64Type,
						"tlsport":  types.Int64Type,
						"autoport": types.StringType,
						"listen":   types.StringType,
					},
				},
			})
		}

		// Create the new devices model
		newDevices := DomainDevicesModel{
			Disks:      disksList,
			Interfaces: interfacesList,
			Graphics:   graphicsObj,
		}

		// Create the devices object
		devicesObj, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"disks": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"device":    types.StringType,
						"source":    types.StringType,
						"volume_id": types.StringType,
						"target":    types.StringType,
						"bus":       types.StringType,
					},
				},
			},
			"interfaces": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"type":  types.StringType,
						"mac":   types.StringType,
						"model": types.StringType,
						"source": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"network":   types.StringType,
								"portgroup": types.StringType,
								"bridge":    types.StringType,
								"dev":       types.StringType,
							},
						},
					},
				},
			},
			"graphics": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"vnc": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"socket":    types.StringType,
							"port":      types.Int64Type,
							"autoport":  types.StringType,
							"websocket": types.Int64Type,
							"listen":    types.StringType,
						},
					},
					"spice": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"port":     types.Int64Type,
							"tlsport":  types.Int64Type,
							"autoport": types.StringType,
							"listen":   types.StringType,
						},
					},
				},
			},
		}, newDevices)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		model.Devices = devicesObj
	}

	return diags
}
