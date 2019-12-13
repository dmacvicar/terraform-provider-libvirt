package ignition

import (
	"encoding/json"

	"github.com/coreos/ignition/config/v2_1/types"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceFilesystem() *schema.Resource {
	return &schema.Resource{
		Exists: resourceFilesystemExists,
		Read:   resourceFilesystemRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"mount": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"format": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"wipe_filesystem": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"label": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"uuid": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"options": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"path": {
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

func resourceFilesystemRead(d *schema.ResourceData, meta interface{}) error {
	id, err := buildFilesystem(d)
	if err != nil {
		return err
	}

	d.SetId(id)
	return nil
}

func resourceFilesystemExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	id, err := buildFilesystem(d)
	if err != nil {
		return false, err
	}

	return id == d.Id(), nil
}

func buildFilesystem(d *schema.ResourceData) (string, error) {
	fs := &types.Filesystem{
		Name: d.Get("name").(string),
	}

	if _, ok := d.GetOk("mount"); ok {
		fs.Mount = &types.Mount{
			Device:         d.Get("mount.0.device").(string),
			Format:         d.Get("mount.0.format").(string),
			WipeFilesystem: d.Get("mount.0.wipe_filesystem").(bool),
		}

		if err := handleReport(fs.Mount.ValidateDevice()); err != nil {
			return "", err
		}

		label, hasLabel := d.GetOk("mount.0.label")
		if hasLabel {
			str := label.(string)
			fs.Mount.Label = &str

			if err := handleReport(fs.Mount.ValidateLabel()); err != nil {
				return "", err
			}
		}

		uuid, hasUUID := d.GetOk("mount.0.uuid")
		if hasUUID {
			str := uuid.(string)
			fs.Mount.UUID = &str
		}

		options, hasOptions := d.GetOk("mount.0.options")
		if hasOptions {
			fs.Mount.Options = castSliceInterfaceToMountOption(options.([]interface{}))
		}
	}

	if p, ok := d.GetOk("path"); ok {
		path := p.(string)
		fs.Path = &path

		if err := handleReport(fs.ValidatePath()); err != nil {
			return "", err
		}
	}

	b, err := json.Marshal(fs)
	if err != nil {
		return "", err
	}
	d.Set("rendered", string(b))

	return hash(string(b)), handleReport(fs.Validate())
}

func castSliceInterfaceToMountOption(i []interface{}) []types.MountOption {
	var o []types.MountOption
	for _, value := range i {
		if value == nil {
			continue
		}

		o = append(o, types.MountOption(value.(string)))
	}

	return o
}
