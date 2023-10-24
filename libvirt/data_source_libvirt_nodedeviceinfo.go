package libvirt

import (
	"fmt"
	"log"
	"strconv"

	"github.com/dmacvicar/terraform-provider-libvirt/libvirt/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// a libvirt nodeinfo datasource
//
// Datasource example:
//
//	data "libvirt_nodeinfo" "info" {
//	}
//
//	output "cpus" {
//	  value = data.libvirt_nodeinfo.info.cpus
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

	xml, err := virConn.NodeDeviceGetXMLDesc(device.Name, 0)
	if err != nil {
		return fmt.Errorf("failed to get XML for node device: %v", err)
	}

	d.Set("xml", xml)
	//d.Set("devices", )
	d.SetId(strconv.Itoa(hashcode.String(fmt.Sprintf("%v", xml))))

	return nil
}
