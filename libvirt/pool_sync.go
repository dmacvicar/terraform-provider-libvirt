package libvirt

import "sync"

type LVirtPool struct {
	sync.Mutex
	locked bool
}

// LVirtPoolSync makes possible to synchronize operations against libvirt
// pools. Doing pool.Refresh() operations while uploading or removing a volume
// into the pool causes errors inside of libvirtd.
type LVirtPoolSync struct {
	sync.Mutex
	pools map[string]*LVirtPool
}

// NewLVirtPoolSync allocates a new instance of LibVirtPoolSync.
func NewLVirtPoolSync() *LVirtPoolSync {
	return &LVirtPoolSync{
		pools: make(map[string]*LVirtPool),
	}
}

// AcquireLock acquires a lock for the specified pool. If the mutex is already
// locked, the method returns false, and does not block. It returns true, if
// the lock could be acquired.
func (ps *LVirtPoolSync) AcquireLock(pool string) bool {
	ps.Lock()
	defer ps.Unlock()

	if ps.pools[pool] == nil {
		ps.pools[pool] = &LVirtPool{}
	} else {
		if ps.pools[pool].locked {
			return false
		}
	}

	ps.pools[pool].Lock()
	ps.pools[pool].locked = true

	return true
}

// ReleaseLock releases the look for the specified pool.
func (ps *LVirtPoolSync) ReleaseLock(pool string) {
	ps.Lock()
	defer ps.Unlock()

	if ps.pools[pool] == nil {
		return
	}

	ps.pools[pool].Unlock()
	ps.pools[pool].locked = false
}
