package libvirt

import (
	"encoding/xml"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/libvirt/libvirt-go-xml"
)

// TerraformInstanceXML type
type TerraformInstanceXML struct {
	XMLName xml.Name          `xml:"https://terraform.io instance"`
	Tags    []TerraformTagXML `xml:"tag"`
}

// TerraformTagXML type
type TerraformTagXML struct {
	Key   string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

func tagsToXML(tags map[string]interface{}, metadata *TerraformInstanceXML) (out string, err error) {
	// Overwrite existing tags while keeping additional metadata
	metadata.Tags = []TerraformTagXML{}
	for key, value := range tags {
		metadata.Tags = append(metadata.Tags, TerraformTagXML{
			Key:   key,
			Value: value.(string),
		})
	}
	var bytesOut []byte
	if bytesOut, err = xml.Marshal(metadata); err != nil {
		return "", fmt.Errorf("Failed to marshal metadata XML: %s", err)
	}
	return string(bytesOut), nil
}

func setMetadata(d *schema.ResourceData, domain *libvirtxml.Domain) (err error) {
	metadata := TerraformInstanceXML{}
	if err = xml.Unmarshal([]byte(domain.Metadata.XML), &metadata); domain.Metadata.XML != "" && err != nil {
		return fmt.Errorf("XML Unmarshal Error: %s", err)
	}
	if _, ok := d.GetOk("tags"); ok {
		tags := d.Get("tags").(map[string]interface{})
		var out string
		if out, err = tagsToXML(tags, &metadata); err != nil {
			return err
		}
		domain.Metadata = &libvirtxml.DomainMetadata{
			XML: out,
		}
		log.Printf("[DEBUG] Created Metadata XML: %s", domain.Metadata.XML)
	}
	return nil
}

func getMetadata(d *schema.ResourceData, domain *libvirtxml.Domain) error {
	if domain.Metadata != nil && domain.Metadata.XML != "" {
		metadata := TerraformInstanceXML{}
		if err := xml.Unmarshal([]byte(domain.Metadata.XML), &metadata); err != nil {
			return fmt.Errorf("XML Unmarshal Error: %s", err)
		}
		out := make(map[string]string)
		for _, tag := range metadata.Tags {
			out[tag.Key] = tag.Value
		}
		d.Set("tags", out)
	}

	return nil
}
