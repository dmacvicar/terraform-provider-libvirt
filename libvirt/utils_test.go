package libvirt

import (
	"net"
	"testing"
)

func TestDiskLetterForIndex(t *testing.T) {

	diskNumbers := []int{0, 1, 2, 3, 4, 16, 24, 25, 26, 30, 300}
	names := []string{"a", "b", "c", "d", "e", "q", "y", "z", "aa", "ae", "ko"}

	for i, diskNumber := range diskNumbers {
		ret := DiskLetterForIndex(diskNumber)
		if ret != names[i] {
			t.Errorf("Expected %s, got %s for disk %d", names[i], ret, diskNumber)
		}
	}
}

func TestIPsRange(t *testing.T) {
	_, net, err := net.ParseCIDR("192.168.18.1/24")
	if err != nil {
		t.Errorf("When parsing network: %s", err)
	}

	start, end := NetworkRange(net)
	if start.String() != "192.168.18.0" {
		t.Errorf("unexpected range start for '%s': %s", net, start)
	}
	if end.String() != "192.168.18.255" {
		t.Errorf("unexpected range start for '%s': %s", net, start)
	}
}
