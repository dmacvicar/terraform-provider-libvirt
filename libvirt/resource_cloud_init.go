package libvirt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCloudInit() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudInitCreate,
		Read:   resourceCloudInitRead,
		Delete: resourceCloudInitDelete,
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
				Optional: true,
				ForceNew: true,
			},
			"user_data": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
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

func resourceCloudInitCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] creating cloudinit")
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	cloudInit := newCloudInitDef()
	cloudInit.Metadata.LocalHostname = d.Get("local_hostname").(string)
	cloudInit.UserDataRaw = d.Get("user_data").(string)

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

	// make sure we record the id even if the rest of this gets interrupted
	d.Partial(true) // make sure we record the id even if the rest of this gets interrupted
	d.Set("id", key)
	d.SetPartial("id")
	// TODO: at this point we have collected more things than the ID, so let's save as many things as we can
	d.Partial(false)

	return resourceCloudInitRead(d, meta)
}

func resourceCloudInitRead(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	ci, err := newCloudInitDefFromRemoteISO(virConn, d.Id())
	d.Set("pool", ci.PoolName)
	d.Set("name", ci.Name)
	d.Set("local_hostname", ci.Metadata.LocalHostname)
	d.Set("user_data", ci.UserDataRaw)

	if err != nil {
		return fmt.Errorf("Error while retrieving remote ISO: %s", err)
	}

	if len(ci.UserData.SSHAuthorizedKeys) == 1 {
		d.Set("ssh_authorized_key", ci.UserData.SSHAuthorizedKeys[0])
	}

	return nil
}

func resourceCloudInitDelete(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	key, err := getCloudInitVolumeKeyFromTerraformID(d.Id())
	if err != nil {
		return err
	}

	return RemoveVolume(virConn, key)
}
