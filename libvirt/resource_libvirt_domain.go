package libvirt

import (
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform/helper/schema"
	libvirt "github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
)

type pendingMapping struct {
	mac      string
	hostname string
	network  *libvirt.Network
}

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
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"metadata": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"vcpu": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
				ForceNew: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  512,
				ForceNew: true,
			},
			"firmware": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"nvram": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"file": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"template": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"running": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: false,
				Required: false,
			},
			"cloudinit": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"coreos_ignition": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "",
			},
			"filesystem": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"accessmode": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "mapped",
						},
						"source": {
							Type:     schema.TypeString,
							Required: true,
						},
						"target": {
							Type:     schema.TypeString,
							Required: true,
						},
						"readonly": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},
			"disk": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"volume_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"file": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"scsi": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"wwn": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"network_interface": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"network_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"bridge": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"vepa": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"macvtap": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"passthrough": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"hostname": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"mac": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"wait_for_lease": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"addresses": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"graphics": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "spice",
						},
						"autoport": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"listen_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "none",
						},
					},
				},
			},
			"console": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"source_path": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"target_port": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"target_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"cpu": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"autostart": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"machine": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"arch": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"boot_device": {
				Type:     schema.TypeList,
				Optional: true,
				Required: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dev": {
							Type:     schema.TypeList,
							Optional: true,
							Required: false,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"emulator": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"kernel": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
			"initrd": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
			"cmdline": {
				Type:     schema.TypeList,
				Optional: true,
				Required: false,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
			},
			"qemu_agent": {
				Type:     schema.TypeBool,
				Optional: true,
				Required: false,
				Default:  false,
				ForceNew: false,
			},
			"xml": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
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

func resourceLibvirtDomainExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	log.Printf("[DEBUG] Check if resource libvirt_domain exists")

	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return false, fmt.Errorf(LibVirtConIsNil)
	}

	domain, err := virConn.LookupDomainByUUIDString(d.Id())
	if err != nil {
		if err.(libvirt.Error).Code == libvirt.ERR_NO_DOMAIN {
			return false, nil
		}
		return false, err
	}
	defer domain.Free()

	return true, nil
}

func resourceLibvirtDomainCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Create resource libvirt_domain")

	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	domainDef, err := newDomainDefForConnection(virConn, d)
	if err != nil {
		return err
	}

	if name, ok := d.GetOk("name"); ok {
		domainDef.Name = name.(string)
	}

	if cpuMode, ok := d.GetOk("cpu.mode"); ok {
		domainDef.CPU = &libvirtxml.DomainCPU{
			Mode: cpuMode.(string),
		}
	}

	domainDef.Memory = &libvirtxml.DomainMemory{
		Value: uint(d.Get("memory").(int)),
		Unit:  "MiB",
	}
	domainDef.VCPU = &libvirtxml.DomainVCPU{
		Value: d.Get("vcpu").(int),
	}

	domainDef.OS.Kernel = d.Get("kernel").(string)
	domainDef.OS.Initrd = d.Get("initrd").(string)
	domainDef.OS.Type.Arch = d.Get("arch").(string)
	domainDef.OS.Type.Machine = d.Get("machine").(string)
	domainDef.Devices.Emulator = d.Get("emulator").(string)

	arch, err := getHostArchitecture(virConn)
	if err != nil {
		return fmt.Errorf("Error retrieving host architecture: %s", err)
	}

	if err := setGraphics(d, &domainDef, arch); err != nil {
		return err
	}

	setConsoles(d, &domainDef)
	setCmdlineArgs(d, &domainDef)
	setFirmware(d, &domainDef)
	setBootDevices(d, &domainDef)

	if err := setCoreOSIgnition(d, &domainDef); err != nil {
		return err
	}

	if err := setDisks(d, &domainDef, virConn); err != nil {
		return err
	}

	if err := setFilesystems(d, &domainDef); err != nil {
		return err
	}

	if err := setCloudinit(d, &domainDef, virConn); err != nil {
		return err
	}

	var waitForLeases []*libvirtxml.DomainInterface
	partialNetIfaces := make(map[string]*pendingMapping, d.Get("network_interface.#").(int))

	if err := setNetworkInterfaces(d, &domainDef, virConn, partialNetIfaces, &waitForLeases); err != nil {
		return err
	}

	connectURI, err := virConn.GetURI()
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt connection URI: %s", err)
	}
	log.Printf("[INFO] Creating libvirt domain at %s", connectURI)

	data, err := xmlMarshallIndented(domainDef)
	if err != nil {
		return fmt.Errorf("Error serializing libvirt domain: %s", err)
	}
	log.Printf("[DEBUG] Generated XML for libvirt domain:\n%s", data)

	data, err = transformResourceXML(data, d)
	if err != nil {
		return fmt.Errorf("Error applying XSLT stylesheet: %s", err)
	}

	domain, err := virConn.DomainDefineXML(data)
	if err != nil {
		return fmt.Errorf("Error defining libvirt domain: %s", err)
	}

	if autostart, ok := d.GetOk("autostart"); ok {
		err = domain.SetAutostart(autostart.(bool))
		if err != nil {
			return fmt.Errorf("Error setting autostart for domain: %s", err)
		}
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

	// the domain ID must always be saved, otherwise it won't be possible to cleanup a domain
	// if something bad happens at provisioning time
	d.Partial(true)
	d.Set("id", id)
	d.SetPartial("id")
	d.Partial(false)

	log.Printf("[INFO] Domain ID: %s", d.Id())

	if len(waitForLeases) > 0 {
		err = domainWaitForLeases(domain, waitForLeases, d.Timeout(schema.TimeoutCreate), d)
		if err != nil {
			return err
		}
	}

	err = resourceLibvirtDomainRead(d, meta)
	if err != nil {
		return err
	}

	// we must read devices again in order to set some missing ip/MAC/host mappings
	for i := 0; i < d.Get("network_interface.#").(int); i++ {
		prefix := fmt.Sprintf("network_interface.%d", i)

		mac := strings.ToUpper(d.Get(prefix + ".mac").(string))

		// if we were waiting for an IP address for this MAC, go ahead.
		if pending, ok := partialNetIfaces[mac]; ok {
			// we should have the address now
			addressesI, ok := d.GetOk(prefix + ".addresses")
			if !ok {
				return fmt.Errorf("Did not obtain the IP address for MAC=%s", mac)
			}
			for _, addressI := range addressesI.([]interface{}) {
				address := addressI.(string)
				log.Printf("[INFO] Finally adding IP/MAC/host=%s/%s/%s", address, mac, pending.hostname)
				updateOrAddHost(pending.network, address, mac, pending.hostname)
				if err != nil {
					return fmt.Errorf("Could not add IP/MAC/host=%s/%s/%s: %s", address, mac, pending.hostname, err)
				}
			}
		}
	}

	destroyDomainByUserRequest(d, domain)
	return nil
}

func resourceLibvirtDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Update resource libvirt_domain")

	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}
	domain, err := virConn.LookupDomainByUUIDString(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt domain: %s", err)
	}
	defer domain.Free()

	domainRunningNow, err := domainIsRunning(*domain)
	if err != nil {
		return err
	}
	if !domainRunningNow {
		err = domain.Create()
		if err != nil {
			return fmt.Errorf("Error creating libvirt domain: %s", err)
		}
	}

	d.Partial(true)

	if d.HasChange("cloudinit") {
		cloudinitID, err := getCloudInitVolumeKeyFromTerraformID(d.Get("cloudinit").(string))
		if err != nil {
			return err
		}
		disk, err := newDiskForCloudInit(virConn, cloudinitID)
		if err != nil {
			return err
		}

		data, err := xml.Marshal(disk)
		if err != nil {
			return fmt.Errorf("Error serializing cloudinit disk: %s", err)
		}

		err = domain.UpdateDeviceFlags(
			string(data),
			libvirt.DOMAIN_DEVICE_MODIFY_CONFIG|libvirt.DOMAIN_DEVICE_MODIFY_CURRENT|libvirt.DOMAIN_DEVICE_MODIFY_LIVE)
		if err != nil {
			return fmt.Errorf("Error while changing the cloudinit volume: %s", err)
		}

		d.SetPartial("cloudinit")
	}

	if d.HasChange("autostart") {
		err = domain.SetAutostart(d.Get("autostart").(bool))
		if err != nil {
			return fmt.Errorf("Error setting autostart for domain: %s", err)
		}
		d.SetPartial("autostart")
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
				return fmt.Errorf("Error retrieving network name: %s", err)
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
				if err := updateOrAddHost(network, ip.String(), mac, hostname); err != nil {
					return err
				}
			}
		}
	}

	d.Partial(false)

	return nil
}

func resourceLibvirtDomainRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Read resource libvirt_domain")

	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	domain, err := virConn.LookupDomainByUUIDString(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt domain: %s", err)
	}
	defer domain.Free()

	xmlDesc, err := domain.GetXMLDesc(0)
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt domain XML description: %s", err)
	}

	log.Printf("[DEBUG] read: obtained XML desc for domain:\n%s", xmlDesc)

	domainDef, err := newDomainDefForConnection(virConn, d)
	if err != nil {
		return err
	}

	err = xml.Unmarshal([]byte(xmlDesc), &domainDef)
	if err != nil {
		return fmt.Errorf("Error reading libvirt domain XML description: %s", err)
	}

	autostart, err := domain.GetAutostart()
	if err != nil {
		return fmt.Errorf("Error reading domain autostart setting: %s", err)
	}

	d.Set("name", domainDef.Name)
	d.Set("vcpu", domainDef.VCPU)
	d.Set("memory", domainDef.Memory)
	d.Set("firmware", domainDef.OS.Loader)
	d.Set("nvram", domainDef.OS.NVRam)
	d.Set("cpu", domainDef.CPU)
	d.Set("arch", domainDef.OS.Type.Arch)
	d.Set("autostart", autostart)

	cmdLines, err := splitKernelCmdLine(domainDef.OS.Cmdline)
	if err != nil {
		return err
	}
	d.Set("cmdline", cmdLines)
	d.Set("kernel", domainDef.OS.Kernel)
	d.Set("initrd", domainDef.OS.Initrd)

	caps, err := getHostCapabilities(virConn)
	if err != nil {
		return err
	}
	machine, err := getOriginalMachineName(caps, domainDef.OS.Type.Arch, domainDef.OS.Type.Type,
		domainDef.OS.Type.Machine)
	if err != nil {
		return err
	}
	d.Set("machine", machine)

	// Emulator is the same as the default don't set it in domainDef
	// or it will show as changed
	d.Set("emulator", domainDef.Devices.Emulator)
	var (
		disks []map[string]interface{}
		disk  map[string]interface{}
	)
	for _, diskDef := range domainDef.Devices.Disks {
		// network drives do not have a volume associated
		if diskDef.Source.Network != nil {
			if len(diskDef.Source.Network.Hosts) < 1 {
				return fmt.Errorf("Network disk does not contain any hosts")
			}
			url, err := url.Parse(fmt.Sprintf("%s://%s:%s%s",
				diskDef.Source.Network.Protocol,
				diskDef.Source.Network.Hosts[0].Name,
				diskDef.Source.Network.Hosts[0].Port,
				diskDef.Source.Network.Name))
			if err != nil {
				return err
			}
			disk = map[string]interface{}{
				"url": url.String(),
			}
		} else if diskDef.Device == "cdrom" {
			disk = map[string]interface{}{
				"file": diskDef.Source.File,
			}
		} else if diskDef.Source.File != nil {
			// LEGACY way of handling volumes using "file", which we replaced
			// by the diskdef.Source.Volume once we realized it existed.
			// This code will be removed in future versions of the provider.
			virVol, err := virConn.LookupStorageVolByPath(diskDef.Source.File.File)
			if err != nil {
				return fmt.Errorf("Error retrieving volume for disk: %s", err)
			}
			defer virVol.Free()

			virVolKey, err := virVol.GetKey()
			if err != nil {
				return fmt.Errorf("Error retrieving volume for disk: %s", err)
			}

			disk = map[string]interface{}{
				"volume_id": virVolKey,
			}
		} else {
			pool, err := virConn.LookupStoragePoolByName(diskDef.Source.Volume.Pool)
			if err != nil {
				return fmt.Errorf("Error retrieving pool for disk: %s", err)
			}
			defer pool.Free()

			virVol, err := pool.LookupStorageVolByName(diskDef.Source.Volume.Volume)
			if err != nil {
				return fmt.Errorf("Error retrieving volume for disk: %s", err)
			}
			defer virVol.Free()

			virVolKey, err := virVol.GetKey()
			if err != nil {
				return fmt.Errorf("Error retrieving volume key for disk: %s", err)
			}

			disk = map[string]interface{}{
				"volume_id": virVolKey,
			}
		}

		disks = append(disks, disk)
	}
	d.Set("disks", disks)
	var filesystems []map[string]interface{}
	for _, fsDef := range domainDef.Devices.Filesystems {
		fs := map[string]interface{}{
			"accessmode": fsDef.AccessMode,
			"source":     fsDef.Source.Mount.Dir,
			"target":     fsDef.Target.Dir,
			"readonly":   fsDef.ReadOnly,
		}
		filesystems = append(filesystems, fs)
	}
	d.Set("filesystems", filesystems)

	// lookup interfaces with addresses
	ifacesWithAddr, err := domainGetIfacesInfo(*domain, d)
	if err != nil {
		return fmt.Errorf("Error retrieving interface addresses: %s", err)
	}

	addressesForMac := func(mac string) []string {
		// look for an ip address and try to match it with the mac address
		// not sure if using the target device name is a better idea here
		var addrs []string
		for _, ifaceWithAddr := range ifacesWithAddr {
			if strings.ToUpper(ifaceWithAddr.Hwaddr) == mac {
				for _, addr := range ifaceWithAddr.Addrs {
					addrs = append(addrs, addr.Addr)
				}
			}
		}
		return addrs
	}

	var netIfaces []map[string]interface{}
	for i, networkInterfaceDef := range domainDef.Devices.Interfaces {
		// we need it to read old values
		prefix := fmt.Sprintf("network_interface.%d", i)

		mac := strings.ToUpper(networkInterfaceDef.MAC.Address)
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

		if networkInterfaceDef.Source.Network != nil {
			network, err := virConn.LookupNetworkByName(networkInterfaceDef.Source.Network.Network)
			if err != nil {
				return fmt.Errorf("Can't retrieve network ID for '%s'", networkInterfaceDef.Source.Network.Network)
			}
			defer network.Free()

			netIface["network_id"], err = network.GetUUIDString()
			if err != nil {
				return fmt.Errorf("Can't retrieve network ID for '%s'", networkInterfaceDef.Source.Network.Network)
			}

			networkDef, err := getXMLNetworkDefFromLibvirt(network)
			if err != nil {
				return err
			}

			netIface["network_name"] = networkInterfaceDef.Source.Network.Network

			// try to look for this MAC in the DHCP configuration for this VM
			if HasDHCP(networkDef) {
			hostnameSearch:
				for _, ip := range networkDef.IPs {
					if ip.DHCP != nil {
						for _, host := range ip.DHCP.Hosts {
							if strings.ToUpper(host.MAC) == netIface["mac"] {
								log.Printf("[DEBUG] read: hostname for '%s': '%s'", netIface["mac"], host.Name)
								netIface["hostname"] = host.Name
								break hostnameSearch
							}
						}
					}
				}
			}
		} else if networkInterfaceDef.Source.Bridge != nil {
			netIface["bridge"] = networkInterfaceDef.Source.Bridge.Bridge
		} else if networkInterfaceDef.Source.Direct != nil {
			switch networkInterfaceDef.Source.Direct.Mode {
			case "vepa":
				netIface["vepa"] = networkInterfaceDef.Source.Direct.Dev
			case "bridge":
				netIface["macvtap"] = networkInterfaceDef.Source.Direct.Dev
			case "passthrough":
				netIface["passthrough"] = networkInterfaceDef.Source.Direct.Dev
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
		return fmt.Errorf(LibVirtConIsNil)
	}

	log.Printf("[DEBUG] Deleting domain %s", d.Id())

	domain, err := virConn.LookupDomainByUUIDString(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt domain: %s", err)
	}
	defer domain.Free()

	xmlDesc, err := domain.GetXMLDesc(0)
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt domain XML description: %s", err)
	}

	domainDef, err := newDomainDefForConnection(virConn, d)
	if err != nil {
		return err
	}

	err = xml.Unmarshal([]byte(xmlDesc), &domainDef)
	if err != nil {
		return fmt.Errorf("Error reading libvirt domain XML description: %s", err)
	}

	state, _, err := domain.GetState()
	if err != nil {
		return fmt.Errorf("Couldn't get info about domain: %s", err)
	}

	if state == libvirt.DOMAIN_RUNNING || state == libvirt.DOMAIN_PAUSED {
		if err := domain.Destroy(); err != nil {
			return fmt.Errorf("Couldn't destroy libvirt domain: %s", err)
		}
	}

	if err := domain.UndefineFlags(libvirt.DOMAIN_UNDEFINE_NVRAM); err != nil {
		if e := err.(libvirt.Error); e.Code == libvirt.ERR_NO_SUPPORT || e.Code == libvirt.ERR_INVALID_ARG {
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

func newDiskForCloudInit(virConn *libvirt.Connect, volumeKey string) (libvirtxml.DomainDisk, error) {
	disk := libvirtxml.DomainDisk{
		Device: "cdrom",
		Target: &libvirtxml.DomainDiskTarget{
			Dev: "hda",
			Bus: "ide",
		},
		Driver: &libvirtxml.DomainDiskDriver{
			Name: "qemu",
			Type: "raw",
		},
	}

	diskVolume, err := virConn.LookupStorageVolByKey(volumeKey)
	if err != nil {
		return disk, fmt.Errorf("Can't retrieve volume %s: %v", volumeKey, err)
	}
	diskVolumeFile, err := diskVolume.GetPath()
	if err != nil {
		return disk, fmt.Errorf("Error retrieving volume file: %s", err)
	}

	disk.Source = &libvirtxml.DomainDiskSource{
		File: &libvirtxml.DomainDiskSourceFile{
			File: diskVolumeFile,
		},
	}

	return disk, nil
}

func setCoreOSIgnition(d *schema.ResourceData, domainDef *libvirtxml.Domain) error {
	if ignition, ok := d.GetOk("coreos_ignition"); ok {
		ignitionKey, err := getIgnitionVolumeKeyFromTerraformID(ignition.(string))
		if err != nil {
			return err
		}

		domainDef.QEMUCommandline = &libvirtxml.DomainQEMUCommandline{
			Args: []libvirtxml.DomainQEMUCommandlineArg{
				{
					Value: "-fw_cfg",
				},
				{
					Value: fmt.Sprintf("name=opt/com.coreos/config,file=%s", ignitionKey),
				},
			},
		}
	}

	return nil
}

func setGraphics(d *schema.ResourceData, domainDef *libvirtxml.Domain, arch string) error {
	if arch == "s390x" || arch == "ppc64" {
		domainDef.Devices.Graphics = nil
		return nil
	}

	prefix := "graphics.0"
	if _, ok := d.GetOk(prefix); ok {
		domainDef.Devices.Graphics = []libvirtxml.DomainGraphic{{}}
		graphicsType, ok := d.GetOk(prefix + ".type")
		if !ok {
			return fmt.Errorf("Missing graphics type for domain")
		}

		autoport := d.Get(prefix + ".autoport").(bool)
		listener := libvirtxml.DomainGraphicListener{}

		if listenType, ok := d.GetOk(prefix + ".listen_type"); ok {
			switch listenType {
			case "address":
				listener.Address = &libvirtxml.DomainGraphicListenerAddress{}
			case "network":
				listener.Network = &libvirtxml.DomainGraphicListenerNetwork{}
			case "socket":
				listener.Socket = &libvirtxml.DomainGraphicListenerSocket{}
			}
		} else {
			listenType = "none"
		}

		switch graphicsType {
		case "spice":
			domainDef.Devices.Graphics[0] = libvirtxml.DomainGraphic{
				Spice: &libvirtxml.DomainGraphicSpice{},
			}
			domainDef.Devices.Graphics[0].Spice.AutoPort = formatBoolYesNo(autoport)
			domainDef.Devices.Graphics[0].Spice.Listeners = []libvirtxml.DomainGraphicListener{
				listener,
			}
		case "vnc":
			domainDef.Devices.Graphics[0] = libvirtxml.DomainGraphic{
				VNC: &libvirtxml.DomainGraphicVNC{},
			}
			domainDef.Devices.Graphics[0].VNC.AutoPort = formatBoolYesNo(autoport)
			domainDef.Devices.Graphics[0].VNC.Listeners = []libvirtxml.DomainGraphicListener{
				listener,
			}
		default:
			return fmt.Errorf("This provider only supports vnc/spice as graphics type. Provided: '%s'", graphicsType)
		}
	}
	return nil
}

func setCmdlineArgs(d *schema.ResourceData, domainDef *libvirtxml.Domain) {
	var cmdlineArgs []string
	for i := 0; i < d.Get("cmdline.#").(int); i++ {
		for k, v := range d.Get(fmt.Sprintf("cmdline.%d", i)).(map[string]interface{}) {
			var cmd string
			if k == "_" {
				// keyless cmd (eg: nosplash)
				cmd = fmt.Sprintf("%v", v)
			} else {
				cmd = fmt.Sprintf("%s=%v", k, v)
			}
			cmdlineArgs = append(cmdlineArgs, cmd)
		}
	}
	sort.Strings(cmdlineArgs)
	domainDef.OS.Cmdline = strings.Join(cmdlineArgs, " ")
}

func setFirmware(d *schema.ResourceData, domainDef *libvirtxml.Domain) error {
	if firmware, ok := d.GetOk("firmware"); ok {
		firmwareFile := firmware.(string)
		domainDef.OS.Loader = &libvirtxml.DomainLoader{
			Path:     firmwareFile,
			Readonly: "yes",
			Type:     "pflash",
			Secure:   "no",
		}

		if _, ok := d.GetOk("nvram.0"); ok {
			nvramFile := d.Get("nvram.0.file").(string)
			nvramTemplateFile := ""
			if nvramTemplate, ok := d.GetOk("nvram.0.template"); ok {
				nvramTemplateFile = nvramTemplate.(string)
			}
			domainDef.OS.NVRam = &libvirtxml.DomainNVRam{
				NVRam:    nvramFile,
				Template: nvramTemplateFile,
			}
		}
	}

	return nil
}

func setBootDevices(d *schema.ResourceData, domainDef *libvirtxml.Domain) {
	for i := 0; i < d.Get("boot_device.#").(int); i++ {
		if bootMap, ok := d.GetOk(fmt.Sprintf("boot_device.%d.dev", i)); ok {
			for _, dev := range bootMap.([]interface{}) {
				domainDef.OS.BootDevices = append(domainDef.OS.BootDevices,
					libvirtxml.DomainBootDevice{
						Dev: dev.(string),
					})
			}
		}
	}
}

func setConsoles(d *schema.ResourceData, domainDef *libvirtxml.Domain) {
	for i := 0; i < d.Get("console.#").(int); i++ {
		console := libvirtxml.DomainConsole{}
		prefix := fmt.Sprintf("console.%d", i)
		consoleTargetPortInt, err := strconv.Atoi(d.Get(prefix + ".target_port").(string))
		if err == nil {
			consoleTargetPort := uint(consoleTargetPortInt)
			console.Target = &libvirtxml.DomainConsoleTarget{
				Port: &consoleTargetPort,
			}
		}
		if sourcePath, ok := d.GetOk(prefix + ".source_path"); ok {
			console.Source = &libvirtxml.DomainChardevSource{
				Dev: &libvirtxml.DomainChardevSourceDev{
					Path: sourcePath.(string),
				},
			}
		}
		if targetType, ok := d.GetOk(prefix + ".target_type"); ok {
			if console.Target == nil {
				console.Target = &libvirtxml.DomainConsoleTarget{}
			}
			console.Target.Type = targetType.(string)
		}
		domainDef.Devices.Consoles = append(domainDef.Devices.Consoles, console)
	}
}

func setDisks(d *schema.ResourceData, domainDef *libvirtxml.Domain, virConn *libvirt.Connect) error {
	var scsiDisk = false
	for i := 0; i < d.Get("disk.#").(int); i++ {
		disk := newDefDisk(i)

		prefix := fmt.Sprintf("disk.%d", i)
		if d.Get(prefix + ".scsi").(bool) {
			disk.Target.Bus = "scsi"
			scsiDisk = true
			if wwn, ok := d.GetOk(prefix + ".wwn"); ok {
				disk.WWN = wwn.(string)
			} else {
				disk.WWN = randomWWN(10)
			}
		}

		if volumeKey, ok := d.GetOk(prefix + ".volume_id"); ok {
			diskVolume, err := virConn.LookupStorageVolByKey(volumeKey.(string))
			if err != nil {
				return fmt.Errorf("Can't retrieve volume %s: %v", volumeKey.(string), err)
			}

			diskVolumeName, err := diskVolume.GetName()
			if err != nil {
				return fmt.Errorf("Can't retrieve name for volume %s", volumeKey.(string))
			}

			diskPool, err := diskVolume.LookupPoolByVolume()
			if err != nil {
				return fmt.Errorf("Can't retrieve pool for volume %s", volumeKey.(string))
			}

			diskPoolName, err := diskPool.GetName()
			if err != nil {
				return fmt.Errorf("Can't retrieve name for pool of volume %s", volumeKey.(string))
			}

			// find out the format of the volume in order to set the appropriate
			// driver
			volumeDef, err := newDefVolumeFromLibvirt(diskVolume)
			if err != nil {
				return err
			}
			if volumeDef.Target != nil && volumeDef.Target.Format != nil && volumeDef.Target.Format.Type != "" {
				if volumeDef.Target.Format.Type == "qcow2" {
					log.Print("[DEBUG] Setting disk driver to 'qcow2' to match disk volume format")
					disk.Driver = &libvirtxml.DomainDiskDriver{
						Name: "qemu",
						Type: "qcow2",
					}
				}
				if volumeDef.Target.Format.Type == "raw" {
					log.Print("[DEBUG] Setting disk driver to 'raw' to match disk volume format")
					disk.Driver = &libvirtxml.DomainDiskDriver{
						Name: "qemu",
						Type: "raw",
					}
				}
			} else {
				log.Printf("[WARN] Disk volume has no format specified: %s", volumeKey.(string))
			}

			disk.Source = &libvirtxml.DomainDiskSource{
				Volume: &libvirtxml.DomainDiskSourceVolume{
					Pool:   diskPoolName,
					Volume: diskVolumeName,
				},
			}
		} else if rawURL, ok := d.GetOk(prefix + ".url"); ok {
			// Support for remote, read-only http disks
			// useful for booting CDs
			url, err := url.Parse(rawURL.(string))
			if err != nil {
				return err
			}

			disk.Source = &libvirtxml.DomainDiskSource{
				Network: &libvirtxml.DomainDiskSourceNetwork{
					Protocol: url.Scheme,
					Name:     url.Path,
					Hosts: []libvirtxml.DomainDiskSourceHost{
						{
							Name: url.Hostname(),
							Port: url.Port(),
						},
					},
				},
			}

			if strings.HasSuffix(url.Path, ".iso") {
				disk.Device = "cdrom"
			}
			if !strings.HasSuffix(url.Path, ".qcow2") {
				disk.Driver.Type = "raw"
			}
		} else if file, ok := d.GetOk(prefix + ".file"); ok {
			// support for local disks, e.g. CDs
			disk.Source = &libvirtxml.DomainDiskSource{
				File: &libvirtxml.DomainDiskSourceFile{
					File: file.(string),
				},
			}

			if strings.HasSuffix(file.(string), ".iso") {
				disk.Device = "cdrom"
				disk.Target = &libvirtxml.DomainDiskTarget{
					Dev: "hda",
					Bus: "ide",
				}
				disk.Driver = &libvirtxml.DomainDiskDriver{
					Name: "qemu",
					Type: "raw",
				}
			}
		}

		domainDef.Devices.Disks = append(domainDef.Devices.Disks, disk)
	}

	log.Printf("[DEBUG] scsiDisk: %t", scsiDisk)
	if scsiDisk {
		domainDef.Devices.Controllers = append(domainDef.Devices.Controllers,
			libvirtxml.DomainController{
				Type:  "scsi",
				Model: "virtio-scsi",
			})
	}

	return nil
}

func setFilesystems(d *schema.ResourceData, domainDef *libvirtxml.Domain) error {
	for i := 0; i < d.Get("filesystem.#").(int); i++ {
		fs := newFilesystemDef()

		prefix := fmt.Sprintf("filesystem.%d", i)
		if accessMode, ok := d.GetOk(prefix + ".accessmode"); ok {
			fs.AccessMode = accessMode.(string)
		}
		if sourceDir, ok := d.GetOk(prefix + ".source"); ok {
			fs.Source = &libvirtxml.DomainFilesystemSource{
				Mount: &libvirtxml.DomainFilesystemSourceMount{
					Dir: sourceDir.(string),
				},
			}
		} else {
			return fmt.Errorf("Filesystem entry must have a 'source' set")
		}
		if targetDir, ok := d.GetOk(prefix + ".target"); ok {
			fs.Target = &libvirtxml.DomainFilesystemTarget{
				Dir: targetDir.(string),
			}
		} else {
			return fmt.Errorf("Filesystem entry must have a 'target' set")
		}
		if d.Get(prefix + ".readonly").(bool) {
			fs.ReadOnly = &libvirtxml.DomainFilesystemReadOnly{}
		} else {
			fs.ReadOnly = nil
		}

		domainDef.Devices.Filesystems = append(domainDef.Devices.Filesystems, fs)
	}
	log.Printf("filesystems: %+v\n", domainDef.Devices.Filesystems)
	return nil
}

func setCloudinit(d *schema.ResourceData, domainDef *libvirtxml.Domain, virConn *libvirt.Connect) error {
	if cloudinit, ok := d.GetOk("cloudinit"); ok {
		cloudinitID, err := getCloudInitVolumeKeyFromTerraformID(cloudinit.(string))
		if err != nil {
			return err
		}
		disk, err := newDiskForCloudInit(virConn, cloudinitID)
		if err != nil {
			return err
		}
		domainDef.Devices.Disks = append(domainDef.Devices.Disks, disk)
	}

	return nil
}

func setNetworkInterfaces(d *schema.ResourceData, domainDef *libvirtxml.Domain,
	virConn *libvirt.Connect, partialNetIfaces map[string]*pendingMapping,
	waitForLeases *[]*libvirtxml.DomainInterface) error {
	for i := 0; i < d.Get("network_interface.#").(int); i++ {
		prefix := fmt.Sprintf("network_interface.%d", i)

		netIface := libvirtxml.DomainInterface{
			Model: &libvirtxml.DomainInterfaceModel{
				Type: "virtio",
			},
		}

		// calculate the MAC address
		var mac string
		if macI, ok := d.GetOk(prefix + ".mac"); ok {
			mac = strings.ToUpper(macI.(string))
		} else {
			var err error
			mac, err = randomMACAddress()
			if err != nil {
				return fmt.Errorf("Error generating mac address: %s", err)
			}
		}
		netIface.MAC = &libvirtxml.DomainInterfaceMAC{
			Address: mac,
		}

		// this is not passed to libvirt, but used by waitForAddress
		if waitForLease, ok := d.GetOk(prefix + ".wait_for_lease"); ok {
			if waitForLease.(bool) {
				*waitForLeases = append(*waitForLeases, &netIface)
			}
		}

		// connect to the interface to the network... first, look for the network
		if n, ok := d.GetOk(prefix + ".network_name"); ok {
			// when using a "network_name" we do not try to do anything: we just
			// connect to that network
			netIface.Source = &libvirtxml.DomainInterfaceSource{
				Network: &libvirtxml.DomainInterfaceSourceNetwork{
					Network: n.(string),
				},
			}
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
			networkDef, err := getXMLNetworkDefFromLibvirt(network)

			// only for DHCP, we update the host table of the network
			if HasDHCP(networkDef) {
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

						log.Printf("[INFO] Adding IP/MAC/host=%s/%s/%s to %s", ip.String(), mac, hostname, networkName)
						if err := updateOrAddHost(network, ip.String(), mac, hostname); err != nil {
							return err
						}
					}
				} else {
					// no IPs provided: if the hostname has been provided, wait until we get an IP
					wait := false
					for _, iface := range *waitForLeases {
						if iface == &netIface {
							wait = true
							break
						}
					}
					if !wait {
						return fmt.Errorf("Cannot map '%s': we are not waiting for DHCP lease and no IP has been provided", hostname)
					}
					// the resource specifies a hostname but not an IP, so we must wait until we
					// have a valid lease and then read the IP we have been assigned, so we can
					// do the mapping
					log.Printf("[DEBUG] Do not have an IP for '%s' yet: will wait until DHCP provides one...", hostname)
					partialNetIfaces[strings.ToUpper(mac)] = &pendingMapping{
						mac:      strings.ToUpper(mac),
						hostname: hostname,
						network:  network,
					}
				}
			}

			netIface.Source = &libvirtxml.DomainInterfaceSource{
				Network: &libvirtxml.DomainInterfaceSourceNetwork{
					Network: networkName,
				},
			}
		} else if bridgeNameI, ok := d.GetOk(prefix + ".bridge"); ok {
			netIface.Source = &libvirtxml.DomainInterfaceSource{
				Bridge: &libvirtxml.DomainInterfaceSourceBridge{
					Bridge: bridgeNameI.(string),
				},
			}
		} else if devI, ok := d.GetOk(prefix + ".vepa"); ok {
			netIface.Source = &libvirtxml.DomainInterfaceSource{
				Direct: &libvirtxml.DomainInterfaceSourceDirect{
					Dev:  devI.(string),
					Mode: "vepa",
				},
			}
		} else if devI, ok := d.GetOk(prefix + ".macvtap"); ok {
			netIface.Source = &libvirtxml.DomainInterfaceSource{
				Direct: &libvirtxml.DomainInterfaceSourceDirect{
					Dev:  devI.(string),
					Mode: "bridge",
				},
			}
		} else if devI, ok := d.GetOk(prefix + ".passthrough"); ok {
			netIface.Source = &libvirtxml.DomainInterfaceSource{
				Direct: &libvirtxml.DomainInterfaceSourceDirect{
					Dev:  devI.(string),
					Mode: "passthrough",
				},
			}
		} else {
			// no network has been specified: we are on our own
		}

		domainDef.Devices.Interfaces = append(domainDef.Devices.Interfaces, netIface)
	}

	return nil
}

func destroyDomainByUserRequest(d *schema.ResourceData, domain *libvirt.Domain) error {
	if d.Get("running").(bool) {
		return nil
	}

	domainID, err := domain.GetUUIDString()

	if err != nil {
		return fmt.Errorf("Error retrieving libvirt domain id: %s", err)
	}

	log.Printf("Destroying libvirt domain %s", domainID)
	state, _, err := domain.GetState()
	if err != nil {
		return fmt.Errorf("Couldn't get info about domain: %s", err)
	}

	if state == libvirt.DOMAIN_RUNNING || state == libvirt.DOMAIN_PAUSED {
		if err := domain.Destroy(); err != nil {
			return fmt.Errorf("Couldn't destroy libvirt domain: %s", err)
		}
	}

	return nil
}
