package libvirt

import (
	libvirt "github.com/libvirt/libvirt-go"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
	"os"
)

func newFilesystemDef() libvirtxml.DomainFilesystem {
	return libvirtxml.DomainFilesystem{
		Type:       "mount",  // This is the only type used by qemu/kvm
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
				Type:    "hvm",
				Arch:    "x86_64",
				Machine: "pc",
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
				libvirtxml.DomainGraphic{
					Type:     "spice",
					AutoPort: "yes",
				},
			},
			Channels: []libvirtxml.DomainChannel{
				libvirtxml.DomainChannel{
					Type: "unix",
					Source: &libvirtxml.DomainChardevSource{
						Mode: "bind",
					},
					Target: &libvirtxml.DomainChannelTarget{
						Type: "virtio",
						Name: "org.qemu.guest_agent.0",
					},
				},
			},
			RNGs: []libvirtxml.DomainRNG{
				libvirtxml.DomainRNG{
					Model: "virtio",
					Backend: &libvirtxml.DomainRNGBackend{
						Model: "random",
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

	return domainDef
}

func newDomainDefForConnection(virConn *libvirt.Connect) (libvirtxml.Domain, error) {
	d := newDomainDef()
	caps, err := getHostCapabilities(virConn)
	if err != nil {
		return d, err
	}
	guest, err := getGuestForArchType(caps, d.OS.Type.Arch, d.OS.Type.Type)
	if err != nil {
		return d, err
	}

	d.Devices.Emulator = guest.Arch.Emulator

	canonicalmachine, err := getCanonicalMachineName(caps, d.OS.Type.Arch, d.OS.Type.Type, d.OS.Type.Machine)
	if err != nil {
		return d, err
	}
	d.OS.Type.Machine = canonicalmachine
	return d, nil
}
