package libvirt

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"time"

	libvirt "github.com/dmacvicar/libvirt-go"
	"github.com/davecgh/go-spew/spew"
)

var diskLetters []rune = []rune("abcdefghijklmnopqrstuvwxyz")

func DiskLetterForIndex(i int) string {

	q := i / len(diskLetters)
	r := i % len(diskLetters)
	letter := diskLetters[r]

	if q == 0 {
		return fmt.Sprintf("%c", letter)
	}

	return fmt.Sprintf("%s%c", DiskLetterForIndex(q-1), letter)
}

// wait for success and timeout after 5 minutes.
func WaitForSuccess(errorMessage string, f func() error) error {
	start := time.Now()
	for {
		err := f()
		if err == nil {
			return nil
		}
		log.Printf("[DEBUG] %s. Re-trying.\n", err)

		time.Sleep(1 * time.Second)
		if time.Since(start) > 5*time.Minute {
			return fmt.Errorf("%s: %s", errorMessage, err)
		}
	}
}

// return an indented XML
func xmlMarshallIndented(b interface{}) (string, error) {
	buf := new(bytes.Buffer)
	enc := xml.NewEncoder(buf)
	enc.Indent("  ", "    ")
	if err := enc.Encode(b); err != nil {
		fmt.Errorf("could not marshall this:\n%s", spew.Sdump(b))
	}
	return buf.String(), nil
}

// Remove the volume identified by `key` from libvirt
func RemoveVolume(virConn *libvirt.VirConnection, key string) error {
	volume, err := virConn.LookupStorageVolByKey(key)
	if err != nil {
		return fmt.Errorf("Can't retrieve volume %s", key)
	}
	defer volume.Free()

	// Refresh the pool of the volume so that libvirt knows it is
	// not longer in use.
	volPool, err := volume.LookupPoolByVolume()
	if err != nil {
		return fmt.Errorf("Error retrieving pool for volume: %s", err)
	}
	defer volPool.Free()

	WaitForSuccess("Error refreshing pool for volume", func() error {
		return volPool.Refresh(0)
	})

	// Workaround for redhat#1293804
	// https://bugzilla.redhat.com/show_bug.cgi?id=1293804#c12
	// Does not solve the problem but it makes it happen less often.
	_, err = volume.GetXMLDesc(0)
	if err != nil {
		return fmt.Errorf("Can't retrieve volume %s XML desc: %s", key, err)
	}

	err = volume.Delete(0)
	if err != nil {
		return fmt.Errorf("Can't delete volume %s: %s", key, err)
	}

	return nil

}
