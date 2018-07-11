package ignition

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/coreos/ignition/config/v2_1/types"
)

var configReferenceResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"source": &schema.Schema{
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"verification": &schema.Schema{
			Type:     schema.TypeString,
			ForceNew: true,
			Optional: true,
		},
	},
}

func resourceConfig() *schema.Resource {
	return &schema.Resource{
		Exists: resourceIgnitionFileExists,
		Read:   resourceIgnitionFileRead,
		Schema: map[string]*schema.Schema{
			"disks": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arrays": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"filesystems": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"files": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"directories": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"links": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"systemd": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"networkd": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"users": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"groups": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"replace": &schema.Schema{
				Type:     schema.TypeList,
				ForceNew: true,
				Optional: true,
				MaxItems: 1,
				Elem:     configReferenceResource,
			},
			"append": &schema.Schema{
				Type:     schema.TypeList,
				ForceNew: true,
				Optional: true,
				Elem:     configReferenceResource,
			},
			"rendered": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceIgnitionFileRead(d *schema.ResourceData, meta interface{}) error {
	rendered, err := renderConfig(d, globalCache)
	if err != nil {
		return err
	}

	if err := d.Set("rendered", rendered); err != nil {
		return err
	}

	d.SetId(hash(rendered))
	return nil
}

func resourceIgnitionFileExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	rendered, err := renderConfig(d, globalCache)
	if err != nil {
		return false, err
	}

	return hash(rendered) == d.Id(), nil
}

func renderConfig(d *schema.ResourceData, c *cache) (string, error) {
	i, err := buildConfig(d, c)
	if err != nil {
		return "", err
	}

	bytes, err := json.Marshal(i)

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func buildConfig(d *schema.ResourceData, c *cache) (*types.Config, error) {
	var err error
	config := &types.Config{}
	config.Ignition, err = buildIgnition(d)
	if err != nil {
		return nil, err
	}

	config.Storage, err = buildStorage(d, c)
	if err != nil {
		return nil, err
	}

	config.Systemd, err = buildSystemd(d, c)
	if err != nil {
		return nil, err
	}

	config.Networkd, err = buildNetworkd(d, c)
	if err != nil {
		return nil, err
	}

	config.Passwd, err = buildPasswd(d, c)
	if err != nil {
		return nil, err
	}

	return config, handleReport(config.Validate())
}

func buildIgnition(d *schema.ResourceData) (types.Ignition, error) {
	var err error

	i := types.Ignition{}
	i.Version = types.MaxVersion.String()

	rr := d.Get("replace.0").(map[string]interface{})
	if len(rr) != 0 {
		i.Config.Replace, err = buildConfigReference(rr)
		if err != nil {
			return i, err
		}
	}

	ar := d.Get("append").([]interface{})
	if len(ar) != 0 {
		for _, rr := range ar {
			r, err := buildConfigReference(rr.(map[string]interface{}))
			if err != nil {
				return i, err
			}

			i.Config.Append = append(i.Config.Append, *r)
		}
	}

	return i, nil
}

func buildConfigReference(raw map[string]interface{}) (*types.ConfigReference, error) {
	r := &types.ConfigReference{}
	r.Source = raw["source"].(string)

	hash := raw["verification"].(string)
	if hash != "" {
		r.Verification.Hash = &hash
	}

	return r, nil
}

func buildStorage(d *schema.ResourceData, c *cache) (types.Storage, error) {
	storage := types.Storage{}

	for _, id := range d.Get("disks").([]interface{}) {
		if id == nil {
			continue
		}
		d, ok := c.disks[id.(string)]
		if !ok {
			return storage, fmt.Errorf("invalid disk %q, unknown disk id", id)
		}

		storage.Disks = append(storage.Disks, *d)
	}

	for _, id := range d.Get("arrays").([]interface{}) {
		if id == nil {
			continue
		}
		a, ok := c.arrays[id.(string)]
		if !ok {
			return storage, fmt.Errorf("invalid raid %q, unknown raid id", id)
		}

		storage.Raid = append(storage.Raid, *a)
	}

	for _, id := range d.Get("filesystems").([]interface{}) {
		if id == nil {
			continue
		}
		f, ok := c.filesystems[id.(string)]
		if !ok {
			return storage, fmt.Errorf("invalid filesystem %q, unknown filesystem id", id)
		}

		storage.Filesystems = append(storage.Filesystems, *f)
	}

	for _, id := range d.Get("files").([]interface{}) {
		if id == nil {
			continue
		}
		f, ok := c.files[id.(string)]
		if !ok {
			return storage, fmt.Errorf("invalid file %q, unknown file id", id)
		}

		storage.Files = append(storage.Files, *f)
	}

	for _, id := range d.Get("directories").([]interface{}) {
		if id == nil {
			continue
		}
		f, ok := c.directories[id.(string)]
		if !ok {
			return storage, fmt.Errorf("invalid file %q, unknown directory id", id)
		}

		storage.Directories = append(storage.Directories, *f)
	}

	for _, id := range d.Get("links").([]interface{}) {
		if id == nil {
			continue
		}
		f, ok := c.links[id.(string)]
		if !ok {
			return storage, fmt.Errorf("invalid file %q, unknown link id", id)
		}

		storage.Links = append(storage.Links, *f)
	}

	return storage, nil

}

func buildSystemd(d *schema.ResourceData, c *cache) (types.Systemd, error) {
	systemd := types.Systemd{}

	for _, id := range d.Get("systemd").([]interface{}) {
		if id == nil {
			continue
		}

		u, ok := c.systemdUnits[id.(string)]
		if !ok {
			return systemd, fmt.Errorf("invalid systemd unit %q, unknown systemd unit id", id)
		}

		systemd.Units = append(systemd.Units, *u)
	}

	return systemd, nil

}

func buildNetworkd(d *schema.ResourceData, c *cache) (types.Networkd, error) {
	networkd := types.Networkd{}

	for _, id := range d.Get("networkd").([]interface{}) {
		if id == nil {
			continue
		}

		u, ok := c.networkdUnits[id.(string)]
		if !ok {
			return networkd, fmt.Errorf("invalid networkd unit %q, unknown networkd unit id", id)
		}

		networkd.Units = append(networkd.Units, *u)
	}

	return networkd, nil
}

func buildPasswd(d *schema.ResourceData, c *cache) (types.Passwd, error) {
	passwd := types.Passwd{}

	for _, id := range d.Get("users").([]interface{}) {
		if id == nil {
			continue
		}

		u, ok := c.users[id.(string)]
		if !ok {
			return passwd, fmt.Errorf("invalid user %q, unknown user id", id)
		}

		passwd.Users = append(passwd.Users, *u)
	}

	for _, id := range d.Get("groups").([]interface{}) {
		if id == nil {
			continue
		}

		g, ok := c.groups[id.(string)]
		if !ok {
			return passwd, fmt.Errorf("invalid group %q, unknown group id", id)
		}

		passwd.Groups = append(passwd.Groups, *g)
	}

	return passwd, nil

}
