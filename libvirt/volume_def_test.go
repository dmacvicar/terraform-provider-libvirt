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

func TestVolumeUnmarshal(t *testing.T) {
	xmlDesc := `
	<volume type='file'>
	  <name>caasp_master.img</name>
	  <key>/home/user/.libvirt/images/image.img</key>
	  <source>
	  </source>
	  <capacity unit='bytes'>42949672960</capacity>
	  <allocation unit='bytes'>900800512</allocation>
	  <target>
	    <path>/home/user/.libvirt/images/image.img</path>
	    <format type='qcow2'/>
	    <permissions>
	      <mode>0644</mode>
	      <owner>480</owner>
	      <group>473</group>
	    </permissions>
	    <timestamps>
	      <atime>1488789260.012293492</atime>
	      <mtime>1488802938.454893390</mtime>
	      <ctime>1488802938.454893390</ctime>
	    </timestamps>
	  </target>
	  <backingStore>
	    <path>/home/user/.libvirt/images/image.img</path>
	    <format type='qcow2'/>
	    <permissions>
	      <mode>0644</mode>
	      <owner>480</owner>
	      <group>473</group>
	    </permissions>
	    <timestamps>
	      <atime>1488541864.606322102</atime>
	      <mtime>1488541858.638308597</mtime>
	      <ctime>1488541864.526321921</ctime>
	    </timestamps>
	  </backingStore>
	</volume>
	`

	_, err := newDefVolumeFromXML(xmlDesc)
	if err != nil {
		t.Fatalf("could not unmarshall volume definition:\n%s", err)
	}
}

func TestDefaultVolumeMarshall(t *testing.T) {
	b := newDefVolume()

	buf := new(bytes.Buffer)
	enc := xml.NewEncoder(buf)
	enc.Indent("  ", "    ")
	if err := enc.Encode(b); err != nil {
		t.Fatalf("could not marshall this:\n%s", spew.Sdump(b))
	}
}
