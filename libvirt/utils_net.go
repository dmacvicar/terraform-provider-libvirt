package libvirt

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

const (
	maxIfaceNum = 100
)

// randomMACAddress returns a randomized MAC address
// with libvirt prefix
func randomMACAddress() (string, error) {
	buf := make([]byte, 3)
	rand.Seed(time.Now().UnixNano())
	if _, err := rand.Read(buf); err != nil {
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

	return fmt.Sprintf("52:54:00:%02x:%02x:%02x",
		buf[0], buf[1], buf[2]), nil
}

// randomPort returns a random port
func randomPort() int {
	const minPort = 1024
	const maxPort = 65535

	rand.Seed(time.Now().UnixNano())
	return rand.Intn(maxPort-minPort) + minPort
}

func getNetMaskWithMax16Bits(m net.IPMask) net.IPMask {
	ones, bits := m.Size()

	if bits-ones > 16 {
		if bits == 128 {
			// IPv6 Mask with max 16 bits
			return net.CIDRMask(128-16, 128)
		}

		// IPv4 Mask with max 16 bits
		return net.CIDRMask(32-16, 32)
	}

	return m
}

func getLastIP(network *net.IPNet, netIP net.IP) net.IP {
	lastIP := make(net.IP, len(netIP))

	// intermediate network mask with max 16 bits for hosts
	// We need a mask with max 16 bits since libvirt only supports 65535) IP's per subnet
	// 2^16 = 65536 (minus broadcast and .1)
	intMask := getNetMaskWithMax16Bits(network.Mask)
	for i, netIPByte := range netIP {
		lastIP[i] = netIPByte | ^intMask[i]
	}

	return lastIP
}

// networkRange calculates the first and last IP addresses in an IPNet
func networkRange(network *net.IPNet) (firstIP net.IP, lastIP net.IP) {
	netIP := network.IP.To4()
	if netIP == nil {
		netIP = network.IP.To16()
	}

	return netIP.Mask(network.Mask), getLastIP(network, netIP)
}
