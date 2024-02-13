package libvirt

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceIgnition() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIgnitionCreate,
		ReadContext:   resourceIgnitionRead,
		DeleteContext: resourceIgnitionDelete,
		Schema: map[string]*schema.Schema{
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
			"content": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceIgnitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] creating ignition file")
	client := meta.(*Client)
	if client.libvirt == nil {
		return diag.Errorf(LibVirtConIsNil)
	}

	ignition := newIgnitionDef()

	ignition.Name = d.Get("name").(string)
	ignition.PoolName = d.Get("pool").(string)
	ignition.Content = d.Get("content").(string)

	log.Printf("[INFO] ignition: %+v", ignition)

	key, err := ignition.CreateAndUpload(client)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(key)

	return resourceIgnitionRead(ctx, d, meta)
}

func resourceIgnitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return diag.Errorf(LibVirtConIsNil)
	}

	ign, err := newIgnitionDefFromRemoteVol(virConn, d.Id())
	d.Set("pool", ign.PoolName)
	d.Set("name", ign.Name)

	if err != nil {
		return diag.Errorf("error while retrieving remote volume: %s", err)
	}

	return nil
}

func resourceIgnitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	if client.libvirt == nil {
		return diag.Errorf(LibVirtConIsNil)
	}

	key, err := getIgnitionVolumeKeyFromTerraformID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(volumeDelete(ctx, client, key))
}
