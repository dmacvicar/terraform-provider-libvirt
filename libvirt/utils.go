package libvirt

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
	libvirt "github.com/digitalocean/go-libvirt"
)

var diskLetters = []rune("abcdefghijklmnopqrstuvwxyz")

// LibVirtConIsNil is a global string error msg.
const LibVirtConIsNil string = "the libvirt connection was nil"

// diskLetterForIndex return diskLetters for index.
func diskLetterForIndex(i int) string {

	q := i / len(diskLetters)
	r := i % len(diskLetters)
	letter := diskLetters[r]

	if q == 0 {
		return fmt.Sprintf("%c", letter)
	}

	return fmt.Sprintf("%s%c", diskLetterForIndex(q-1), letter)
}

// WaitSleepInterval time.
var WaitSleepInterval = 1 * time.Second

// WaitTimeout time.
var WaitTimeout = 5 * time.Minute

// waitForSuccess wait for success and timeout after 5 minutes.
func waitForSuccess(errorMessage string, f func() error) error {
	start := time.Now()
	for {
		err := f()
		if err == nil {
			return nil
		}
		log.Printf("[DEBUG] %s. Re-trying.\n", err)

		time.Sleep(WaitSleepInterval)
		if time.Since(start) > WaitTimeout {
			return fmt.Errorf("%s: %w", errorMessage, err)
		}
	}
}

// return an indented XML.
func xmlMarshallIndented(b interface{}) (string, error) {
	buf := new(bytes.Buffer)
	enc := xml.NewEncoder(buf)
	enc.Indent("  ", "    ")
	if err := enc.Encode(b); err != nil {
		return "", fmt.Errorf("could not marshall this:\n%s", spew.Sdump(b))
	}
	return buf.String(), nil
}

// formatBoolYesNo is similar to strconv.FormatBool with yes/no instead of true/false.
func formatBoolYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// analog to internal libvirt.checkError
// IsNotFound in libvirt-go should be enhanced to detect the other types
// (pool, network).
func isError(err error, errorCode libvirt.ErrorNumber) bool {
	var perr libvirt.Error
	if errors.As(err, &perr) {
		return  perr.Code == uint32(errorCode)
	}
	return false
}
