package libvirt

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestTransformXML(t *testing.T) {
	const xslt = `
<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
  <xsl:output method="xml" version="1.0" encoding="UTF-8" indent="no" omit-xml-declaration="yes"/>
  <xsl:strip-space elements="*"/>
  <xsl:template match="@*|node()">
    <xsl:copy>
      <xsl:apply-templates select="@*|node()"/>
    </xsl:copy>
  </xsl:template>
  <xsl:template match="@format[parent::book]">
    <xsl:attribute name="format">
      <xsl:value-of select="'kindle'"/>
    </xsl:attribute>
  </xsl:template>
</xsl:stylesheet>
`
	const inXML string = "<books><book format=\"paper\"/></books>"
	const outXML string = "<books><book format=\"kindle\"/></books>"

	result, err := transformXML(inXML, xslt)
	assert.Nil(t, err)
	assert.Equal(t, outXML, result)
}

func TestTransformXMLEmptyXSLTNoOp(t *testing.T) {
	const xslt = ""
	const inXML string = "<books><book format=\"paper\"/></books>"

	result, err := transformXML(inXML, xslt)
	assert.Nil(t, err)
	assert.Equal(t, inXML, result)
}

func TestXSLTDiffSupressFunc(t *testing.T) {
	const inXML string = `    <foo>

      <la two="two" one="one">foo this is a test</la>

  <be>bebe  </be>
  </foo>
`
	const outXML string = `<?xml version="1.0"?>
<foo><la two="two" one="one">foo this is a test</la><be>bebe  </be></foo>
`

	assert.True(t, xsltDiffSupressFunc("K", inXML, outXML, &schema.ResourceData{}))
}
