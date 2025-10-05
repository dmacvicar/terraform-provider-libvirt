// Package provider implements the Terraform provider for libvirt.
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
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
func domainModelToXML(model *DomainResourceModel) (*libvirtxml.Domain, error) {
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

	// Set title
	if !model.Title.IsNull() && !model.Title.IsUnknown() {
		domain.Title = model.Title.ValueString()
	}

	// Set description
	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		domain.Description = model.Description.ValueString()
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

	return domain, nil
}

// xmlToDomainModel converts libvirtxml.Domain to a DomainResourceModel
func xmlToDomainModel(domain *libvirtxml.Domain, model *DomainResourceModel) {
	model.Name = types.StringValue(domain.Name)
	model.Type = types.StringValue(domain.Type)

	if domain.UUID != "" {
		model.UUID = types.StringValue(domain.UUID)
		model.ID = types.StringValue(domain.UUID)
	}

	if domain.Title != "" {
		model.Title = types.StringValue(domain.Title)
	}

	if domain.Description != "" {
		model.Description = types.StringValue(domain.Description)
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

	if domain.VCPU != nil {
		model.VCPU = types.Int64Value(int64(domain.VCPU.Value))
	}

	if domain.OnPoweroff != "" {
		model.OnPoweroff = types.StringValue(domain.OnPoweroff)
	}
	if domain.OnReboot != "" {
		model.OnReboot = types.StringValue(domain.OnReboot)
	}
	if domain.OnCrash != "" {
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
}
