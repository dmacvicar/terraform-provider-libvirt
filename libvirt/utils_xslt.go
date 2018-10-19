package libvirt

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

// this function applies a XSLT transform to the xml data
func transformXML(xml string, xslt string) (string, error) {
	xsltFile, err := ioutil.TempFile("", "terraform-provider-libvirt-xslt")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(xsltFile.Name()) // clean up

	// we trim the xslt as it may contain space before the xml declaration
	// because of HCL heredoc
	if _, err := xsltFile.Write([]byte(strings.TrimSpace(xslt))); err != nil {
		log.Fatal(err)
	}

	if err := xsltFile.Close(); err != nil {
		log.Fatal(err)
	}

	xmlFile, err := ioutil.TempFile("", "terraform-provider-libvirt-xml")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(xmlFile.Name()) // clean up

	if _, err := xmlFile.Write([]byte(xml)); err != nil {
		log.Fatal(err)
	}

	if err := xmlFile.Close(); err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("xsltproc", xsltFile.Name(), xmlFile.Name())
	transformedXML, err := cmd.Output()
	if err != nil {
		return xml, err
	}
	log.Printf("[DEBUG] Transformed XML with user specified XSLT:\n%s", transformedXML)
	return string(transformedXML), nil
}

// this function applies a XSLT transform to the xml data
// and is to be reused in all resource types
// your resource need to have a xml.xslt element in the schema
func transformResourceXML(xml string, d *schema.ResourceData) (string, error) {
	xslt, ok := d.GetOk("xml.0.xslt")
	if !ok {
		return xml, nil
	}

	return transformXML(xml, xslt.(string))
}
