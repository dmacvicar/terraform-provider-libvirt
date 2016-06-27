package libvirt

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func resourceCloudInitISO() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudInitISOCreate,
		Read:   resourceCloudInitISORead,
		Delete: resourceCloudInitISODelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"pool": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
				ForceNew: true,
			},
			"local_hostname": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ssh_authorized_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceCloudInitISOCreate(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	cloudInit := newCloudInitDef()
	cloudInit.Metadata.LocalHostname = d.Get("local_hostname").(string)

	if _, ok := d.GetOk("ssh_authorized_key"); ok {
		sshKey := d.Get("ssh_authorized_key").(string)
		cloudInit.UserData.SSHAuthorizedKeys = append(
			cloudInit.UserData.SSHAuthorizedKeys,
			sshKey)
	}

	cloudInit.Name = d.Get("name").(string)
	cloudInit.PoolName = d.Get("pool").(string)

	log.Printf("[INFO] cloudInit: %+v", cloudInit)

	key, err := cloudInit.CreateAndUpload(virConn)
	if err != nil {
		return err
	}
	d.SetId(key)

	return resourceCloudInitISORead(d, meta)
}

func resourceCloudInitISORead(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	ci, err := newCloudInitDefFromRemoteISO(virConn, d.Id())
	d.Set("pool", ci.PoolName)
	d.Set("name", ci.Name)
	d.Set("local_hostname", ci.Metadata.LocalHostname)

	if err != nil {
		return fmt.Errorf("Error while retrieving remote ISO: %s", err)
	}

	if len(ci.UserData.SSHAuthorizedKeys) == 1 {
		d.Set("ssh_authorized_key", ci.UserData.SSHAuthorizedKeys[0])
	}

	return nil
}

func resourceCloudInitISODelete(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	return RemoveVolume(virConn, d.Id())
}
