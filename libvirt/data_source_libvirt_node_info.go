package libvirt

import (
	"log"
	"fmt"
	"encoding/xml"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// a libvirt based host capability data source
//
// Example usage:
//
//	data "libvirt_nodeinfo" "host_machine" {
//        uri = "qemu+ssh://target-machine/system?"
//	}
//
//      locals {
//      	arch = data.external.native_machine_details.result["arch"]
//      	images = {
//      		"debian-x86_64" = "/srv/images/debian-12-backports-generic-amd64.qcow2",
//    			"debian-aarch64"  = "/srv/images/debian-12-backports-generic-arm64.qcow2",
//      	}
//      image_path = local.images["debian-${local.arch}"]
//      }

func datasourceLibvirtNodeInfo() *schema.Resource {
	return &schema.Resource{
		Read: resourceLibvirtNodeInfoRead,
		Schema: map[string]*schema.Schema{
			"host" : {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"live_migration_support" : {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"live_migration_transports" : {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{ Type: schema.TypeString, },
				Computed: true,
			},
			"arch" : {
				Type:     schema.TypeString,
				Computed: true,
			},
			"model" : {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vendor" : {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sockets" : {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"dies" : {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"cores" : {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"threads" : {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"features": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"topology" : {
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true, // or Required: true, based on your use case
						},
						"memory": {
							Type:     schema.TypeInt,
							Optional: true, // or Required: true, based on your use case
						},
						"cpu": {
							Type:     schema.TypeList,
							Optional: true, // or Required: true, based on your use case
							Elem:     &schema.Schema{
								Type: schema.TypeMap,
								Elem: schema.TypeInt,

							},
						},
					},
				},
				Computed: true,
			},
		},
	}
}

type Host struct {
	XMLName  xml.Name `xml:"host"`
	UUID     string   `xml:"uuid"`
	CPU      CPU      `xml:"cpu"`
	Migration MigrationFeatures `xml:"migration_features"`
	Topology Topology `xml:"topology"`
}

type CPU struct {
	Arch      string       `xml:"arch"`
	Model     string       `xml:"model"`
	Vendor    string       `xml:"vendor"`
	Topology  CPUtopology  `xml:"topology"`
	Features  []struct{
		Name string `xml:"name,attr"`
	}    `xml:"feature"`
}

type MigrationFeatures struct {
	XMLName    xml.Name `xml:"migration_features"`
	Live       *bool `xml:"live"`
	Transports []string `xml:"uri_transports>uri_transport"`
}

type CPUtopology struct {
	Sockets int `xml:"sockets,attr"`
	Dies    int `xml:"dies,attr"`
	Cores   int `xml:"cores,attr"`
	Threads int `xml:"threads,attr"`
}

type Topology struct {
	Cells []Cell `xml:"cells>cell"`
}

type Cell struct {
	ID     int `xml:"id,attr"`
	Memory Memory `xml:"memory"`
	CPUs   CPUs   `xml:"cpus"`
}

type Memory struct {
	Unit  string `xml:"unit,attr"`
	Value int    `xml:",chardata"`
}

type CPUs struct {
	Num  string `xml:"num,attr"`
	CPU  []CPUInfo `xml:"cpu"`
}

type CPUInfo struct {
	ID        string `xml:"id,attr"`
	SocketID  string `xml:"socket_id,attr"`
	DieID     string `xml:"die_id,attr"`
	CoreID    string `xml:"core_id,attr"`
	Siblings  string `xml:"siblings,attr"`
}

// Data represents the top-level structure of the XML that includes the <host> node
type Capabilities struct {
	Host Host `xml:"host"`
}

func convertUnit( unit string ) (int){
	switch unit {
	case "KiB":
		return 1024
	case "MiB":
		return 1024*1024
	default:
		return 1
	}
}

func resourceLibvirtNodeInfoRead(d *schema.ResourceData, meta interface{}) error {

	uri := d.Get("host").(string)
	virConn, err := meta.(*Client).Connection(&uri)
	if virConn == nil {
		return fmt.Errorf("unable to connect for nodeinfo read: %v", err)
	}

	caps, err := virConn.ConnectGetCapabilities()
	if err != nil {
		return err
	}

	var capabilities Capabilities
	err = xml.Unmarshal([]byte(caps), &capabilities)
	if err != nil {
		return fmt.Errorf("Error unmarshalling XML: %v", err)
	}

	d.SetId(capabilities.Host.UUID)

	d.Set("arch", capabilities.Host.CPU.Arch)
	d.Set("model", capabilities.Host.CPU.Model)
	d.Set("vendor", capabilities.Host.CPU.Vendor)
	d.Set("sockets", capabilities.Host.CPU.Topology.Sockets)
	d.Set("dies", capabilities.Host.CPU.Topology.Dies)
	d.Set("cores", capabilities.Host.CPU.Topology.Cores)
	d.Set("threads", capabilities.Host.CPU.Topology.Threads)

	d.Set("live_migration_support", capabilities.Host.Migration.Live != nil)
	if err := d.Set("live_migration_transports", capabilities.Host.Migration.Transports); err != nil {
		return fmt.Errorf("error setting features: %s", err)
	}

	features := []string{}
	// Iterate over the cpu.Features slice
	for _, feature := range capabilities.Host.CPU.Features {
		// Append the 'Name' field of each 'Feature' struct to the slice
		features = append(features, feature.Name)
	}

	if err := d.Set("features", features); err != nil {
		return fmt.Errorf("error setting features: %s", err)
	}

	// topology := make([]map[string]interface{}, 0)
	// for _, t := range capabilities.Host.Topology {
	// 	topologyItem := make(map[string]interface{})
	// 	if t.ID != nil { // Assuming `ID`, `Memory`, and `CPU` are fields in your topology data structure
	// 		topologyItem["id"] = *t.ID
	// 	}
	// 	if t.Memory != nil {
	// 		topologyItem["memory"] = *t.Memory
	// 	}
	// 	if t.CPU != nil {
	// 		cpuList := make([]map[string]int, 0)
	// 		for _, cpu := range *t.CPU { // Assuming `CPU` is a slice of maps or similar structure
	// 			cpuItem := make(map[string]int)
	// 			for key, value := range cpu {
	// 				cpuItem[key] = value
	// 			}
	// 			cpuList = append(cpuList, cpuItem)
	// 		}
	// 		topologyItem["cpu"] = cpuList
	// 	}
	// 	topology = append(topology, topologyItem)
	// }

	d.Set("topology", nil)
	log.Printf("[DEBUG] Host topology not yet implemented")

	return nil
}
