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

type domain struct {
	XMLName xml.Name `xml:"domain"`
	Name string `xml:"name"`
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
		},
	}
}

func resourceLibvirtDomainCreate(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	domainDef := domain{
		Name: d.Get("name").(string),
	}

	log.Printf("[INFO] Creating virtual machine")

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

	log.Printf("[INFO] Virtual Machine ID: %s", d.Id())

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

