package libvirt

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func consoleSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"type": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true,
		},
		"source_path": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
			ForceNew: true,
		},
		"target_port": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true,
		},
		"target_type": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
			ForceNew: true,
		},
	}
}
