package libvirt

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	libvirt "github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
)

// deprecated, now defaults to not use it, but we warn the user
const skipQemuAgentEnvVar = "TF_SKIP_QEMU_AGENT"

// if explicitly enabled
const useQemuAgentEnvVar = "TF_USE_QEMU_AGENT"

const domWaitLeaseStillWaiting = "waiting-addresses"
const domWaitLeaseDone = "all-addresses-obtained"

var errDomainInvalidState = errors.New("invalid state for domain")

func domainWaitForLeases(domain *libvirt.Domain, waitForLeases []*libvirtxml.DomainInterface,
	timeout time.Duration, rd *schema.ResourceData) error {
	waitFunc := func() (interface{}, string, error) {

		state, err := domainGetState(*domain)
		if err != nil {
			return false, "", err
		}

		for _, fatalState := range []string{"crashed", "shutoff", "shutdown", "pmsuspended"} {
			if state == fatalState {
				return false, "", errDomainInvalidState
			}
		}

		if state != "running" {
			return false, domWaitLeaseStillWaiting, nil
		}

		// check we have IPs for all the interfaces we are waiting for
		for _, iface := range waitForLeases {
			found, ignore, err := domainIfaceHasAddress(*domain, *iface, rd)
			if err != nil {
				return false, "", err
			}
			if ignore {
				log.Printf("[DEBUG] we don't care about the IP address for %+v", iface)
				continue
			}
			if !found {
				log.Printf("[DEBUG] IP address not found for iface=%+v: will try in a while", strings.ToUpper(iface.MAC.Address))
				return false, domWaitLeaseStillWaiting, nil
			}
		}

		log.Printf("[DEBUG] all the %d IP addresses obtained for the domain", len(waitForLeases))
		return true, domWaitLeaseDone, nil
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{domWaitLeaseStillWaiting},
		Target:     []string{domWaitLeaseDone},
		Refresh:    waitFunc,
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      5 * time.Second,
	}

	_, err := stateConf.WaitForState()
	log.Print("[DEBUG] wait-for-leases was successful")
	return err
}

func domainIfaceHasAddress(domain libvirt.Domain, iface libvirtxml.DomainInterface, rd *schema.ResourceData) (found bool, ignore bool, err error) {

	mac := strings.ToUpper(iface.MAC.Address)
	if mac == "" {
		log.Printf("[DEBUG] Can't wait without a MAC address: ignoring interface %+v.\n", iface)
		// we can't get the ip without a mac address
		return false, true, nil
	}

	log.Printf("[DEBUG] waiting for network address for iface=%s\n", mac)
	ifacesWithAddr, err := domainGetIfacesInfo(domain, rd)
	if err != nil {
		return false, false, fmt.Errorf("Error retrieving interface addresses: %s", err)
	}
	log.Printf("[DEBUG] ifaces with addresses: %+v\n", ifacesWithAddr)

	for _, ifaceWithAddr := range ifacesWithAddr {
		if mac == strings.ToUpper(ifaceWithAddr.Hwaddr) {
			log.Printf("[DEBUG] found IPs for MAC=%+v: %+v\n", mac, ifaceWithAddr.Addrs)
			return true, false, nil
		}
	}

	log.Printf("[DEBUG] %+v doesn't have IP address(es) yet...\n", mac)
	return false, false, nil
}

func domainGetState(domain libvirt.Domain) (string, error) {
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

func domainIsRunning(domain libvirt.Domain) (bool, error) {
	state, _, err := domain.GetState()
	if err != nil {
		return false, fmt.Errorf("Couldn't get state of domain: %s", err)
	}

	return state == libvirt.DOMAIN_RUNNING, nil
}

func domainGetIfacesInfo(domain libvirt.Domain, rd *schema.ResourceData) ([]libvirt.DomainInterface, error) {
	domainRunningNow, err := domainIsRunning(domain)
	if err != nil {
		return []libvirt.DomainInterface{}, err
	}
	if !domainRunningNow {
		log.Print("[DEBUG] no interfaces could be obtained: domain not running")
		return []libvirt.DomainInterface{}, nil
	}

	qemuAgentEnabled := rd.Get("qemu_agent").(bool)
	if qemuAgentEnabled {
		// get all the interfaces using the qemu-agent, this includes also
		// interfaces that are not attached to networks managed by libvirt
		// (eg. bridges, macvtap,...)
		log.Print("[DEBUG] fetching networking interfaces using qemu-agent")
		interfaces := qemuAgentGetInterfacesInfo(&domain, true)
		if len(interfaces) > 0 {
			// the agent will always return all the interfaces, both the
			// ones managed by libvirt and the ones attached to bridge interfaces
			// or macvtap. Hence it has the highest priority
			return interfaces, nil
		}
	} else {
		log.Printf("[DEBUG] qemu-agent is not used")
	}
	var interfaces []libvirt.DomainInterface

	// get all the interfaces attached to libvirt networks
	log.Print("[DEBUG] no interfaces could be obtained with qemu-agent: falling back to the libvirt API")

	interfaces, err = domain.ListAllInterfaceAddresses(libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
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
	log.Printf("[DEBUG] Interfaces info obtained with libvirt API:\n%s\n", spew.Sdump(interfaces))

	return interfaces, nil
}

func newDiskForCloudInit(virConn *libvirt.Connect, volumeKey string) (libvirtxml.DomainDisk, error) {
	disk := libvirtxml.DomainDisk{
		Device: "cdrom",
		Target: &libvirtxml.DomainDiskTarget{
			// Last device letter possible with a single IDE controller on i440FX
			Dev: "hdd",
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

func setCoreOSIgnition(d *schema.ResourceData, domainDef *libvirtxml.Domain, virConn *libvirt.Connect, arch string) error {
	if ignition, ok := d.GetOk("coreos_ignition"); ok {
		ignitionKey, err := getIgnitionVolumeKeyFromTerraformID(ignition.(string))
		if err != nil {
			return err
		}

		switch arch {
		case "i686", "x86_64", "aarch64":
			// QEMU and the Linux kernel support the use of the Firmware
			// Configuration Device on these architectures. Ignition will use
			// this mechanism to read its configuration from the hypervisor.

			// `fw_cfg_name` stands for firmware config is defined by a key and a value
			// credits for this cryptic name: https://github.com/qemu/qemu/commit/81b2b81062612ebeac4cd5333a3b15c7d79a5a3d
			if fwCfg, ok := d.GetOk("fw_cfg_name"); ok {
				domainDef.QEMUCommandline = &libvirtxml.DomainQEMUCommandline{
					Args: []libvirtxml.DomainQEMUCommandlineArg{
						{
							Value: "-fw_cfg",
						},
						{
							Value: fmt.Sprintf("name=%s,file=%s", fwCfg, ignitionKey),
						},
					},
				}
			}
		case "s390", "s390x", "ppc64", "ppc64le":
			// System Z and PowerPC do not support the Firmware Configuration
			// device. After a discussion about the best way to support a similar
			// method for qemu in https://github.com/coreos/ignition/issues/928,
			// decided on creating a virtio-blk device with a serial of ignition
			// which contains the ignition config and have ignition support for
			// reading from the device which landed in https://github.com/coreos/ignition/pull/936
			igndisk := libvirtxml.DomainDisk{
				Device: "disk",
				Source: &libvirtxml.DomainDiskSource{
					File: &libvirtxml.DomainDiskSourceFile{
						File: ignitionKey,
					},
				},
				Target: &libvirtxml.DomainDiskTarget{
					Dev: "vdb",
					Bus: "virtio",
				},
				Driver: &libvirtxml.DomainDiskDriver{
					Name: "qemu",
					Type: "raw",
				},
				ReadOnly: &libvirtxml.DomainDiskReadOnly{},
				Serial:   "ignition",
			}
			domainDef.Devices.Disks = append(domainDef.Devices.Disks, igndisk)
		default:
			return fmt.Errorf("Ignition not supported on %q", arch)
		}
	}

	return nil
}

func setVideo(d *schema.ResourceData, domainDef *libvirtxml.Domain) error {
	prefix := "video.0"
	if _, ok := d.GetOk(prefix); ok {
		domainDef.Devices.Videos = append(domainDef.Devices.Videos, libvirtxml.DomainVideo{
			Model: libvirtxml.DomainVideoModel{
				Type: d.Get(prefix + ".type").(string),
			},
		})
	}

	return nil
}

func setGraphics(d *schema.ResourceData, domainDef *libvirtxml.Domain, arch string) error {
	// For s390x, ppc64 and ppc64le spice is not supported
	if arch == "s390x" || strings.HasPrefix(arch, "ppc64") {
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
				listenAddress := d.Get(prefix + ".listen_address")
				listener.Address = &libvirtxml.DomainGraphicListenerAddress{
					Address: listenAddress.(string),
				}
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
		if targetType, ok := d.GetOk(prefix + ".target_type"); ok {
			if console.Target == nil {
				console.Target = &libvirtxml.DomainConsoleTarget{}
			}
			console.Target.Type = targetType.(string)
		}
		switch d.Get(prefix + ".type").(string) {
		case "tcp":
			sourceHost := d.Get(prefix + ".source_host")
			sourceService := d.Get(prefix + ".source_service")
			console.Source = &libvirtxml.DomainChardevSource{
				TCP: &libvirtxml.DomainChardevSourceTCP{
					Mode:    "bind",
					Host:    sourceHost.(string),
					Service: sourceService.(string),
				},
			}
			console.Protocol = &libvirtxml.DomainChardevProtocol{
				Type: "telnet",
			}
		case "pty":
			fallthrough
		default:
			if sourcePath, ok := d.GetOk(prefix + ".source_path"); ok {
				console.Source = &libvirtxml.DomainChardevSource{
					Dev: &libvirtxml.DomainChardevSourceDev{
						Path: sourcePath.(string),
					},
				}
			}
		}
		domainDef.Devices.Consoles = append(domainDef.Devices.Consoles, console)
	}
}

func setDisks(d *schema.ResourceData, domainDef *libvirtxml.Domain, virConn *libvirt.Connect) error {
	var scsiDisk = false
	var numOfISOs = 0

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
				disk.Target = &libvirtxml.DomainDiskTarget{
					Dev: fmt.Sprintf("hd%s", diskLetterForIndex(numOfISOs)),
					Bus: "ide",
				}
				disk.Driver = &libvirtxml.DomainDiskDriver{
					Name: "qemu",
				}
				numOfISOs++
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
					Dev: fmt.Sprintf("hd%s", diskLetterForIndex(numOfISOs)),
					Bus: "ide",
				}
				disk.Driver = &libvirtxml.DomainDiskDriver{
					Name: "qemu",
					Type: "raw",
				}

				numOfISOs++
			}

			if !strings.HasSuffix(file.(string), ".qcow2") {
				disk.Driver.Type = "raw"
			}
		} else if blockDev, ok := d.GetOk(prefix + ".block_device"); ok {
			disk.Source = &libvirtxml.DomainDiskSource{
				Block: &libvirtxml.DomainDiskSourceBlock{
					Dev: blockDev.(string),
				},
			}

			disk.Driver.Type = "raw"
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
		var network *libvirt.Network
		var err error

		if networkName, ok := d.GetOk(prefix + ".network_name"); ok {
			network, err = virConn.LookupNetworkByName(networkName.(string))
			if err != nil {
				return fmt.Errorf("Can't retrieve network '%s'", networkName.(string))
			}
			defer network.Free()

		} else if networkUUID, ok := d.GetOk(prefix + ".network_id"); ok {
			// when using a "network_id" we are referring to a "network resource"
			// we have defined somewhere else...
			network, err = virConn.LookupNetworkByUUIDString(networkUUID.(string))
			if err != nil {
				return fmt.Errorf("Can't retrieve network ID %s", networkUUID)
			}
			defer network.Free()

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

		// if we got a network
		if network != nil {
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
					if wait {
						// the resource specifies a hostname but not an IP, so we must wait until we
						// have a valid lease and then read the IP we have been assigned, so we can
						// do the mapping
						log.Printf("[DEBUG] Do not have an IP for '%s' yet: will wait until DHCP provides one...", hostname)
						if err != nil {
							return err
						}
						partialNetIfaces[strings.ToUpper(mac)] = &pendingMapping{
							mac:         strings.ToUpper(mac),
							hostname:    hostname,
							networkName: networkName,
						}
					}
				}
			}

			netIface.Source = &libvirtxml.DomainInterfaceSource{
				Network: &libvirtxml.DomainInterfaceSourceNetwork{
					Network: networkName,
				},
			}
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
