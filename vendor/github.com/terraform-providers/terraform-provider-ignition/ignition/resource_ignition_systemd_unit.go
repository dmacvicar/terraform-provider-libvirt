package ignition

import (
	"encoding/json"

	"github.com/coreos/ignition/config/v2_1/types"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceSystemdUnit() *schema.Resource {
	return &schema.Resource{
		Exists: resourceSystemdUnitExists,
		Read:   resourceSystemdUnitRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"mask": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"content": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"dropin": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
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
					},
				},
			},
			"rendered": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSystemdUnitRead(d *schema.ResourceData, meta interface{}) error {
	id, err := buildSystemdUnit(d)
	if err != nil {
		return err
	}

	d.SetId(id)
	return nil
}

func resourceSystemdUnitExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	id, err := buildSystemdUnit(d)
	if err != nil {
		return false, err
	}

	return id == d.Id(), nil
}

func buildSystemdUnit(d *schema.ResourceData) (string, error) {
	enabled := d.Get("enabled").(bool)
	unit := &types.Unit{
		Name:     d.Get("name").(string),
		Contents: d.Get("content").(string),
		Enabled:  &enabled,
		Mask:     d.Get("mask").(bool),
	}

	if err := handleReport(unit.ValidateName()); err != nil {
		return "", err
	}

	if err := handleReport(unit.ValidateContents()); err != nil {
		return "", err
	}

	for _, raw := range d.Get("dropin").([]interface{}) {
		value := raw.(map[string]interface{})

		d := types.Dropin{
			Name:     value["name"].(string),
			Contents: value["content"].(string),
		}

		if err := handleReport(d.Validate()); err != nil {
			return "", err
		}

		unit.Dropins = append(unit.Dropins, d)
	}

	b, err := json.Marshal(unit)
	if err != nil {
		return "", err
	}
	d.Set("rendered", string(b))

	return hash(string(b)), nil
}
