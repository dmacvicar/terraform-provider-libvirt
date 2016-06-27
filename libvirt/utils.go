package libvirt

import (
	"bytes"
	"crypto/rand"
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

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

func RandomMACAddress() (string, error) {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	// set local bit and unicast
	buf[0] = (buf[0] | 2) & 0xfe
	// Set the local bit
	buf[0] |= 2

	mac := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
	return strings.ToUpper(mac), nil
}

// Calculates the first and last IP addresses in an IPNet
func NetworkRange(network *net.IPNet) (net.IP, net.IP) {
	netIP := network.IP.To4()
	lastIP := net.IPv4(0, 0, 0, 0).To4()
	if netIP == nil {
		netIP = network.IP.To16()
		lastIP = net.IPv6zero.To16()
	}
	firstIP := netIP.Mask(network.Mask)
	for i := 0; i < len(lastIP); i++ {
		lastIP[i] = netIP[i] | ^network.Mask[i]
	}
	return firstIP, lastIP
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
