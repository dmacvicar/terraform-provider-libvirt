package libvirt

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func networkInterfaceCommonSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"network_id": &schema.Schema{
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			Computed:      true,
		},
		"network_name": &schema.Schema{
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			Computed:      true,
		},
		"bridge": &schema.Schema{
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
		},
		"vepa": &schema.Schema{
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
		},
		"macvtap": &schema.Schema{
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			Computed:      true,
		},
		"passthrough": &schema.Schema{
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
		},
		"hostname": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: false,
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
		"addresses": &schema.Schema{
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
