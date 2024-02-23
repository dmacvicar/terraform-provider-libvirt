package libvirt

import (
	"log"
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
			"libvirt_node_info":                        datasourceLibvirtNodeInfo(),
			"libvirt_node_device_info":                 datasourceLibvirtNodeDeviceInfo(),
			"libvirt_node_devices":                     datasourceLibvirtNodeDevices(),
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
	uri := d.Get("uri").(string)
	log.Printf("[DEBUG] configuring provider - default URI is '%v'", uri)
	client := &Client{
		defaultURI: uri,
		connections: make(map[string]*Connection),
	}

	return client, nil
}
