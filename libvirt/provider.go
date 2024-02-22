package libvirt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider libvirt.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"uri": {
				Type:        schema.TypeString,
				Optional:    true,
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
			"libvirt_nodeinfo":                         datasourceLibvirtNodeInfo(),
			"libvirt_network_dns_host_template":        datasourceLibvirtNetworkDNSHostTemplate(),
			"libvirt_network_dns_srv_template":         datasourceLibvirtNetworkDNSSRVTemplate(),
			"libvirt_network_dnsmasq_options_template": datasourceLibvirtNetworkDnsmasqOptionsTemplate(),
		},

		ConfigureFunc: providerConfigure,
	}
}

// maintain a list of all providers so that we can clean their resources up on exit
var globalProviderList = []*Client {}

// CleanupLibvirtConnections closes libvirt clients for all URIs.
func CleanupLibvirtConnections() {
	for _, provider := range globalProviderList {
		provider.Close()
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	// don't do much of anything since we connect on demand and to potentially multiple targets
	client := &Client{
		defaultURI: d.Get("uri").(string),
		connections: make(map[string]*Connection),
	}

	return client, nil
}
