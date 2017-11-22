package libvirt

import (
	"testing"
)

func TestAcquireLock(t *testing.T) {
	ps := NewLVirtPoolSync()

	if !ps.AcquireLock("test") {
		t.Errorf("lock not found")
	}
}

func TestReleaseLock(t *testing.T) {
	ps := NewLVirtPoolSync()

	if !ps.AcquireLock("test") {
		t.Errorf("lock not found")
	}

	ps.ReleaseLock("test")
}

func TestReleaseNotExistingLock(t *testing.T) {
	ps := NewLVirtPoolSync()

	ps.ReleaseLock("test")
	// moreover there should be no runtime error because
	// we are not trying to unlock a not-locked mutex
}
