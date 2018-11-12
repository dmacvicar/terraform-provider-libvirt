package libvirt

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	libvirt "github.com/libvirt/libvirt-go"
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
				ForceNew: true,
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
				ForceNew: true,
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
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						"hosts": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ip": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"mac": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"name": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
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

func resourceLibvirtNetworkUpdate(d *schema.ResourceData, meta interface{}) error {
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

	active, err := network.IsActive()
	if err != nil {
		return fmt.Errorf("Error by getting network's status during update: %s", err)
	}

	if !active {
		log.Printf("[DEBUG] Activating network")
		if err := network.Create(); err != nil {
			return fmt.Errorf("Error by activating network during update: %s", err)
		}
	}

	if d.HasChange("autostart") {
		err = network.SetAutostart(d.Get("autostart").(bool))
		if err != nil {
			return fmt.Errorf("Error updating autostart for network: %s", err)
		}
		d.SetPartial("autostart")
	}

	err = resourceLibvirtNetworkUpdateDNSHosts(d, network)
	if err != nil {
		return fmt.Errorf("update DNS hosts: %s", err)
	}

	d.Partial(false)
	return nil
}

func resourceLibvirtNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	// see https://libvirt.org/formatnetwork.html
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	networkDef := newNetworkDef()
	networkDef.Name = d.Get("name").(string)

	if domain, ok := d.GetOk("domain"); ok {
		networkDef.Domain = &libvirtxml.NetworkDomain{
			Name: domain.(string),
		}

		if dnsLocalOnly, ok := d.GetOk(dnsPrefix + ".local_only"); ok {
			if dnsLocalOnly.(bool) {
				networkDef.Domain.LocalOnly = "yes" // this "boolean" must be "yes"|"no"
			}
		}
	}

	// use a bridge provided by the user, or create one otherwise (libvirt will assign on automatically when empty)
	bridgeName := ""
	if b, ok := d.GetOk("bridge"); ok {
		bridgeName = b.(string)
	}
	networkDef.Bridge = &libvirtxml.NetworkBridge{
		Name: bridgeName,
		STP:  "on",
	}

	// check the network mode
	networkDef.Forward = &libvirtxml.NetworkForward{
		Mode: strings.ToLower(d.Get("mode").(string)),
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
		err := setDhcpByCIDRAdressesSubnets(d, &networkDef)
		if err != nil {
			return fmt.Errorf("Could not set DHCP from adresses '%s'", err)
		}
		var dnsForwarders []libvirtxml.NetworkDNSForwarder
		if dnsForwardCount, ok := d.GetOk(dnsPrefix + ".forwarders.#"); ok {
			for i := 0; i < dnsForwardCount.(int); i++ {
				forward := libvirtxml.NetworkDNSForwarder{}
				forwardPrefix := fmt.Sprintf(dnsPrefix+".forwarders.%d", i)
				if address, ok := d.GetOk(forwardPrefix + ".address"); ok {
					ip := net.ParseIP(address.(string))
					if ip == nil {
						return fmt.Errorf("Could not parse address '%s'", address)
					}
					forward.Addr = ip.String()
				}
				if domain, ok := d.GetOk(forwardPrefix + ".domain"); ok {
					forward.Domain = domain.(string)
				}
				dnsForwarders = append(dnsForwarders, forward)
			}
		}

		var dnsSRVs []libvirtxml.NetworkDNSSRV
		if dnsSRVCount, ok := d.GetOk(dnsPrefix + ".srvs.#"); ok {
			for i := 0; i < dnsSRVCount.(int); i++ {
				srv := libvirtxml.NetworkDNSSRV{}
				srvPrefix := fmt.Sprintf(dnsPrefix+".srvs.%d", i)
				if service, ok := d.GetOk(srvPrefix + ".service"); ok {
					srv.Service = service.(string)
				}
				if protocol, ok := d.GetOk(srvPrefix + ".protocol"); ok {
					srv.Protocol = protocol.(string)
				}
				if domain, ok := d.GetOk(srvPrefix + ".domain"); ok {
					srv.Domain = domain.(string)
				}
				if target, ok := d.GetOk(srvPrefix + ".target"); ok {
					srv.Target = target.(string)
				}
				if port, ok := d.GetOk(srvPrefix + ".port"); ok {
					p, err := strconv.Atoi(port.(string))
					if err != nil {
						return fmt.Errorf("Could not convert port '%s' to int", port)
					}
					srv.Port = uint(p)
				}
				if weight, ok := d.GetOk(srvPrefix + ".weight"); ok {
					w, err := strconv.Atoi(weight.(string))
					if err != nil {
						return fmt.Errorf("Could not convert weight '%s' to int", weight)
					}
					srv.Weight = uint(w)
				}
				if priority, ok := d.GetOk(srvPrefix + ".priority"); ok {
					w, err := strconv.Atoi(priority.(string))
					if err != nil {
						return fmt.Errorf("Could not convert priority '%s' to int", priority)
					}
					srv.Priority = uint(w)
				}
				dnsSRVs = append(dnsSRVs, srv)
			}
		}

		dnsHostsMap := map[string][]string{}
		if dnsHostCount, ok := d.GetOk(dnsPrefix + ".hosts.#"); ok {
			for i := 0; i < dnsHostCount.(int); i++ {
				hostPrefix := fmt.Sprintf(dnsPrefix+".hosts.%d", i)

				address := d.Get(hostPrefix + ".ip").(string)
				if net.ParseIP(address) == nil {
					return fmt.Errorf("Could not parse address '%s'", address)
				}

				dnsHostsMap[address] = append(dnsHostsMap[address], d.Get(hostPrefix+".hostname").(string))
			}
		}

		var dnsHosts []libvirtxml.NetworkDNSHost
		for ip, hostnames := range dnsHostsMap {
			dnsHostnames := []libvirtxml.NetworkDNSHostHostname{}
			for _, hostname := range hostnames {
				dnsHostnames = append(dnsHostnames, libvirtxml.NetworkDNSHostHostname{Hostname: hostname})
			}
			dnsHosts = append(dnsHosts, libvirtxml.NetworkDNSHost{
				IP:        ip,
				Hostnames: dnsHostnames,
			})
		}

		if len(dnsForwarders) > 0 || len(dnsSRVs) > 0 || len(dnsHosts) > 0 {
			dns := libvirtxml.NetworkDNS{
				Forwarders: dnsForwarders,
				Host:       dnsHosts,
				SRVs:       dnsSRVs,
			}
			networkDef.DNS = &dns
		}

	} else if networkDef.Forward.Mode == netModeBridge {
		if bridgeName == "" {
			return fmt.Errorf("'bridge' must be provided when using the bridged network mode")
		}
		// Bridges cannot forward
		networkDef.Forward = nil
	} else {
		return fmt.Errorf("unsupported network mode '%s'", networkDef.Forward.Mode)
	}

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

func resourceLibvirtNetworkRead(d *schema.ResourceData, meta interface{}) error {
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
