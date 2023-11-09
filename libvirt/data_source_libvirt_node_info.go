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
func datasourceLibvirtNodeInfo() *schema.Resource {
	return &schema.Resource{
		Read: resourceLibvirtNodeInfoRead,
		Schema: map[string]*schema.Schema{
			"cpu_model": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"memory_total_kb": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"cpu_cores_total": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"numa_nodes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"cpu_sockets": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"cpu_cores_per_socket": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"cpu_threads_per_core": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceLibvirtNodeInfoRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Read data source libvirt_nodeinfo")

	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	model, memory, cpus, _, nodes, sockets, cores, threads, err := virConn.NodeGetInfo()

	if err != nil {
		return fmt.Errorf("failed to retrieve domains: %w", err)
	}

	d.Set("cpu_model", int8ToString(model))
	d.Set("cpu_cores_total", cpus)
	d.Set("cpu_cores_per_socket", cores)
	d.Set("numa_nodes", nodes)
	d.Set("cpu_sockets", sockets)
	d.Set("cpu_threads_per_core", threads)
	d.Set("numa_nodes", nodes)
	d.Set("memory_total_kb", memory)
	d.SetId(strconv.Itoa(hashcode.String(fmt.Sprintf("%v%v%v%v%v%v%v", model, memory, cpus, nodes, sockets, cores, threads))))

	return nil
}

func int8ToString(bs [32]int8) string {
	ba := []uint8{}
	for _, b := range bs {
		if b == 0 {
			break
		}
		ba = append(ba, uint8(b))
	}
	return string(ba)
}
