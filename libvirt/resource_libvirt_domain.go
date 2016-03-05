package libvirt

import (
	"encoding/xml"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	//"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	libvirt "gopkg.in/alexzorin/libvirt-go.v2"
)

func resourceLibvirtDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceLibvirtDomainCreate,
		Read:   resourceLibvirtDomainRead,
		Update: resourceLibvirtDomainUpdate,
		Delete: resourceLibvirtDomainDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"vcpu": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"memory": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  512,
			},
			"base_image": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceLibvirtDomainCreate(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	// setup base image and disks
	baseImageSpec := d.Get("base_image").(string)
	// if the base image was specifies as $pool/$name
	components := strings.Split(baseImageSpec, "/")
	storagePool := "default"
	baseImage := ""
	if len(components) > 1 {
		storagePool = components[0]
		baseImage = components[1]
	} else {
		baseImage = components[0]
	}

	pool, err := virConn.LookupStoragePoolByName(storagePool)
	if err != nil {
		return fmt.Errorf("can't find storage pool '%s'", storagePool)
	}
	baseVol, err := pool.LookupStorageVolByName(baseImage)
	if err != nil {
		return fmt.Errorf("can't find image '%s' in pool '%s'", baseImage, storagePool)
	}

	// create the volume
	rootVolumeDef := defVolume{}
	rootVolumeDef.Name = "__terraform_" + d.Get("name").(string) + "-rootdisk"
	rootVolumeDef.Target.Format.Type = "qcow2"
	// use the base image as backing store
	rootVolumeDef.BackingStore.Format.Type = "qcow2"
	baseVolPath, err := baseVol.GetPath()
	if err != nil {
		return fmt.Errorf("can't get name for base image '%s'", baseImage)
	}
	rootVolumeDef.BackingStore.Path = baseVolPath
	rootVolumeDefXml, err := xml.Marshal(rootVolumeDef)
	if err != nil {
		return fmt.Errorf("Error serializing libvirt volume: %s", err)
	}

	// create the volume
	rootVolume, err := pool.StorageVolCreateXML(string(rootVolumeDefXml), 0)
	if err != nil {
		return fmt.Errorf("Error creating libvirt volume: %s", err)
	}

	// create the disk
	rootDisk := defDisk{}
	rootDisk.Type = "volume"
	rootDisk.Device = "disk"
	rootDisk.Format.Type = "qcow2"

	rootVolumeName, err := rootVolume.GetName()
	if err != nil {
		return fmt.Errorf("Error retrieving volume name: %s", err)
	}

	rootDisk.Source.Volume = rootVolumeName
	rootDisk.Source.Pool = storagePool
	rootDisk.Target.Dev = "sda"
	rootDisk.Target.Bus = "virtio"

	domainDef := newDomainDef()
	domainDef.Name = d.Get("name").(string)
	domainDef.Memory.Amount = d.Get("memory").(int)
	domainDef.VCpu.Amount = d.Get("vcpu").(int)
	domainDef.Devices.RootDisk = rootDisk

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

	id, err := domain.GetID()
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt domain id: %s", err)
	}
	d.SetId(fmt.Sprintf("%d", id))

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

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	domain, err := virConn.LookupDomainById(uint32(id))
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt domain: %s", err)
	}

	name, err := domain.GetName()
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt domain name: %s", err)
	}
	d.Set("name", name)

	return nil
}

func resourceLibvirtDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	return fmt.Errorf("Couldn't update libvirt domain")
}

func resourceLibvirtDomainDelete(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	domain, err := virConn.LookupDomainById(uint32(id))
	if err != nil {
		return fmt.Errorf("Error retrieving libvirt domain: %s", err)
	}

	err = domain.Destroy()
	if err != nil {
		return fmt.Errorf("Couldn't destroy libvirt domain: %s", err)
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
