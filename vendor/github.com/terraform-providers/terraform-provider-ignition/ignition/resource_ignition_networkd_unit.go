package ignition

import (
	"encoding/json"

	"github.com/coreos/ignition/config/v2_1/types"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceNetworkdUnit() *schema.Resource {
	return &schema.Resource{
		Exists: resourceNetworkdUnitExists,
		Read:   resourceNetworkdUnitRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"content": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"rendered": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceNetworkdUnitRead(d *schema.ResourceData, meta interface{}) error {
	id, err := buildNetworkdUnit(d)
	if err != nil {
		return err
	}

	d.SetId(id)
	return nil
}

func resourceNetworkdUnitExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	id, err := buildNetworkdUnit(d)
	if err != nil {
		return false, err
	}

	return id == d.Id(), nil
}

func buildNetworkdUnit(d *schema.ResourceData) (string, error) {
	unit := &types.Networkdunit{
		Name:     d.Get("name").(string),
		Contents: d.Get("content").(string),
	}

	b, err := json.Marshal(unit)
	if err != nil {
		return "", err
	}
	d.Set("rendered", string(b))

	return hash(string(b)), handleReport(unit.Validate())
}
