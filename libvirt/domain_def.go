package libvirt

import (
	"encoding/xml"
)

type defDomain struct {
	XMLName  xml.Name  `xml:"domain"`
	Name     string    `xml:"name"`
	Type     string    `xml:"type,attr"`
	Os       defOs     `xml:"os"`
	Memory   defMemory `xml:"memory"`
	VCpu     defVCpu   `xml:"vcpu"`
	Metadata struct {
		XMLName          xml.Name `xml:"metadata"`
		TerraformLibvirt defMetadata
	}
	Features struct {
		Acpi string `xml:"acpi"`
		Apic string `xml:"apic"`
		Pae  string `xml:"pae"`
	} `xml:"features"`
	Devices struct {
		Disks             []defDisk             `xml:"disk"`
		NetworkInterfaces []defNetworkInterface `xml:"interface"`
		Spice             struct {
			Type     string `xml:"type,attr"`
			Autoport string `xml:"autoport,attr"`
		} `xml:"graphics"`
		// QEMU guest agent channel
		QemuGAChannel struct {
			Type   string `xml:"type,attr"`
			Source struct {
				Mode string `xml:"mode,attr"`
			} `xml:"source"`
			Target struct {
				Type string `xml:"type,attr"`
				Name string `xml:"name,attr"`
			} `xml:"target"`
		} `xml:"channel"`
	} `xml:"devices"`
}

type defMetadata struct {
	XMLName xml.Name `xml:"http://github.com/dmacvicar/terraform-provider-libvirt/ user_data"`
	Xml     string   `xml:",cdata"`
}

type defOs struct {
	Type defOsType `xml:"type"`
}

type defOsType struct {
	Arch    string `xml:"arch,attr,omitempty"`
	Machine string `xml:"machine,attr,omitempty"`
	Name    string `xml:"chardata"`
}

type defMemory struct {
	Unit   string `xml:"unit,attr"`
	Amount int    `xml:"chardata"`
}

type defVCpu struct {
	Placement string `xml:"unit,attr"`
	Amount    int    `xml:"chardata"`
}

// Creates a domain definition with the defaults
// the provider uses
func newDomainDef() defDomain {
	// libvirt domain definition
	domainDef := defDomain{}
	domainDef.Type = "kvm"

	domainDef.Os = defOs{}
	domainDef.Os.Type = defOsType{}
	domainDef.Os.Type.Name = "hvm"

	domainDef.Memory = defMemory{}
	domainDef.Memory.Unit = "MiB"
	domainDef.Memory.Amount = 512

	domainDef.VCpu = defVCpu{}
	domainDef.VCpu.Placement = "static"
	domainDef.VCpu.Amount = 1

	domainDef.Devices.Spice.Type = "spice"
	domainDef.Devices.Spice.Autoport = "yes"

	domainDef.Devices.QemuGAChannel.Type = "unix"
	domainDef.Devices.QemuGAChannel.Source.Mode = "bind"
	domainDef.Devices.QemuGAChannel.Target.Type = "virtio"
	domainDef.Devices.QemuGAChannel.Target.Name = "org.qemu.guest_agent.0"

	return domainDef
}
