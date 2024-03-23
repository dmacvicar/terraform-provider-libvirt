package libvirt

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// a libvirt volume datasource.
//
// Datasource example:
//
// data "libvirt_volume" "cloudimg" {
//    pool = "image_pool"
//	  name = "ubuntu-22.04-server-cloudimg.img"
// }

func datasourceLibvirtVolume() *schema.Resource {
	return &schema.Resource{
		Read:   resourceLibvirtVolumeRead,
		Exists: resourceLibvirtVolumeExists,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"pool": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
			},
			"source": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"format": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"base_volume_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"base_volume_pool": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
			},
			"base_volume_name": {
				Type:     schema.TypeString,
				Optional: false,
				Required: true,
			},
			"xml": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"xslt": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}
