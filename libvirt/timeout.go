package libvirt

import (
	"time"
)

const (
	resourceStateTimeout    = 1 * time.Minute
	resourceStateDelay      = 200 * time.Millisecond
	resourceStateMinTimeout = 100 * time.Millisecond
)
