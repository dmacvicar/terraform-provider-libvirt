package libvirt

import (
	"encoding/xml"
)

type defDomain struct {
	XMLName  xml.Name  `xml:"domain"`
	Name     string    `xml:"name"`
	Type     string    `xml:"type,attr"`
	Xmlns    string    `xml:"xmlns:qemu,attr"`
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
		Console           []defConsole          `xml:"console"`
		Graphics          struct {
			Type     string `xml:"type,attr"`
			Autoport string `xml:"autoport,attr"`
			Listen   struct {
				Type string `xml:"type,attr"`
			} `xml:"listen"`
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
		Rng struct {
			Model   string `xml:"model,attr"`
			Backend struct {
				Model string `xml:"model,attr"`
			} `xml:"backend"`
		} `xml:"rng"`
	} `xml:"devices"`
	CmdLine struct {
		XMLName xml.Name `xml:"qemu:commandline"`
		Cmd     []defCmd `xml:"qemu:arg"`
	}
}

type defMetadata struct {
	XMLName      xml.Name `xml:"http://github.com/dmacvicar/terraform-provider-libvirt/ user_data"`
	Xml          string   `xml:",cdata"`
	IgnitionFile string   `xml:",ignition_file,omitempty"`
}

type defOs struct {
	Type   defOsType  `xml:"type"`
	Loader *defLoader `xml:"loader,omitempty"`
	NvRam  *defNvRam  `xml:"nvram,omitempty"`
}

type defOsType struct {
	Arch    string `xml:"arch,attr,omitempty"`
	Machine string `xml:"machine,attr,omitempty"`
	Name    string `xml:",chardata"`
}

type defMemory struct {
	Unit   string `xml:"unit,attr"`
	Amount int    `xml:",chardata"`
}

type defVCpu struct {
	Placement string `xml:"unit,attr"`
	Amount    int    `xml:",chardata"`
}

type defCmd struct {
	Value string `xml:"value,attr"`
}

type defLoader struct {
	ReadOnly string `xml:"readonly,attr,omitempty"`
	Type     string `xml:"type,attr,omitempty"`
	File     string `xml:",chardata"`
}

// <nvram>/var/lib/libvirt/qemu/nvram/sled12sp1_VARS.fd</nvram>
type defNvRam struct {
	File string `xml:",chardata"`
}

type defConsole struct {
	Type   string `xml:"type,attr"`
	Source struct {
		Path string `xml:"path,attr,omitempty"`
	} `xml:"source"`
	Target struct {
		Type string `xml:"type,attr,omitempty"`
		Port string `xml:"port,attr"`
	} `xml:"target"`
}

// Creates a domain definition with the defaults
// the provider uses
func newDomainDef() defDomain {
	// libvirt domain definition
	domainDef := defDomain{}
	domainDef.Type = "kvm"
	domainDef.Xmlns = ""

	domainDef.Os = defOs{}
	domainDef.Os.Type = defOsType{}
	domainDef.Os.Type.Name = "hvm"

	domainDef.Memory = defMemory{}
	domainDef.Memory.Unit = "MiB"
	domainDef.Memory.Amount = 512

	domainDef.VCpu = defVCpu{}
	domainDef.VCpu.Placement = "static"
	domainDef.VCpu.Amount = 1

	domainDef.Devices.Graphics.Type = "spice"
	domainDef.Devices.Graphics.Autoport = "yes"
	domainDef.Devices.Graphics.Listen.Type = "none"

	domainDef.Devices.QemuGAChannel.Type = "unix"
	domainDef.Devices.QemuGAChannel.Source.Mode = "bind"
	domainDef.Devices.QemuGAChannel.Target.Type = "virtio"
	domainDef.Devices.QemuGAChannel.Target.Name = "org.qemu.guest_agent.0"

	domainDef.Devices.Rng.Model = "virtio"
	domainDef.Devices.Rng.Backend.Model = "random"

	return domainDef
}

