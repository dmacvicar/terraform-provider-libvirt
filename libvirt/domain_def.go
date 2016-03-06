package libvirt

import (
	"encoding/xml"
)

type defDisk struct {
	XMLName xml.Name `xml:"disk"`
	Type    string   `xml:"type,attr"`
	Device  string   `xml:"device,attr"`
	Format  struct {
		Type string `xml:"type,attr"`
	} `xml:"format"`
	Source struct {
		Pool   string `xml:"pool,attr"`
		Volume string `xml:"volume,attr"`
	} `xml:"source"`
	Target struct {
		Dev string `xml:"dev,attr"`
		Bus string `xml:"bus,attr"`
	} `xml:"target"`
}

type defDomain struct {
	XMLName xml.Name  `xml:"domain"`
	Name    string    `xml:"name"`
	Type    string    `xml:"type,attr"`
	Os      defOs     `xml:"os"`
	Memory  defMemory `xml:"memory"`
	VCpu    defVCpu   `xml:"vcpu"`
	Devices struct {
		RootDisk defDisk `xml:"disk"`
		Spice    struct {
			Type     string `xml:"type,attr"`
			Autoport string `xml:"autoport,attr"`
		} `xml:"graphics"`
	} `xml:"devices"`
}

type defOs struct {
	Type defOsType `xml:"type"`
}

type defOsType struct {
	Arch    string `xml:"arch,attr"`
	Machine string `xml:"machine,attr"`
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
	domainDef.Os.Type.Arch = "x86_64"
	domainDef.Os.Type.Machine = "pc-i440fx-2.4"
	domainDef.Os.Type.Name = "hvm"

	domainDef.Memory = defMemory{}
	domainDef.Memory.Unit = "MiB"
	domainDef.Memory.Amount = 512

	domainDef.VCpu = defVCpu{}
	domainDef.VCpu.Placement = "static"
	domainDef.VCpu.Amount = 1

	domainDef.Devices.Spice.Type = "spice"
	domainDef.Devices.Spice.Autoport = "yes"

	return domainDef
}
