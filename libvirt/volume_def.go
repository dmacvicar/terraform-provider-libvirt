package libvirt

import (
	"encoding/xml"
)

type defBackingStore struct {
	Path   string `xml:"path"`
	Format struct {
		Type string `xml:"type,attr"`
	} `xml:"format"`
}

type defVolume struct {
	XMLName xml.Name `xml:"volume"`
	Name    string   `xml:"name"`
	Target  struct {
		Format struct {
			Type string `xml:"type,attr"`
		} `xml:"format"`
	} `xml:"target"`
	Allocation int `xml:"allocation"`
	Capacity struct {
		Unit string `xml:"unit,attr"`
		Amount int `xml:"chardata"`
	} `xml:"capacity"`
	BackingStore *defBackingStore `xml:"backingStore,omitempty"`
}

func newDefVolume() defVolume {
	volumeDef := defVolume{}
	volumeDef.Target.Format.Type = "qcow2"
	volumeDef.Capacity.Unit = "GB"
	volumeDef.Capacity.Amount = 1
	return volumeDef
}
