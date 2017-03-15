package libvirt

import (
	"encoding/xml"
	"fmt"
	"math"
	"strconv"
	"time"

	libvirt "github.com/dmacvicar/libvirt-go"
)

type UnixTimestamp struct{ time.Time }

func (t *UnixTimestamp) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var content string
	if err := d.DecodeElement(&content, &start); err != nil {
		return err
	}

	ts, err := strconv.ParseFloat(content, 64)
	if err != nil {
		return err
	}
	s, ns := math.Modf(ts)
	*t = UnixTimestamp{time.Time(time.Unix(int64(s), int64(ns)))}
	return nil
}

func (t *UnixTimestamp) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	s := t.UTC().Unix()
	ns := t.UTC().UnixNano()
	return e.EncodeElement(fmt.Sprintf("%d.%d", s, ns), start)
}

type defTimestamps struct {
	Change       *UnixTimestamp `xml:"ctime,omitempty"`
	Modification *UnixTimestamp `xml:"mtime,omitempty"`
	Access       *UnixTimestamp `xml:"atime,omitempty"`
}

type defBackingStore struct {
	Path   string `xml:"path"`
	Format struct {
		Type string `xml:"type,attr"`
	} `xml:"format"`
	Timestamps *defTimestamps `xml:"timestamps,omitempty"`
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
		Timestamps *defTimestamps `xml:"timestamps,omitempty"`
	} `xml:"target"`
	Allocation int `xml:"allocation"`
	Capacity   struct {
		Unit   string `xml:"unit,attr"`
		Amount uint64 `xml:"chardata"`
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

// Creates a volume definition from a XML
func newDefVolumeFromXML(s string) (defVolume, error) {
	var volumeDef defVolume
	err := xml.Unmarshal([]byte(s), &volumeDef)
	if err != nil {
		return defVolume{}, err
	}
	return volumeDef, nil
}

func newDefVolumeFromLibvirt(volume *libvirt.VirStorageVol) (defVolume, error) {
	name, err := volume.GetName()
	if err != nil {
		return defVolume{}, fmt.Errorf("could not get name for volume: %s.", err)
	}
	volumeDefXml, err := volume.GetXMLDesc(0)
	if err != nil {
		return defVolume{}, fmt.Errorf("could not get XML description for volume %s: %s.", name, err)
	}
	volumeDef, err := newDefVolumeFromXML(volumeDefXml)
	if err != nil {
		return defVolume{}, fmt.Errorf("could not get a volume definition from XML for %s: %s.", volumeDef.Name, err)
	}
	return volumeDef, nil
}
