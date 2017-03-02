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

func TestDefaultDiskMarshall(t *testing.T) {
	b := newDefDisk()
	prettyB := spew.Sdump(b)
	t.Logf("Parsed default disk:\n%s", prettyB)

	buf := new(bytes.Buffer)
	enc := xml.NewEncoder(buf)
	enc.Indent("  ", "    ")
	if err := enc.Encode(b); err != nil {
		t.Fatalf("could not marshall this:\n%s", spew.Sdump(b))
	}
	t.Logf("Marshalled default disk:\n%s", buf.String())
}

func TestDefaultCDROMMarshall(t *testing.T) {
	b := newCDROM()
	prettyB := spew.Sdump(b)
	t.Logf("Parsed default cdrom:\n%s", prettyB)

	buf := new(bytes.Buffer)
	enc := xml.NewEncoder(buf)
	enc.Indent("  ", "    ")
	if err := enc.Encode(b); err != nil {
		t.Fatalf("could not marshall this:\n%s", spew.Sdump(b))
	}
	t.Logf("Marshalled default cdrom:\n%s", buf.String())
}
