package libvirt

import (
	"encoding/xml"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/digitalocean/go-libvirt"
	"libvirt.org/go/libvirtxml"
)

func newDefPoolFromLibvirt(virConn *libvirt.Libvirt, pool libvirt.StoragePool) (libvirtxml.StoragePool, diag.Diagnostics) {
	poolDefXML, err := virConn.StoragePoolGetXMLDesc(pool, 0)
	if err != nil {
		return libvirtxml.StoragePool{}, diag.Errorf("could not get XML description for pool %s: %s", pool.Name, err)
	}
	poolDef, err := newDefPoolFromXML(poolDefXML)
	if err != nil {
		return libvirtxml.StoragePool{}, diag.Errorf("could not get a pool definition from XML for %s: %s", pool.Name, err)
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
