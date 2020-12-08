package libvirt

import (
	"encoding/xml"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	libvirt "github.com/libvirt/libvirt-go"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

// from existing domain return its  XMLdefintion
func getXMLDomainDefFromLibvirt(domain *libvirt.Domain) (libvirtxml.Domain, error) {
	domainXMLDesc, err := domain.GetXMLDesc(0)
	if err != nil {
		return libvirtxml.Domain{}, fmt.Errorf("Error retrieving libvirt domain XML description: %s", err)
	}

	domainDef := newDomainDef()
	err = xml.Unmarshal([]byte(domainXMLDesc), &domainDef)
	if err != nil {
		return libvirtxml.Domain{}, fmt.Errorf("Error reading libvirt domain XML description: %s", err)
	}

	return domainDef, nil
}

// note source and target are not initialized
func newFilesystemDef() libvirtxml.DomainFilesystem {
	return libvirtxml.DomainFilesystem{
		AccessMode: "mapped", // A safe default value
		ReadOnly:   &libvirtxml.DomainFilesystemReadOnly{},
	}
}

// Creates a domain definition with the defaults
// the provider uses
func newDomainDef() libvirtxml.Domain {
	domainDef := libvirtxml.Domain{
		OS: &libvirtxml.DomainOS{
			Type: &libvirtxml.DomainOSType{
				Type: "hvm",
			},
		},
		Memory: &libvirtxml.DomainMemory{
			Unit:  "MiB",
			Value: 512,
		},
		VCPU: &libvirtxml.DomainVCPU{
			Placement: "static",
			Value:     1,
		},
		CPU: &libvirtxml.DomainCPU{},
		Devices: &libvirtxml.DomainDeviceList{
			Graphics: []libvirtxml.DomainGraphic{
				{
					Spice: &libvirtxml.DomainGraphicSpice{
						AutoPort: "yes",
					},
				},
			},
			Channels: []libvirtxml.DomainChannel{
				{
					Source: &libvirtxml.DomainChardevSource{
						UNIX: &libvirtxml.DomainChardevSourceUNIX{},
					},
					Target: &libvirtxml.DomainChannelTarget{
						VirtIO: &libvirtxml.DomainChannelTargetVirtIO{
							Name: "org.qemu.guest_agent.0",
						},
					},
				},
			},
		},
		Features: &libvirtxml.DomainFeatureList{
			PAE:  &libvirtxml.DomainFeature{},
			ACPI: &libvirtxml.DomainFeature{},
			APIC: &libvirtxml.DomainFeatureAPIC{},
		},
	}

	if v := os.Getenv("TERRAFORM_LIBVIRT_TEST_DOMAIN_TYPE"); v != "" {
		domainDef.Type = v
	} else {
		domainDef.Type = "kvm"
	}

	// FIXME: We should allow setting this from configuration as well.
	rngDev := os.Getenv("TF_LIBVIRT_RNG_DEV")
	if rngDev == "" {
		rngDev = "/dev/urandom"
	}

	domainDef.Devices.RNGs = []libvirtxml.DomainRNG{
		{
			Model: "virtio",
			Backend: &libvirtxml.DomainRNGBackend{
				Random: &libvirtxml.DomainRNGBackendRandom{Device: rngDev},
			},
		},
	}

	return domainDef
}

func newDomainDefForConnection(virConn *libvirt.Connect, rd *schema.ResourceData) (libvirtxml.Domain, error) {
	d := newDomainDef()

	if arch, ok := rd.GetOk("arch"); ok {
		d.OS.Type.Arch = arch.(string)
	} else {
		arch, err := getHostArchitecture(virConn)
		if err != nil {
			return d, err
		}
		d.OS.Type.Arch = arch
	}

	if d.OS.Type.Arch == "aarch64" {
		// for aarch64 speciffying this will automatically select the firmware and NVRAM file
		// reference: https://libvirt.org/formatdomain.html#bios-bootloader
		d.OS.Firmware = "efi"
	}

	caps, err := getHostCapabilities(virConn)
	if err != nil {
		return d, err
	}
	guest, err := getGuestForArchType(caps, d.OS.Type.Arch, d.OS.Type.Type)
	if err != nil {
		return d, err
	}

	if emulator, ok := rd.GetOk("emulator"); ok {
		d.Devices.Emulator = emulator.(string)
	} else {
		d.Devices.Emulator = guest.Arch.Emulator
	}

	if machine, ok := rd.GetOk("machine"); ok {
		d.OS.Type.Machine = machine.(string)
	} else if len(guest.Arch.Machines) > 0 {
		d.OS.Type.Machine = guest.Arch.Machines[0].Name
	}

	canonicalmachine, err := getCanonicalMachineName(caps, d.OS.Type.Arch, d.OS.Type.Type, d.OS.Type.Machine)
	if err != nil {
		return d, err
	}
	d.OS.Type.Machine = canonicalmachine
	return d, nil
}
