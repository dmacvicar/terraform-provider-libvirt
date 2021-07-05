package libvirt

import (
	"fmt"
	"log"
	"net"
	"time"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

const (
	networkFamilyIPv4 = "ipv4"
	networkFamilyIPv6 = "ipv6"
)

func resourceLibvirtNetworkV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceLibvirtNetworkV2Create,
		Read:   resourceLibvirtNetworkV2Read,
		Delete: resourceLibvirtNetworkV2Delete,
		Exists: resourceLibvirtNetworkV2Exists,
		Update: resourceLibvirtNetworkV2Update,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"autostart": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"bridge": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"domain": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Required: false,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"local_only": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"dns": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"forwarder": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"addr": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc:     validation.IsIPAddress,
										StateFunc:        networkAddressStateFunc,
										DiffSuppressFunc: networkAddressDiffSuppressFunc,
									},
									"domain": {
										Type:     schema.TypeString,
										Optional: true,
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
						"host": {
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
									// TODO make this a list?
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
			"dnsmasq_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"options": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: false,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"option_name": {
										Type: schema.TypeString,
										// This should be required, but Terraform does validation too early
										// and therefore doesn't recognize that this is set when assigning from
										// a rendered dnsmasq_options template.
										Optional: true,
									},
									"option_value": {
										Type: schema.TypeString,
										// This should be required, but Terraform does validation too early
										// and therefore doesn't recognize that this is set when assigning from
										// a rendered dnsmasq_options template.
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"forward": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Required: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// can be "none", "nat" (default), "route", "bridge"
						"mode": {
							Type:     schema.TypeString,
							Optional: true,
							Required: false,
							Default:  netModeNat,
						},
					},
				},
			},
			"ip": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:             schema.TypeString,
							Optional:         true,
							Required:         false,
							ValidateFunc:     validation.IsIPAddress,
							StateFunc:        networkAddressStateFunc,
							DiffSuppressFunc: networkAddressDiffSuppressFunc,
						},
						"family": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      networkFamilyIPv4,
							ValidateFunc: validation.StringInSlice([]string{networkFamilyIPv4, networkFamilyIPv6}, false),
						},
						"dhcp": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Required: false,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"range": {
										Type:     schema.TypeList,
										Optional: true,
										Required: false,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"start": {
													Type:     schema.TypeString,
													Required: true,
												},
												"end": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"host": {
										Type:     schema.TypeList,
										Optional: true,
										Required: false,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mac": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.IsMACAddress,
												},
												"id": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"name": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"ip": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.IsIPAddress,
												},
											},
										},
									},
								},
							},
						},
						"netmask": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"prefix": {
							Type:     schema.TypeInt,
							Optional: true,
							ValidateFunc: validation.IntBetween(1, net.IPv6len * 8),
						},
					},
				},
			},
			"mtu": {
				Type:     schema.TypeInt,
				Optional: true,
				Required: false,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"route": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsIPAddress,
						},
						"prefix": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"gateway": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsIPAddress,
						},
						"metric": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"netmask": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsCIDR,
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

func resourceLibvirtNetworkV2Exists(d *schema.ResourceData, meta interface{}) (bool, error) {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return false, fmt.Errorf(LibVirtConIsNil)
	}

	var uuid libvirt.UUID
	uuid = parseUUID(d.Id())
	_, err := virConn.NetworkLookupByUUID(uuid)
	if err != nil {
		// If the network couldn't be found, don't return an error otherwise
		// Terraform won't create it again.
		if lverr, ok := err.(libvirt.Error); ok && lverr.Code == uint32(libvirt.ErrNoNetwork) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// resourceLibvirtNetworkV2Update updates dynamically some attributes in the network
func resourceLibvirtNetworkV2Update(d *schema.ResourceData, meta interface{}) error {
	// check the list of things that can be changed dynamically
	// in https://wiki.libvirt.org/page/Networking#virsh_net-update
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	network, err := virConn.NetworkLookupByUUID(parseUUID(d.Id()))
	if err != nil {
		return fmt.Errorf("Can't retrieve network with ID '%s' during update: %s", d.Id(), err)
	}

	d.Partial(true)

	activeInt, err := virConn.NetworkIsActive(network)
	if err != nil {
		return fmt.Errorf("Error when getting network %s status during update: %s", network.Name, err)
	}

	active := activeInt == 1
	if !active {
		log.Printf("[DEBUG] Activating network %s", network.Name)
		if err := virConn.NetworkCreate(network); err != nil {
			return fmt.Errorf("Error when activating network %s during update: %s", network.Name, err)
		}
	}

	if d.HasChange("autostart") {
		err = virConn.NetworkSetAutostart(network, bool2int(d.Get("autostart").(bool)))
		if err != nil {
			return fmt.Errorf("Error updating autostart for network %s: %s", network.Name, err)
		}
	}

	// detect changes in the DNS entries in this network
	err = updateDNSHosts(d, meta, network)
	if err != nil {
		return fmt.Errorf("Error updating DNS hosts for network %s: %s", network.Name, err)
	}

	d.Partial(false)
	return nil
}

// resourceLibvirtNetworkV2Create creates a libvirt network from the resource definition

func resourceLibvirtNetworkV2Create(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt

	// see https://libvirt.org/formatnetwork.html
	networkDef := libvirtxml.Network{}

	if v, ok := d.GetOk("bridge"); ok {
		networkDef.Bridge = expandNetworkV2Bridge(v.([]interface{}))
	}

	if v, ok := d.GetOk("domain"); ok {
		networkDef.Domain = expandNetworkV2Domain(v.([]interface{}))
	}

	if v, ok := d.GetOk("forward"); ok {
		networkDef.Forward = expandNetworkV2Forward(v.([]interface{}))
	}

	if v, ok := d.GetOk("ip"); ok {
		networkDef.IPs = expandNetworkV2IPs(v.([]interface{}))
	}

	if v, ok := d.GetOk("name"); ok {
		networkDef.Name = v.(string)
	}

	if v, ok := d.GetOk("route"); ok {
		networkDef.Routes = expandNetworkV2Routes(v.([]interface{}))
	}

	// once we have the network defined, connect to libvirt and create it from the XML serialization
	log.Printf("[INFO] Creating libvirt network")

	data, err := xmlMarshallIndented(networkDef)
	if err != nil {
		return fmt.Errorf("Error serializing libvirt network: %s", err)
	}
	log.Printf("[DEBUG] Generated XML for libvirt network:\n%s", data)

	data, err = transformResourceXML(data, d)
	if err != nil {
		return fmt.Errorf("Error applying XSLT stylesheet: %s", err)
	}

	network, err := func() (libvirt.Network, error) {
		// define only one network at a time
		// see https://gitlab.com/libvirt/libvirt/-/issues/78
		meta.(*Client).networkMutex.Lock()
		defer meta.(*Client).networkMutex.Unlock()

		log.Printf("[DEBUG] Creating libvirt network: %s", data)
		return virConn.NetworkDefineXML(data)
	}()

	if err != nil {
		return fmt.Errorf("Error defining libvirt network: %s - %s", err, data)
	}

	err = virConn.NetworkCreate(network)
	if err != nil {
		// in some cases, the network creation fails but an artifact is created
		// an 'broken network". Remove the network in case of failure
		// see https://github.com/dmacvicar/terraform-provider-libvirt/issues/739
		// don't handle the error for destroying
		virConn.NetworkDestroy(network)
		virConn.NetworkUndefine(network)
		return fmt.Errorf("Error creating libvirt network: %s", err)
	}
	id := uuidString(network.UUID)
	d.SetId(id)

	log.Printf("[INFO] Created network %s [%s]", networkDef.Name, d.Id())

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"BUILD"},
		Target:     []string{"ACTIVE"},
		Refresh:    waitForNetworkActive(virConn, network),
		Timeout:    1 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for network to reach ACTIVE state: %s", err)
	}

	/*
		if v, ok := d.GetOk("autostart"); ok {
			err = virConn.NetworkSetAutostart(network, bool2int(v.(bool)))
			if err != nil {
				return fmt.Errorf("Error setting autostart for network: %s", err)
			}
		}
	*/
	return resourceLibvirtNetworkV2Read(d, meta)
}

// resourceLibvirtNetworkV2Read gets the current resource from libvirt and creates
// the corresponding `schema.ResourceData`
func resourceLibvirtNetworkV2Read(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Read resource libvirt_network")

	virConn := meta.(*Client).libvirt

	var uuid libvirt.UUID
	uuid = parseUUID(d.Id())
	network, err := virConn.NetworkLookupByUUID(uuid)
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt network: %s", err)
	}

	networkDef, err := getXMLNetworkDefFromLibvirt(virConn, network)
	if err != nil {
		return fmt.Errorf("Error reading libvirt network XML description: %s", err)
	}

	autostart, err := virConn.NetworkGetAutostart(network)
	if err != nil {
		return fmt.Errorf("Error reading network autostart setting: %s", err)
	}
	d.Set("autostart", autostart > 0)

	if networkDef.Bridge != nil {
		d.Set("bridge", flattenNetworkV2Bridge(networkDef.Bridge))
	}

	if networkDef.Domain != nil {
		d.Set("domain", flattenNetworkV2Domain(networkDef.Domain))
	}

	if len(networkDef.IPs) > 0 {
		d.Set("ip", flattenNetworkV2IPs(networkDef.IPs))
	}

	if networkDef.Name != "" {
		d.Set("name", networkDef.Name)
	}

	if len(networkDef.Routes) > 0 {
		d.Set("routes", flattenNetworkV2Routes(networkDef.Routes))
	}

	// TODO: get any other parameters from the network and save them
	log.Printf("[DEBUG] Network ID %s successfully read", d.Id())
	return nil
}

func resourceLibvirtNetworkV2Delete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Deleting network ID %s", d.Id())

	virConn := meta.(*Client).libvirt

	var uuid libvirt.UUID
	uuid = parseUUID(d.Id())

	network, err := virConn.NetworkLookupByUUID(uuid)
	if err != nil {
		if lverr, ok := err.(libvirt.Error); ok && lverr.Code == uint32(libvirt.ErrNoNetwork) {
			d.SetId("")
		}
		return fmt.Errorf("When destroying libvirt network: error retrieving %s", err)
	}

	activeInt, err := virConn.NetworkIsActive(network)
	if err != nil {
		return fmt.Errorf("Couldn't determine if network is active: %s", err)
	}

	// network is active, so we need to destroy it and undefine it
	if activeInt == 1 {
		if err := virConn.NetworkDestroy(network); err != nil {
			return fmt.Errorf("When destroying libvirt network: %s", err)
		}
	}

	if err := virConn.NetworkUndefine(network); err != nil {
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
