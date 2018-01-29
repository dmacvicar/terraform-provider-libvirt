package libvirt

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func networkInterfaceCommonSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"network_id": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
			Computed: true,
		},
		"network_name": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
			Computed: true,
		},
		"bridge": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"vepa": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"macvtap": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"passthrough": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"hostname": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: false,
		},
		"mac": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"wait_for_lease": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"addresses": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			ForceNew: false,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}
