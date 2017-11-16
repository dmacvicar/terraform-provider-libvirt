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
				Required: false,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"running": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: false,
			},
			"dns_forwarder": {
				Type:     schema.TypeList,
				Optional: true,
				Required: false,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: dnsForwarderSchema(),
				},
			},
		},
	}
}

func dnsForwarderSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
		return err
	}
	defer network.Free()

	d.Partial(true)

	active, err := network.IsActive()
	if err != nil {
		return err
	}

	if !active {
		log.Printf("[DEBUG] Activating network")
		if err := network.Create(); err != nil {
			return err
		}
		d.Set("running", true)
		d.SetPartial("running")
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
	networkDef.Domain = &libvirtxml.NetworkDomain{
		Name: d.Get("domain").(string),
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

		// some network modes require a DHCP/DNS server
		// set the addresses for DHCP
		if addresses, ok := d.GetOk("addresses"); ok {
			ipsPtrsLst := []libvirtxml.NetworkIP{}
			for _, addressI := range addresses.([]interface{}) {
				address := addressI.(string)
				_, ipNet, err := net.ParseCIDR(address)
				if err != nil {
					return fmt.Errorf("Error parsing addresses definition '%s': %s", address, err)
				}
				ones, bits := ipNet.Mask.Size()
				family := "ipv4"
				if bits == (net.IPv6len * 8) {
					family = "ipv6"
				}
				ipsRange := 2 ^ bits - 2 ^ ones
				if ipsRange < 4 {
					return fmt.Errorf("Netmask seems to be too strict: only %d IPs available (%s)", ipsRange-3, family)
				}

				// we should calculate the range served by DHCP. For example, for
				// 192.168.121.0/24 we will serve 192.168.121.2 - 192.168.121.254
				start, end := NetworkRange(ipNet)

				// skip the .0, (for the network),
				start[len(start)-1]++

				// assign the .1 to the host interface
				dni := libvirtxml.NetworkIP{
					Address: start.String(),
					Prefix:  strconv.Itoa(ones),
					Family:  family,
				}

				start[len(start)-1]++ // then skip the .1
				end[len(end)-1]--     // and skip the .255 (for broadcast)

				dni.DHCP = &libvirtxml.NetworkDHCP{
					Ranges: []libvirtxml.NetworkDHCPRange{
						{
							Start: start.String(),
							End:   end.String(),
						},
					},
				}
				ipsPtrsLst = append(ipsPtrsLst, dni)
			}
			networkDef.IPs = ipsPtrsLst
		}

		if dnsForwardCount, ok := d.GetOk("dns_forwarder.#"); ok {
			dns := libvirtxml.NetworkDNS{
				Forwarders: []libvirtxml.NetworkDNSForwarder{},
			}

			for i := 0; i < dnsForwardCount.(int); i++ {
				forward := libvirtxml.NetworkDNSForwarder{}
				forwardPrefix := fmt.Sprintf("dns_forwarder.%d", i)
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
				dns.Forwarders = append(dns.Forwarders, forward)
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

	log.Printf("[DEBUG] Creating libvirt network at %s: %s", connectURI, data)
	network, err := virConn.NetworkDefineXML(data)
	if err != nil {
		return fmt.Errorf("Error defining libvirt network: %s - %s", err, data)
	}
	err = network.Create()
	if err != nil {
		return fmt.Errorf("Error crearing libvirt network: %s", err)
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

	networkDef, err := newDefNetworkfromLibvirt(network)
	if err != nil {
		return fmt.Errorf("Error reading libvirt network XML description: %s", err)
	}

	d.Set("name", networkDef.Name)
	d.Set("bridge", networkDef.Bridge.Name)

	// Domain as won't be present for bridged networks
	if networkDef.Domain != nil {
		d.Set("domain", networkDef.Domain.Name)
	}

	active, err := network.IsActive()
	if err != nil {
		return err
	}
	d.Set("running", active)

	addresses := []string{}
	for _, address := range networkDef.IPs {
		// we get the host interface IP (ie, 10.10.8.1) but we want the network CIDR (ie, 10.10.8.0/24)
		// so we need some transformations...
		addr := net.ParseIP(address.Address)
		if addr == nil {
			return fmt.Errorf("Error parsing IP '%s': %s", address, err)
		}
		bits := net.IPv6len * 8
		if addr.To4() != nil {
			bits = net.IPv4len * 8
		}

		prefix, _ := strconv.Atoi(address.Prefix)
		mask := net.CIDRMask(prefix, bits)
		network := addr.Mask(mask)
		addresses = append(addresses, fmt.Sprintf("%s/%s", network, address.Prefix))
	}
	if len(addresses) > 0 {
		d.Set("addresses", addresses)
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

func waitForNetworkActive(network libvirt.Network) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		active, err := network.IsActive()
		if err != nil {
			return nil, "", err
		}
		if active {
			return network, "ACTIVE", nil
		}
		return network, "BUILD", err
	}
}

// wait for network to be up and timeout after 5 minutes.
func waitForNetworkDestroyed(virConn *libvirt.Connect, uuid string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("Waiting for network %s to be destroyed", uuid)
		network, err := virConn.LookupNetworkByUUIDString(uuid)
		if err.(libvirt.Error).Code == libvirt.ERR_NO_NETWORK {
			return virConn, "NOT-EXISTS", nil
		}
		defer network.Free()
		return virConn, "ACTIVE", err
	}
}
