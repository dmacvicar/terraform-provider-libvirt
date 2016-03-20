package libvirt

import (
	"encoding/xml"
	"github.com/hashicorp/terraform/helper/schema"
)

type defNetworkInterface struct {
	XMLName xml.Name `xml:"interface"`
	Type    string   `xml:"type,attr"`
	Mac     struct {
		Address string `xml:"address,attr,omitempty"`
	} `xml:"mac,omitempty"`
	Source struct {
		Network string `xml:"network,attr"`
	} `xml:"source"`
	Model struct {
		Type string `xml:"type,attr"`
	} `xml:"model"`
}

func networkInterfaceCommonSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"network": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"mac": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
	}
}

func newDefNetworkInterface() defNetworkInterface {
	iface := defNetworkInterface{}
	iface.Type = "network"
	//iface.Mac.Address = "52:54:00:36:c0:65"
	iface.Source.Network = "default"
	iface.Model.Type = "virtio"
	return iface
}
