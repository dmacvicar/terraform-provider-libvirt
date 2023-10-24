package libvirt

import (
	"encoding/xml"
	"fmt"
	"log"
	"strconv"

	"github.com/dmacvicar/terraform-provider-libvirt/libvirt/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// strutures to parse device XML
type Device struct {
	Device     xml.Name   `xml:"device"`
	Name       string     `xml:"name"`
	Path       string     `xml:"path"`
	Parent     string     `xml:"parent"`
	Capability Capability `xml:"capability"`
}

type Capability struct {
	Capability xml.Name `xml:"capability"`
	Type       string   `xml:"type,attr"`
	// PCI
	Class      string     `xml:"class"`
	Domain     int        `xml:"domain"`
	Bus        int        `xml:"bus"`
	Slot       int        `xml:"slot"`
	Function   int        `xml:"function"`
	Product    Product    `xml:"product"`
	Vendor     Vendor     `xml:"vendor"`
	IommuGroup IommuGroup `xml:"iommuGroup"`
	// STORAGE
	Block            string `xml:"block"`
	DriveType        string `xml:"drive_type"`
	Model            string `xml:"model"`
	Serial           string `xml:"serial"`
	Size             int    `xml:"size"`
	LogicalBlockSize int    `xml:"logical_block_size"`
	NumBlocks        int    `xml:"num_blocks"`
	// USB
	Number   int `xml:"number"`
	Subclass int `xml:"subclass"`
	Protocol int `xml:"protocol"`
}

type Product struct {
	Name string `xml:",chardata"`
	Id   string `xml:"id,attr"`
}

type Vendor struct {
	Name string `xml:",chardata"`
	Id   string `xml:"id,attr"`
}

type IommuGroup struct {
	IommuGroup xml.Name  `xml:"iommuGroup"`
	Number     int       `xml:"number,attr"`
	Addresses  []Address `xml:"address"`
}

type Address struct {
	Address  xml.Name `xml:"address"`
	Domain   string   `xml:"domain,attr"`
	Bus      string   `xml:"bus,attr"`
	Slot     string   `xml:"slot,attr"`
	Function string   `xml:"function,attr"`
}

// a libvirt nodeinfo datasource
//
// Datasource example:
//
//	data "libvirt_node_device_info" "info" {
//	}
//
//	output "xml" {
//	  value = data.libvirt_node_device_info.info.xml
//	}
func datasourceLibvirtNodeDeviceInfo() *schema.Resource {
	return &schema.Resource{
		Read: resourceLibvirtNodeDeviceInfoRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"xml": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capability": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func resourceLibvirtNodeDeviceInfoRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Read data source libvirt_nodedevices")

	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	var device_name string

	if name, ok := d.GetOk("name"); ok {
		device_name = name.(string)
		log.Printf("[DEBUG] Got name: %s", device_name)
	}

	device, err := virConn.NodeDeviceLookupByName(device_name)
	if err != nil {
		return fmt.Errorf("failed to lookup node device: %v", err)
	}

	xml_desc, err := virConn.NodeDeviceGetXMLDesc(device.Name, 0)
	if err != nil {
		return fmt.Errorf("failed to get XML for node device: %v", err)
	}

	var device_xml Device

	err = xml.Unmarshal([]byte(xml_desc), &device_xml)
	if err != nil {
		log.Fatalf("failed to unmarshal device_xml into XML: %v", err)
	}

	capability := map[string]interface{}{}
	if device_xml.Capability.Type == "pci" {
		capability["type"] = device_xml.Capability.Type
		capability["class"] = device_xml.Capability.Class
		capability["domain"] = fmt.Sprintf("%d", device_xml.Capability.Domain)
		capability["bus"] = fmt.Sprintf("%d", device_xml.Capability.Bus)
		capability["slot"] = fmt.Sprintf("%d", device_xml.Capability.Slot)
		capability["function"] = fmt.Sprintf("%d", device_xml.Capability.Function)
		capability["product_id"] = device_xml.Capability.Product.Id
		capability["product_name"] = device_xml.Capability.Product.Name
		capability["vendor_id"] = device_xml.Capability.Vendor.Id
		capability["vendor_name"] = device_xml.Capability.Vendor.Name
		capability["iommu_group_number"] = fmt.Sprintf("%d", device_xml.Capability.IommuGroup.Number)
	}

	if device_xml.Capability.Type == "storage" {
		capability["type"] = device_xml.Capability.Type
		capability["block"] = device_xml.Capability.Block
		capability["drive_type"] = device_xml.Capability.DriveType
		capability["mode"] = device_xml.Capability.Model
		capability["serial"] = device_xml.Capability.Serial
		capability["size"] = fmt.Sprintf("%d", device_xml.Capability.Size)
		capability["logical_block_size"] = fmt.Sprintf("%d", device_xml.Capability.LogicalBlockSize)
		capability["num_blocks"] = fmt.Sprintf("%d", device_xml.Capability.NumBlocks)
	}

	if device_xml.Capability.Type == "usb" {
		capability["type"] = device_xml.Capability.Type
		capability["number"] = fmt.Sprintf("%d", device_xml.Capability.Number)
		capability["class"] = device_xml.Capability.Class
		capability["subclass"] = fmt.Sprintf("%d", device_xml.Capability.Subclass)
		capability["protocol"] = fmt.Sprintf("%d", device_xml.Capability.Protocol)
	}

	d.Set("xml", xml_desc)
	d.Set("path", device_xml.Path)
	d.Set("parent", device_xml.Parent)
	d.Set("capability", capability)
	d.SetId(strconv.Itoa(hashcode.String(fmt.Sprintf("%v", xml_desc))))

	return nil
}
