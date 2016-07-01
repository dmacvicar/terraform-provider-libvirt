package libvirt

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"uri": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LIBVIRT_DEFAULT_URI", nil),
				Description: "libvirt connection URI for operations. See https://libvirt.org/uri.html",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"libvirt_domain":    resourceLibvirtDomain(),
			"libvirt_volume":    resourceLibvirtVolume(),
			"libvirt_network":   resourceLibvirtNetwork(),
			"libvirt_cloudinit": resourceCloudInit(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Uri: d.Get("uri").(string),
	}

	return config.Client()
}
