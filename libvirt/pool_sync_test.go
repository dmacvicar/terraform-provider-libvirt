package libvirt

import (
	"testing"
)

func TestAcquireLock(t *testing.T) {
	ps := NewLVirtPoolSync()

	ps.AcquireLock("test")

	_, found := ps.PoolLocks["test"]

	if !found {
		t.Errorf("lock not found")
	}
}

func TestReleaseLock(t *testing.T) {
	ps := NewLVirtPoolSync()

	ps.AcquireLock("test")

	_, found := ps.PoolLocks["test"]
	if !found {
		t.Errorf("lock not found")
	}

	ps.ReleaseLock("test")
	_, found = ps.PoolLocks["test"]
	if !found {
		t.Errorf("lock not found")
	}
}

func TestReleaseNotExistingLock(t *testing.T) {
	ps := NewLVirtPoolSync()

	ps.ReleaseLock("test")
	_, found := ps.PoolLocks["test"]
	if found {
		t.Errorf("lock found")
	}
	// moreover there should be no runtime error because
	// we are not trying to unlock a not-locked mutex
}
