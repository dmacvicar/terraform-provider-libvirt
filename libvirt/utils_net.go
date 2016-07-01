package libvirt

import (
	"crypto/rand"
	"fmt"
	"net"
)

const (
	maxIfaceNum = 100
)

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

	// avoid libvirt-reserved addresses
	if buf[0] == 0xfe {
		buf[0] = 0xee
	}

	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		buf[0], buf[1], buf[2], buf[3], buf[4], buf[5]), nil
}

func FreeNetworkInterface(basename string) (string, error) {
	for i := 0; i < maxIfaceNum; i++ {
		ifaceName := fmt.Sprintf("%s%d", basename, i)
		_, err := net.InterfaceByName(ifaceName)
		if err != nil {
			return ifaceName, nil
		}
	}
	return "", fmt.Errorf("could not obtain a free network interface")
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
