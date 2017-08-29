package libvirt

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	libvirt "github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
)

type DomainMeta struct {
	domain *libvirt.Domain
	ifaces chan libvirtxml.DomainInterface
}

var PoolSync = NewLibVirtPoolSync()

func init() {
	spew.Config.Indent = "\t"
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}

func resourceLibvirtDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceLibvirtDomainCreate,
		Read:   resourceLibvirtDomainRead,
		Delete: resourceLibvirtDomainDelete,
		Update: resourceLibvirtDomainUpdate,
		Exists: resourceLibvirtDomainExists,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
		},
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
				Type:     schema.TypeMap,
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
			"coreos_ignition": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "",
			},
			"filesystem": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Required: false,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
			},
			"disk": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Required: false,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
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
			"graphics": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Required: false,
			},
			"console": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Required: false,
				Elem: &schema.Resource{
					Schema: consoleSchema(),
				},
			},
			"cpu": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Required: false,
				ForceNew: true,
			},
			"autostart": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Required: false,
			},
			"machine": &schema.Schema{
			        Type: schema.TypeString,
				Optional: true,
			},
			"arch": &schema.Schema{
			        Type: schema.TypeString,
				Optional: true,
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
		return fmt.Errorf("The libvirt connection was nil.")
	}

	domainDef := newDomainDef()

	if name, ok := d.GetOk("name"); ok {
		domainDef.Name = name.(string)
	}

	if ignition, ok := d.GetOk("coreos_ignition"); ok {
		//var fw_cfg []defCmd
		var fw_cfg []libvirtxml.DomainQEMUCommandlineArg
		ignitionKey, err := getIgnitionVolumeKeyFromTerraformID(ignition.(string))
		if err != nil {
			return err
		}
		ign_str := fmt.Sprintf("name=opt/com.coreos/config,file=%s", ignitionKey)
		fw_cfg = append(fw_cfg, libvirtxml.DomainQEMUCommandlineArg{"-fw_cfg"})
		fw_cfg = append(fw_cfg, libvirtxml.DomainQEMUCommandlineArg{ign_str})
		QemuCmdline := &libvirtxml.DomainQEMUCommandline{
			Args: fw_cfg,
		}
		domainDef.QEMUCommandline = QemuCmdline
	}

	arch, err := getHostArchitecture(virConn)
	if err != nil {
		return fmt.Errorf("Error retrieving host architecture: %s", err)
	}

	if arch == "s390x" || arch == "ppc64" {
		domainDef.Devices.Graphics = nil
	} else {
		if graphics, ok := d.GetOk("graphics"); ok {
			graphics_map := graphics.(map[string]interface{})
			domainDef.Devices.Graphics = []libvirtxml.DomainGraphic{
				libvirtxml.DomainGraphic{},
			}
			if graphics_type, ok := graphics_map["type"]; ok {
				domainDef.Devices.Graphics[0].Type = graphics_type.(string)
			}
			if autoport, ok := graphics_map["autoport"]; ok {
				domainDef.Devices.Graphics[0].AutoPort = autoport.(string)
			}
			if listen_type, ok := graphics_map["listen_type"]; ok {
				domainDef.Devices.Graphics[0].Listeners = []libvirtxml.DomainGraphicListener{
					libvirtxml.DomainGraphicListener{
						Type: listen_type.(string),
					},
				}
			}
		}
	}

	if cpu, ok := d.GetOk("cpu"); ok {
		cpu_map := cpu.(map[string]interface{})
		if cpu_mode, ok := cpu_map["mode"]; ok {
			domainDef.CPU = &libvirtxml.DomainCPU{
				Mode: cpu_mode.(string),
			}
		}
	}

	domainDef.OS.Type.Arch = d.Get("arch").(string)
	domainDef.OS.Type.Machine = d.Get("machine").(string)
	
	if firmware, ok := d.GetOk("firmware"); ok {
		firmwareFile := firmware.(string)
		if _, err := os.Stat(firmwareFile); os.IsNotExist(err) {
			return fmt.Errorf("Could not find firmware file '%s'.", firmwareFile)
		}
		domainDef.OS.Loader = &libvirtxml.DomainLoader{
			Path:     firmwareFile,
			Readonly: "yes",
			Type:     "pflash",
			Secure:   "no",
		}

		if nvram, ok := d.GetOk("nvram"); ok {
			nvram_map := nvram.(map[string]interface{})

			nvramFile := nvram_map["file"].(string)
			if _, err := os.Stat(nvramFile); os.IsNotExist(err) {
				return fmt.Errorf("Could not find nvram file '%s'.", nvramFile)
			}
			nvramTemplateFile := ""
			if nvram_template, ok := nvram_map["template"]; ok {
				nvramTemplateFile = nvram_template.(string)
				if _, err := os.Stat(nvramTemplateFile); os.IsNotExist(err) {
					return fmt.Errorf("Could not find nvram template file '%s'.", nvramTemplateFile)
				}
			}
			domainDef.OS.NVRam = &libvirtxml.DomainNVRam{
				NVRam:    nvramFile,
				Template: nvramTemplateFile,
			}
		}
	}

	domainDef.Memory = &libvirtxml.DomainMemory{
		Value: uint(d.Get("memory").(int)),
		Unit:  "MiB",
	}
	domainDef.VCPU = &libvirtxml.DomainVCPU{
		Value: d.Get("vcpu").(int),
	}

	var consoles []libvirtxml.DomainConsole
	for i := 0; i < d.Get("console.#").(int); i++ {
		console := libvirtxml.DomainConsole{}
		consolePrefix := fmt.Sprintf("console.%d", i)
		console.Type = d.Get(consolePrefix + ".type").(string)
		consoleTargetPortInt, err := strconv.Atoi(d.Get(consolePrefix + ".target_port").(string))
		if err == nil {
			consoleTargetPort := uint(consoleTargetPortInt)
			console.Target = &libvirtxml.DomainConsoleTarget{
				Port: &consoleTargetPort,
			}
		}
		if source_path, ok := d.GetOk(consolePrefix + ".source_path"); ok {
			console.Source = &libvirtxml.DomainChardevSource{
				Path: source_path.(string),
			}
		}
		if target_type, ok := d.GetOk(consolePrefix + ".target_type"); ok {
			if console.Target == nil {
				console.Target = &libvirtxml.DomainConsoleTarget{}
			}
			console.Target.Type = target_type.(string)
		}
		consoles = append(consoles, console)
	}
	domainDef.Devices.Consoles = consoles

	disksCount := d.Get("disk.#").(int)
	var disks []libvirtxml.DomainDisk
	var scsiDisk bool = false
	for i := 0; i < disksCount; i++ {
		disk := libvirtxml.DomainDisk{
			Type:   "file",
			Device: "disk",
			Target: &libvirtxml.DomainDiskTarget{
				Bus: "virtio",
				Dev: fmt.Sprintf("vd%s", DiskLetterForIndex(i)),
			},
			Driver: &libvirtxml.DomainDiskDriver{
				Name: "qemu",
				Type: "qcow2",
			},
		}

		diskKey := fmt.Sprintf("disk.%d", i)
		diskMap := d.Get(diskKey).(map[string]interface{})
		volumeKey := diskMap["volume_id"].(string)
		if _, ok := diskMap["scsi"].(string); ok {
			disk.Target.Bus = "scsi"
			scsiDisk = true
			if wwn, ok := diskMap["wwn"].(string); ok {
				disk.WWN = wwn
			} else {
				disk.WWN = randomWWN(10)
			}
		}
		diskVolume, err := virConn.LookupStorageVolByKey(volumeKey)
		if err != nil {
			return fmt.Errorf("Can't retrieve volume %s", volumeKey)
		}
		diskVolumeFile, err := diskVolume.GetPath()
		if err != nil {
			return fmt.Errorf("Error retrieving volume file: %s", err)
		}

		disk.Source = &libvirtxml.DomainDiskSource{
			File: diskVolumeFile,
		}

		disks = append(disks, disk)
	}

	log.Printf("[DEBUG] scsiDisk: %s", scsiDisk)
	if scsiDisk {
		controller := libvirtxml.DomainController{Type: "scsi", Model: "virtio-scsi"}
		domainDef.Devices.Controllers = append(domainDef.Devices.Controllers, controller)
	}

	filesystemsCount := d.Get("filesystem.#").(int)
	var filesystems []libvirtxml.DomainFilesystem
	for i := 0; i < filesystemsCount; i++ {
		fs := newFilesystemDef()

		fsKey := fmt.Sprintf("filesystem.%d", i)
		fsMap := d.Get(fsKey).(map[string]interface{})
		if accessMode, ok := fsMap["accessmode"]; ok {
			fs.AccessMode = accessMode.(string)
		}
		if sourceDir, ok := fsMap["source"]; ok {
			fs.Source = &libvirtxml.DomainFilesystemSource{
				Dir: sourceDir.(string),
			}
		} else {
			return fmt.Errorf("Filesystem entry must have a 'source' set")
		}
		if targetDir, ok := fsMap["target"]; ok {
			fs.Target = &libvirtxml.DomainFilesystemTarget{
				Dir: targetDir.(string),
			}
		} else {
			return fmt.Errorf("Filesystem entry must have a 'target' set")
		}
		if readonly, ok := fsMap["readonly"]; ok {
			if readonly.(string) == "1" {
				fs.ReadOnly = &libvirtxml.DomainFilesystemReadOnly{}
			} else {
				fs.ReadOnly = nil
			}
		}

		filesystems = append(filesystems, fs)
	}
	log.Printf("filesystems: %+v\n", filesystems)

	type pendingMapping struct {
		mac      string
		hostname string
		network  *libvirt.Network
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
	netIfaces := make([]libvirtxml.DomainInterface, 0, netIfacesCount)
	partialNetIfaces := make(map[string]pendingMapping, netIfacesCount)
	waitForLeases = make(map[libvirtxml.DomainInterface]bool, netIfacesCount)
	for i := 0; i < netIfacesCount; i++ {
		prefix := fmt.Sprintf("network_interface.%d", i)

		netIface := libvirtxml.DomainInterface{}
		netIface.Model = &libvirtxml.DomainInterfaceModel{
			Type: "virtio",
		}

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
		netIface.MAC = &libvirtxml.DomainInterfaceMAC{
			Address: mac,
		}

		// this is not passed to libvirt, but used by waitForAddress
		waitForLeases[netIface] = false
		if waitForLease, ok := d.GetOk(prefix + ".wait_for_lease"); ok {
			waitForLeases[netIface] = waitForLease.(bool)
		}

		// connect to the interface to the network... first, look for the network
		if n, ok := d.GetOk(prefix + ".network_name"); ok {
			// when using a "network_name" we do not try to do anything: we just
			// connect to that network
			netIface.Type = "network"
			netIface.Source = &libvirtxml.DomainInterfaceSource{
				Network: n.(string),
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
			networkDef, err := newDefNetworkfromLibvirt(network)
			if !HasDHCP(networkDef) {
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
					if err := addHost(network, ip.String(), mac, hostname); err != nil {
						return err
					}
				}
			} else {
				// no IPs provided: if the hostname has been provided, wait until we get an IP
				if len(hostname) > 0 {
					if !waitForLeases[netIface] {
						return fmt.Errorf("Cannot map '%s': we are not waiting for lease and no IP has been provided", hostname)
					}
					// the resource specifies a hostname but not an IP, so we must wait until we
					// have a valid lease and then read the IP we have been assigned, so we can
					// do the mapping
					log.Printf("[DEBUG] Will wait for an IP for hostname '%s'...", hostname)
					partialNetIfaces[strings.ToUpper(mac)] = pendingMapping{
						mac:      strings.ToUpper(mac),
						hostname: hostname,
						network:  network,
					}
				} else {
					// neither an IP or a hostname has been provided: so nothing must be forced
				}
			}
			netIface.Type = "network"
			netIface.Source = &libvirtxml.DomainInterfaceSource{
				Network: networkName,
			}
		} else if bridgeNameI, ok := d.GetOk(prefix + ".bridge"); ok {
			netIface.Type = "bridge"
			netIface.Source = &libvirtxml.DomainInterfaceSource{
				Bridge: bridgeNameI.(string),
			}
		} else if devI, ok := d.GetOk(prefix + ".vepa"); ok {
			netIface.Type = "direct"
			netIface.Source = &libvirtxml.DomainInterfaceSource{
				Dev:  devI.(string),
				Mode: "vepa",
			}
		} else if devI, ok := d.GetOk(prefix + ".macvtap"); ok {
			netIface.Type = "direct"
			netIface.Source = &libvirtxml.DomainInterfaceSource{
				Dev:  devI.(string),
				Mode: "bridge",
			}
		} else if devI, ok := d.GetOk(prefix + ".passthrough"); ok {
			netIface.Type = "direct"
			netIface.Source = &libvirtxml.DomainInterfaceSource{
				Dev:  devI.(string),
				Mode: "passthrough",
			}
		} else {
			// no network has been specified: we are on our own
		}

		netIfaces = append(netIfaces, netIface)
	}

	domainDef.Devices.Disks = disks
	domainDef.Devices.Filesystems = filesystems
	domainDef.Devices.Interfaces = netIfaces

	connectURI, err := virConn.GetURI()
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt connection URI: %s", err)
	}
	log.Printf("[INFO] Creating libvirt domain at %s", connectURI)

	data, err := xmlMarshallIndented(domainDef)
	if err != nil {
		return fmt.Errorf("Error serializing libvirt domain: %s", err)
	}

	log.Printf("[DEBUG] Creating libvirt domain with XML:\n%s", data)

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

	domainMeta := DomainMeta{
		domain,
		make(chan libvirtxml.DomainInterface, len(netIfaces)),
	}

	// populate interface channels
	for _, iface := range netIfaces {
		domainMeta.ifaces <- iface
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"blocked"},
		Target:     []string{"running"},
		Refresh:    resourceLibvirtDomainStateRefreshFunc(d, &domainMeta),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	_, err = stateConf.WaitForState()
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
	domain, err := virConn.LookupDomainByUUIDString(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt domain: %s", err)
	}
	defer domain.Free()

	running, err := isDomainRunning(*domain)
	if err != nil {
		return err
	}
	if !running {
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
				if err := updateHost(network, ip.String(), mac, hostname); err != nil {
					return err
				}
			}
		}
	}

	d.Partial(false)

	// TODO
	return nil
}

func resourceLibvirtDomainRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Read resource libvirt_domain")

	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
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

	domainDef := newDomainDef()
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
	d.Set("autostart", autostart)

	running, err := isDomainRunning(*domain)
	if err != nil {
		return err
	}
	d.Set("running", running)

	disks := make([]map[string]interface{}, 0)
	for _, diskDef := range domainDef.Devices.Disks {
		var virVol *libvirt.StorageVol
		if len(diskDef.Source.File) > 0 {
			virVol, err = virConn.LookupStorageVolByPath(diskDef.Source.File)
		} else {
			virPool, err := virConn.LookupStoragePoolByName(diskDef.Source.Pool)
			if err != nil {
				return fmt.Errorf("Error retrieving pool for disk: %s", err)
			}
			defer virPool.Free()

			virVol, err = virPool.LookupStorageVolByName(diskDef.Source.Volume)
		}

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

	filesystems := make([]map[string]interface{}, 0)
	for _, fsDef := range domainDef.Devices.Filesystems {
		fs := map[string]interface{}{
			"accessmode": fsDef.AccessMode,
			"source":     fsDef.Source.Dir,
			"target":     fsDef.Target.Dir,
			"readonly":   fsDef.ReadOnly,
		}
		filesystems = append(filesystems, fs)
	}
	d.Set("filesystems", filesystems)

	// look interfaces with addresses
	ifacesWithAddr, err := getDomainInterfaces(*domain)
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

				networkDef, err := newDefNetworkfromLibvirt(network)
				if err != nil {
					return err
				}

				netIface["network_name"] = networkInterfaceDef.Source.Network

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

	domain, err := virConn.LookupDomainByUUIDString(d.Id())
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

func resourceLibvirtDomainStateRefreshFunc(
	d *schema.ResourceData, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		domain := meta.(*DomainMeta).domain

		state, err := getDomainState(*domain)
		if err != nil {
			return nil, "", err
		}

		if state == "running" {
			// domain and interface(s) are up, we're done
			if len(meta.(*DomainMeta).ifaces) == 0 {
				return meta, state, nil
			}

			// set state to "blocked" since we still have interfaces to check
			state = "blocked"

			iface := <-meta.(*DomainMeta).ifaces
			found, ignore, err := hasNetworkAddress(iface, *domain)
			if err != nil {
				return nil, "", err
			}

			if found {
				return meta, state, nil
			} else if !found && !ignore {
				// re-add the interface and deal with it later
				meta.(*DomainMeta).ifaces <- iface
			}
		}

		return meta, state, nil
	}
}

func hasNetworkAddress(iface libvirtxml.DomainInterface,
	domain libvirt.Domain) (found bool, ignore bool, err error) {

	if !waitForLeases[iface] {
		return false, true, nil
	}

	mac := strings.ToUpper(iface.MAC.Address)
	if mac == "" {
		log.Printf("[DEBUG] Can't wait without a mac address.\n")
		// we can't get the ip without a mac address
		return false, true, nil
	}

	log.Printf("[DEBUG] waiting for network address for interface with hwaddr: '%s'\n", iface.MAC.Address)
	ifacesWithAddr, err := getDomainInterfaces(domain)
	if err != nil {
		return false, false, fmt.Errorf("Error retrieving interface addresses: %s", err)
	}
	log.Printf("[DEBUG] ifaces with addresses: %+v\n", ifacesWithAddr)

	for _, ifaceWithAddr := range ifacesWithAddr {
		// found
		if mac == strings.ToUpper(ifaceWithAddr.Hwaddr) {
			return true, false, nil
		}
	}

	return false, false, nil
}

func getDomainState(domain libvirt.Domain) (string, error) {
	state, _, err := domain.GetState()
	if err != nil {
		return "", err
	}

	var stateStr string

	switch state {
	case libvirt.DOMAIN_NOSTATE:
		stateStr = "nostate"
	case libvirt.DOMAIN_RUNNING:
		stateStr = "running"
	case libvirt.DOMAIN_BLOCKED:
		stateStr = "blocked"
	case libvirt.DOMAIN_PAUSED:
		stateStr = "paused"
	case libvirt.DOMAIN_SHUTDOWN:
		stateStr = "shutdown"
	case libvirt.DOMAIN_CRASHED:
		stateStr = "crashed"
	case libvirt.DOMAIN_PMSUSPENDED:
		stateStr = "pmsuspended"
	case libvirt.DOMAIN_SHUTOFF:
		stateStr = "shutoff"
	default:
		stateStr = fmt.Sprintf("unknown: %v", state)
	}

	return stateStr, nil
}

func isDomainRunning(domain libvirt.Domain) (bool, error) {
	state, _, err := domain.GetState()
	if err != nil {
		return false, fmt.Errorf("Couldn't get state of domain: %s", err)
	}

	return state == libvirt.DOMAIN_RUNNING, nil
}

func newDiskForCloudInit(virConn *libvirt.Connect, volumeKey string) (libvirtxml.DomainDisk, error) {
	disk := libvirtxml.DomainDisk{
		Type:   "file",
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
		return disk, fmt.Errorf("Can't retrieve volume %s", volumeKey)
	}
	diskVolumeFile, err := diskVolume.GetPath()
	if err != nil {
		return disk, fmt.Errorf("Error retrieving volume file: %s", err)
	}

	disk.Source = &libvirtxml.DomainDiskSource{
		File: diskVolumeFile,
	}

	return disk, nil
}

func getDomainInterfaces(domain libvirt.Domain) ([]libvirt.DomainInterface, error) {

	// get all the interfaces using the qemu-agent, this includes also
	// interfaces that are not attached to networks managed by libvirt
	// (eg. bridges, macvtap,...)
	interfaces := getDomainInterfacesViaQemuAgent(&domain, true)
	if len(interfaces) > 0 {
		// the agent will always return all the interfaces, both the
		// ones managed by libvirt and the ones attached to bridge interfaces
		// or macvtap. Hence it has the highest priority
		return interfaces, nil
	}

	log.Print("[DEBUG] fetching networking interfaces using libvirt API")

	// get all the interfaces attached to libvirt networks
	interfaces, err := domain.ListAllInterfaceAddresses(libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
	if err != nil {
		switch err.(type) {
		default:
			return interfaces, fmt.Errorf("Error retrieving interface addresses: %s", err)
		case libvirt.Error:
			virErr := err.(libvirt.Error)
			if virErr.Code != libvirt.ERR_OPERATION_INVALID || virErr.Domain != libvirt.FROM_QEMU {
				return interfaces, fmt.Errorf("Error retrieving interface addresses: %s", err)
			}
		}
	}
	log.Printf("[DEBUG] Interfaces: %s", spew.Sdump(interfaces))

	return interfaces, nil
}
