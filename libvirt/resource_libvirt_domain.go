package libvirt

import (
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/dmacvicar/terraform-provider-libvirt/libvirt/helper/suppress"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	libvirt "github.com/libvirt/libvirt-go"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

type pendingMapping struct {
	mac         string
	hostname    string
	networkName string
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
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
			"fw_cfg_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "opt/com.coreos/config",
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
				Type:       schema.TypeList,
				Optional:   true,
				ForceNew:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
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
						},
						"wwn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"block_device": {
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
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							DiffSuppressFunc: suppress.CaseDifference,
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
						"listen_address": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "127.0.0.1",
						},
					},
				},
			},
			"video": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "cirrus",
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
						"source_host": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "127.0.0.1",
						},
						"source_service": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "0",
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
	domainDef.Description = d.Get("description").(string)

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

	setVideo(d, &domainDef)
	setConsoles(d, &domainDef)
	setCmdlineArgs(d, &domainDef)
	setFirmware(d, &domainDef)
	setBootDevices(d, &domainDef)

	if err := setCoreOSIgnition(d, &domainDef, virConn, arch); err != nil {
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
			ipNotFoundMsg := "Error: couldn't retrieve IP address of domain." +
				"Please check following: \n" +
				"1) is the domain running proplerly? \n" +
				"2) has the network interface an IP address? \n" +
				"3) Networking issues on your libvirt setup? \n " +
				"4) is DHCP enabled on this Domain's network? \n" +
				"5) if you use bridge network, the domain should have the pkg qemu-agent installed \n" +
				"IMPORTANT: This error is not a terraform libvirt-provider error, but an error caused by your KVM/libvirt infrastructure configuration/setup"
			return fmt.Errorf("%s \n %s", ipNotFoundMsg, err)
		}
	}

	// We save runnig state to not mix what we have and what we want
	requiredStatus := d.Get("running")

	err = resourceLibvirtDomainRead(d, meta)
	if err != nil {
		return err
	}

	d.Set("running", requiredStatus)

	// we must read devices again in order to set some missing ip/MAC/host mappings
	for i := 0; i < d.Get("network_interface.#").(int); i++ {
		prefix := fmt.Sprintf("network_interface.%d", i)
		mac := strings.ToUpper(d.Get(prefix + ".mac").(string))

		// if we were waiting for an IP address for this MAC, go ahead.
		if pending, ok := partialNetIfaces[mac]; ok {
			// we should have the address now
			addressesI, ok := d.GetOk(prefix + ".addresses")
			if !ok {
				log.Printf("Did not obtain the IP address for MAC=%s", mac)
				continue
			}

			network, err := virConn.LookupNetworkByName(pending.networkName)
			if err != nil {
				log.Printf("Can't retrieve network '%s'", pending.networkName)
				continue
			}

			for _, addressI := range addressesI.([]interface{}) {
				address := addressI.(string)
				log.Printf("[INFO] Finally adding IP/MAC/host=%s/%s/%s", address, mac, pending.hostname)

				err = updateOrAddHost(network, address, mac, pending.hostname)
				if err != nil {
					log.Printf("Could not add IP/MAC/host=%s/%s/%s: %s", address, mac, pending.hostname, err)
				}
			}
			network.Free()
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

	domainRunningNow, err := domainIsRunning(*domain)
	if err != nil {
		return fmt.Errorf("Error reading domain running state : %s", err)
	}

	d.Set("name", domainDef.Name)
	d.Set("description", domainDef.Description)
	d.Set("vcpu", domainDef.VCPU)
	d.Set("memory", domainDef.Memory)
	d.Set("firmware", domainDef.OS.Loader)
	d.Set("nvram", domainDef.OS.NVRam)
	d.Set("cpu", domainDef.CPU)
	d.Set("arch", domainDef.OS.Type.Arch)
	d.Set("autostart", autostart)
	d.Set("running", domainRunningNow)

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
		} else if diskDef.Source.Block != nil {
			disk = map[string]interface{}{
				"block_device": diskDef.Source.Block,
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
		netIface["hostname"] = d.Get(prefix + ".hostname").(string)
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
