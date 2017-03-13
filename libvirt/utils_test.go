package libvirt

import (
	"errors"
	"net"
	"testing"
	"time"
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

func TestWaitForSuccessEverythingFine(t *testing.T) {
	wait_sleep := WAIT_SLEEP_INTERVAL
	wait_timeout := WAIT_TIMEOUT
	defer func() {
		WAIT_SLEEP_INTERVAL = wait_sleep
		WAIT_TIMEOUT = wait_timeout
	}()

	WAIT_TIMEOUT = 1 * time.Second
	WAIT_SLEEP_INTERVAL = 1 * time.Nanosecond

	err := WaitForSuccess(
		"boom",
		func() error {
			return nil
		})

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

func TestWaitForSuccessBrokenFunction(t *testing.T) {
	wait_sleep := WAIT_SLEEP_INTERVAL
	wait_timeout := WAIT_TIMEOUT
	defer func() {
		WAIT_SLEEP_INTERVAL = wait_sleep
		WAIT_TIMEOUT = wait_timeout
	}()

	WAIT_TIMEOUT = 1 * time.Second
	WAIT_SLEEP_INTERVAL = 1 * time.Nanosecond

	err := WaitForSuccess(
		"boom",
		func() error {
			return errors.New("something went wrong")
		})

	if err == nil {
		t.Error("expected error")
	}
}
