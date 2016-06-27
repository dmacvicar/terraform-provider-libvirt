package libvirt

import (
	"encoding/xml"
	"github.com/hashicorp/terraform/helper/schema"
)

// An interface definition, as returned/understood by libvirt
// (see https://libvirt.org/formatdomain.html#elementsNICS)
//
// Something like:
//   <interface type='network'>
//       <source network='default'/>
//   </interface>
//
type defNetworkInterface struct {
	XMLName xml.Name `xml:"interface"`
	Type    string   `xml:"type,attr"`
	Mac     struct {
		Address string `xml:"address,attr"`
	} `xml:"mac"`
	Source struct {
		Network string `xml:"network,attr"`
	} `xml:"source"`
	Model struct {
		Type string `xml:"type,attr"`
	} `xml:"model"`
	waitForLease bool
}

func networkInterfaceCommonSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"network_id": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"hostname": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"mac": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"wait_for_lease": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
		},
		"address": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
	}
}

func newDefNetworkInterface() defNetworkInterface {
	iface := defNetworkInterface{}
	iface.Type = "network"
	//iface.Mac.Address = "52:54:00:36:c0:65"
	iface.Source.Network = "default"
	iface.Model.Type = "virtio"
	iface.waitForLease = false
	return iface
}
