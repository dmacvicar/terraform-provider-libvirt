package libvirt

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"strings"
	"time"

	libvirt "github.com/dmacvicar/libvirt-go"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceLibvirtDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceLibvirtDomainCreate,
		Read:   resourceLibvirtDomainRead,
		Delete: resourceLibvirtDomainDelete,
		Update: resourceLibvirtDomainUpdate,
		Exists: resourceLibvirtDomainExists,
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
			"running": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: false,
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

func resourceLibvirtDomainExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return false, fmt.Errorf("The libvirt connection was nil.")
	}
	domain, err := virConn.LookupByUUIDString(d.Id())
	defer domain.Free()
	return err == nil, err
}

func resourceLibvirtDomainCreate(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	domainDef := newDomainDef()
	if name, ok := d.GetOk("name"); ok {
		domainDef.Name = name.(string)
	}
	domainDef.Memory.Amount = d.Get("memory").(int)
	domainDef.VCpu.Amount = d.Get("vcpu").(int)

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
			netIface.Mac.Address = strings.ToUpper(mac.(string))
		} else {
			var err error
			netIface.Mac.Address, err = RandomMACAddress()
			if err != nil {
				return fmt.Errorf("Error generating mac address: %s", err)
			}
		}

		// this is not passes to libvirt, but used by waitForAddress
		if waitForLease, ok := d.GetOk(prefix + ".wait_for_lease"); ok {
			netIface.waitForLease = waitForLease.(bool)
		}

		networkName := "default"
		if n, ok := d.GetOk(prefix + ".network_name"); ok {
			// when using a "network_name" we do not try to do anything: we just
			// connect to that network
			networkName = n.(string)
		} else if networkUUID, ok := d.GetOk(prefix + ".network_id"); ok {
			network, err := virConn.LookupNetworkByUUIDString(networkUUID.(string))
			if err != nil {
				return fmt.Errorf("Can't retrieve network ID %s", networkUUID)
			}
			networkName, err = network.GetName()
			if err != nil {
				return fmt.Errorf("Error retrieving volume name: %s", err)
			}
			getOrEmpty := func(field string) string {
				if r, ok := d.GetOk(prefix + "." + field); ok {
					return r.(string)
				} else {
					return ""
				}
			}

			// try to associate the IP / hostname to the MAC in that network
			address, hostname := getOrEmpty("address"), getOrEmpty("hostname")
			if len(address) > 0 {
				// TODO: check the address is in the valid ranges for this network

				if len(hostname) == 0 {
					// try to use the name from the domain definition when no hostname has been specified
					hostname = domainDef.Name
				}
			} else {
				if len(hostname) > 0 {
					return fmt.Errorf("Cannot set hostname for '%s' when the 'address' has not been specified", d.Id())
				}
			}
			log.Printf("[INFO] Adding ip/MAC/host=%s/%s/%s to %s", address, netIface.Mac.Address, hostname, networkName)
			addHost(&network, address, netIface.Mac.Address, hostname)
		}

		netIface.Source.Network = networkName
		netIfaces = append(netIfaces, netIface)
	}

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

	log.Printf("[DEBUG] Creating libvirt domain with XML:\n%s", string(data))

	domain, err := virConn.DomainDefineXML(string(data))
	if err != nil {
		return fmt.Errorf("Error defining libvirt domain: %s", err)
	}

	err = domain.Create()
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

	err = waitForNetworkAddresses(netIfaces, domain)
	if err != nil {
		return err
	}

	return resourceLibvirtDomainRead(d, meta)
}

func resourceLibvirtDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	domain, err := virConn.LookupByUUIDString(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt domain: %s", err)
	}
	defer domain.Free()

	running, err := isDomainRunning(domain)
	if err != nil {
		return err
	}
	if !running {
		err = domain.Create()
		if err != nil {
			return fmt.Errorf("Error crearing libvirt domain: %s", err)
		}
	}

	return nil
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

	running, err := isDomainRunning(domain)
	if err != nil {
		return err
	}
	d.Set("running", running)

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
			return fmt.Errorf("Error retrieving volume for disk: %s", err)
		}

		disk := map[string]interface{}{
			"volume_id": virVolKey,
		}
		disks = append(disks, disk)
	}
	d.Set("disks", disks)

	// look interfaces with addresses
	ifacesWithAddr, err := domain.ListAllInterfaceAddresses(libvirt.VIR_DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
	if err != nil {
		switch err.(type) {
		default:
			return fmt.Errorf("Error retrieving interface addresses: %s", err)
		case libvirt.VirError:
			virErr := err.(libvirt.VirError)
			if virErr.Code != libvirt.VIR_ERR_OPERATION_INVALID || virErr.Domain != libvirt.VIR_FROM_QEMU {
				return fmt.Errorf("Error retrieving interface addresses: %s", err)
			}
		}
	}

	netIfaces := make([]map[string]interface{}, 0)
	for i, networkInterfaceDef := range domainDef.Devices.NetworkInterfaces {
		// we need it to read old values
		prefix := fmt.Sprintf("network_interface.%d", i)

		if networkInterfaceDef.Type != "network" {
			log.Printf("[DEBUG] ignoring interface of type '%s'", networkInterfaceDef.Type)
			continue
		}

		netIface := map[string]interface{}{
			"network_id": networkInterfaceDef.Source.Network,
			"mac":        networkInterfaceDef.Mac.Address,
		}

		log.Printf("[INFO] Reading interface %+v", netIface)

		netIfaceAddrs := make([]string, 0)
		// look for an ip address and try to match it with the mac address
		// not sure if using the target device name is a better idea here
		for _, ifaceWithAddr := range ifacesWithAddr {
			if ifaceWithAddr.Hwaddr == networkInterfaceDef.Mac.Address {
				for _, addr := range ifaceWithAddr.Addrs {
					netIfaceAddr := fmt.Sprintf("%s/%d", addr.Addr, addr.Prefix)
					netIfaceAddrs = append(netIfaceAddrs, netIfaceAddr)
				}
			}
		}

		log.Printf("[DEBUG] %d addresses for %s\n", len(netIfaceAddrs), networkInterfaceDef.Mac.Address)
		netIface["address"] = netIfaceAddrs

		// pass on old wait_for_lease value
		if waitForLease, ok := d.GetOk(prefix + ".wait_for_lease"); ok {
			netIface["wait_for_lease"] = waitForLease
		}

		netIfaces = append(netIfaces, netIface)
	}
	d.Set("network_interface", netIfaces)

	if len(ifacesWithAddr) > 0 {
		d.SetConnInfo(map[string]string{
			"type": "ssh",
			"host": ifacesWithAddr[0].Addrs[0].Addr,
		})
	}
	return nil
}

func resourceLibvirtDomainDelete(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	log.Printf("[DEBUG] Deleting domain %s", d.Id())

	domain, err := virConn.LookupByUUIDString(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt domain: %s", err)
	}
	defer domain.Free()

	state, err := domain.GetState()
	if err != nil {
		return fmt.Errorf("Couldn't get info about domain: %s", err)
	}

	if state[0] == libvirt.VIR_DOMAIN_RUNNING || state[0] == libvirt.VIR_DOMAIN_PAUSED {
		if err := domain.Destroy(); err != nil {
			return fmt.Errorf("Couldn't destroy libvirt domain: %s", err)
		}
	}

	if err := domain.Undefine(); err != nil {
		return fmt.Errorf("Couldn't undefine libvirt domain: %s", err)
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

func waitForNetworkAddresses(ifaces []defNetworkInterface, domain libvirt.VirDomain) error {
	log.Printf("[DEBUG] waiting for network addresses.\n")
	// wait for network interfaces with 'wait_for_lease' to get an address
	for _, iface := range ifaces {
		if !iface.waitForLease {
			continue
		}

		mac := strings.ToUpper(iface.Mac.Address)
		if mac == "" {
			log.Printf("[DEBUG] Can't wait without a mac address.\n")
			// we can't get the ip without a mac address
			continue
		}

		// loop until address appear, with timeout
		start := time.Now()

	waitLoop:
		for {
			log.Printf("[DEBUG] waiting for network address for interface with hwaddr: '%s'\n", iface.Mac.Address)
			ifacesWithAddr, err := domain.ListAllInterfaceAddresses(libvirt.VIR_DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
			if err != nil {
				return fmt.Errorf("Error retrieving interface addresses: %s", err)
			}
			log.Printf("[DEBUG] ifaces with addresses: %+v\n", ifacesWithAddr)

			for _, ifaceWithAddr := range ifacesWithAddr {
				// found
				if mac == strings.ToUpper(ifaceWithAddr.Hwaddr) {
					break waitLoop
				}
			}

			time.Sleep(1 * time.Second)
			if time.Since(start) > 5*time.Minute {
				return fmt.Errorf("Timeout waiting for interface addresses")
			}
		}
	}

	return nil
}

func isDomainRunning(domain libvirt.VirDomain) (bool, error) {
	state, err := domain.GetState()
	if err != nil {
		return false, fmt.Errorf("Couldn't get state of domain: %s", err)
	}

	return state[0] == libvirt.VIR_DOMAIN_RUNNING, nil
}

func getHostXMLDesc(ip, mac, name string) string {
	var b bytes.Buffer
	b.WriteString("<host ")
	if len(ip) > 0 {
		b.WriteString(fmt.Sprintf(" ip=\"%s\"", ip))
	}
	if len(mac) > 0 {
		b.WriteString(fmt.Sprintf(" mac=\"%s\"", mac))
	}
	if len(name) > 0 {
		b.WriteString(fmt.Sprintf(" name=\"%s\"", name))
	}
	b.WriteString(" />")
	return b.String()
}

// Adds a new static host to the network
func addHost(n *libvirt.VirNetwork, ip, mac, name string) error {
	return n.UpdateXMLDesc(getHostXMLDesc(ip, mac, name),
		libvirt.VIR_NETWORK_UPDATE_COMMAND_ADD_LAST, libvirt.VIR_NETWORK_SECTION_IP_DHCP_HOST)
}

// Removes a static host from the network
func removeHost(n *libvirt.VirNetwork, ip, mac, name string) error {
	return n.UpdateXMLDesc(getHostXMLDesc(ip, mac, name),
		libvirt.VIR_NETWORK_UPDATE_COMMAND_DELETE, libvirt.VIR_NETWORK_SECTION_IP_DHCP_HOST)
}
