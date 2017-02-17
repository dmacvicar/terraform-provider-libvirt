package libvirt

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func consoleSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"type": &schema.Schema{
			Type: schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true,
		},
		"source_path": &schema.Schema{
			Type: schema.TypeString,
			Optional: true,
			Required: false,
			ForceNew: true,
		},
		"target_port": &schema.Schema{
			Type: schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true,
		},
		"target_type": &schema.Schema{
			Type: schema.TypeString,
			Optional: true,
			Required: false,
			ForceNew: true,
		},
	}
}
