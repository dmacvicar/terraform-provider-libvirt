package libvirt

import (
	"fmt"
	"log"
	"strconv"
	"encoding/xml"
	//"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	libvirt "gopkg.in/alexzorin/libvirt-go.v2"
)

type defDomain struct {
	XMLName xml.Name `xml:"domain"`
	Name string `xml:"name"`
	Type string `xml:"type,attr"`
	Os defOs `xml:"os"`
	Memory defMemory `xml:"memory"`
	VCpu defVCpu `xml:"vcpu"`
}

type defOs struct {
	Type defOsType `xml:"type"`
}

type defOsType struct {
	Arch string `xml:"arch,attr"`
	Machine string `xml:"machine,attr"`
	Name string `xml:"chardata"`
}

type defMemory struct {
	Unit string `xml:"unit,attr"`
	Amount int `xml:"chardata"`
}

type defVCpu struct {
	Placement string `xml:"unit,attr"`
	Amount int `xml:"chardata"`
}

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
			    Default: 1,
			},

			"memory": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			    Default: 512,
			},
		},
	}
}

func resourceLibvirtDomainCreate(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	domainDef := defDomain{
		Name: d.Get("name").(string),
		Type: "kvm",
		Os: defOs{
			defOsType{
				Arch: "x86_64",
				Machine: "pc-i440fx-2.4",
				Name: "hvm",
			},
		},
		Memory: defMemory{
			Unit: "MiB",
			Amount: d.Get("memory").(int),
		},
		VCpu: defVCpu{
			Placement: "static",
			Amount: d.Get("vcpu").(int),
		},
	}

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

