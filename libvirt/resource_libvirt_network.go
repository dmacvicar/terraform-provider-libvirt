package libvirt

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
)

const (
	netModeIsolated = "none"
	netModeNat      = "nat"
	netModeRoute    = "route"
	netModeBridge   = "bridge"
	dnsPrefix       = "dns.0"
)

// a libvirt network resource
//
// Resource example:
//
// resource "libvirt_network" "k8snet" {
//    name = "k8snet"
//    domain = "k8s.local"
//    mode = "nat"
//    addresses = ["10.17.3.0/24"]
// }
//
// "addresses" can contain (0 or 1) ipv4 and (0 or 1) ipv6 subnets
// "mode" can be one of: "nat" (default), "isolated"
//
func resourceLibvirtNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceLibvirtNetworkCreate,
		Read:   resourceLibvirtNetworkRead,
		Delete: resourceLibvirtNetworkDelete,
		Exists: resourceLibvirtNetworkExists,
		Update: resourceLibvirtNetworkUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"mode": { // can be "none", "nat" (default), "route", "bridge"
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  netModeNat,
			},
			"bridge": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: false,
			},
			"mtu": {
				Type:     schema.TypeInt,
				Optional: true,
				Required: false,
			},
			"addresses": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"autostart": {
				Type:     schema.TypeBool,
				Optional: true,
				Required: false,
			},
			"dns": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Default:  true,
							Optional: true,
							Required: false,
						},
						"local_only": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
							Required: false,
						},
						"forwarders": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"address": {
										Type:     schema.TypeString,
										Optional: true,
										Required: false,
										ForceNew: true,
									},
									"domain": {
										Type:     schema.TypeString,
										Optional: true,
										Required: false,
										ForceNew: true,
									},
								},
							},
						},
						"srvs": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"service": {
										Type: schema.TypeString,
										// This should be required, but Terraform does validation too early
										// and therefore doesn't recognize that this is set when assigning from
										// a rendered dns_host template.
										Optional: true,
										ForceNew: true,
									},
									"protocol": {
										Type: schema.TypeString,
										// This should be required, but Terraform does validation too early
										// and therefore doesn't recognize that this is set when assigning from
										// a rendered dns_host template.
										Optional: true,
										ForceNew: true,
									},
									"domain": {
										Type:     schema.TypeString,
										Optional: true,
										Required: false,
										ForceNew: true,
									},
									"target": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"port": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"weight": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"priority": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
						"hosts": {
							Type:     schema.TypeList,
							ForceNew: false,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ip": {
										Type: schema.TypeString,
										// This should be required, but Terraform does validation too early
										// and therefore doesn't recognize that this is set when assigning from
										// a rendered dns_host template.
										Optional: true,
									},
									"hostname": {
										Type: schema.TypeString,
										// This should be required, but Terraform does validation too early
										// and therefore doesn't recognize that this is set when assigning from
										// a rendered dns_host template.
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"dhcp": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Default:  true,
							Optional: true,
							Required: false,
						},
					},
				},
			},
			"routes": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr": {
							Type:     schema.TypeString,
							Required: true,
						},
						"gateway": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"xml": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"xslt": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
		},
	}
}

func resourceLibvirtNetworkExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return false, fmt.Errorf(LibVirtConIsNil)
	}
	network, err := virConn.LookupNetworkByUUIDString(d.Id())
	if err != nil {
		// If the network couldn't be found, don't return an error otherwise
		// Terraform won't create it again.
		if lverr, ok := err.(libvirt.Error); ok && lverr.Code == libvirt.ERR_NO_NETWORK {
			return false, nil
		}
		return false, err
	}
	defer network.Free()

	return err == nil, err
}

// resourceLibvirtNetworkUpdate updates dynamically some attributes in the network
func resourceLibvirtNetworkUpdate(d *schema.ResourceData, meta interface{}) error {
	// check the list of things that can be changed dynamically
	// in https://wiki.libvirt.org/page/Networking#virsh_net-update
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}
	network, err := virConn.LookupNetworkByUUIDString(d.Id())
	if err != nil {
		return fmt.Errorf("Can't retrieve network with ID '%s' during update: %s", d.Id(), err)
	}
	defer network.Free()

	d.Partial(true)

	networkName, err := network.GetName()
	if err != nil {
		return err
	}

	active, err := network.IsActive()
	if err != nil {
		return fmt.Errorf("Error when getting network %s status during update: %s", networkName, err)
	}

	if !active {
		log.Printf("[DEBUG] Activating network %s", networkName)
		if err := network.Create(); err != nil {
			return fmt.Errorf("Error when activating network %s during update: %s", networkName, err)
		}
	}

	if d.HasChange("autostart") {
		err = network.SetAutostart(d.Get("autostart").(bool))
		if err != nil {
			return fmt.Errorf("Error updating autostart for network %s: %s", networkName, err)
		}
		d.SetPartial("autostart")
	}

	// detect changes in the DNS entries in this network
	err = updateDNSHosts(d, network)
	if err != nil {
		return fmt.Errorf("Error updating DNS hosts for network %s: %s", networkName, err)
	}

	// detect changes in the bridge
	if d.HasChange("bridge") {
		networkBridge := getBridgeFromResource(d)

		data, err := xmlMarshallIndented(networkBridge)
		if err != nil {
			return fmt.Errorf("Error serializing update for network %s: %s", networkName, err)
		}

		log.Printf("[DEBUG] Updating bridge for libvirt network '%s' with XML: %s", networkName, networkBridge.Name)
		err = network.Update(libvirt.NETWORK_UPDATE_COMMAND_MODIFY, libvirt.NETWORK_SECTION_BRIDGE, -1,
			data, libvirt.NETWORK_UPDATE_AFFECT_LIVE|libvirt.NETWORK_UPDATE_AFFECT_CONFIG)
		if err != nil {
			return fmt.Errorf("Error when updating bridge in %s: %s", networkName, err)
		}

		d.SetPartial("bridge")
	}

	// detect changes in the domain
	if d.HasChange("domain") {
		networkDomain := getDomainFromResource(d)

		data, err := xmlMarshallIndented(networkDomain)
		if err != nil {
			return fmt.Errorf("serialize update: %s", err)
		}

		log.Printf("[DEBUG] Updating domain for libvirt network '%s' with XML: %s", networkName, data)
		err = network.Update(libvirt.NETWORK_UPDATE_COMMAND_MODIFY, libvirt.NETWORK_SECTION_DOMAIN, -1,
			data, libvirt.NETWORK_UPDATE_AFFECT_LIVE|libvirt.NETWORK_UPDATE_AFFECT_CONFIG)
		if err != nil {
			return fmt.Errorf("Error when updating domain in %s: %s", networkName, err)
		}

		d.SetPartial("domain")
	}

	d.Partial(false)
	return nil
}

// resourceLibvirtNetworkCreate creates a libvirt network from the resource definition
func resourceLibvirtNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	// see https://libvirt.org/formatnetwork.html
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	networkDef := newNetworkDef()
	networkDef.Name = d.Get("name").(string)
	networkDef.Domain = getDomainFromResource(d)

	// use a bridge provided by the user, or create one otherwise (libvirt will assign on automatically when empty)
	networkDef.Bridge = getBridgeFromResource(d)

	networkDef.MTU = getMTUFromResource(d)

	// check the network mode
	networkDef.Forward = &libvirtxml.NetworkForward{
		Mode: getNetModeFromResource(d),
	}
	if networkDef.Forward.Mode == netModeIsolated || networkDef.Forward.Mode == netModeNat || networkDef.Forward.Mode == netModeRoute {

		if networkDef.Forward.Mode == netModeIsolated {
			// there is no forwarding when using an isolated network
			networkDef.Forward = nil
		} else if networkDef.Forward.Mode == netModeRoute {
			// there is no NAT when using a routed network
			networkDef.Forward.NAT = nil
		}

		// if addresses are given set dhcp for these
		ips, err := getIPsFromResource(d)
		if err != nil {
			return fmt.Errorf("Could not set DHCP from adresses '%s'", err)
		}
		networkDef.IPs = ips

		dnsEnabled, err := getDNSEnableFromResource(d)
		if err != nil {
			return err
		}

		dnsForwarders, err := getDNSForwardersFromResource(d)
		if err != nil {
			return err
		}

		dnsSRVs, err := getDNSSRVFromResource(d)
		if err != nil {
			return err
		}

		dnsHosts, err := getDNSHostsFromResource(d)
		if err != nil {
			return err
		}

		dns := libvirtxml.NetworkDNS{
			Enable:     dnsEnabled,
			Forwarders: dnsForwarders,
			Host:       dnsHosts,
			SRVs:       dnsSRVs,
		}
		networkDef.DNS = &dns

	} else if networkDef.Forward.Mode == netModeBridge {
		if networkDef.Bridge.Name == "" {
			return fmt.Errorf("'bridge' must be provided when using the bridged network mode")
		}
		networkDef.Bridge.STP = ""
	} else {
		return fmt.Errorf("unsupported network mode '%s'", networkDef.Forward.Mode)
	}

	// parse any static routes
	routes, err := getRoutesFromResource(d)
	if err != nil {
		return err
	}
	networkDef.Routes = routes

	// once we have the network defined, connect to libvirt and create it from the XML serialization
	connectURI, err := virConn.GetURI()
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt connection URI: %s", err)
	}
	log.Printf("[INFO] Creating libvirt network at %s", connectURI)

	data, err := xmlMarshallIndented(networkDef)
	if err != nil {
		return fmt.Errorf("Error serializing libvirt network: %s", err)
	}
	log.Printf("[DEBUG] Generated XML for libvirt network:\n%s", data)

	data, err = transformResourceXML(data, d)
	if err != nil {
		return fmt.Errorf("Error applying XSLT stylesheet: %s", err)
	}

	log.Printf("[DEBUG] Creating libvirt network at %s: %s", connectURI, data)
	network, err := virConn.NetworkDefineXML(data)
	if err != nil {
		return fmt.Errorf("Error defining libvirt network: %s - %s", err, data)
	}
	err = network.Create()
	if err != nil {
		return fmt.Errorf("Error clearing libvirt network: %s", err)
	}
	defer network.Free()

	id, err := network.GetUUIDString()
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt network id: %s", err)
	}
	d.SetId(id)

	// make sure we record the id even if the rest of this gets interrupted
	d.Partial(true)
	d.Set("id", id)
	d.SetPartial("id")
	d.Partial(false)

	log.Printf("[INFO] Created network %s [%s]", networkDef.Name, d.Id())

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"BUILD"},
		Target:     []string{"ACTIVE"},
		Refresh:    waitForNetworkActive(*network),
		Timeout:    1 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for network to reach ACTIVE state: %s", err)
	}

	if autostart, ok := d.GetOk("autostart"); ok {
		err = network.SetAutostart(autostart.(bool))
		if err != nil {
			return fmt.Errorf("Error setting autostart for network: %s", err)
		}
	}

	return resourceLibvirtNetworkRead(d, meta)
}

// resourceLibvirtNetworkRead gets the current resource from libvirt and creates
// the corresponding `schema.ResourceData`
func resourceLibvirtNetworkRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Read resource libvirt_network")

	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	network, err := virConn.LookupNetworkByUUIDString(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt network: %s", err)
	}
	defer network.Free()

	networkDef, err := getXMLNetworkDefFromLibvirt(network)
	if err != nil {
		return fmt.Errorf("Error reading libvirt network XML description: %s", err)
	}

	d.Set("name", networkDef.Name)
	d.Set("bridge", networkDef.Bridge.Name)

	if networkDef.MTU != nil {
		d.Set("mtu", networkDef.MTU.Size)
	}

	if networkDef.Forward != nil {
		d.Set("mode", networkDef.Forward.Mode)
	}

	// Domain as won't be present for bridged networks
	if networkDef.Domain != nil {
		d.Set("domain", networkDef.Domain.Name)
		d.Set(dnsPrefix+".local_only", strings.ToLower(networkDef.Domain.LocalOnly) == "yes")
	}

	autostart, err := network.GetAutostart()
	if err != nil {
		return fmt.Errorf("Error reading network autostart setting: %s", err)
	}
	d.Set("autostart", autostart)

	// read add the IP addresses
	addresses := []string{}
	for _, address := range networkDef.IPs {
		// we get the host interface IP (ie, 10.10.8.1) but we want the network CIDR (ie, 10.10.8.0/24)
		// so we need some transformations...
		addr := net.ParseIP(address.Address)
		if addr == nil {
			return fmt.Errorf("Error parsing IP '%s': %s", address.Address, err)
		}
		bits := net.IPv6len * 8
		if addr.To4() != nil {
			bits = net.IPv4len * 8
		}

		mask := net.CIDRMask(int(address.Prefix), bits)
		network := addr.Mask(mask)
		addresses = append(addresses, fmt.Sprintf("%s/%d", network, address.Prefix))
	}
	if len(addresses) > 0 {
		d.Set("addresses", addresses)
	}

	// set as DHCP=enabled if at least one of the IPs has a DHCP configuration
	dhcpEnabled := false
	for _, address := range networkDef.IPs {
		if address.DHCP != nil {
			dhcpEnabled = true
			break
		}
	}
	d.Set("dhcp.0.enabled", dhcpEnabled)

	// read the DNS configuration
	if networkDef.DNS != nil {
		for i, forwarder := range networkDef.DNS.Forwarders {
			key := fmt.Sprintf(dnsPrefix+".forwarders.%d", i)
			if len(forwarder.Addr) > 0 {
				d.Set(key+".address", forwarder.Addr)
			}
			if len(forwarder.Domain) > 0 {
				d.Set(key+".domain", forwarder.Domain)
			}
		}
	}

	// and the static routes
	if len(networkDef.Routes) > 0 {
		for i, route := range networkDef.Routes {
			routePrefix := fmt.Sprintf("routes.%d", i)
			d.Set(routePrefix+".gateway", route.Gateway)

			cidr := fmt.Sprintf("%s/%d", route.Address, route.Prefix)
			d.Set(routePrefix+".cidr", cidr)
		}
	}

	// TODO: get any other parameters from the network and save them

	log.Printf("[DEBUG] Network ID %s successfully read", d.Id())
	return nil
}

func resourceLibvirtNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}
	log.Printf("[DEBUG] Deleting network ID %s", d.Id())

	network, err := virConn.LookupNetworkByUUIDString(d.Id())
	if err != nil {
		return fmt.Errorf("When destroying libvirt network: error retrieving %s", err)
	}
	defer network.Free()

	active, err := network.IsActive()
	if err != nil {
		return fmt.Errorf("Couldn't determine if network is active: %s", err)
	}
	if !active {
		// we have to restart an inactive network, otherwise it won't be
		// possible to remove it.
		if err := network.Create(); err != nil {
			return fmt.Errorf("Cannot restart an inactive network %s", err)
		}
	}

	if err := network.Destroy(); err != nil {
		return fmt.Errorf("When destroying libvirt network: %s", err)
	}

	if err := network.Undefine(); err != nil {
		return fmt.Errorf("Couldn't undefine libvirt network: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"ACTIVE"},
		Target:     []string{"NOT-EXISTS"},
		Refresh:    waitForNetworkDestroyed(virConn, d.Id()),
		Timeout:    1 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for network to reach NOT-EXISTS state: %s", err)
	}
	return nil
}
