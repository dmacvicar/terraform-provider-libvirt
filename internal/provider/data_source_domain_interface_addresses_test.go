package provider

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/digitalocean/go-libvirt"
	golibvirt "github.com/digitalocean/go-libvirt"
)

// runRetryTest executes getInterfacesWithRetry inside a synctest environment,
// tracks the number of times the mock function is called, and returns the results.
func runRetryTest(t *testing.T, fn getInterfacesFunc, timeout time.Duration) (int, []libvirt.DomainInterface, error) {
	t.Helper()
	var ifaces []libvirt.DomainInterface
	var err error
	var calls int

	// Wrap the provided fn to track calls
	countingFn := func() ([]libvirt.DomainInterface, error) {
		calls++
		return fn()
	}

	synctest.Test(t, func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			ifaces, err = getInterfacesWithRetry(context.Background(), countingFn, libvirt.Domain{ID: 1, UUID: libvirt.UUID{}, Name: "test"}, timeout)()
		}()

		wg.Wait()
		synctest.Wait()
	})

	return calls, ifaces, err
}

func Test_getInterfacesWithRetry(t *testing.T) {
	t.Run("responsive", func(t *testing.T) {
		fn := func() ([]libvirt.DomainInterface, error) {
			return []libvirt.DomainInterface{{Name: "test"}}, nil
		}

		calls, ifaces, err := runRetryTest(t, fn, time.Minute)

		if calls != 1 {
			t.Errorf("expected 1 call, got %d", calls)
		}
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if len(ifaces) != 1 || ifaces[0].Name != "test" {
			t.Errorf("expected 1 interface named 'test', got %v", ifaces)
		}
	})

	t.Run("unresponsive_timeout", func(t *testing.T) {
		fn := func() ([]libvirt.DomainInterface, error) {
			return nil, golibvirt.Error{
				Code:    uint32(libvirt.ErrAgentUnresponsive),
				Message: "agent is unresponsive",
			}
		}

		calls, _, err := runRetryTest(t, fn, time.Minute)

		if calls != 12 { // 1m timeout / 5s ticker = 12 calls
			t.Errorf("expected 12 calls, got %d", calls)
		}
		if err == nil || !strings.HasPrefix(err.Error(), "timed out waiting for interfaces") {
			t.Errorf("expected timeout error, got %v", err)
		}
		var libvirtErr golibvirt.Error
		if !errors.As(err, &libvirtErr) || libvirtErr.Code != uint32(libvirt.ErrAgentUnresponsive) {
			t.Errorf("expected wrapped ErrAgentUnresponsive, got %v", err)
		}
	})

	t.Run("generic_error_timeout", func(t *testing.T) {
		fn := func() ([]libvirt.DomainInterface, error) {
			return nil, errors.New("BOOM")
		}

		calls, _, err := runRetryTest(t, fn, time.Minute)

		if calls != 12 {
			t.Errorf("expected 12 calls, got %d", calls)
		}
		if err == nil || !strings.HasPrefix(err.Error(), "timed out waiting for interfaces") || !strings.HasSuffix(err.Error(), "BOOM") {
			t.Errorf("expected timeout wrapping 'BOOM', got %v", err)
		}
		var libvirtErr golibvirt.Error
		if errors.As(err, &libvirtErr) {
			t.Errorf("expected non-golibvirt.Error, got %v", err)
		}
	})

	t.Run("unresponsive_generic_error_response", func(t *testing.T) {
		localCalls := 0
		fn := func() ([]libvirt.DomainInterface, error) {
			localCalls++
			if localCalls == 1 {
				return nil, golibvirt.Error{Code: uint32(libvirt.ErrAgentUnresponsive)}
			}
			if localCalls == 2 {
				return nil, errors.New("BOOM")
			}
			return []libvirt.DomainInterface{{Name: "test"}}, nil
		}

		calls, ifaces, err := runRetryTest(t, fn, time.Minute)

		if calls != 3 {
			t.Errorf("expected 3 calls, got %d", calls)
		}
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if len(ifaces) != 1 || ifaces[0].Name != "test" {
			t.Errorf("expected 1 interface named 'test', got %v", ifaces)
		}
	})
}
