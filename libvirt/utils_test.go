package libvirt

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform/terraform"
)

// getResourceFromTerraformState is helper function for getting a resource by name
// from terraform states produced during testacc
func getResourceFromTerraformState(resourceName string, state *terraform.State) (*terraform.ResourceState, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("Not found: %s", resourceName)
	}

	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("No libvirt resource key ID is set")
	}
	return rs, nil

}

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
	_, net, err := net.ParseCIDR("192.168.18.1/24")
	if err != nil {
		t.Errorf("When parsing network: %s", err)
	}

	start, end := networkRange(net)
	if start.String() != "192.168.18.0" {
		t.Errorf("unexpected range start for '%s': %s", net, start)
	}
	if end.String() != "192.168.18.255" {
		t.Errorf("unexpected range start for '%s': %s", net, start)
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
