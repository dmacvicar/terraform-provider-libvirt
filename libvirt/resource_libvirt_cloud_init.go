package libvirt

import (
	"context"
	"log"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudInitDisk() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudInitDiskCreate,
		ReadContext:   resourceCloudInitDiskRead,
		DeleteContext: resourceCloudInitDiskDelete,
		Schema: map[string]*schema.Schema{
			"host" : {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"pool": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
				ForceNew: true,
			},
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"meta_data": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"network_config": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceCloudInitDiskCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] creating cloudinit")

	uri := d.Get("host").(string)
	client := meta.(*Client)
	virConn, err := meta.(*Client).Connection(&uri)
	if virConn == nil {
		return diag.Errorf("unable to connect for cloud-init creation: %v", err)
	}

	cloudInit := newCloudInitDef()
	cloudInit.UserData = d.Get("user_data").(string)
	cloudInit.MetaData = d.Get("meta_data").(string)
	cloudInit.NetworkConfig = d.Get("network_config").(string)
	cloudInit.Name = d.Get("name").(string)
	cloudInit.PoolName = d.Get("pool").(string)

	log.Printf("[INFO] cloudInit: %+v", cloudInit)

	iso, err := cloudInit.CreateIso()
	if err != nil {
		return diag.FromErr(err)
	}

	client.poolMutexKV.Lock(cloudInit.PoolName)
	key, err := cloudInit.UploadIso(virConn, iso)
	if err != nil {
		return diag.FromErr(err)
	}
	client.poolMutexKV.Unlock(cloudInit.PoolName)
	d.SetId(key)

	return resourceCloudInitDiskRead(ctx, d, meta)
}

func resourceCloudInitDiskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	uri := d.Get("host").(string)
	virConn, err := meta.(*Client).Connection(&uri)
	if virConn == nil {
		return diag.Errorf("unable to connect for cloud-init read: %v", err)
	}

	ci, err := newCloudInitDefFromRemoteISO(ctx, virConn, d.Id())
	if err != nil {
		if isError(err, libvirt.ErrNoStorageVol) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error while retrieving remote ISO: %s", err)
	}
	d.Set("pool", ci.PoolName)
	d.Set("name", ci.Name)
	d.Set("user_data", ci.UserData)
	d.Set("meta_data", ci.MetaData)
	d.Set("network_config", ci.NetworkConfig)
	return nil
}

func resourceCloudInitDiskDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	uri := d.Get("host").(string)
	virConn, err := meta.(*Client).Connection(&uri)
	if virConn == nil {
		return diag.Errorf("unable to connect for cloud-init deletion: %v", err)
	}

	key, err := getCloudInitVolumeKeyFromTerraformID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	poolName := d.Get("pool").(string)

	client.poolMutexKV.Lock(poolName)
	res := volumeDelete(ctx, virConn, key)
	client.poolMutexKV.Unlock(poolName)

	return diag.FromErr(res)
}
