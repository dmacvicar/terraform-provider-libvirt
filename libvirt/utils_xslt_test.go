package libvirt

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestTransformXML(t *testing.T) {
	const inXML string = `    <foo>

      <la two="two" one="one">foo this is a test</la>

  <be>bebe  </be>
  </foo>
`
	const outXML string = `<?xml version="1.0"?>
<foo><la two="two" one="one">foo this is a test</la><be>bebe  </be></foo>
`

	result, err := transformXML(inXML, identitySpaceStripXSLT)
	assert.Nil(t, err)
	assert.Equal(t, outXML, result)
}
