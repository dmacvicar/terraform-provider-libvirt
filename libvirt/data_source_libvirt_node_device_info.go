package libvirt

import (
	"encoding/xml"
	"fmt"
	"log"
	"strconv"

	"github.com/dmacvicar/terraform-provider-libvirt/libvirt/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// strutures to parse device XML just enough to get Capability.Type.
type DeviceGeneric struct {
	Device     xml.Name `xml:"device"`
	Name       string   `xml:"name"`
	Path       string   `xml:"path"`
	Parent     string   `xml:"parent"`
	Capability struct {
		Capability xml.Name `xml:"capability"`
		Type       string   `xml:"type,attr"`
	} `xml:"capability"`
}

type DeviceSystem struct {
	Device     xml.Name `xml:"device"`
	Name       string   `xml:"name"`
	Path       string   `xml:"path"`
	Parent     string   `xml:"parent"`
	Capability struct {
		Type     string `xml:"type,attr"`
		Product  string `xml:"product"`
		Hardware struct {
			Vendor  string `xml:"vendor"`
			Version string `xml:"version"`
			Serial  string `xml:"serial"`
			UUID    string `xml:"uuid"`
		} `xml:"hardware"`
		Firmware struct {
			Vendor      string `xml:"vendor"`
			Version     string `xml:"version"`
			ReleaseDate string `xml:"release_date"`
		} `xml:"firmware"`
	} `xml:"capability"`
}

type DevicePCI struct {
	Device     xml.Name `xml:"device"`
	Name       string   `xml:"name"`
	Path       string   `xml:"path"`
	Parent     string   `xml:"parent"`
	Capability struct {
		Type     string `xml:"type,attr"`
		Class    string `xml:"class"`
		Domain   int    `xml:"domain"`
		Bus      int    `xml:"bus"`
		Slot     int    `xml:"slot"`
		Function int    `xml:"function"`
		Product  struct {
			Name string `xml:",chardata"`
			ID   string `xml:"id,attr"`
		} `xml:"product"`
		Vendor struct {
			Name string `xml:",chardata"`
			ID   string `xml:"id,attr"`
		} `xml:"vendor"`
		IommuGroup struct {
			IommuGroup xml.Name `xml:"iommuGroup"`
			Number     int      `xml:"number,attr"`
			Addresses  []struct {
				Address  xml.Name `xml:"address"`
				Domain   string   `xml:"domain,attr"`
				Bus      string   `xml:"bus,attr"`
				Slot     string   `xml:"slot,attr"`
				Function string   `xml:"function,attr"`
			} `xml:"address"`
		} `xml:"iommuGroup"`
	} `xml:"capability"`
}

type DeviceUSBDevice struct {
	Device     xml.Name `xml:"device"`
	Name       string   `xml:"name"`
	Path       string   `xml:"path"`
	Parent     string   `xml:"parent"`
	Capability struct {
		Type    string `xml:"type,attr"`
		Bus     string `xml:"bus"`
		Device  string `xml:"device"`
		Product struct {
			Name string `xml:",chardata"`
			ID   string `xml:"id,attr"`
		} `xml:"product"`
		Vendor struct {
			Name string `xml:",chardata"`
			ID   string `xml:"id,attr"`
		} `xml:"vendor"`
	} `xml:"capability"`
}

type DeviceUSB struct {
	Device     xml.Name `xml:"device"`
	Name       string   `xml:"name"`
	Path       string   `xml:"path"`
	Parent     string   `xml:"parent"`
	Capability struct {
		Type        string `xml:"type,attr"`
		Number      int    `xml:"number"`
		Class       int    `xml:"class"`
		Subclass    int    `xml:"subclass"`
		Protocol    int    `xml:"protocol"`
		Description string `xml:"description"`
	} `xml:"capability"`
}

type DeviceNet struct {
	Device     xml.Name `xml:"device"`
	Name       string   `xml:"name"`
	Path       string   `xml:"path"`
	Parent     string   `xml:"parent"`
	Capability struct {
		Type      string `xml:"type,attr"`
		Interface string `xml:"interface"`
		Address   string `xml:"address"`
		Link      struct {
			Speed string `xml:"speed,attr"`
			State string `xml:"state,attr"`
		} `xml:"link"`
		Feature struct {
			Name []string `xml:"name,attr"`
		} `xml:"feature"`
		Cappability struct {
			Type string `xml:"type,attr"`
		} `xml:"capability"`
	} `xml:"capability"`
}

type DeviceSCSI struct {
	Device     xml.Name `xml:"device"`
	Name       string   `xml:"name"`
	Path       string   `xml:"path"`
	Parent     string   `xml:"parent"`
	Capability struct {
		Type     string `xml:"type,attr"`
		Host     int    `xml:"host"`
		Bus      int    `xml:"bus"`
		Target   int    `xml:"target"`
		Lun      int    `xml:"lun"`
		ScsiType string `xml:"type"`
	} `xml:"capability"`
}

type DeviceSCSIHost struct {
	Device     xml.Name `xml:"device"`
	Name       string   `xml:"name"`
	Path       string   `xml:"path"`
	Parent     string   `xml:"parent"`
	Capability struct {
		Type     string `xml:"type,attr"`
		Host     int    `xml:"host"`
		UniqueID int    `xml:"unique_id"`
	} `xml:"capability"`
}

type DeviceStorage struct {
	Device     xml.Name `xml:"device"`
	Name       string   `xml:"name"`
	Path       string   `xml:"path"`
	Parent     string   `xml:"parent"`
	Capability struct {
		Type             string `xml:"type,attr"`
		Block            string `xml:"block"`
		DriveType        string `xml:"drive_type"`
		Model            string `xml:"model"`
		Serial           string `xml:"serial"`
		Size             int    `xml:"size"`
		LogicalBlockSize int    `xml:"logical_block_size"`
		NumBlocks        int    `xml:"num_blocks"`
	} `xml:"capability"`
}

type DeviceDRM struct {
	Device  xml.Name `xml:"device"`
	Name    string   `xml:"name"`
	Path    string   `xml:"path"`
	Parent  string   `xml:"parent"`
	Devnode []struct {
		Type string `xml:"type,attr"`
		Path string `xml:",chardata"`
	} `xml:"devnode"`
	Capability struct {
		Type    string `xml:"type,attr"`
		DRMType string `xml:"type"`
	} `xml:"capability"`
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
			"devnode": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": { // drm
							Type:     schema.TypeString,
							Computed: true,
						},
						"path": { // drm
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"capability": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"class": { // pci, usb
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"domain": { // pci
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"bus": { // pci, usb_device
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"slot": { // pci
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"function": { // pci
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"block": { // storage
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"drive_type": { // storage
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"model": { // storage
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"serial": { // storage
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"size": { // storage
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"logical_block_size": { // storage
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"num_blocks": { // storage
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"hardware": { // system
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"firmware": { // system
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"product": { // pci, system, usb_device
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"vendor": { // pci, usb_device
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"iommu_group": { // pci
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"number": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"addresses": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeMap,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
									},
								},
							},
						},
						"device": { // usb_device
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"number": { // usb
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"subclass": { // usb
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"protocol": { // usb
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"description": { // usb
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"interface": { // net
							Type:     schema.TypeString,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"address": { // net
							Type:     schema.TypeString,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"link": { // net
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"capability": { // net
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"features": { // net
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"host": { // scsi, scsi_host
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"target": { // scsi
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"lun": { // scsi
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"scsi_type": { // scsi
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"unique_id": { // scsi_host
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"drm_type": { // drm
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
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

	var deviceName string

	if name, ok := d.GetOk("name"); ok {
		deviceName = name.(string)
		log.Printf("[DEBUG] Got name: %s", deviceName)
	}

	device, err := virConn.NodeDeviceLookupByName(deviceName)
	if err != nil {
		return fmt.Errorf("failed to lookup node device: %w", err)
	}

	xmlDesc, err := virConn.NodeDeviceGetXMLDesc(device.Name, 0)
	if err != nil {
		return fmt.Errorf("failed to get XML for node device: %w", err)
	}

	deviceXML := DeviceGeneric{}
	capability := []map[string]interface{}{}

	err = xml.Unmarshal([]byte(xmlDesc), &deviceXML)
	if err != nil {
		log.Fatalf("failed to unmarshal XML into deviceXML: %v", err)
	}
	log.Printf("[DEBUG] Parsed device generic into deviceXML: %#v", deviceXML)

	switch deviceXML.Capability.Type {
	////////////////// SYSTEM //////////////////
	case "system":
		deviceXML := DeviceSystem{}
		err = xml.Unmarshal([]byte(xmlDesc), &deviceXML)
		if err != nil {
			log.Fatalf("failed to unmarshal device system XML into deviceXML: %v", err)
		}
		log.Printf("[DEBUG] Parsed device system into deviceXML: %v", deviceXML)

		c := map[string]interface{}{
			// clash with pci.product which is a map, converted to map
			"product": map[string]interface{}{
				"name": deviceXML.Capability.Product,
			},
			"hardware": map[string]interface{}{
				"vendor":  deviceXML.Capability.Hardware.Vendor,
				"version": deviceXML.Capability.Hardware.Version,
				"serial":  deviceXML.Capability.Hardware.Serial,
				"uuid":    deviceXML.Capability.Hardware.UUID,
			},
			"firmware": map[string]interface{}{
				"vendor":       deviceXML.Capability.Firmware.Vendor,
				"version":      deviceXML.Capability.Firmware.Version,
				"release_date": deviceXML.Capability.Firmware.ReleaseDate,
			},
		}

		capability = append(capability, c)
		log.Printf("[DEBUG] Set capability type system to : %v", capability)

	///////////////// PCI //////////////////
	case "pci":
		deviceXML := DevicePCI{}
		err = xml.Unmarshal([]byte(xmlDesc), &deviceXML)
		if err != nil {
			log.Fatalf("failed to unmarshal device pci XML into deviceXML: %v", err)
		}
		log.Printf("[DEBUG] Parsed device pci into deviceXML: %v", deviceXML)

		iommuGroup := []map[string]interface{}{}
		addresses := []map[string]interface{}{}

		for _, v := range deviceXML.Capability.IommuGroup.Addresses {
			m := map[string]interface{}{
				"domain":   v.Domain,
				"bus":      v.Bus,
				"slot":     v.Slot,
				"function": v.Function,
			}
			addresses = append(addresses, m)
		}

		ig := map[string]interface{}{
			"number":    fmt.Sprintf("%d", deviceXML.Capability.IommuGroup.Number),
			"addresses": addresses,
		}
		c := map[string]interface{}{
			"type":     deviceXML.Capability.Type,
			"class":    deviceXML.Capability.Class,
			"domain":   fmt.Sprintf("%d", deviceXML.Capability.Domain),
			"bus":      fmt.Sprintf("%d", deviceXML.Capability.Bus),
			"slot":     fmt.Sprintf("%d", deviceXML.Capability.Slot),
			"function": fmt.Sprintf("%d", deviceXML.Capability.Function),
			"product": map[string]interface{}{
				"id":   deviceXML.Capability.Product.ID,
				"name": deviceXML.Capability.Product.Name,
			},
			"vendor": map[string]interface{}{
				"id":   deviceXML.Capability.Vendor.ID,
				"name": deviceXML.Capability.Vendor.Name,
			},
			"iommu_group": append(iommuGroup, ig),
		}
		log.Printf("[DEBUG] iommuGroup device type pci to : %v", deviceXML.Capability.IommuGroup.Addresses)
		capability = append(capability, c)
		log.Printf("[DEBUG] Set capability type pci to : %v", capability)

	////////////////// USB Device //////////////////
	case "usb_device":
		deviceXML := DeviceUSBDevice{}
		err = xml.Unmarshal([]byte(xmlDesc), &deviceXML)
		if err != nil {
			log.Fatalf("failed to unmarshal device usb_device XML into deviceXML: %v", err)
		}
		log.Printf("[DEBUG] Parsed device usb_device into deviceXML: %v", deviceXML)

		c := map[string]interface{}{
			"type":   deviceXML.Capability.Type,
			"bus":    deviceXML.Capability.Bus,
			"device": deviceXML.Capability.Device,
			"product": map[string]interface{}{
				"id":   deviceXML.Capability.Product.ID,
				"name": deviceXML.Capability.Product.Name,
			},
			"vendor": map[string]interface{}{
				"id":   deviceXML.Capability.Vendor.ID,
				"name": deviceXML.Capability.Vendor.Name,
			},
		}
		capability = append(capability, c)
		log.Printf("[DEBUG] Set capability type usb_device to : %v", capability)

	////////////////// USB Host //////////////////
	case "usb":
		deviceXML := DeviceUSB{}
		err = xml.Unmarshal([]byte(xmlDesc), &deviceXML)
		if err != nil {
			log.Fatalf("failed to unmarshal device usb XML into deviceXML: %v", err)
		}
		log.Printf("[DEBUG] Parsed device usb into deviceXML: %v", deviceXML)

		c := map[string]interface{}{
			"type":        deviceXML.Capability.Type,
			"number":      fmt.Sprintf("%d", deviceXML.Capability.Number),
			"class":       fmt.Sprintf("%d", deviceXML.Capability.Class),
			"subclass":    fmt.Sprintf("%d", deviceXML.Capability.Subclass),
			"protocol":    fmt.Sprintf("%d", deviceXML.Capability.Protocol),
			"description": deviceXML.Capability.Description,
		}
		capability = append(capability, c)
		log.Printf("[DEBUG] Set capability type usb to : %v", capability)

	////////////////// STORAGE //////////////////
	case "storage":
		deviceXML := DeviceStorage{}
		err = xml.Unmarshal([]byte(xmlDesc), &deviceXML)
		if err != nil {
			log.Fatalf("failed to unmarshal device storage XML into deviceXML: %v", err)
		}
		log.Printf("[DEBUG] Parsed device storage into deviceXML: %v", deviceXML)

		c := map[string]interface{}{
			"type":               deviceXML.Capability.Type,
			"block":              deviceXML.Capability.Block,
			"drive_type":         deviceXML.Capability.DriveType,
			"model":              deviceXML.Capability.Model,
			"serial":             deviceXML.Capability.Serial,
			"size":               fmt.Sprintf("%d", deviceXML.Capability.Size),
			"logical_block_size": fmt.Sprintf("%d", deviceXML.Capability.LogicalBlockSize),
			"num_blocks":         fmt.Sprintf("%d", deviceXML.Capability.NumBlocks),
		}

		capability = append(capability, c)
		log.Printf("[DEBUG] Set capability type storage to : %v", capability)

	////////////////// NET //////////////////
	case "net":
		deviceXML := DeviceNet{}
		err = xml.Unmarshal([]byte(xmlDesc), &deviceXML)
		if err != nil {
			log.Fatalf("failed to unmarshal device net XML into deviceXML: %v", err)
		}
		log.Printf("[DEBUG] Parsed device net into deviceXML: %v", deviceXML)

		c := map[string]interface{}{
			"type":      deviceXML.Capability.Type,
			"interface": deviceXML.Capability.Interface,
			"address":   deviceXML.Capability.Address,
			"link": map[string]interface{}{
				"speed": deviceXML.Capability.Link.Speed,
				"state": deviceXML.Capability.Link.State,
			},
			"features": deviceXML.Capability.Feature.Name,
			"capability": map[string]interface{}{
				"type": deviceXML.Capability.Cappability.Type,
			},
		}

		capability = append(capability, c)
		log.Printf("[DEBUG] Set capability type net to : %v", capability)

	////////////////// SCSI //////////////////
	case "scsi":
		deviceXML := DeviceSCSI{}
		err = xml.Unmarshal([]byte(xmlDesc), &deviceXML)
		if err != nil {
			log.Fatalf("failed to unmarshal device scsi XML into deviceXML: %v", err)
		}
		log.Printf("[DEBUG] Parsed device scsi into deviceXML: %v", deviceXML)

		c := map[string]interface{}{
			"type":      deviceXML.Capability.Type,
			"host":      fmt.Sprintf("%d", deviceXML.Capability.Host),
			"bus":       fmt.Sprintf("%d", deviceXML.Capability.Bus),
			"target":    fmt.Sprintf("%d", deviceXML.Capability.Target),
			"lun":       fmt.Sprintf("%d", deviceXML.Capability.Lun),
			"scsi_type": deviceXML.Capability.ScsiType,
		}

		capability = append(capability, c)
		log.Printf("[DEBUG] Set capability type scsi to : %v", capability)

	////////////////// SCSI Host //////////////////
	case "scsi_host":
		deviceXML := DeviceSCSIHost{}
		err = xml.Unmarshal([]byte(xmlDesc), &deviceXML)
		if err != nil {
			log.Fatalf("failed to unmarshal device scsi_host XML into deviceXML: %v", err)
		}
		log.Printf("[DEBUG] Parsed device scsi_host into deviceXML: %v", deviceXML)

		c := map[string]interface{}{
			"type":      deviceXML.Capability.Type,
			"host":      fmt.Sprintf("%d", deviceXML.Capability.Host),
			"unique_id": fmt.Sprintf("%d", deviceXML.Capability.UniqueID),
		}

		capability = append(capability, c)
		log.Printf("[DEBUG] Set capability type scsi_host to : %v", capability)

	////////////////// SCSI Host //////////////////
	case "drm":
		deviceXML := DeviceDRM{}
		err = xml.Unmarshal([]byte(xmlDesc), &deviceXML)
		if err != nil {
			log.Fatalf("failed to unmarshal device drm XML into deviceXML: %v", err)
		}
		log.Printf("[DEBUG] Parsed device drm into deviceXML: %v", deviceXML)

		c := map[string]interface{}{
			"type":     deviceXML.Capability.Type,
			"drm_type": deviceXML.Capability.DRMType,
		}

		capability = append(capability, c)

		devnode := []map[string]interface{}{}
		for _, v := range deviceXML.Devnode {
			m := map[string]interface{}{
				"type": v.Type,
				"path": v.Path,
			}
			devnode = append(devnode, m)
		}

		log.Printf("[DEBUG] Set capability type drm to : %v", capability)
		d.Set("devnode", devnode)
		log.Printf("[DEBUG] Set devnode type drm to : %v", deviceXML.Devnode)

	default:
		return fmt.Errorf("unknown device capability type: %v", deviceXML.Capability.Type)
	}

	d.Set("xml", xmlDesc)
	d.Set("path", deviceXML.Path)
	d.Set("parent", deviceXML.Parent)
	log.Printf("[DEBUG] d.Set capability type %s : %s", capability[0]["type"], capability)
	d.Set("capability", capability)
	d.SetId(strconv.Itoa(hashcode.String(fmt.Sprintf("%v", xmlDesc))))

	return nil
}
