package ignition

import (
	"encoding/json"

	"github.com/coreos/ignition/config/v2_1/types"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceDirectory() *schema.Resource {
	return &schema.Resource{
		Exists: resourceDirectoryExists,
		Read:   resourceDirectoryRead,
		Schema: map[string]*schema.Schema{
			"filesystem": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"path": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mode": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"uid": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"gid": {
				Type:     schema.TypeInt,
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

func resourceDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	id, err := buildDirectory(d)
	if err != nil {
		return err
	}

	d.SetId(id)
	return nil
}

func resourceDirectoryExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	id, err := buildDirectory(d)
	if err != nil {
		return false, err
	}

	return id == d.Id(), nil
}

func buildDirectory(d *schema.ResourceData) (string, error) {
	dir := &types.Directory{}
	dir.Filesystem = d.Get("filesystem").(string)
	if err := handleReport(dir.ValidateFilesystem()); err != nil {
		return "", err
	}

	dir.Path = d.Get("path").(string)
	if err := handleReport(dir.ValidatePath()); err != nil {
		return "", err
	}

	dir.Mode = d.Get("mode").(int)
	if err := handleReport(dir.ValidateMode()); err != nil {
		return "", err
	}

	uid := d.Get("uid").(int)
	if uid != 0 {
		dir.User = types.NodeUser{ID: &uid}
	}

	gid := d.Get("gid").(int)
	if gid != 0 {
		dir.Group = types.NodeGroup{ID: &gid}
	}

	b, err := json.Marshal(dir)
	if err != nil {
		return "", err
	}
	d.Set("rendered", string(b))

	return hash(string(b)), nil
}
