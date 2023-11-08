package libvirt

import (
	"encoding/xml"
	"fmt"
	"log"
	"strconv"

	"github.com/dmacvicar/terraform-provider-libvirt/libvirt/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// strutures to parse device XML just enough to get Capability.Type
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
			Id   string `xml:"id,attr"`
		} `xml:"product"`
		Vendor struct {
			Name string `xml:",chardata"`
			Id   string `xml:"id,attr"`
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
			Id   string `xml:"id,attr"`
		} `xml:"product"`
		Vendor struct {
			Name string `xml:",chardata"`
			Id   string `xml:"id,attr"`
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
						// "iommu_group": { // pci
						// 	Type:     schema.TypeMap,
						// 	Optional: true,
						// 	Elem: &schema.Schema{
						// 		Type: schema.TypeString,
						// 	},
						// },
						"iommu_group": {
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

// https://github.com/hashicorp/terraform-plugin-sdk/issues/726

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

	device_xml := DeviceGeneric{}
	capability := []map[string]interface{}{}

	err = xml.Unmarshal([]byte(xml_desc), &device_xml)
	if err != nil {
		log.Fatalf("failed to unmarshal XML into device_xml: %v", err)
	}
	log.Printf("[DEBUG] Parsed device generic into device_xml: %v", device_xml)

	switch device_xml.Capability.Type {
	////////////////// SYSTEM //////////////////
	case "system":
		device_xml := DeviceSystem{}
		err = xml.Unmarshal([]byte(xml_desc), &device_xml)
		if err != nil {
			log.Fatalf("failed to unmarshal device system XML into device_xml: %v", err)
		}
		log.Printf("[DEBUG] Parsed device system into device_xml: %v", device_xml)

		c := map[string]interface{}{
			// clash with pci.product which is a map, converted to map.
			"product": map[string]interface{}{
				"name": device_xml.Capability.Product,
			},
			"hardware": map[string]interface{}{
				"vendor":  device_xml.Capability.Hardware.Vendor,
				"version": device_xml.Capability.Hardware.Version,
				"serial":  device_xml.Capability.Hardware.Serial,
				"uuid":    device_xml.Capability.Hardware.UUID,
			},
			"firmware": map[string]interface{}{
				"vendor":       device_xml.Capability.Firmware.Vendor,
				"version":      device_xml.Capability.Firmware.Version,
				"release_date": device_xml.Capability.Firmware.ReleaseDate,
			},
		}

		capability = append(capability, c)
		log.Printf("[DEBUG] Set capability type system to : %v", capability)

	///////////////// PCI //////////////////
	case "pci":
		device_xml := DevicePCI{}
		err = xml.Unmarshal([]byte(xml_desc), &device_xml)
		if err != nil {
			log.Fatalf("failed to unmarshal device pci XML into device_xml: %v", err)
		}
		log.Printf("[DEBUG] Parsed device pci into device_xml: %v", device_xml)

		iommu_group := []map[string]interface{}{}
		addresses := []map[string]interface{}{}

		for _, v := range device_xml.Capability.IommuGroup.Addresses {
			m := map[string]interface{}{
				"domain":   v.Domain,
				"bus":      v.Bus,
				"slot":     v.Slot,
				"function": v.Function,
			}
			addresses = append(addresses, m)
		}

		ig := map[string]interface{}{
			"number":    fmt.Sprintf("%d", device_xml.Capability.IommuGroup.Number),
			"addresses": addresses,
		}
		c := map[string]interface{}{
			"type":     device_xml.Capability.Type,
			"class":    device_xml.Capability.Class,
			"domain":   fmt.Sprintf("%d", device_xml.Capability.Domain),
			"bus":      fmt.Sprintf("%d", device_xml.Capability.Bus),
			"slot":     fmt.Sprintf("%d", device_xml.Capability.Slot),
			"function": fmt.Sprintf("%d", device_xml.Capability.Function),
			"product": map[string]interface{}{
				"id":   device_xml.Capability.Product.Id,
				"name": device_xml.Capability.Product.Name,
			},
			"vendor": map[string]interface{}{
				"id":   device_xml.Capability.Vendor.Id,
				"name": device_xml.Capability.Vendor.Name,
			},
			"iommu_group": append(iommu_group, ig),
		}
		log.Printf("[DEBUG] iommu_group device type pci to : %v", device_xml.Capability.IommuGroup.Addresses)
		capability = append(capability, c)
		log.Printf("[DEBUG] Set capability type pci to : %v", capability)

	////////////////// USB Device //////////////////
	case "usb_device":
		device_xml := DeviceUSBDevice{}
		err = xml.Unmarshal([]byte(xml_desc), &device_xml)
		if err != nil {
			log.Fatalf("failed to unmarshal device usb_device XML into device_xml: %v", err)
		}
		log.Printf("[DEBUG] Parsed device usb_device into device_xml: %v", device_xml)

		c := map[string]interface{}{
			"type":   device_xml.Capability.Type,
			"bus":    device_xml.Capability.Bus,
			"device": device_xml.Capability.Device,
			"product": map[string]interface{}{
				"id":   device_xml.Capability.Product.Id,
				"name": device_xml.Capability.Product.Name,
			},
			"vendor": map[string]interface{}{
				"id":   device_xml.Capability.Vendor.Id,
				"name": device_xml.Capability.Vendor.Name,
			},
		}
		capability = append(capability, c)
		log.Printf("[DEBUG] Set capability type usb_device to : %v", capability)

	////////////////// USB Host //////////////////
	case "usb":
		device_xml := DeviceUSB{}
		err = xml.Unmarshal([]byte(xml_desc), &device_xml)
		if err != nil {
			log.Fatalf("failed to unmarshal device usb XML into device_xml: %v", err)
		}
		log.Printf("[DEBUG] Parsed device usb into device_xml: %v", device_xml)

		c := map[string]interface{}{
			"type":        device_xml.Capability.Type,
			"number":      fmt.Sprintf("%d", device_xml.Capability.Number),
			"class":       fmt.Sprintf("%d", device_xml.Capability.Class),
			"subclass":    fmt.Sprintf("%d", device_xml.Capability.Subclass),
			"protocol":    fmt.Sprintf("%d", device_xml.Capability.Protocol),
			"description": device_xml.Capability.Description,
		}
		capability = append(capability, c)
		log.Printf("[DEBUG] Set capability type usb to : %v", capability)

	////////////////// STORAGE //////////////////
	case "storage":
		device_xml := DeviceStorage{}
		err = xml.Unmarshal([]byte(xml_desc), &device_xml)
		if err != nil {
			log.Fatalf("failed to unmarshal device storage XML into device_xml: %v", err)
		}
		log.Printf("[DEBUG] Parsed device storage into device_xml: %v", device_xml)

		c := map[string]interface{}{
			"type":               device_xml.Capability.Type,
			"block":              device_xml.Capability.Block,
			"drive_type":         device_xml.Capability.DriveType,
			"model":              device_xml.Capability.Model,
			"serial":             device_xml.Capability.Serial,
			"size":               fmt.Sprintf("%d", device_xml.Capability.Size),
			"logical_block_size": fmt.Sprintf("%d", device_xml.Capability.LogicalBlockSize),
			"num_blocks":         fmt.Sprintf("%d", device_xml.Capability.NumBlocks),
		}

		capability = append(capability, c)
		log.Printf("[DEBUG] Set capability type storage to : %v", capability)

	////////////////// NET //////////////////
	case "net":
		device_xml := DeviceNet{}
		err = xml.Unmarshal([]byte(xml_desc), &device_xml)
		if err != nil {
			log.Fatalf("failed to unmarshal device net XML into device_xml: %v", err)
		}
		log.Printf("[DEBUG] Parsed device net into device_xml: %v", device_xml)

		c := map[string]interface{}{
			"type":      device_xml.Capability.Type,
			"interface": device_xml.Capability.Interface,
			"address":   device_xml.Capability.Address,
			"link": map[string]interface{}{
				"speed": device_xml.Capability.Link.Speed,
				"state": device_xml.Capability.Link.State,
			},
			"features": device_xml.Capability.Feature.Name,
			"capability": map[string]interface{}{
				"type": device_xml.Capability.Cappability.Type,
			},
		}

		capability = append(capability, c)
		log.Printf("[DEBUG] Set capability type net to : %v", capability)

	////////////////// SCSI //////////////////
	case "scsi":
		device_xml := DeviceSCSI{}
		err = xml.Unmarshal([]byte(xml_desc), &device_xml)
		if err != nil {
			log.Fatalf("failed to unmarshal device scsi XML into device_xml: %v", err)
		}
		log.Printf("[DEBUG] Parsed device scsi into device_xml: %v", device_xml)

		c := map[string]interface{}{
			"type":      device_xml.Capability.Type,
			"host":      fmt.Sprintf("%d", device_xml.Capability.Host),
			"bus":       fmt.Sprintf("%d", device_xml.Capability.Bus),
			"target":    fmt.Sprintf("%d", device_xml.Capability.Target),
			"lun":       fmt.Sprintf("%d", device_xml.Capability.Lun),
			"scsi_type": device_xml.Capability.ScsiType,
		}

		capability = append(capability, c)
		log.Printf("[DEBUG] Set capability type scsi to : %v", capability)

	////////////////// SCSI Host //////////////////
	case "scsi_host":
		device_xml := DeviceSCSIHost{}
		err = xml.Unmarshal([]byte(xml_desc), &device_xml)
		if err != nil {
			log.Fatalf("failed to unmarshal device scsi_host XML into device_xml: %v", err)
		}
		log.Printf("[DEBUG] Parsed device scsi_host into device_xml: %v", device_xml)

		c := map[string]interface{}{
			"type":      device_xml.Capability.Type,
			"host":      fmt.Sprintf("%d", device_xml.Capability.Host),
			"unique_id": fmt.Sprintf("%d", device_xml.Capability.UniqueID),
		}

		capability = append(capability, c)
		log.Printf("[DEBUG] Set capability type scsi_host to : %v", capability)

	////////////////// SCSI Host //////////////////
	case "drm":
		device_xml := DeviceDRM{}
		err = xml.Unmarshal([]byte(xml_desc), &device_xml)
		if err != nil {
			log.Fatalf("failed to unmarshal device drm XML into device_xml: %v", err)
		}
		log.Printf("[DEBUG] Parsed device drm into device_xml: %v", device_xml)

		c := map[string]interface{}{
			"type":     device_xml.Capability.Type,
			"drm_type": device_xml.Capability.DRMType,
		}

		capability = append(capability, c)

		devnode := []map[string]interface{}{}
		for _, v := range device_xml.Devnode {
			m := map[string]interface{}{
				"type": v.Type,
				"path": v.Path,
			}
			devnode = append(devnode, m)
		}

		log.Printf("[DEBUG] Set capability type drm to : %v", capability)
		d.Set("devnode", devnode)
		log.Printf("[DEBUG] Set devnode type drm to : %v", device_xml.Devnode)

	default:
		log.Fatalf("Unknown device capability type: %v", device_xml.Capability.Type)
	}

	d.Set("xml", xml_desc)
	d.Set("path", device_xml.Path)
	d.Set("parent", device_xml.Parent)
	log.Printf("[DEBUG] d.Set capability type %s : %s", capability[0]["type"], capability)
	d.Set("capability", capability)
	d.SetId(strconv.Itoa(hashcode.String(fmt.Sprintf("%v", xml_desc))))

	return nil
}
