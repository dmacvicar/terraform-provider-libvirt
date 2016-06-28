package libvirt

import (
	"fmt"
	"log"
	"time"

	libvirt "github.com/dmacvicar/libvirt-go"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"net"
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
// "addresses" can contain (0 or 1) ipv4 and (0 or 1) ipv6 ranges
// "mode" can be one of: "nat" (default), "isolated"
//
func resourceLibvirtNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceLibvirtNetworkCreate,
		Read:   resourceLibvirtNetworkRead,
		Update: resourceLibvirtNetworkUpdate,
		Delete: resourceLibvirtNetworkDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"mode": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"bridge": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"addresses": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Required: false,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceLibvirtNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	// TODO: we should create different networks depending on the "mode" ("nat", etc)
	networkDef := newNetworkDef()
	networkDef.Name = d.Get("name").(string)
	networkDef.Domain = &defNetworkDomain{
		Name: d.Get("domain").(string),
	}
	if bridgeName, ok := d.GetOk("bridge"); ok {
		networkDef.Bridge = &defNetworkBridge{
			Name: bridgeName.(string),
		}
	}
	if addresses, ok := d.GetOk("addresses"); ok {
		ipsPtrsLst := []*defNetworkIp{}
		for _, addressI := range addresses.([]interface{}) {
			address := addressI.(string)
			_, net, err := net.ParseCIDR(address)
			if err != nil {
				return fmt.Errorf("Error parsing addresses definition '%s': %s", address, err)
			}
			ones, bits := net.Mask.Size()
			family := "ipv4"
			if bits == 64 { // TODO: use some constant
				family = "ipv6"
			}
			dni := defNetworkIp{
				Address: net.IP.String(),
				Prefix:  ones,
				Family:  family,
			}

			// we always start DHCP if we have addresses
			// TODO: maybe we should not enforce this in the future...
			// we should calculate the range served by DHCP. For example, for
			// 192.168.121.0/24 we will serve 192.168.121.1 - 192.168.121.254
			start, end := NetworkRange(net)

			// skip the .0 and .255
			start[len(start)-1]++
			end[len(start)-1]--

			dni.Dhcp = &defNetworkIpDhcp{
				Ranges: []*defNetworkIpDhcpRange{
					&defNetworkIpDhcpRange{
						Start: start.String(),
						End:   end.String(),
					},
				},
			}

			// TODO: check there is
			ipsPtrsLst = append(ipsPtrsLst, &dni)
		}
		networkDef.Ips = ipsPtrsLst

	}

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

	log.Printf("[INFO] Created network %s [%s]", networkDef.Name, d.Id())

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"BUILD"},
		Target:     []string{"ACTIVE"},
		Refresh:    waitForNetworkActive(network),
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
		return fmt.Errorf("The libvirt connection was nil.")
	}

	network, err := virConn.LookupNetworkByUUIDString(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt network: %s", err)
	}
	defer network.Free()

	xmlDesc, err := network.GetXMLDesc(0)
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt network XML description: %s", err)
	}
	log.Printf("[DEBUG] obtained network with XML definition: %s\n", xmlDesc)

	networkDef, err := newNetworkDefFromXML(xmlDesc)
	if err != nil {
		return fmt.Errorf("Error reading libvirt network XML description: %s", err)
	}

	d.Set("name", networkDef.Name)
	d.Set("domain", networkDef.Domain.Name)

	addresses := []string{}
	for _, address := range networkDef.Ips {
		addresses = append(addresses, fmt.Sprintf("%s/%d", address.Address, address.Prefix))
	}
	if len(addresses) > 0 {
		d.Set("addresses", addresses)
	}

	// TODO: read some other parameters from the network and save them

	log.Printf("[DEBUG] Network ID %s successfully read", d.Id())
	return nil
}

func resourceLibvirtNetworkUpdate(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	log.Printf("[DEBUG] Updating network ID %s", d.Id())

	/*

		network, err := virConn.LookupNetworkByUUIDString(d.Id())
		if err != nil {
			return fmt.Errorf("Error retrieving libvirt network: %s", err)
		}
		defer network.Free()

		// TODO

		if d.HasChange("domain") {
			network.UpdateXMLDesc(getHostXMLDesc(ip, mac, name),
				libvirt.VIR_NETWORK_UPDATE_COMMAND_ADD_LAST,
				libvirt.VIR_NETWORK_SECTION_DOMAIN)
		}
		if d.HasChange("addresses") {
		}

		log.Printf("[DEBUG] Updating Network %s with options: %+v", d.Id(), updateOpts)

		_, err = networks.Update(networkingClient, d.Id(), updateOpts).Extract()
		if err != nil {
			return fmt.Errorf("Error updating OpenStack Neutron Network: %s", err)
		}

	*/

	return resourceLibvirtNetworkRead(d, meta)
}

func resourceLibvirtNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}
	log.Printf("[DEBUG] Deleting network ID %s", d.Id())

	network, err := virConn.LookupNetworkByUUIDString(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt network: %s", err)
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
		return fmt.Errorf("Couldn't destroy libvirt network: %s", err)
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

func waitForNetworkActive(network libvirt.VirNetwork) resource.StateRefreshFunc {
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
func waitForNetworkDestroyed(virConn *libvirt.VirConnection, uuid string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("Waiting for network %s to be destroyed", uuid)
		_, err := virConn.LookupNetworkByUUIDString(uuid)
		if err.(libvirt.VirError).Code == libvirt.VIR_ERR_NO_NETWORK {
			return virConn, "NOT-EXISTS", nil
		}
		return virConn, "ACTIVE", err
	}
}
