package libvirt

import (
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
	"os"

	"github.com/davecgh/go-spew/spew"
	libvirt "github.com/dmacvicar/libvirt-go"
	"github.com/hashicorp/terraform/helper/schema"
)

func init() {
	spew.Config.Indent = "\t"
}

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
			"metadata": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: false,
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
			"firmware": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"nvram": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"running": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: false,
			},
			"cloudinit": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
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
				Elem: &schema.Resource{
					Schema: networkInterfaceCommonSchema(),
				},
			},
		},
	}
}

func resourceLibvirtDomainExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	log.Printf("[DEBUG] Check if resource libvirt_domain exists")

	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return false, fmt.Errorf("The libvirt connection was nil.")
	}
	domain, err := virConn.LookupByUUIDString(d.Id())
	defer domain.Free()
	return err == nil, err
}

func resourceLibvirtDomainCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Create resource libvirt_domain")

	// Ensure partial mode to save some relevant keys
	d.Partial(true)

	// the domain ID must always be saved, otherwise it won't be possible to cleanup a domain
	// if something bad happens at provisioning time
	d.SetPartial("id")

	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	domainDef := newDomainDef()
	if name, ok := d.GetOk("name"); ok {
		domainDef.Name = name.(string)
	}

	if metadata, ok := d.GetOk("metadata"); ok {
		domainDef.Metadata.TerraformLibvirt.Xml = metadata.(string)
	}

	if firmware, ok := d.GetOk("firmware"); ok {
		firmwareFile := firmware.(string)
		if _, err := os.Stat(firmwareFile); os.IsNotExist(err) {
			return fmt.Errorf("Could not find firmware file '%s'.", firmwareFile)
		}
		domainDef.Os.Loader = &defLoader{
			File:     firmwareFile,
			ReadOnly: "yes",
			Type:     "pflash",
		}

		if nvram, ok := d.GetOk("nvram"); ok {
			nvramFile := nvram.(string)
			if _, err := os.Stat(nvramFile); os.IsNotExist(err) {
				return fmt.Errorf("Could not find nvram file '%s'.", nvramFile)
			}
			domainDef.Os.NvRam = &defNvRam{
				File: nvramFile,
			}
		}
	}

	domainDef.Memory.Amount = d.Get("memory").(int)
	domainDef.VCpu.Amount = d.Get("vcpu").(int)

	disksCount := d.Get("disk.#").(int)
	var disks []defDisk
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

	type pendingMapping struct {
		mac      string
		hostname string
		network  *libvirt.VirNetwork
	}

	if cloudinit, ok := d.GetOk("cloudinit"); ok {
		cloudinitID, err := getCloudInitVolumeKeyFromTerraformID(cloudinit.(string))
		if err != nil {
			return err
		}
		disk, err := newDiskForCloudInit(virConn, cloudinitID)
		if err != nil {
			return err
		}
		disks = append(disks, disk)
	}

	netIfacesCount := d.Get("network_interface.#").(int)
	netIfaces := make([]defNetworkInterface, 0, netIfacesCount)
	partialNetIfaces := make(map[string]pendingMapping, netIfacesCount)
	for i := 0; i < netIfacesCount; i++ {
		prefix := fmt.Sprintf("network_interface.%d", i)

		netIface := defNetworkInterface{}
		netIface.Model.Type = "virtio"

		// calculate the MAC address
		var mac string
		if macI, ok := d.GetOk(prefix + ".mac"); ok {
			mac = strings.ToUpper(macI.(string))
		} else {
			var err error
			mac, err = RandomMACAddress()
			if err != nil {
				return fmt.Errorf("Error generating mac address: %s", err)
			}
		}
		netIface.Mac.Address = mac

		// this is not passed to libvirt, but used by waitForAddress
		netIface.waitForLease = false
		if waitForLease, ok := d.GetOk(prefix + ".wait_for_lease"); ok {
			netIface.waitForLease = waitForLease.(bool)
		}

		// connect to the interface to the network... first, look for the network
		if n, ok := d.GetOk(prefix + ".network_name"); ok {
			// when using a "network_name" we do not try to do anything: we just
			// connect to that network
			netIface.Type = "network"
			netIface.Source.Network = n.(string)
		} else if networkUUID, ok := d.GetOk(prefix + ".network_id"); ok {
			// when using a "network_id" we are referring to a "network resource"
			// we have defined somewhere else...
			network, err := virConn.LookupNetworkByUUIDString(networkUUID.(string))
			if err != nil {
				return fmt.Errorf("Can't retrieve network ID %s", networkUUID)
			}
			defer network.Free()

			networkName, err := network.GetName()
			if err != nil {
				return fmt.Errorf("Error retrieving network name: %s", err)
			}
			networkDef, err := newDefNetworkfromLibvirt(&network)
			if !networkDef.HasDHCP() {
				continue
			}

			hostname := domainDef.Name
			if hostnameI, ok := d.GetOk(prefix + ".hostname"); ok {
				hostname = hostnameI.(string)
			}
			if addresses, ok := d.GetOk(prefix + ".addresses"); ok {
				// some IP(s) provided
				for _, addressI := range addresses.([]interface{}) {
					address := addressI.(string)
					ip := net.ParseIP(address)
					if ip == nil {
						return fmt.Errorf("Could not parse addresses '%s'", address)
					}
					// TODO: we should check the IP is contained in the DHCP addresses served
					log.Printf("[INFO] Adding IP/MAC/host=%s/%s/%s to %s", ip.String(), mac, hostname, networkName)
					if err := addHost(&network, ip.String(), mac, hostname); err != nil {
						return err
					}
				}
			} else {
				// no IPs provided: if the hostname has been provided, wait until we get an IP
				if len(hostname) > 0 {
					if !netIface.waitForLease {
						return fmt.Errorf("Cannot map '%s': we are not waiting for lease and no IP has been provided", hostname)
					}
					// the resource specifies a hostname but not an IP, so we must wait until we
					// have a valid lease and then read the IP we have been assigned, so we can
					// do the mapping
					log.Printf("[DEBUG] Will wait for an IP for hostname '%s'...", hostname)
					partialNetIfaces[strings.ToUpper(mac)] = pendingMapping{
						mac:      strings.ToUpper(mac),
						hostname: hostname,
						network:  &network,
					}
				} else {
					// neither an IP or a hostname has been provided: so nothing must be forced
				}
			}
			netIface.Type = "network"
			netIface.Source.Network = networkName
		} else if bridgeNameI, ok := d.GetOk(prefix + ".bridge"); ok {
			netIface.Type = "bridge"
			netIface.Source.Bridge = bridgeNameI.(string)
		} else if devI, ok := d.GetOk(prefix + ".vepa"); ok {
			netIface.Type = "direct"
			netIface.Source.Dev = devI.(string)
			netIface.Source.Mode = "vepa"
		} else if devI, ok := d.GetOk(prefix + ".macvtap"); ok {
			netIface.Type = "direct"
			netIface.Source.Dev = devI.(string)
			netIface.Source.Mode = "bridge"
		} else if devI, ok := d.GetOk(prefix + ".passthrough"); ok {
			netIface.Type = "direct"
			netIface.Source.Dev = devI.(string)
			netIface.Source.Mode = "passthrough"
		} else {
			// no network has been specified: we are on our own
		}

		netIfaces = append(netIfaces, netIface)
	}

	domainDef.Devices.Disks = disks
	domainDef.Devices.NetworkInterfaces = netIfaces

	connectURI, err := virConn.GetURI()
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt connection URI: %s", err)
	}
	log.Printf("[INFO] Creating libvirt domain at %s", connectURI)

	data, err := xmlMarshallIndented(domainDef)
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
		return fmt.Errorf("Error creating libvirt domain: %s", err)
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

	err = resourceLibvirtDomainRead(d, meta)
	if err != nil {
		return err
	}

	// we must read devices again in order to set some missing ip/MAC/host mappings
	for i := 0; i < netIfacesCount; i++ {
		prefix := fmt.Sprintf("network_interface.%d", i)

		mac := strings.ToUpper(d.Get(prefix + ".mac").(string))

		// if we were waiting for an IP address for this MAC, go ahead.
		if pending, ok := partialNetIfaces[mac]; ok {
			// we should have the address now
			if addressesI, ok := d.GetOk(prefix + ".addresses"); !ok {
				return fmt.Errorf("Did not obtain the IP address for MAC=%s", mac)
			} else {
				for _, addressI := range addressesI.([]interface{}) {
					address := addressI.(string)
					log.Printf("[INFO] Finally adding IP/MAC/host=%s/%s/%s", address, mac, pending.hostname)
					addHost(pending.network, address, mac, pending.hostname)
					if err != nil {
						return fmt.Errorf("Could not add IP/MAC/host=%s/%s/%s: %s", address, mac, pending.hostname, err)
					}
				}
			}
		}
	}

	return nil
}

func resourceLibvirtDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Update resource libvirt_domain")

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
			return fmt.Errorf("Error creating libvirt domain: %s", err)
		}
	}

	if d.HasChange("metadata") {
		metadata := defMetadata{}
		metadata.Xml = d.Get("metadata").(string)
		metadataToXml, err := xml.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("Error serializing libvirt metadata: %s", err)
		}

		err = domain.SetMetadata(libvirt.VIR_DOMAIN_METADATA_ELEMENT,
			string(metadataToXml),
			"terraform-libvirt",
			"http://github.com/dmacvicar/terraform-provider-libvirt/",
			libvirt.VIR_DOMAIN_AFFECT_LIVE|libvirt.VIR_DOMAIN_AFFECT_CONFIG)
		if err != nil {
			return fmt.Errorf("Error changing domain metadata: %s", err)
		}
	}

	if d.HasChange("cloudinit") {
		cloudinit, err := newDiskForCloudInit(virConn, d.Get("cloudinit").(string))
		if err != nil {
			return err
		}

		data, err := xml.Marshal(cloudinit)
		if err != nil {
			return fmt.Errorf("Error serializing cloudinit disk: %s", err)
		}

		err = domain.UpdateDeviceFlags(
			string(data),
			libvirt.VIR_DOMAIN_AFFECT_CONFIG|libvirt.VIR_DOMAIN_AFFECT_CURRENT|libvirt.VIR_DOMAIN_AFFECT_LIVE)
		if err != nil {
			return fmt.Errorf("Error while changing the cloudinit volume: %s", err)
		}
	}

	netIfacesCount := d.Get("network_interface.#").(int)
	for i := 0; i < netIfacesCount; i++ {
		prefix := fmt.Sprintf("network_interface.%d", i)
		if d.HasChange(prefix+".hostname") || d.HasChange(prefix+".addresses") || d.HasChange(prefix+".mac") {
			networkUUID, ok := d.GetOk(prefix + ".network_id")
			if !ok {
				continue
			}
			network, err := virConn.LookupNetworkByUUIDString(networkUUID.(string))
			if err != nil {
				return fmt.Errorf("Can't retrieve network ID %s", networkUUID)
			}
			defer network.Free()

			networkName, err := network.GetName()
			if err != nil {
				return fmt.Errorf("Error retrieving volume name: %s", err)
			}
			hostname := d.Get(prefix + ".hostname").(string)
			mac := d.Get(prefix + ".mac").(string)
			addresses := d.Get(prefix + ".addresses")
			for _, addressI := range addresses.([]interface{}) {
				address := addressI.(string)
				ip := net.ParseIP(address)
				if ip == nil {
					return fmt.Errorf("Could not parse addresses '%s'", address)
				}
				log.Printf("[INFO] Updating IP/MAC/host=%s/%s/%s in '%s' network", ip.String(), mac, hostname, networkName)
				if err := updateHost(&network, ip.String(), mac, hostname); err != nil {
					return err
				}
			}
		}
	}

	// TODO
	return nil
}

func resourceLibvirtDomainRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Read resource libvirt_domain")

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

	log.Printf("[DEBUG] read: obtained XML desc for domain:\n%s", xmlDesc)

	domainDef := newDomainDef()
	err = xml.Unmarshal([]byte(xmlDesc), &domainDef)
	if err != nil {
		return fmt.Errorf("Error reading libvirt domain XML description: %s", err)
	}

	d.Set("name", domainDef.Name)
	d.Set("metadata", domainDef.Metadata.TerraformLibvirt.Xml)
	d.Set("vpu", domainDef.VCpu)
	d.Set("memory", domainDef.Memory)
	d.Set("firmware", domainDef.Os.Loader)
	d.Set("nvram", domainDef.Os.NvRam)

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
	ifacesWithAddr, err := getDomainInterfaces(&domain)
	if err != nil {
		return fmt.Errorf("Error retrieving interface addresses: %s", err)
	}

	addressesForMac := func(mac string) []string {
		// look for an ip address and try to match it with the mac address
		// not sure if using the target device name is a better idea here
		addrs := make([]string, 0)
		for _, ifaceWithAddr := range ifacesWithAddr {
			if strings.ToUpper(ifaceWithAddr.Hwaddr) == mac {
				for _, addr := range ifaceWithAddr.Addrs {
					addrs = append(addrs, addr.Addr)
				}
			}
		}
		return addrs
	}

	netIfaces := make([]map[string]interface{}, 0)
	for i, networkInterfaceDef := range domainDef.Devices.NetworkInterfaces {
		// we need it to read old values
		prefix := fmt.Sprintf("network_interface.%d", i)

		mac := strings.ToUpper(networkInterfaceDef.Mac.Address)
		netIface := map[string]interface{}{
			"network_id":     "",
			"network_name":   "",
			"bridge":         "",
			"vepa":           "",
			"macvtap":        "",
			"passthrough":    "",
			"mac":            mac,
			"hostname":       "",
			"wait_for_lease": false,
		}

		netIface["wait_for_lease"] = d.Get(prefix + ".wait_for_lease").(bool)
		netIface["addresses"] = addressesForMac(mac)
		log.Printf("[DEBUG] read: addresses for '%s': %+v", mac, netIface["addresses"])

		switch networkInterfaceDef.Type {
		case "network":
			{
				network, err := virConn.LookupNetworkByName(networkInterfaceDef.Source.Network)
				if err != nil {
					return fmt.Errorf("Can't retrieve network ID for '%s'", networkInterfaceDef.Source.Network)
				}
				defer network.Free()

				netIface["network_id"], err = network.GetUUIDString()
				if err != nil {
					return fmt.Errorf("Can't retrieve network ID for '%s'", networkInterfaceDef.Source.Network)
				}

				networkDef, err := newDefNetworkfromLibvirt(&network)
				if err != nil {
					return err
				}

				netIface["network_name"] = networkInterfaceDef.Source.Network

				// try to look for this MAC in the DHCP configuration for this VM
				if networkDef.HasDHCP() {
				hostnameSearch:
					for _, ip := range networkDef.Ips {
						if ip.Dhcp != nil {
							for _, host := range ip.Dhcp.Hosts {
								if strings.ToUpper(host.Mac) == netIface["mac"] {
									log.Printf("[DEBUG] read: hostname for '%s': '%s'", netIface["mac"], host.Name)
									netIface["hostname"] = host.Name
									break hostnameSearch
								}
							}
						}
					}
				}

			}
		case "bridge":
			netIface["bridge"] = networkInterfaceDef.Source.Bridge
		case "direct":
			{
				switch networkInterfaceDef.Source.Mode {
				case "vepa":
					netIface["vepa"] = networkInterfaceDef.Source.Dev
				case "bridge":
					netIface["macvtap"] = networkInterfaceDef.Source.Dev
				case "passthrough":
					netIface["passthrough"] = networkInterfaceDef.Source.Dev
				}
			}
		}
		netIfaces = append(netIfaces, netIface)
	}
	log.Printf("[DEBUG] read: ifaces for '%s':\n%s", domainDef.Name, spew.Sdump(netIfaces))
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
	log.Printf("[DEBUG] Delete resource libvirt_domain")

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

	if err := domain.UndefineFlags(libvirt.VIR_DOMAIN_UNDEFINE_NVRAM); err != nil {
		if e := err.(libvirt.VirError); e.Code == libvirt.VIR_ERR_NO_SUPPORT || e.Code == libvirt.VIR_ERR_INVALID_ARG {
			log.Printf("libvirt does not support undefine flags: will try again without flags")
			if err := domain.Undefine(); err != nil {
				return fmt.Errorf("Couldn't undefine libvirt domain: %s", err)
			}
		} else {
			return fmt.Errorf("Couldn't undefine libvirt domain with flags: %s", err)
		}
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
			ifacesWithAddr, err := getDomainInterfaces(&domain)
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

func newDiskForCloudInit(virConn *libvirt.VirConnection, volumeKey string) (defDisk, error) {
	disk := newCDROM()

	diskVolume, err := virConn.LookupStorageVolByKey(volumeKey)
	if err != nil {
		return disk, fmt.Errorf("Can't retrieve volume %s", volumeKey)
	}
	diskVolumeName, err := diskVolume.GetName()
	if err != nil {
		return disk, fmt.Errorf("Error retrieving volume name: %s", err)
	}
	diskPool, err := diskVolume.LookupPoolByVolume()
	if err != nil {
		return disk, fmt.Errorf("Error retrieving pool for volume: %s", err)
	}
	diskPoolName, err := diskPool.GetName()
	if err != nil {
		return disk, fmt.Errorf("Error retrieving pool name: %s", err)
	}

	disk.Source.Volume = diskVolumeName
	disk.Source.Pool = diskPoolName

	return disk, nil
}

func getDomainInterfaces(domain *libvirt.VirDomain) ([]libvirt.VirDomainInterface, error) {

	// get all the interfaces using the qemu-agent, this includes also
	// interfaces that are not attached to networks managed by libvirt
	// (eg. bridges, macvtap,...)
	interfaces := getDomainInterfacesViaQemuAgent(domain, true)
	if len(interfaces) > 0 {
		// the agent will always return all the interfaces, both the
		// ones managed by libvirt and the ones attached to bridge interfaces
		// or macvtap. Hence it has the highest priority
		return interfaces, nil
	}

	log.Print("[DEBUG] fetching networking interfaces using libvirt API")

	// get all the interfaces attached to libvirt networks
	interfaces, err := domain.ListAllInterfaceAddresses(libvirt.VIR_DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
	if err != nil {
		switch err.(type) {
		default:
			return interfaces, fmt.Errorf("Error retrieving interface addresses: %s", err)
		case libvirt.VirError:
			virErr := err.(libvirt.VirError)
			if virErr.Code != libvirt.VIR_ERR_OPERATION_INVALID || virErr.Domain != libvirt.VIR_FROM_QEMU {
				return interfaces, fmt.Errorf("Error retrieving interface addresses: %s", err)
			}
		}
	}
	log.Printf("[DEBUG] Interfaces: %s", spew.Sdump(interfaces))

	return interfaces, nil
}
