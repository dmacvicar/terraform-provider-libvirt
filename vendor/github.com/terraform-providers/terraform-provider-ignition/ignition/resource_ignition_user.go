package ignition

import (
	"github.com/coreos/ignition/config/v2_1/types"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Exists: resourceUserExists,
		Read:   resourceUserRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"password_hash": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ssh_authorized_keys": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"uid": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"gecos": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"home_dir": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"no_create_home": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"primary_group": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"groups": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"no_user_group": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"no_log_init": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"shell": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"system": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUserRead(d *schema.ResourceData, meta interface{}) error {
	id, err := buildUser(d, globalCache)
	if err != nil {
		return err
	}

	d.SetId(id)
	return nil
}

func resourceUserExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	id, err := buildUser(d, globalCache)
	if err != nil {
		return false, err
	}

	return id == d.Id(), nil
}

func buildUser(d *schema.ResourceData, c *cache) (string, error) {
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

	return c.addUser(&user), handleReport(user.Validate())
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
