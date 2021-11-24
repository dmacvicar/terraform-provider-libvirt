package libvirt

import (
	"fmt"
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
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("LIBVIRT_DEFAULT_URI", nil),
				Description: "libvirt connection URI for operations. See https://libvirt.org/uri.html",
			},
			"host": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "default",
							ForceNew: true,
						},
						"az": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "default",
							ForceNew: true,
						},
						"uri": {
							Type:        schema.TypeString,
							Required:    true,
							DefaultFunc: schema.EnvDefaultFunc("LIBVIRT_DEFAULT_URI", nil),
							Description: "libvirt connection URI for operations. See https://libvirt.org/uri.html",
						},
					},
				},
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
		// TODO: Confirm appropriate IsAlive() validation
		err := client.libvirt.ConnectClose()
		if err != nil {
			log.Printf("[ERROR] cannot close libvirt connection: %v", err)
		}
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	// check for mandatory requirements
	globalURI := d.Get("uri").(string)
	hostCount := d.Get("host.#").(int)
	if globalURI == "" && hostCount == 0 {
		return nil, fmt.Errorf("The libvirt provider must feature either the 'uri' or one to several 'host' parameters")
	}

	// build up configuration object
	config := Config{
		URI: globalURI,
	}
	for i := 0; i < hostCount; i++ {
		prefix := fmt.Sprintf("host.%d", i)
		configHost := ConfigHost{
			Region: d.Get(prefix + ".region").(string),
			AZ:     d.Get(prefix + ".az").(string),
			URI:    d.Get(prefix + ".uri").(string),
		}
		config.Hosts = append(config.Hosts, configHost)
	}
	if false {
		return nil, fmt.Errorf("Config '%v'", config)
	}

	// build up list of client connections
	clients, err := config.Clients()
	if err != nil {
		return nil, err
	}

	return clients, nil
}
