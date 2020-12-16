package libvirt

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

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
			"libvirt_domain":         resourceLibvirtDomain(),
			"libvirt_volume":         resourceLibvirtVolume(),
			"libvirt_network":        resourceLibvirtNetwork(),
			"libvirt_pool":           resourceLibvirtPool(),
			"libvirt_cloudinit_disk": resourceCloudInitDisk(),
			"libvirt_ignition":       resourceIgnition(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"libvirt_network_dns_host_template":        datasourceLibvirtNetworkDNSHostTemplate(),
			"libvirt_network_dns_srv_template":         datasourceLibvirtNetworkDNSSRVTemplate(),
			"libvirt_network_dnsmasq_options_template": datasourceLibvirtNetworkDnsmasqOptionsTemplate(),
		},

		ConfigureFunc: providerConfigure,
	}
}

// uri -> client for multi instance support
// (we share the same client for the same uri)
var globalClientMap = make(map[string]*Client)

// CleanupLibvirtConnections closes libvirt clients for all URIs
func CleanupLibvirtConnections() {
	for uri, client := range globalClientMap {
		log.Printf("[DEBUG] cleaning up connection for URI: %s", uri)
		alive, err := client.libvirt.IsAlive()
		if err != nil {
			log.Printf("[ERROR] cannot determine libvirt connection status: %v", err)
		}
		if alive {
			ret, err := client.libvirt.Close()
			if err != nil {
				log.Printf("[ERROR] cannot close libvirt connection %d - %v", ret, err)
			}
		}
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		URI: d.Get("uri").(string),
	}
	log.Printf("[DEBUG] Configuring provider for '%s': %v", config.URI, d)

	if client, ok := globalClientMap[config.URI]; ok {
		log.Printf("[DEBUG] Reusing client for uri: '%s'", config.URI)
		return client, nil
	}

	client, err := config.Client()
	if err != nil {
		return nil, err
	}
	globalClientMap[config.URI] = client

	return client, nil
}
