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
		Permissions struct {
			Mode int `xml:"mode,omitempty"`
		} `xml:"permissions,omitempty"`
	} `xml:"target"`
	Allocation int `xml:"allocation"`
	Capacity   struct {
		Unit   string `xml:"unit,attr"`
		Amount int64  `xml:"chardata"`
	} `xml:"capacity"`
	BackingStore *defBackingStore `xml:"backingStore,omitempty"`
}

func newDefVolume() defVolume {
	volumeDef := defVolume{}
	volumeDef.Target.Format.Type = "qcow2"
	volumeDef.Target.Permissions.Mode = 644
	volumeDef.Capacity.Unit = "bytes"
	volumeDef.Capacity.Amount = 1
	return volumeDef
}
