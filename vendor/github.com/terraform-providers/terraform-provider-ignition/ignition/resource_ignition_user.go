package ignition

import (
	"encoding/json"

	"github.com/coreos/ignition/config/v2_1/types"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		Exists: resourceUserExists,
		Read:   resourceUserRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"password_hash": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ssh_authorized_keys": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"uid": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"gecos": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"home_dir": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"no_create_home": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"primary_group": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"groups": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"no_user_group": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"no_log_init": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"shell": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"system": {
				Type:     schema.TypeBool,
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

func resourceUserRead(d *schema.ResourceData, meta interface{}) error {
	id, err := buildUser(d)
	if err != nil {
		return err
	}

	d.SetId(id)
	return nil
}

func resourceUserExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	id, err := buildUser(d)
	if err != nil {
		return false, err
	}

	return id == d.Id(), nil
}

func buildUser(d *schema.ResourceData) (string, error) {
	user := types.PasswdUser{
		Name:         d.Get("name").(string),
		UID:          getInt(d, "uid"),
		Gecos:        d.Get("gecos").(string),
		HomeDir:      d.Get("home_dir").(string),
		NoCreateHome: d.Get("no_create_home").(bool),
		PrimaryGroup: d.Get("primary_group").(string),
		Groups:       castSliceInterfaceToPasswdUserGroup(d.Get("groups").([]interface{})),
		NoUserGroup:  d.Get("no_user_group").(bool),
		NoLogInit:    d.Get("no_log_init").(bool),
		Shell:        d.Get("shell").(string),
		System:       d.Get("system").(bool),
		SSHAuthorizedKeys: castSliceInterfaceToSSHAuthorizedKey(
			d.Get("ssh_authorized_keys").([]interface{}),
		),
	}

	pwd := d.Get("password_hash").(string)
	if pwd != "" {
		user.PasswordHash = &pwd
	}

	b, err := json.Marshal(user)
	if err != nil {
		return "", err
	}
	d.Set("rendered", string(b))

	return hash(string(b)), handleReport(user.Validate())
}

func castSliceInterfaceToPasswdUserGroup(i []interface{}) []types.PasswdUserGroup {
	var res []types.PasswdUserGroup
	for _, g := range i {
		if g == nil {
			continue
		}

		res = append(res, types.PasswdUserGroup(g.(string)))
	}
	return res
}

func castSliceInterfaceToSSHAuthorizedKey(i []interface{}) []types.SSHAuthorizedKey {
	var res []types.SSHAuthorizedKey
	for _, k := range i {
		if k == nil {
			continue
		}

		res = append(res, types.SSHAuthorizedKey(k.(string)))
	}
	return res
}
