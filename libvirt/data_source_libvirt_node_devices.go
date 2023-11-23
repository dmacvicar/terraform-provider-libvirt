package libvirt

import (
	"fmt"
	"log"
	"sort"
	"strconv"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/dmacvicar/terraform-provider-libvirt/libvirt/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// a libvirt nodeinfo datasource
//
// Datasource example:
//
//	data "libvirt_node_devices" "list" {
//	}
//
//	output "cpus" {
//	  value = data.libvirt_node_devices.list.devices
//	}
func datasourceLibvirtNodeDevices() *schema.Resource {
	return &schema.Resource{
		Read: resourceLibvirtNodeDevicesRead,
		Schema: map[string]*schema.Schema{
			"capability": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"devices": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
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

	var cap libvirt.OptString

	if capability, ok := d.GetOk("capability"); ok {
		cap = append(cap, capability.(string))
		log.Printf("[DEBUG] Got capability: %v", cap)
	}

	rnum, err := virConn.NodeNumOfDevices(libvirt.OptString(cap), 0)
	if err != nil {
		return fmt.Errorf("failed to retrieve number of devices: %w", err)
	}

	devices, err := virConn.NodeListDevices(cap, rnum, 0)
	if err != nil {
		return fmt.Errorf("failed to retrieve list of node devices: %w", err)
	}

	sort.Strings(devices)
	d.Set("devices", devices)
	d.SetId(strconv.Itoa(hashcode.String(fmt.Sprintf("%v", devices))))

	return nil
}
