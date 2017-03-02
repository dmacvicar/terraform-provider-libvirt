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

func TestDefaultVolumeMarshall(t *testing.T) {
	b := newDefVolume()
	prettyB := spew.Sdump(b)
	t.Logf("Parsed default volume:\n%s", prettyB)

	buf := new(bytes.Buffer)
	enc := xml.NewEncoder(buf)
	enc.Indent("  ", "    ")
	if err := enc.Encode(b); err != nil {
		t.Fatalf("could not marshall this:\n%s", spew.Sdump(b))
	}
	t.Logf("Marshalled default volume:\n%s", buf.String())
}
