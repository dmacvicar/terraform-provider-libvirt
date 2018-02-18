package libvirt

import (
	"github.com/hashicorp/terraform/helper/mutexkv"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"

	libvirt "github.com/libvirt/libvirt-go"
)

// Global poolMutexKV
var poolMutexKV = mutexkv.NewMutexKV()
var LibvirtClient *libvirt.Connect

// Provider libvirt
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"uri": {
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
			"libvirt_ignition":  resourceIgnition(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		URI: d.Get("uri").(string),
	}

	return config.Client()
}
