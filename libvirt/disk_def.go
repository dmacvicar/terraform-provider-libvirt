package libvirt

import (
	"encoding/xml"
	"math/rand"
	"time"
)

const OUI = "05abcd"

type defDisk struct {
	XMLName xml.Name `xml:"disk"`
	Type    string   `xml:"type,attr"`
	Device  string   `xml:"device,attr"`
	Wwn     string   `xml:"wwn,omitempty"`
	Format  struct {
		Type string `xml:"type,attr"`
	} `xml:"format"`
	Source struct {
		File   string `xml:"file,attr,omitempty"`
		// retain Pool/Volume for compatibility with existing tfstate
		Pool   string `xml:"pool,attr,omitempty"`
		Volume string `xml:"volume,attr,omitempty"`
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

func newDefDisk() defDisk {
	disk := defDisk{}
	disk.Type = "file"
	disk.Device = "disk"
	disk.Format.Type = "qcow2"
	disk.Target.Bus = "virtio"

	disk.Driver.Name = "qemu"
	disk.Driver.Type = "qcow2"

	return disk
}

func newCDROM() defDisk {
	disk := defDisk{}
	disk.Type = "file"
	disk.Device = "cdrom"
	disk.Target.Dev = "hda"
	disk.Target.Bus = "ide"

	disk.Driver.Name = "qemu"
	disk.Driver.Type = "raw"

	return disk
}

func randomWWN(strlen int) string {
	const chars = "abcdef0123456789"
	rand.Seed(time.Now().UTC().UnixNano())
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return OUI + string(result)
}
