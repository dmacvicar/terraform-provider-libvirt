package libvirt

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	identitySpaceStripXSLT = `
<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
  <xsl:strip-space elements="*" />
  <xsl:template match="@*|node()">
    <xsl:copy>
      <xsl:apply-templates select="@*|node()"/>
    </xsl:copy>
  </xsl:template>
</xsl:stylesheet>
`
)

// This is the function we use to detect if the XSLT attribute itself changed
// As we don't want to recreate resources when the XSLT is changed with whitespace,
// we specify the diff suppress function as the result of applying the identity
// transform to the xslt, stripping whitespace
// See https://www.terraform.io/docs/extend/schemas/schema-behaviors.html#diffsuppressfunc
func xsltDiffSupressFunc(_, old, new string, _ *schema.ResourceData) bool {
	oldStrip, err := transformXML(old, identitySpaceStripXSLT)
	if err != nil {
		// fail, just use normal equality
		log.Printf("[ERROR] Couldn't normalize XSLT stylesheet")
		return old == new
	}
	newStrip, err := transformXML(new, identitySpaceStripXSLT)
	if err != nil {
		// fail, just use normal equality
		log.Printf("[ERROR] Couldn't normalize XSLT stylesheet")
		return old == new
	}
	return oldStrip == newStrip
}

// this function applies a XSLT transform to the xml data.
func transformXML(xmlS string, xsltS string) (string, error) {
	// empty xslt is a no-op
	if strings.TrimSpace(xsltS) == "" {
		return xmlS, nil
	}

	xsltFile, err := os.CreateTemp("", "terraform-provider-libvirt-xslt")
	if err != nil {
		return "", err
	}
	defer os.Remove(xsltFile.Name()) // clean up

	// we trim the xslt as it may contain space before the xml declaration
	// because of HCL heredoc
	if _, err := xsltFile.Write([]byte(strings.TrimSpace(xsltS))); err != nil {
		return "", err
	}

	if err := xsltFile.Close(); err != nil {
		return "", err
	}

	xmlFile, err := os.CreateTemp("", "terraform-provider-libvirt-xml")
	if err != nil {
		return "", err
	}
	defer os.Remove(xmlFile.Name()) // clean up

	if _, err := xmlFile.Write([]byte(xmlS)); err != nil {
		return "", err
	}

	if err := xmlFile.Close(); err != nil {
		return "", err
	}

	//nolint:gosec // G204 not sure why gosec complains
	cmd := exec.Command("xsltproc",
		"--nomkdir",
		"--nonet",
		"--nowrite",
		xsltFile.Name(),
		xmlFile.Name())
	transformedXML, err := cmd.Output()
	if err != nil {
		log.Printf("[ERROR] Failed to run xsltproc (is it installed?)")
		return xmlS, err
	}
	log.Printf("[DEBUG] Transformed XML with user specified XSLT:\n%s", transformedXML)

	return string(transformedXML), err
}

// this function applies a XSLT transform to the xml data
// and is to be reused in all resource types
// your resource need to have a xml.xslt element in the schema.
func transformResourceXML(xml string, d *schema.ResourceData) (string, error) {
	xslt, ok := d.GetOk("xml.0.xslt")
	if !ok {
		return xml, nil
	}

	return transformXML(xml, xslt.(string))
}
