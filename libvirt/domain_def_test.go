package libvirt

import (
	"bytes"
	"encoding/xml"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func init() {
	spew.Config.Indent = "\t"
}

func TestDefaultDomainMarshall(t *testing.T) {
	b := newDomainDef()
	prettyB := spew.Sdump(b)
	t.Logf("Parsed default domain:\n%s", prettyB)

	buf := new(bytes.Buffer)
	enc := xml.NewEncoder(buf)
	enc.Indent("  ", "    ")
	if err := enc.Encode(b); err != nil {
		t.Fatalf("could not marshall this:\n%s", spew.Sdump(b))
	}
	t.Logf("Marshalled default domain:\n%s", buf.String())
}
