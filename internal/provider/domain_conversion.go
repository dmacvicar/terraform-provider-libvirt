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
	if !model.OS.IsNull() && !model.OS.IsUnknown() {
		var osModel DomainOSModel
		diags := model.OS.As(ctx, &osModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, fmt.Errorf("failed to extract OS: %v", diags.Errors())
		}

		os := &libvirtxml.DomainOS{}

		// OS type is required
		if osModel.Type.IsNull() || osModel.Type.IsUnknown() {
			return nil, fmt.Errorf("os.type is required")
		}

		os.Type = &libvirtxml.DomainOSType{
			Type: osModel.Type.ValueString(),
		}

		// Optional OS type attributes
		if !osModel.Arch.IsNull() && !osModel.Arch.IsUnknown() {
			os.Type.Arch = osModel.Arch.ValueString()
		}
		if !osModel.Machine.IsNull() && !osModel.Machine.IsUnknown() {
			os.Type.Machine = osModel.Machine.ValueString()
		}

		// Boot devices
		if !osModel.BootDevices.IsNull() && !osModel.BootDevices.IsUnknown() {
			var bootDevices []types.String
			osModel.BootDevices.ElementsAs(context.Background(), &bootDevices, false)
			for _, dev := range bootDevices {
				if !dev.IsNull() && !dev.IsUnknown() {
					os.BootDevices = append(os.BootDevices, libvirtxml.DomainBootDevice{
						Dev: dev.ValueString(),
					})
				}
			}
		}

		// Direct kernel boot
		if !osModel.Kernel.IsNull() && !osModel.Kernel.IsUnknown() {
			os.Kernel = osModel.Kernel.ValueString()
		}
		if !osModel.Initrd.IsNull() && !osModel.Initrd.IsUnknown() {
			os.Initrd = osModel.Initrd.ValueString()
		}
		if !osModel.KernelArgs.IsNull() && !osModel.KernelArgs.IsUnknown() {
			os.Cmdline = osModel.KernelArgs.ValueString()
		}

		// UEFI loader
		if !osModel.LoaderPath.IsNull() && !osModel.LoaderPath.IsUnknown() {
			loader := &libvirtxml.DomainLoader{
				Path: osModel.LoaderPath.ValueString(),
			}
			if !osModel.LoaderReadOnly.IsNull() && !osModel.LoaderReadOnly.IsUnknown() {
				readOnly := "no"
				if osModel.LoaderReadOnly.ValueBool() {
					readOnly = "yes"
				}
				loader.Readonly = readOnly
			}
			if !osModel.LoaderType.IsNull() && !osModel.LoaderType.IsUnknown() {
				loader.Type = osModel.LoaderType.ValueString()
			}
			os.Loader = loader
		}

		// NVRAM
		if !osModel.NVRAM.IsNull() && !osModel.NVRAM.IsUnknown() {
			var nvramModel DomainNVRAMModel
			diags := osModel.NVRAM.As(ctx, &nvramModel, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				return nil, fmt.Errorf("failed to extract nvram: %v", diags.Errors())
			}

			// Create NVRAM if any NVRAM field is specified
			if !nvramModel.Path.IsNull() && !nvramModel.Path.IsUnknown() ||
				!nvramModel.Template.IsNull() && !nvramModel.Template.IsUnknown() ||
				!nvramModel.Format.IsNull() && !nvramModel.Format.IsUnknown() ||
				!nvramModel.TemplateFormat.IsNull() && !nvramModel.TemplateFormat.IsUnknown() {

				nvram := &libvirtxml.DomainNVRam{}

				// Set path if specified
				if !nvramModel.Path.IsNull() && !nvramModel.Path.IsUnknown() {
					nvram.NVRam = nvramModel.Path.ValueString()
				}

				// Set template if specified
				if !nvramModel.Template.IsNull() && !nvramModel.Template.IsUnknown() {
					nvram.Template = nvramModel.Template.ValueString()
				}

				// Set format if specified
				if !nvramModel.Format.IsNull() && !nvramModel.Format.IsUnknown() {
					nvram.Format = nvramModel.Format.ValueString()
				}

				// Set template format if specified
				if !nvramModel.TemplateFormat.IsNull() && !nvramModel.TemplateFormat.IsUnknown() {
					nvram.TemplateFormat = nvramModel.TemplateFormat.ValueString()
				}

	
				os.NVRam = nvram
			}
		}

		domain.OS = os
	}

	// Set features
	if !model.Features.IsNull() && !model.Features.IsUnknown() {
		var featuresModel DomainFeaturesModel
		diags := model.Features.As(ctx, &featuresModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, fmt.Errorf("failed to extract features: %v", diags.Errors())
		}

		features := &libvirtxml.DomainFeatureList{}

		// PAE - presence-based
		if !featuresModel.PAE.IsNull() && !featuresModel.PAE.IsUnknown() {
			if featuresModel.PAE.ValueBool() {
				features.PAE = &libvirtxml.DomainFeature{}
			}
		}

		// ACPI - presence-based
		if !featuresModel.ACPI.IsNull() && !featuresModel.ACPI.IsUnknown() {
			if featuresModel.ACPI.ValueBool() {
				features.ACPI = &libvirtxml.DomainFeature{}
			}
		}

		// APIC - presence-based
		if !featuresModel.APIC.IsNull() && !featuresModel.APIC.IsUnknown() {
			if featuresModel.APIC.ValueBool() {
				features.APIC = &libvirtxml.DomainFeatureAPIC{}
			}
		}

		// Viridian - presence-based
		if !featuresModel.Viridian.IsNull() && !featuresModel.Viridian.IsUnknown() {
			if featuresModel.Viridian.ValueBool() {
				features.Viridian = &libvirtxml.DomainFeature{}
			}
		}

		// PrivNet - presence-based
		if !featuresModel.PrivNet.IsNull() && !featuresModel.PrivNet.IsUnknown() {
			if featuresModel.PrivNet.ValueBool() {
				features.PrivNet = &libvirtxml.DomainFeature{}
			}
		}

		// HAP - state-based
		if !featuresModel.HAP.IsNull() && !featuresModel.HAP.IsUnknown() {
			features.HAP = &libvirtxml.DomainFeatureState{
				State: featuresModel.HAP.ValueString(),
			}
		}

		// PMU - state-based
		if !featuresModel.PMU.IsNull() && !featuresModel.PMU.IsUnknown() {
			features.PMU = &libvirtxml.DomainFeatureState{
				State: featuresModel.PMU.ValueString(),
			}
		}

		// VMPort - state-based
		if !featuresModel.VMPort.IsNull() && !featuresModel.VMPort.IsUnknown() {
			features.VMPort = &libvirtxml.DomainFeatureState{
				State: featuresModel.VMPort.ValueString(),
			}
		}

		// PVSpinlock - state-based
		if !featuresModel.PVSpinlock.IsNull() && !featuresModel.PVSpinlock.IsUnknown() {
			features.PVSpinlock = &libvirtxml.DomainFeatureState{
				State: featuresModel.PVSpinlock.ValueString(),
			}
		}

		// VMCoreInfo - state-based
		if !featuresModel.VMCoreInfo.IsNull() && !featuresModel.VMCoreInfo.IsUnknown() {
			features.VMCoreInfo = &libvirtxml.DomainFeatureState{
				State: featuresModel.VMCoreInfo.ValueString(),
			}
		}

		// HTM - state-based
		if !featuresModel.HTM.IsNull() && !featuresModel.HTM.IsUnknown() {
			features.HTM = &libvirtxml.DomainFeatureState{
				State: featuresModel.HTM.ValueString(),
			}
		}

		// NestedHV - state-based
		if !featuresModel.NestedHV.IsNull() && !featuresModel.NestedHV.IsUnknown() {
			features.NestedHV = &libvirtxml.DomainFeatureState{
				State: featuresModel.NestedHV.ValueString(),
			}
		}

		// CCFAssist - state-based
		if !featuresModel.CCFAssist.IsNull() && !featuresModel.CCFAssist.IsUnknown() {
			features.CCFAssist = &libvirtxml.DomainFeatureState{
				State: featuresModel.CCFAssist.ValueString(),
			}
		}

		// RAS - state-based
		if !featuresModel.RAS.IsNull() && !featuresModel.RAS.IsUnknown() {
			features.RAS = &libvirtxml.DomainFeatureState{
				State: featuresModel.RAS.ValueString(),
			}
		}

		// PS2 - state-based
		if !featuresModel.PS2.IsNull() && !featuresModel.PS2.IsUnknown() {
			features.PS2 = &libvirtxml.DomainFeatureState{
				State: featuresModel.PS2.ValueString(),
			}
		}

		// IOAPIC - driver attribute
		if !featuresModel.IOAPICDriver.IsNull() && !featuresModel.IOAPICDriver.IsUnknown() {
			features.IOAPIC = &libvirtxml.DomainFeatureIOAPIC{
				Driver: featuresModel.IOAPICDriver.ValueString(),
			}
		}

		// GIC - version attribute
		if !featuresModel.GICVersion.IsNull() && !featuresModel.GICVersion.IsUnknown() {
			features.GIC = &libvirtxml.DomainFeatureGIC{
				Version: featuresModel.GICVersion.ValueString(),
			}
		}

		domain.Features = features
	}

	// Set CPU
	if !model.CPU.IsNull() && !model.CPU.IsUnknown() {
		var cpuModel DomainCPUModel
		diags := model.CPU.As(ctx, &cpuModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, fmt.Errorf("failed to extract cpu: %v", diags.Errors())
		}

		cpu := &libvirtxml.DomainCPU{}

		if !cpuModel.Mode.IsNull() && !cpuModel.Mode.IsUnknown() {
			cpu.Mode = cpuModel.Mode.ValueString()
		}

		if !cpuModel.Match.IsNull() && !cpuModel.Match.IsUnknown() {
			cpu.Match = cpuModel.Match.ValueString()
		}

		if !cpuModel.Check.IsNull() && !cpuModel.Check.IsUnknown() {
			cpu.Check = cpuModel.Check.ValueString()
		}

		if !cpuModel.Migratable.IsNull() && !cpuModel.Migratable.IsUnknown() {
			cpu.Migratable = cpuModel.Migratable.ValueString()
		}

		if !cpuModel.DeprecatedFeatures.IsNull() && !cpuModel.DeprecatedFeatures.IsUnknown() {
			cpu.DeprecatedFeatures = cpuModel.DeprecatedFeatures.ValueString()
		}

		if !cpuModel.Model.IsNull() && !cpuModel.Model.IsUnknown() {
			cpu.Model = &libvirtxml.DomainCPUModel{
				Value: cpuModel.Model.ValueString(),
			}
		}

		if !cpuModel.Vendor.IsNull() && !cpuModel.Vendor.IsUnknown() {
			cpu.Vendor = cpuModel.Vendor.ValueString()
		}

		domain.CPU = cpu
	}

	// Set Clock
	if !model.Clock.IsNull() && !model.Clock.IsUnknown() {
		var clockModel DomainClockModel
		diags := model.Clock.As(ctx, &clockModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, fmt.Errorf("failed to extract clock: %v", diags.Errors())
		}

		clock := &libvirtxml.DomainClock{}

		if !clockModel.Offset.IsNull() && !clockModel.Offset.IsUnknown() {
			clock.Offset = clockModel.Offset.ValueString()
		}

		if !clockModel.Basis.IsNull() && !clockModel.Basis.IsUnknown() {
			clock.Basis = clockModel.Basis.ValueString()
		}

		if !clockModel.Adjustment.IsNull() && !clockModel.Adjustment.IsUnknown() {
			clock.Adjustment = clockModel.Adjustment.ValueString()
		}

		if !clockModel.TimeZone.IsNull() && !clockModel.TimeZone.IsUnknown() {
			clock.TimeZone = clockModel.TimeZone.ValueString()
		}

		// Convert timers
		if !clockModel.Timers.IsNull() && !clockModel.Timers.IsUnknown() {
			var timerModels []DomainTimerModel
			diags := clockModel.Timers.ElementsAs(ctx, &timerModels, false)
			if diags.HasError() {
				return nil, fmt.Errorf("failed to extract timers: %v", diags.Errors())
			}

			for _, timerModel := range timerModels {
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
				if !timerModel.CatchUp.IsNull() && !timerModel.CatchUp.IsUnknown() {
					var catchupModel DomainTimerCatchUpModel
					diags := timerModel.CatchUp.As(ctx, &catchupModel, basetypes.ObjectAsOptions{})
					if diags.HasError() {
						return nil, fmt.Errorf("failed to extract timer catchup: %v", diags.Errors())
					}

					catchup := &libvirtxml.DomainTimerCatchUp{}

					if !catchupModel.Threshold.IsNull() && !catchupModel.Threshold.IsUnknown() {
						catchup.Threshold = uint(catchupModel.Threshold.ValueInt64())
					}

					if !catchupModel.Slew.IsNull() && !catchupModel.Slew.IsUnknown() {
						catchup.Slew = uint(catchupModel.Slew.ValueInt64())
					}

					if !catchupModel.Limit.IsNull() && !catchupModel.Limit.IsUnknown() {
						catchup.Limit = uint(catchupModel.Limit.ValueInt64())
					}

					timer.CatchUp = catchup
				}

				clock.Timer = append(clock.Timer, timer)
			}
		}

		domain.Clock = clock
	}

	// Set PM
	if !model.PM.IsNull() && !model.PM.IsUnknown() {
		var pmModel DomainPMModel
		diags := model.PM.As(ctx, &pmModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, fmt.Errorf("failed to extract pm: %v", diags.Errors())
		}

		pm := &libvirtxml.DomainPM{}

		if !pmModel.SuspendToMem.IsNull() && !pmModel.SuspendToMem.IsUnknown() {
			pm.SuspendToMem = &libvirtxml.DomainPMPolicy{
				Enabled: pmModel.SuspendToMem.ValueString(),
			}
		}

		if !pmModel.SuspendToDisk.IsNull() && !pmModel.SuspendToDisk.IsUnknown() {
			pm.SuspendToDisk = &libvirtxml.DomainPMPolicy{
				Enabled: pmModel.SuspendToDisk.ValueString(),
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

			// Set source - either from volume_id, direct file path, or block device
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
					disk.Source = &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: volumeDef.Target.Path,
						},
					}
				} else {
					return nil, fmt.Errorf("volume %s has no target path", volumeKey)
				}
			} else if !diskModel.Source.IsNull() && !diskModel.Source.IsUnknown() {
				// File-based disk
				disk.Source = &libvirtxml.DomainDiskSource{
					File: &libvirtxml.DomainDiskSourceFile{
						File: diskModel.Source.ValueString(),
					},
				}
			} else if !diskModel.BlockDevice.IsNull() && !diskModel.BlockDevice.IsUnknown() {
				// Block device disk
				disk.Source = &libvirtxml.DomainDiskSource{
					Block: &libvirtxml.DomainDiskSourceBlock{
						Dev: diskModel.BlockDevice.ValueString(),
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

			// Set WWN if specified
			if !diskModel.WWN.IsNull() && !diskModel.WWN.IsUnknown() {
				disk.WWN = diskModel.WWN.ValueString()
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

				case "direct":
					direct := &libvirtxml.DomainInterfaceSourceDirect{}
					if !ifaceModel.Source.Dev.IsNull() && !ifaceModel.Source.Dev.IsUnknown() {
						direct.Dev = ifaceModel.Source.Dev.ValueString()
					}
					if !ifaceModel.Source.Mode.IsNull() && !ifaceModel.Source.Mode.IsUnknown() {
						direct.Mode = ifaceModel.Source.Mode.ValueString()
					}
					source.Direct = direct
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

		// Process filesystems
		if !devices.Filesystems.IsNull() && !devices.Filesystems.IsUnknown() {
			var filesystems []DomainFilesystemModel
			diags := devices.Filesystems.ElementsAs(ctx, &filesystems, false)
			if diags.HasError() {
				return nil, fmt.Errorf("failed to extract filesystems: %v", diags.Errors())
			}

			for _, fsModel := range filesystems {
				fs := libvirtxml.DomainFilesystem{
					Source: &libvirtxml.DomainFilesystemSource{
						Mount: &libvirtxml.DomainFilesystemSourceMount{
							Dir: fsModel.Source.ValueString(),
						},
					},
					Target: &libvirtxml.DomainFilesystemTarget{
						Dir: fsModel.Target.ValueString(),
					},
				}

				// Set access mode (default to "mapped" if not specified)
				if !fsModel.AccessMode.IsNull() && !fsModel.AccessMode.IsUnknown() {
					fs.AccessMode = fsModel.AccessMode.ValueString()
				} else {
					fs.AccessMode = "mapped"
				}

				// Set readonly
				if !fsModel.ReadOnly.IsNull() && !fsModel.ReadOnly.IsUnknown() && fsModel.ReadOnly.ValueBool() {
					fs.ReadOnly = &libvirtxml.DomainFilesystemReadOnly{}
				}

				domain.Devices.Filesystems = append(domain.Devices.Filesystems, fs)
			}
		}

		// Process video
		if !devices.Video.IsNull() && !devices.Video.IsUnknown() {
			var videoModel DomainVideoModel
			diags := devices.Video.As(ctx, &videoModel, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				return nil, fmt.Errorf("failed to extract video: %v", diags.Errors())
			}

			video := libvirtxml.DomainVideo{
				Model: libvirtxml.DomainVideoModel{},
			}

			// Set video type
			if !videoModel.Type.IsNull() && !videoModel.Type.IsUnknown() {
				video.Model.Type = videoModel.Type.ValueString()
			}

			domain.Devices.Videos = append(domain.Devices.Videos, video)
		}

		// Process emulator
		if !devices.Emulator.IsNull() && !devices.Emulator.IsUnknown() {
			domain.Devices.Emulator = devices.Emulator.ValueString()
		}
	// Process consoles
	if !devices.Consoles.IsNull() && !devices.Consoles.IsUnknown() {
		var consoles []DomainConsoleModel
		diags := devices.Consoles.ElementsAs(ctx, &consoles, false)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to extract consoles: %v", diags.Errors())
		}

		for _, consoleModel := range consoles {
			console := libvirtxml.DomainConsole{}

			// Set source based on type
			sourceType := "pty" // default
			if !consoleModel.Type.IsNull() && !consoleModel.Type.IsUnknown() {
				sourceType = consoleModel.Type.ValueString()
			}

			switch sourceType {
			case "pty":
				console.Source = &libvirtxml.DomainChardevSource{
					Pty: &libvirtxml.DomainChardevSourcePty{},
				}
				if !consoleModel.SourcePath.IsNull() && !consoleModel.SourcePath.IsUnknown() {
					console.Source.Pty.Path = consoleModel.SourcePath.ValueString()
				}
			case "file":
				console.Source = &libvirtxml.DomainChardevSource{
					File: &libvirtxml.DomainChardevSourceFile{},
				}
				if !consoleModel.SourcePath.IsNull() && !consoleModel.SourcePath.IsUnknown() {
					console.Source.File.Path = consoleModel.SourcePath.ValueString()
				}
			}

			// Set target
			if !consoleModel.TargetType.IsNull() || !consoleModel.TargetPort.IsNull() {
				console.Target = &libvirtxml.DomainConsoleTarget{}
				if !consoleModel.TargetType.IsNull() && !consoleModel.TargetType.IsUnknown() {
					console.Target.Type = consoleModel.TargetType.ValueString()
				}
				if !consoleModel.TargetPort.IsNull() && !consoleModel.TargetPort.IsUnknown() {
					port := uint(consoleModel.TargetPort.ValueInt64())
					console.Target.Port = &port
				}
			}

			domain.Devices.Consoles = append(domain.Devices.Consoles, console)
		}
	}

	// Process serials
	if !devices.Serials.IsNull() && !devices.Serials.IsUnknown() {
		var serials []DomainSerialModel
		diags := devices.Serials.ElementsAs(ctx, &serials, false)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to extract serials: %v", diags.Errors())
		}

		for _, serialModel := range serials {
			serial := libvirtxml.DomainSerial{}

			// Set source based on type
			sourceType := "pty" // default
			if !serialModel.Type.IsNull() && !serialModel.Type.IsUnknown() {
				sourceType = serialModel.Type.ValueString()
			}

			switch sourceType {
			case "pty":
				serial.Source = &libvirtxml.DomainChardevSource{
					Pty: &libvirtxml.DomainChardevSourcePty{},
				}
				if !serialModel.SourcePath.IsNull() && !serialModel.SourcePath.IsUnknown() {
					serial.Source.Pty.Path = serialModel.SourcePath.ValueString()
				}
			case "file":
				serial.Source = &libvirtxml.DomainChardevSource{
					File: &libvirtxml.DomainChardevSourceFile{},
				}
				if !serialModel.SourcePath.IsNull() && !serialModel.SourcePath.IsUnknown() {
					serial.Source.File.Path = serialModel.SourcePath.ValueString()
				}
			}

			// Set target
			if !serialModel.TargetType.IsNull() || !serialModel.TargetPort.IsNull() {
				serial.Target = &libvirtxml.DomainSerialTarget{}
				if !serialModel.TargetType.IsNull() && !serialModel.TargetType.IsUnknown() {
					serial.Target.Type = serialModel.TargetType.ValueString()
				}
				if !serialModel.TargetPort.IsNull() && !serialModel.TargetPort.IsUnknown() {
					port := uint(serialModel.TargetPort.ValueInt64())
					serial.Target.Port = &port
				}
			}

			domain.Devices.Serials = append(domain.Devices.Serials, serial)
		}
	}

	// Process RNGs
	if !devices.RNGs.IsNull() && !devices.RNGs.IsUnknown() {
		var rngs []DomainRNGModel
		diags := devices.RNGs.ElementsAs(ctx, &rngs, false)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to extract rngs: %v", diags.Errors())
		}

		for _, rngModel := range rngs {
			rng := libvirtxml.DomainRNG{}

			// Set model (default to virtio)
			if !rngModel.Model.IsNull() && !rngModel.Model.IsUnknown() {
				rng.Model = rngModel.Model.ValueString()
			} else {
				rng.Model = "virtio"
			}

			// Set backend (Random device)
			device := "/dev/urandom" // default
			if !rngModel.Device.IsNull() && !rngModel.Device.IsUnknown() {
				device = rngModel.Device.ValueString()
			}

			rng.Backend = &libvirtxml.DomainRNGBackend{
				Random: &libvirtxml.DomainRNGBackendRandom{
					Device: device,
				},
			}

			domain.Devices.RNGs = append(domain.Devices.RNGs, rng)
		}
	}

	// Process TPMs
	if !devices.TPMs.IsNull() && !devices.TPMs.IsUnknown() {
		var tpms []DomainTPMModel
		diags := devices.TPMs.ElementsAs(ctx, &tpms, false)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to extract tpms: %v", diags.Errors())
		}

		for _, tpmModel := range tpms {
			tpm := libvirtxml.DomainTPM{}

			// Set model
			if !tpmModel.Model.IsNull() && !tpmModel.Model.IsUnknown() {
				tpm.Model = tpmModel.Model.ValueString()
			}

			// Set backend type (default to emulator)
			backendType := "emulator"
			if !tpmModel.BackendType.IsNull() && !tpmModel.BackendType.IsUnknown() {
				backendType = tpmModel.BackendType.ValueString()
			}

			switch backendType {
			case "passthrough":
				if !tpmModel.BackendDevicePath.IsNull() && !tpmModel.BackendDevicePath.IsUnknown() {
					tpm.Backend = &libvirtxml.DomainTPMBackend{
						Passthrough: &libvirtxml.DomainTPMBackendPassthrough{
							Device: &libvirtxml.DomainTPMBackendDevice{
								Path: tpmModel.BackendDevicePath.ValueString(),
							},
						},
					}
				}
			case "emulator":
				emulator := &libvirtxml.DomainTPMBackendEmulator{}

				if !tpmModel.BackendVersion.IsNull() && !tpmModel.BackendVersion.IsUnknown() {
					emulator.Version = tpmModel.BackendVersion.ValueString()
				}

				if !tpmModel.BackendEncryptionSecret.IsNull() && !tpmModel.BackendEncryptionSecret.IsUnknown() {
					emulator.Encryption = &libvirtxml.DomainTPMBackendEncryption{
						Secret: tpmModel.BackendEncryptionSecret.ValueString(),
					}
				}

				if !tpmModel.BackendPersistentState.IsNull() && !tpmModel.BackendPersistentState.IsUnknown() {
					if tpmModel.BackendPersistentState.ValueBool() {
						emulator.PersistentState = "yes"
					} else {
						emulator.PersistentState = "no"
					}
				}

				tpm.Backend = &libvirtxml.DomainTPMBackend{
					Emulator: emulator,
				}
			}

			domain.Devices.TPMs = append(domain.Devices.TPMs, tpm)
		}
	}
	}

	return domain, nil
}

// xmlToDomainModel converts libvirtxml.Domain to a DomainResourceModel
func xmlToDomainModel(ctx context.Context, domain *libvirtxml.Domain, model *DomainResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	model.Name = types.StringValue(domain.Name)

	// Only set type if user specified it
	if !model.Type.IsNull() && !model.Type.IsUnknown() {
		model.Type = types.StringValue(domain.Type)
	}

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
			model.Memory = types.Int64Value(memoryKiB)
		}
	}

	// Only set current_memory if user specified it
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

	// Only set lifecycle actions if user specified them
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

	if !model.OS.IsNull() && !model.OS.IsUnknown() && domain.OS != nil {
		var origOS DomainOSModel
		d := model.OS.As(ctx, &origOS, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		osModel := DomainOSModel{}

		if domain.OS.Type != nil {
			osModel.Type = types.StringValue(domain.OS.Type.Type)

			// Only set arch if user specified it
			if !origOS.Arch.IsNull() && !origOS.Arch.IsUnknown() && domain.OS.Type.Arch != "" {
				osModel.Arch = types.StringValue(domain.OS.Type.Arch)
			}

			// Preserve user's value to avoid diffs from libvirt expansion (e.g., q35 -> pc-q35-10.1)
			if !origOS.Machine.IsNull() && !origOS.Machine.IsUnknown() {
				osModel.Machine = origOS.Machine
			}
		}

		// Only set boot_devices if user specified them
		if !origOS.BootDevices.IsNull() && !origOS.BootDevices.IsUnknown() {
			osModel.BootDevices = origOS.BootDevices
		} else {
			osModel.BootDevices = types.ListNull(types.StringType)
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
			// Only set loader_type if user originally specified it
			if !origOS.LoaderType.IsNull() && !origOS.LoaderType.IsUnknown() && domain.OS.Loader.Type != "" {
				osModel.LoaderType = types.StringValue(domain.OS.Loader.Type)
			}
		}

		// Handle NVRAM - only if user specified it
		var nvramObj types.Object
		if !origOS.NVRAM.IsNull() && !origOS.NVRAM.IsUnknown() {
			// Extract original NVRAM model to check what user specified
			var origNVRAM DomainNVRAMModel
			d := origOS.NVRAM.As(ctx, &origNVRAM, basetypes.ObjectAsOptions{})
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			nvramModel := DomainNVRAMModel{}

			// Only set path if user specified it and we have NVRam from libvirt
			if !origNVRAM.Path.IsNull() && !origNVRAM.Path.IsUnknown() && domain.OS.NVRam != nil && domain.OS.NVRam.NVRam != "" {
				nvramModel.Path = types.StringValue(domain.OS.NVRam.NVRam)
			}

			// Only set template if user specified it and we have template from libvirt
			if !origNVRAM.Template.IsNull() && !origNVRAM.Template.IsUnknown() && domain.OS.NVRam != nil && domain.OS.NVRam.Template != "" {
				nvramModel.Template = types.StringValue(domain.OS.NVRam.Template)
			}

			// Only set format if user specified it and we have format from libvirt
			if !origNVRAM.Format.IsNull() && !origNVRAM.Format.IsUnknown() && domain.OS.NVRam != nil && domain.OS.NVRam.Format != "" {
				nvramModel.Format = types.StringValue(domain.OS.NVRam.Format)
			}

			// Only set template format if user specified it and we have template format from libvirt
			if !origNVRAM.TemplateFormat.IsNull() && !origNVRAM.TemplateFormat.IsUnknown() && domain.OS.NVRam != nil && domain.OS.NVRam.TemplateFormat != "" {
				nvramModel.TemplateFormat = types.StringValue(domain.OS.NVRam.TemplateFormat)
			}

		
	
			nvram, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
				"path":           types.StringType,
				"template":       types.StringType,
				"format":         types.StringType,
				"template_format": types.StringType,
			}, nvramModel)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			nvramObj = nvram
		} else {
			nvramObj = types.ObjectNull(map[string]attr.Type{
				"path":           types.StringType,
				"template":       types.StringType,
				"format":         types.StringType,
				"template_format": types.StringType,
			})
		}

		osModel.NVRAM = nvramObj

		osObj, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"type":            types.StringType,
			"arch":            types.StringType,
			"machine":         types.StringType,
			"firmware":        types.StringType,
			"boot_devices":    types.ListType{ElemType: types.StringType},
			"kernel":          types.StringType,
			"initrd":          types.StringType,
			"kernel_args":     types.StringType,
			"loader_path":     types.StringType,
			"loader_readonly": types.BoolType,
			"loader_type":     types.StringType,
			"nvram":           types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"path":           types.StringType,
					"template":       types.StringType,
					"format":         types.StringType,
					"template_format": types.StringType,
				},
			},
		}, osModel)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		model.OS = osObj
	}

	// Features - only if user specified it
	if !model.Features.IsNull() && !model.Features.IsUnknown() && domain.Features != nil {
		featuresModel := DomainFeaturesModel{}

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

		featuresObj, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"pae":           types.BoolType,
			"acpi":          types.BoolType,
			"apic":          types.BoolType,
			"viridian":      types.BoolType,
			"privnet":       types.BoolType,
			"hap":           types.StringType,
			"pmu":           types.StringType,
			"vmport":        types.StringType,
			"pvspinlock":    types.StringType,
			"vmcoreinfo":    types.StringType,
			"htm":           types.StringType,
			"nested_hv":     types.StringType,
			"ccf_assist":    types.StringType,
			"ras":           types.StringType,
			"ps2":           types.StringType,
			"ioapic_driver": types.StringType,
			"gic_version":   types.StringType,
		}, featuresModel)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		model.Features = featuresObj
	}

	// Only set CPU if user specified it
	if !model.CPU.IsNull() && !model.CPU.IsUnknown() && domain.CPU != nil {
		// Extract original cpu model to check what user specified
		var origCPU DomainCPUModel
		d := model.CPU.As(ctx, &origCPU, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		cpuModel := DomainCPUModel{}

		// Only set mode if user specified it
		if !origCPU.Mode.IsNull() && !origCPU.Mode.IsUnknown() && domain.CPU.Mode != "" {
			cpuModel.Mode = types.StringValue(domain.CPU.Mode)
		}

		// Only set match if user specified it
		if !origCPU.Match.IsNull() && !origCPU.Match.IsUnknown() && domain.CPU.Match != "" {
			cpuModel.Match = types.StringValue(domain.CPU.Match)
		}

		// Only set check if user specified it
		if !origCPU.Check.IsNull() && !origCPU.Check.IsUnknown() && domain.CPU.Check != "" {
			cpuModel.Check = types.StringValue(domain.CPU.Check)
		}

		// Only set migratable if user specified it
		if !origCPU.Migratable.IsNull() && !origCPU.Migratable.IsUnknown() && domain.CPU.Migratable != "" {
			cpuModel.Migratable = types.StringValue(domain.CPU.Migratable)
		}

		// Only set deprecated_features if user specified it
		if !origCPU.DeprecatedFeatures.IsNull() && !origCPU.DeprecatedFeatures.IsUnknown() && domain.CPU.DeprecatedFeatures != "" {
			cpuModel.DeprecatedFeatures = types.StringValue(domain.CPU.DeprecatedFeatures)
		}

		// Only set model if user specified it
		if !origCPU.Model.IsNull() && !origCPU.Model.IsUnknown() && domain.CPU.Model != nil && domain.CPU.Model.Value != "" {
			cpuModel.Model = types.StringValue(domain.CPU.Model.Value)
		}

		// Only set vendor if user specified it
		if !origCPU.Vendor.IsNull() && !origCPU.Vendor.IsUnknown() && domain.CPU.Vendor != "" {
			cpuModel.Vendor = types.StringValue(domain.CPU.Vendor)
		}

		cpuObj, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"mode":                types.StringType,
			"match":               types.StringType,
			"check":               types.StringType,
			"migratable":          types.StringType,
			"deprecated_features": types.StringType,
			"model":               types.StringType,
			"vendor":              types.StringType,
		}, cpuModel)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		model.CPU = cpuObj
	}

	// Only set clock if user specified it
	if !model.Clock.IsNull() && !model.Clock.IsUnknown() && domain.Clock != nil {
		// Extract original clock to check what user specified
		var origClock DomainClockModel
		d := model.Clock.As(ctx, &origClock, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		clockModel := DomainClockModel{}

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
		if !origClock.Timers.IsNull() && !origClock.Timers.IsUnknown() && len(domain.Clock.Timer) > 0 {
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
					catchupModel := DomainTimerCatchUpModel{}

					if timer.CatchUp.Threshold != 0 {
						catchupModel.Threshold = types.Int64Value(int64(timer.CatchUp.Threshold))
					}

					if timer.CatchUp.Slew != 0 {
						catchupModel.Slew = types.Int64Value(int64(timer.CatchUp.Slew))
					}

					if timer.CatchUp.Limit != 0 {
						catchupModel.Limit = types.Int64Value(int64(timer.CatchUp.Limit))
					}

					catchupObj, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
						"threshold": types.Int64Type,
						"slew":      types.Int64Type,
						"limit":     types.Int64Type,
					}, catchupModel)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					timerModel.CatchUp = catchupObj
				} else {
					timerModel.CatchUp = types.ObjectNull(map[string]attr.Type{
						"threshold": types.Int64Type,
						"slew":      types.Int64Type,
						"limit":     types.Int64Type,
					})
				}

				timers = append(timers, timerModel)
			}

			timersList, d := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":       types.StringType,
					"track":      types.StringType,
					"tickpolicy": types.StringType,
					"frequency":  types.Int64Type,
					"mode":       types.StringType,
					"present":    types.StringType,
					"catchup": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"threshold": types.Int64Type,
							"slew":      types.Int64Type,
							"limit":     types.Int64Type,
						},
					},
				},
			}, timers)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			clockModel.Timers = timersList
		} else {
			clockModel.Timers = types.ListNull(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":       types.StringType,
					"track":      types.StringType,
					"tickpolicy": types.StringType,
					"frequency":  types.Int64Type,
					"mode":       types.StringType,
					"present":    types.StringType,
					"catchup": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"threshold": types.Int64Type,
							"slew":      types.Int64Type,
							"limit":     types.Int64Type,
						},
					},
				},
			})
		}

		clockObj, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"offset":     types.StringType,
			"basis":      types.StringType,
			"adjustment": types.StringType,
			"timezone":   types.StringType,
			"timer": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":       types.StringType,
						"track":      types.StringType,
						"tickpolicy": types.StringType,
						"frequency":  types.Int64Type,
						"mode":       types.StringType,
						"present":    types.StringType,
						"catchup": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"threshold": types.Int64Type,
								"slew":      types.Int64Type,
								"limit":     types.Int64Type,
							},
						},
					},
				},
			},
		}, clockModel)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		model.Clock = clockObj
	}

	// Only set PM if user specified it
	if !model.PM.IsNull() && !model.PM.IsUnknown() && domain.PM != nil {
		pmModel := DomainPMModel{}

		if domain.PM.SuspendToMem != nil && domain.PM.SuspendToMem.Enabled != "" {
			pmModel.SuspendToMem = types.StringValue(domain.PM.SuspendToMem.Enabled)
		}

		if domain.PM.SuspendToDisk != nil && domain.PM.SuspendToDisk.Enabled != "" {
			pmModel.SuspendToDisk = types.StringValue(domain.PM.SuspendToDisk.Enabled)
		}

		pmObj, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"suspend_to_mem":  types.StringType,
			"suspend_to_disk": types.StringType,
		}, pmModel)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		model.PM = pmObj
	}

	// Only set devices if user specified them
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

		// Process disks
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

				// Preserve block_device only when the user specified a block device path
				if !orig.BlockDevice.IsNull() && !orig.BlockDevice.IsUnknown() && disk.Source != nil && disk.Source.Block != nil && disk.Source.Block.Dev != "" {
					diskModel.BlockDevice = types.StringValue(disk.Source.Block.Dev)
				}

				if disk.Target != nil {
					if disk.Target.Dev != "" {
						diskModel.Target = types.StringValue(disk.Target.Dev)
					}
					if !orig.Bus.IsNull() && !orig.Bus.IsUnknown() && disk.Target.Bus != "" {
						diskModel.Bus = types.StringValue(disk.Target.Bus)
					}
				}

				// Set WWN if present
				if !orig.WWN.IsNull() && !orig.WWN.IsUnknown() && disk.WWN != "" {
					diskModel.WWN = types.StringValue(disk.WWN)
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
					} else if iface.Source.Direct != nil {
						ifaceModel.Type = types.StringValue("direct")
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
					} else if iface.Source.Direct != nil {
						if !origModel.Source.Dev.IsNull() && iface.Source.Direct.Dev != "" {
							sourceModel.Dev = types.StringValue(iface.Source.Direct.Dev)
						}
						if !origModel.Source.Mode.IsNull() && iface.Source.Direct.Mode != "" {
							sourceModel.Mode = types.StringValue(iface.Source.Direct.Mode)
						}
					}

					ifaceModel.Source = sourceModel
				}

				interfaces = append(interfaces, ifaceModel)
			}
		}

		// Process graphics
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
				"device":       types.StringType,
				"source":       types.StringType,
				"volume_id":    types.StringType,
				"block_device": types.StringType,
				"target":       types.StringType,
				"bus":          types.StringType,
				"wwn":          types.StringType,
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
						"mode":      types.StringType,
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

		// Process filesystems
		filesystemsType := types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"accessmode": types.StringType,
					"source":     types.StringType,
					"target":     types.StringType,
					"readonly":   types.BoolType,
				},
			},
		}
		filesystemsList := types.ListNull(filesystemsType.ElemType.(types.ObjectType))

		if !model.Devices.IsNull() && !model.Devices.IsUnknown() {
			var existingDevices DomainDevicesModel
			diags.Append(model.Devices.As(ctx, &existingDevices, basetypes.ObjectAsOptions{})...)

			if !existingDevices.Filesystems.IsNull() && !existingDevices.Filesystems.IsUnknown() {
				filesystems := make([]DomainFilesystemModel, 0, len(domain.Devices.Filesystems))
				for _, fs := range domain.Devices.Filesystems {
					fsModel := DomainFilesystemModel{}

					if fs.Source != nil && fs.Source.Mount != nil {
						fsModel.Source = types.StringValue(fs.Source.Mount.Dir)
					}

					if fs.Target != nil {
						fsModel.Target = types.StringValue(fs.Target.Dir)
					}

					if fs.AccessMode != "" {
						fsModel.AccessMode = types.StringValue(fs.AccessMode)
					} else {
						fsModel.AccessMode = types.StringValue("mapped")
					}

					if fs.ReadOnly != nil {
						fsModel.ReadOnly = types.BoolValue(true)
					} else {
						fsModel.ReadOnly = types.BoolValue(false)
					}

					filesystems = append(filesystems, fsModel)
				}

				filesystemsList, d = types.ListValueFrom(ctx, types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"accessmode": types.StringType,
						"source":     types.StringType,
						"target":     types.StringType,
						"readonly":   types.BoolType,
					},
				}, filesystems)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
			}
		}

	// Process video
	var videoObj types.Object
	if !model.Devices.IsNull() && !model.Devices.IsUnknown() {
		var existingDevices DomainDevicesModel
		diags.Append(model.Devices.As(ctx, &existingDevices, basetypes.ObjectAsOptions{})...)

		if !existingDevices.Video.IsNull() && !existingDevices.Video.IsUnknown() && len(domain.Devices.Videos) > 0 {
			video := domain.Devices.Videos[0]
			videoModel := DomainVideoModel{}

			if video.Model.Type != "" {
				videoModel.Type = types.StringValue(video.Model.Type)
			}

			var err diag.Diagnostics
			videoObj, err = types.ObjectValueFrom(ctx, map[string]attr.Type{
				"type": types.StringType,
			}, videoModel)
			diags.Append(err...)
			if diags.HasError() {
				return diags
			}
		} else {
			videoObj = types.ObjectNull(map[string]attr.Type{
				"type": types.StringType,
			})
		}
	} else {
		videoObj = types.ObjectNull(map[string]attr.Type{
			"type": types.StringType,
		})
	}


	// Process emulator
	var emulatorStr types.String
	if !model.Devices.IsNull() && !model.Devices.IsUnknown() {
		var existingDevices DomainDevicesModel
		diags.Append(model.Devices.As(ctx, &existingDevices, basetypes.ObjectAsOptions{})...)

		if !existingDevices.Emulator.IsNull() && !existingDevices.Emulator.IsUnknown() && domain.Devices.Emulator != "" {
			emulatorStr = types.StringValue(domain.Devices.Emulator)
		} else {
			emulatorStr = types.StringNull()
		}
	} else {
		emulatorStr = types.StringNull()
	}
		// Create the new devices model
	// Process consoles
	consolesType := types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type":        types.StringType,
				"source_path": types.StringType,
				"target_type": types.StringType,
				"target_port": types.Int64Type,
			},
		},
	}
	consolesList := types.ListNull(consolesType.ElemType.(types.ObjectType))

	if !model.Devices.IsNull() && !model.Devices.IsUnknown() {
		var existingDevices DomainDevicesModel
		diags.Append(model.Devices.As(ctx, &existingDevices, basetypes.ObjectAsOptions{})...)

		if !existingDevices.Consoles.IsNull() && !existingDevices.Consoles.IsUnknown() && len(domain.Devices.Consoles) > 0 {
			consoles := make([]DomainConsoleModel, 0, len(domain.Devices.Consoles))
			for _, console := range domain.Devices.Consoles {
				consoleModel := DomainConsoleModel{}

				// Determine source type
				if console.Source != nil {
					if console.Source.Pty != nil {
						consoleModel.Type = types.StringValue("pty")
						if console.Source.Pty.Path != "" {
							consoleModel.SourcePath = types.StringValue(console.Source.Pty.Path)
						}
					} else if console.Source.File != nil {
						consoleModel.Type = types.StringValue("file")
						if console.Source.File.Path != "" {
							consoleModel.SourcePath = types.StringValue(console.Source.File.Path)
						}
					}
				}

				// Process target
				if console.Target != nil {
					if console.Target.Type != "" {
						consoleModel.TargetType = types.StringValue(console.Target.Type)
					}
					if console.Target.Port != nil {
						consoleModel.TargetPort = types.Int64Value(int64(*console.Target.Port))
					}
				}

				consoles = append(consoles, consoleModel)
			}

			consolesList, d = types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"type":        types.StringType,
					"source_path": types.StringType,
					"target_type": types.StringType,
					"target_port": types.Int64Type,
				},
			}, consoles)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
		}
	}

	// Process serials
	serialsType := types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type":        types.StringType,
				"source_path": types.StringType,
				"target_type": types.StringType,
				"target_port": types.Int64Type,
			},
		},
	}
	serialsList := types.ListNull(serialsType.ElemType.(types.ObjectType))

	if !model.Devices.IsNull() && !model.Devices.IsUnknown() {
		var existingDevices DomainDevicesModel
		diags.Append(model.Devices.As(ctx, &existingDevices, basetypes.ObjectAsOptions{})...)

		if !existingDevices.Serials.IsNull() && !existingDevices.Serials.IsUnknown() && len(domain.Devices.Serials) > 0 {
			serials := make([]DomainSerialModel, 0, len(domain.Devices.Serials))
			for _, serial := range domain.Devices.Serials {
				serialModel := DomainSerialModel{}

				// Determine source type
				if serial.Source != nil {
					if serial.Source.Pty != nil {
						serialModel.Type = types.StringValue("pty")
						if serial.Source.Pty.Path != "" {
							serialModel.SourcePath = types.StringValue(serial.Source.Pty.Path)
						}
					} else if serial.Source.File != nil {
						serialModel.Type = types.StringValue("file")
						if serial.Source.File.Path != "" {
							serialModel.SourcePath = types.StringValue(serial.Source.File.Path)
						}
					}
				}

				// Process target
				if serial.Target != nil {
					if serial.Target.Type != "" {
						serialModel.TargetType = types.StringValue(serial.Target.Type)
					}
					if serial.Target.Port != nil {
						serialModel.TargetPort = types.Int64Value(int64(*serial.Target.Port))
					}
				}

				serials = append(serials, serialModel)
			}

			serialsList, d = types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"type":        types.StringType,
					"source_path": types.StringType,
					"target_type": types.StringType,
					"target_port": types.Int64Type,
				},
			}, serials)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
		}
	}

	// Process RNGs
	rngsType := types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"model":  types.StringType,
				"device": types.StringType,
			},
		},
	}
	rngsList := types.ListNull(rngsType.ElemType.(types.ObjectType))

	if !model.Devices.IsNull() && !model.Devices.IsUnknown() {
		var existingDevices DomainDevicesModel
		diags.Append(model.Devices.As(ctx, &existingDevices, basetypes.ObjectAsOptions{})...)

		if !existingDevices.RNGs.IsNull() && !existingDevices.RNGs.IsUnknown() && len(domain.Devices.RNGs) > 0 {
			rngs := make([]DomainRNGModel, 0, len(domain.Devices.RNGs))
			for _, rng := range domain.Devices.RNGs {
				rngModel := DomainRNGModel{}

				if rng.Model != "" {
					rngModel.Model = types.StringValue(rng.Model)
				}

				// Extract device path from backend
				if rng.Backend != nil && rng.Backend.Random != nil && rng.Backend.Random.Device != "" {
					rngModel.Device = types.StringValue(rng.Backend.Random.Device)
				}

				rngs = append(rngs, rngModel)
			}

			rngsList, d = types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"model":  types.StringType,
					"device": types.StringType,
				},
			}, rngs)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
		}
	}

	// Process TPMs
	tpmsType := types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"model":                         types.StringType,
				"backend_type":                  types.StringType,
				"backend_device_path":           types.StringType,
				"backend_encryption_secret":     types.StringType,
				"backend_version":               types.StringType,
				"backend_persistent_state":      types.BoolType,
			},
		},
	}
	tpmsList := types.ListNull(tpmsType.ElemType.(types.ObjectType))

	if !model.Devices.IsNull() && !model.Devices.IsUnknown() {
		var existingDevices DomainDevicesModel
		diags.Append(model.Devices.As(ctx, &existingDevices, basetypes.ObjectAsOptions{})...)

		if !existingDevices.TPMs.IsNull() && !existingDevices.TPMs.IsUnknown() && len(domain.Devices.TPMs) > 0 {
			tpms := make([]DomainTPMModel, 0, len(domain.Devices.TPMs))
			for _, tpm := range domain.Devices.TPMs {
				tpmModel := DomainTPMModel{}

				if tpm.Model != "" {
					tpmModel.Model = types.StringValue(tpm.Model)
				}

				// Extract backend information
				if tpm.Backend != nil {
					if tpm.Backend.Passthrough != nil {
						tpmModel.BackendType = types.StringValue("passthrough")
						if tpm.Backend.Passthrough.Device != nil && tpm.Backend.Passthrough.Device.Path != "" {
							tpmModel.BackendDevicePath = types.StringValue(tpm.Backend.Passthrough.Device.Path)
						}
					} else if tpm.Backend.Emulator != nil {
						tpmModel.BackendType = types.StringValue("emulator")
						if tpm.Backend.Emulator.Version != "" {
							tpmModel.BackendVersion = types.StringValue(tpm.Backend.Emulator.Version)
						}
						if tpm.Backend.Emulator.Encryption != nil && tpm.Backend.Emulator.Encryption.Secret != "" {
							tpmModel.BackendEncryptionSecret = types.StringValue(tpm.Backend.Emulator.Encryption.Secret)
						}
						if tpm.Backend.Emulator.PersistentState != "" {
							tpmModel.BackendPersistentState = types.BoolValue(tpm.Backend.Emulator.PersistentState == "yes")
						}
					}
				}

				tpms = append(tpms, tpmModel)
			}

			tpmsList, d = types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"model":                         types.StringType,
					"backend_type":                  types.StringType,
					"backend_device_path":           types.StringType,
					"backend_encryption_secret":     types.StringType,
					"backend_version":               types.StringType,
					"backend_persistent_state":      types.BoolType,
				},
			}, tpms)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
		}
	}

		newDevices := DomainDevicesModel{
			Disks:       disksList,
			Interfaces:  interfacesList,
			Graphics:    graphicsObj,
			Filesystems: filesystemsList,
		Video:       videoObj,
		Emulator:    emulatorStr,
		Consoles:    consolesList,
		Serials:     serialsList,
		RNGs:        rngsList,
		TPMs:        tpmsList,
		}

		// Create the devices object
		devicesObj, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"disks": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"device":       types.StringType,
						"source":       types.StringType,
						"volume_id":    types.StringType,
						"block_device": types.StringType,
						"target":       types.StringType,
						"bus":          types.StringType,
						"wwn":          types.StringType,
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
								"mode":      types.StringType,
							},
						},
					},
				},
			},
			"filesystems": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"accessmode": types.StringType,
						"source":     types.StringType,
						"target":     types.StringType,
						"readonly":   types.BoolType,
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
			"video": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"type": types.StringType,
				},
			},
			"emulator": types.StringType,
			"consoles": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"type":        types.StringType,
						"source_path": types.StringType,
						"target_type": types.StringType,
						"target_port": types.Int64Type,
					},
				},
			},
			"serials": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"type":        types.StringType,
						"source_path": types.StringType,
						"target_type": types.StringType,
						"target_port": types.Int64Type,
					},
				},
			},
			"rngs": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"model":  types.StringType,
						"device": types.StringType,
					},
				},
			},
			"tpms": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"model":                      types.StringType,
						"backend_type":               types.StringType,
						"backend_device_path":        types.StringType,
						"backend_encryption_secret":  types.StringType,
						"backend_version":            types.StringType,
						"backend_persistent_state":   types.BoolType,
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
