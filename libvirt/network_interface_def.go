package libvirt

import (
	"encoding/xml"
	"github.com/hashicorp/terraform/helper/schema"
)

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

func networkAddressCommonSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"type": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"address": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"prefix": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
	}
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
			Computed: true,
			ForceNew: true,
		},
		"wait_for_lease": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
		},
		"address": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem: &schema.Resource{
				Schema: networkAddressCommonSchema(),
			},
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
