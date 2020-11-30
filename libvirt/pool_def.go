package libvirt

import (
	"encoding/xml"
	"fmt"
	libvirt "github.com/digitalocean/go-libvirt"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

func newDefPoolFromLibvirt(virConn *libvirt.Libvirt, pool libvirt.StoragePool) (libvirtxml.StoragePool, error) {
	poolDefXML, err := virConn.StoragePoolGetXMLDesc(pool, 0)
	if err != nil {
		return libvirtxml.StoragePool{}, fmt.Errorf("could not get XML description for pool %s: %w", pool.Name, err)
	}
	poolDef, err := newDefPoolFromXML(poolDefXML)
	if err != nil {
		return libvirtxml.StoragePool{}, fmt.Errorf("could not get a pool definition from XML for %s: %w", pool.Name, err)
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
