package libvirt

import (
	"bytes"
	"errors"
	"log"
	"net"
	"os"
	"testing"
	"time"
)

func TestDiskLetterForIndex(t *testing.T) {

	diskNumbers := []int{0, 1, 2, 3, 4, 16, 24, 25, 26, 30, 300}
	names := []string{"a", "b", "c", "d", "e", "q", "y", "z", "aa", "ae", "ko"}

	for i, diskNumber := range diskNumbers {
		ret := diskLetterForIndex(diskNumber)
		if ret != names[i] {
			t.Errorf("Expected %s, got %s for disk %d", names[i], ret, diskNumber)
		}
	}
}

func TestFormatBoolYesNo(t *testing.T) {

	if formatBoolYesNo(true) != "yes" {
		t.Errorf("Expected 'yes'")
	}

	if formatBoolYesNo(false) != "no" {
		t.Errorf("Expected 'no'")
	}
}

func TestIPsRange(t *testing.T) {
	tt := []struct {
		name         string
		inputCIDR    string
		expectdStart string
		expectedEnd  string
	}{
		{
			name:         "IPv4 range",
			inputCIDR:    "192.168.18.1/24",
			expectdStart: "192.168.18.0",
			expectedEnd:  "192.168.18.255",
		},
		{
			name:         "IPv6 range",
			inputCIDR:    "fdff:beef:beef::1/64",
			expectdStart: "fdff:beef:beef::",
			expectedEnd:  "fdff:beef:beef::ffff",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, net, err := net.ParseCIDR(tc.inputCIDR)
			if err != nil {
				t.Errorf("When parsing network: %s", err)
			}

			start, end := networkRange(net)
			if start.String() != tc.expectdStart {
				t.Errorf("unexpected range start for '%s': %s, expected: %s", net, start, tc.expectdStart)
			}
			if end.String() != tc.expectedEnd {
				t.Errorf("unexpected range end for '%s': %s, expected: %s", net, end, tc.expectedEnd)
			}
		})
	}
}

func TestWaitForSuccessEverythingFine(t *testing.T) {
	waitSleep := WaitSleepInterval
	waitTimeout := WaitTimeout
	defer func() {
		WaitSleepInterval = waitSleep
		WaitTimeout = waitTimeout
	}()

	WaitTimeout = 1 * time.Second
	WaitSleepInterval = 1 * time.Nanosecond

	err := waitForSuccess(
		"boom",
		func() error {
			return nil
		})

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

func TestWaitForSuccessBrokenFunction(t *testing.T) {
	waitSleep := WaitSleepInterval
	waitTimeout := WaitTimeout
	var b bytes.Buffer
	log.SetOutput(&b)
	defer func() {
		WaitSleepInterval = waitSleep
		WaitTimeout = waitTimeout
		log.SetOutput(os.Stderr)
	}()

	WaitTimeout = 1 * time.Second
	WaitSleepInterval = 1 * time.Nanosecond

	err := waitForSuccess(
		"boom",
		func() error {
			return errors.New("something went wrong")
		})

	if err == nil {
		t.Error("expected error")
	}
}
