package libvirt

import (
	"encoding/xml"
	"fmt"
	libvirt "github.com/libvirt/libvirt-go"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

func newDefPoolFromLibvirt(pool *libvirt.StoragePool) (libvirtxml.StoragePool, error) {
	name, err := pool.GetName()
	if err != nil {
		return libvirtxml.StoragePool{}, fmt.Errorf("could not get name for pool: %s", err)
	}
	poolDefXML, err := pool.GetXMLDesc(0)
	if err != nil {
		return libvirtxml.StoragePool{}, fmt.Errorf("could not get XML description for pool %s: %s", name, err)
	}
	poolDef, err := newDefPoolFromXML(poolDefXML)
	if err != nil {
		return libvirtxml.StoragePool{}, fmt.Errorf("could not get a pool definition from XML for %s: %s", name, err)
	}
	return poolDef, nil
}

func newDefPoolFromXML(s string) (libvirtxml.StoragePool, error) {
	var poolDef libvirtxml.StoragePool
	err := xml.Unmarshal([]byte(s), &poolDef)
	if err != nil {
		return libvirtxml.StoragePool{}, err
	}
	return poolDef, nil
}
