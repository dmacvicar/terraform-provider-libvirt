package libvirt

import (
	"fmt"
	"log"
	"strconv"

	libvirt "github.com/digitalocean/go-libvirt"
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
func datasourceLibvirtNodeDevices() *schema.Resource {
	return &schema.Resource{
		Read: resourceLibvirtNodeDevicesRead,
		Schema: map[string]*schema.Schema{
			// "name": {
			// 	Type:     schema.TypeString,
			// 	Required: true,
			// },
			"devices": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceLibvirtNodeDevicesRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Read data source libvirt_nodedevices")

	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	opts := libvirt.OptString{"pci"}

	rnum, err := virConn.NodeNumOfDevices(opts, 0)
	if err != nil {
		return fmt.Errorf("failed to retrieve number of devices: %v", err)
	}

	devices, err := virConn.NodeListDevices(opts, rnum, 0)
	if err != nil {
		return fmt.Errorf("failed to retrieve list of node devices: %v", err)
	}

	d.Set("devices", devices)
	d.SetId(strconv.Itoa(hashcode.String(fmt.Sprintf("%v", devices))))

	return nil
}
