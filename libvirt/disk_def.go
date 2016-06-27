package libvirt

import (
	"encoding/xml"
	"github.com/hashicorp/terraform/helper/schema"
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
	Driver struct {
		Name string `xml:"name,attr"`
		Type string `xml:"type,attr"`
	} `xml:"driver"`
}

func diskCommonSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"volume_id": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
	}
}

func newDefDisk() defDisk {
	disk := defDisk{}
	disk.Type = "volume"
	disk.Device = "disk"
	disk.Format.Type = "qcow2"
	disk.Target.Bus = "virtio"

	disk.Driver.Name = "qemu"
	disk.Driver.Type = "qcow2"

	return disk
}

func newCDROM() defDisk {
	disk := defDisk{}
	disk.Type = "volume"
	disk.Device = "cdrom"
	disk.Target.Dev = "hda"
	disk.Target.Bus = "ide"

	disk.Driver.Name = "qemu"
	disk.Driver.Type = "raw"

	return disk
}
