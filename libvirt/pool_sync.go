package libvirt

import "sync"

// LVirtPoolSync makes possible to synchronize operations
// against libvirt pools.
// Doing pool.Refresh() operations while uploading or removing
// a volume into the pool causes errors inside of libvirtd
type LVirtPoolSync struct {
	PoolLocks     map[string]*sync.Mutex
	poolLocked    map[string]bool
	internalMutex sync.Mutex
}

// NewLVirtPoolSync allocates a new instance of LVirtPoolSync
func NewLVirtPoolSync() *LVirtPoolSync {
	pool := LVirtPoolSync{}
	pool.PoolLocks = make(map[string]*sync.Mutex)
	pool.poolLocked = make(map[string]bool)

	return &pool
}

// AcquireLock acquires a lock for the specified pool. If the mutex is already
// locked, the method returns false, and does not block. It returns true, if
// the lock could be acquired.
func (ps *LVirtPoolSync) AcquireLock(pool string) bool {
	ps.internalMutex.Lock()
	defer ps.internalMutex.Unlock()

	if ps.PoolLocks[pool] == nil {
		ps.PoolLocks[pool] = new(sync.Mutex)
	} else {
		if ps.poolLocked[pool] {
			return false
		}
	}

	ps.PoolLocks[pool].Lock()
	ps.poolLocked[pool] = true

	return true
}

// ReleaseLock releases the look for the specified pool
func (ps *LVirtPoolSync) ReleaseLock(pool string) {
	ps.internalMutex.Lock()
	defer ps.internalMutex.Unlock()

	if ps.PoolLocks[pool] == nil {
		return
	}

	ps.PoolLocks[pool].Unlock()
	ps.poolLocked[pool] = false
}
