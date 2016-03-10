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
}

func diskCommonSchema() map[string]*schema.Schema{
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

	disk.Target.Dev = "sda"
	disk.Target.Bus = "virtio"

	return disk
}
