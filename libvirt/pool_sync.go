package libvirt

import (
	"sync"
)

// LibVirtPoolSync makes possible to synchronize operations
// against libvirt pools.
// Doing pool.Refresh() operations while uploading or removing
// a volume into the pool causes errors inside of libvirtd
type LibVirtPoolSync struct {
	PoolLocks     map[string]*sync.Mutex
	internalMutex sync.Mutex
}

// NewLibVirtPoolSync Allocate a new instance of LibVirtPoolSync
func NewLibVirtPoolSync() LibVirtPoolSync {
	pool := LibVirtPoolSync{}
	pool.PoolLocks = make(map[string]*sync.Mutex)

	return pool
}

// AcquireLock Acquire a lock for the specified pool
func (ps LibVirtPoolSync) AcquireLock(pool string) {
	ps.internalMutex.Lock()
	defer ps.internalMutex.Unlock()

	lock, exists := ps.PoolLocks[pool]
	if !exists {
		lock = new(sync.Mutex)
		ps.PoolLocks[pool] = lock
	}

	lock.Lock()
}

// ReleaseLock Release the look for the specified pool
func (ps LibVirtPoolSync) ReleaseLock(pool string) {
	ps.internalMutex.Lock()
	defer ps.internalMutex.Unlock()

	lock, exists := ps.PoolLocks[pool]
	if !exists {
		return
	}

	lock.Unlock()
}
