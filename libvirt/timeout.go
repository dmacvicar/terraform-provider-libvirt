package libvirt

import (
	"time"
)

const (
	resourceStateTimeout    = 1 * time.Minute
	resourceStateDelay      = 5 * time.Second
	resourceStateMinTimeout = 3 * time.Second
)
