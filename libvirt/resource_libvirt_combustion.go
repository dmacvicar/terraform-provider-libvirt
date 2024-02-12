package libvirt

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCombustion() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCombustionCreate,
		ReadContext:   resourceCombustionRead,
		DeleteContext: resourceCombustionDelete,
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

func resourceCombustionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] creating combustion file")
	client := meta.(*Client)
	if client.libvirt == nil {
		return diag.Errorf(LibVirtConIsNil)
	}

	combustion := newIgnitionDef()

	combustion.Name = d.Get("name").(string)
	combustion.PoolName = d.Get("pool").(string)
	combustion.Content = d.Get("content").(string)
	combustion.Combustion = true

	log.Printf("[INFO] combustion: %+v", combustion)

	key, err := combustion.CreateAndUpload(client)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(key)

	return resourceIgnitionRead(ctx, d, meta)
}

func resourceCombustionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return diag.Errorf(LibVirtConIsNil)
	}

	combustion, err := newIgnitionDefFromRemoteVol(virConn, d.Id())
	d.Set("pool", combustion.PoolName)
	d.Set("name", combustion.Name)

	if err != nil {
		return diag.Errorf("error while retrieving remote volume: %s", err)
	}

	return nil
}

func resourceCombustionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
