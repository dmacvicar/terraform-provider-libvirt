package libvirt

import (
	"encoding/xml"
	"fmt"
	libvirt "github.com/dmacvicar/libvirt-go"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"time"
)

func resourceLibvirtDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceLibvirtDomainCreate,
		Read:   resourceLibvirtDomainRead,
		Delete: resourceLibvirtDomainDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vcpu": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
				ForceNew: true,
			},
			"memory": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  512,
				ForceNew: true,
			},
			"disk": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Required: false,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: diskCommonSchema(),
				},
			},
			"network_interface": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Required: false,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: networkInterfaceCommonSchema(),
				},
			},
		},
	}
}

func resourceLibvirtDomainCreate(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	disksCount := d.Get("disk.#").(int)
	disks := make([]defDisk, 0, disksCount)
	for i := 0; i < disksCount; i++ {
		prefix := fmt.Sprintf("disk.%d", i)
		disk := newDefDisk()
		disk.Target.Dev = fmt.Sprintf("vd%s", DiskLetterForIndex(i))

		volumeKey := d.Get(prefix + ".volume_id").(string)
		diskVolume, err := virConn.LookupStorageVolByKey(volumeKey)
		if err != nil {
			return fmt.Errorf("Can't retrieve volume %s", volumeKey)
		}
		diskVolumeName, err := diskVolume.GetName()
		if err != nil {
			return fmt.Errorf("Error retrieving volume name: %s", err)
		}
		diskPool, err := diskVolume.LookupPoolByVolume()
		if err != nil {
			return fmt.Errorf("Error retrieving pool for volume: %s", err)
		}
		diskPoolName, err := diskPool.GetName()
		if err != nil {
			return fmt.Errorf("Error retrieving pool name: %s", err)
		}

		disk.Source.Volume = diskVolumeName
		disk.Source.Pool = diskPoolName

		disks = append(disks, disk)
	}

	netIfacesCount := d.Get("network_interface.#").(int)
	netIfaces := make([]defNetworkInterface, 0, netIfacesCount)
	for i := 0; i < netIfacesCount; i++ {
		prefix := fmt.Sprintf("network_interface.%d", i)
		netIface := newDefNetworkInterface()

		if mac, ok := d.GetOk(prefix + ".mac"); ok {
			netIface.Mac.Address = mac.(string)
		}

		if network, ok := d.GetOk(prefix + ".network"); ok {
			netIface.Source.Network = network.(string)
		}
		netIfaces = append(netIfaces, netIface)
	}

	domainDef := newDomainDef()
	if name, ok := d.GetOk("name"); ok {
		domainDef.Name = name.(string)
	}

	domainDef.Memory.Amount = d.Get("memory").(int)
	domainDef.VCpu.Amount = d.Get("vcpu").(int)
	domainDef.Devices.Disks = disks
	domainDef.Devices.NetworkInterfaces = netIfaces

	connectURI, err := virConn.GetURI()
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt connection URI: %s", err)
	}
	log.Printf("[INFO] Creating libvirt domain at %s", connectURI)

	data, err := xml.Marshal(domainDef)
	if err != nil {
		return fmt.Errorf("Error serializing libvirt domain: %s", err)
	}

	domain, err := virConn.DomainCreateXML(string(data), libvirt.VIR_DOMAIN_NONE)
	if err != nil {
		return fmt.Errorf("Error crearing libvirt domain: %s", err)
	}
	defer domain.Free()

	id, err := domain.GetUUIDString()
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt domain id: %s", err)
	}
	d.SetId(id)

	log.Printf("[INFO] Domain ID: %s", d.Id())

	err = waitForDomainUp(domain)
	if err != nil {
		return fmt.Errorf("Error waiting for domain to reach RUNNING state: %s", err)
	}

	return resourceLibvirtDomainRead(d, meta)
}

func resourceLibvirtDomainRead(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
 		return fmt.Errorf("The libvirt connection was nil.")
	}

	domain, err := virConn.LookupByUUIDString(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt domain: %s", err)
	}
	defer domain.Free()

	xmlDesc, err := domain.GetXMLDesc(0)
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt domain XML description: %s", err)
	}

	domainDef := newDomainDef()
	err = xml.Unmarshal([]byte(xmlDesc), &domainDef)
	if err != nil {
		return fmt.Errorf("Error reading libvirt domain XML description: %s", err)
	}

	d.Set("name", domainDef.Name)
	d.Set("vpu", domainDef.VCpu)
	d.Set("memory", domainDef.Memory)

	disks := make([]map[string]interface{}, 0)
	for _, diskDef := range domainDef.Devices.Disks {
		virPool, err := virConn.LookupStoragePoolByName(diskDef.Source.Pool)
		if err != nil {
			return fmt.Errorf("Error retrieving pool for disk: %s", err)
		}
		defer virPool.Free()

		virVol, err := virPool.LookupStorageVolByName(diskDef.Source.Volume)
		if err != nil {
			return fmt.Errorf("Error retrieving volume for disk: %s", err)
		}
		defer virVol.Free()

		virVolKey, err := virVol.GetKey()
		if err != nil {
			return fmt.Errorf("Error retrieving volume ke for disk: %s", err)
		}

		disk := map[string]interface{}{
			"volume_id": virVolKey,
		}
		disks = append(disks, disk)
	}
	d.Set("disks", disks)

	// look interfaces with addresses
	start := time.Now()
	var ifacesWithAddr []libvirt.VirDomainInterface
	for {
		ifacesWithAddr, err = domain.ListAllInterfaceAddresses(libvirt.VIR_DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
		if err != nil {
			return fmt.Errorf("Error retrieving interface addresses: %s", err)
		}

		log.Printf("[DEBUG] %d interfaces defined. %d addresses available.", len(domainDef.Devices.NetworkInterfaces), len(ifacesWithAddr))

		// wait for at least one address per interface
		if len(ifacesWithAddr) >= len(domainDef.Devices.NetworkInterfaces) {
			break
		}

		log.Printf("[DEBUG] waiting for addresses...")
		time.Sleep(1 * time.Second)
		if time.Since(start) > 5*time.Minute {
			return fmt.Errorf("Timeout waiting for interface addresses")
		}
	}

	netIfaces := make([]map[string]interface{}, 0)
	for _, networkInterfaceDef := range domainDef.Devices.NetworkInterfaces {

		if networkInterfaceDef.Type != "network" {
			log.Printf("[DEBUG] ignoring interface of type '%s'", networkInterfaceDef.Type)
			continue
		}

		netIface := map[string]interface{}{
			"network": networkInterfaceDef.Source.Network,
			"mac":     networkInterfaceDef.Mac.Address,
		}

		netIfaceAddrs := make([]map[string]interface{}, 0)
		// look for an ip address and try to match it with the mac address
		// not sure if using the target device name is a better idea here
		for _, ifaceWithAddr := range ifacesWithAddr {
			log.Printf("[DEBUG] Trying with: '%s': '%s' against '%s'\n", ifaceWithAddr.Hwaddr, ifaceWithAddr.Addrs[0].Addr, networkInterfaceDef.Mac.Address)
			if ifaceWithAddr.Hwaddr == networkInterfaceDef.Mac.Address {
				for _, addr := range(ifaceWithAddr.Addrs) {
					netIfaceAddr := map[string]interface{}{
						"type": func() string {
							switch addr.Type {
							case libvirt.VIR_IP_ADDR_TYPE_IPV4:
								return "ipv4"
							case libvirt.VIR_IP_ADDR_TYPE_IPV6:
								return "ipv6"
							default:
								return "other"
							}
						}(),
						"address": addr.Addr,
						"prefix": addr.Prefix,
					}
					netIfaceAddrs = append(netIfaceAddrs, netIfaceAddr)
				}
			} else {
				log.Printf("[DEBUG] Unmatched address: '%s': '%s'\n", ifaceWithAddr.Hwaddr, networkInterfaceDef.Mac.Address)
			}
		}

		log.Printf("[DEBUG] %d addresses for %s\n", len(netIfaceAddrs), networkInterfaceDef.Mac.Address)
		netIface["address"] = netIfaceAddrs
		netIfaces = append(netIfaces, netIface)
	}
	d.Set("network_interface", netIfaces)

	if len(ifacesWithAddr) > 0 {
		d.SetConnInfo(map[string]string{
			"type":     "ssh",
  		    "host":     ifacesWithAddr[0].Addrs[0].Addr,
		})
	}
	return nil
}

func resourceLibvirtDomainDelete(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	domain, err := virConn.LookupByUUIDString(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt domain: %s", err)
	}
	defer domain.Free()

	if err := domain.Destroy(); err != nil {
		return fmt.Errorf("Couldn't destroy libvirt domain: %s", err)
	}

	err = waitForDomainDestroyed(virConn, d.Id())
	if err != nil {
		return fmt.Errorf("Error waiting for domain to be destroyed: %s", err)
	}

	return nil
}

// wait for domain to be up and timeout after 5 minutes.
func waitForDomainUp(domain libvirt.VirDomain) error {
	start := time.Now()
	for {
		state, err := domain.GetState()
		if err != nil {
			return err
		}

		running := true
		if state[0] != libvirt.VIR_DOMAIN_RUNNING {
			running = false
		}

		if running {
			return nil
		}
		time.Sleep(1 * time.Second)
		if time.Since(start) > 5*time.Minute {
			return fmt.Errorf("Domain didn't switch to state RUNNING in 5 minutes")
		}
	}
}

// wait for domain to be up and timeout after 5 minutes.
func waitForDomainDestroyed(virConn *libvirt.VirConnection, uuid string) error {
	start := time.Now()
	for {
		log.Printf("Waiting for domain %s to be destroyed", uuid)
		_, err := virConn.LookupByUUIDString(uuid)
		if err.(libvirt.VirError).Code == libvirt.VIR_ERR_NO_DOMAIN {
			return nil
		}

		time.Sleep(1 * time.Second)
		if time.Since(start) > 5*time.Minute {
			return fmt.Errorf("Domain is still there after 5 minutes")
		}
	}
}
