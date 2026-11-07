package bytesize

import (
	"errors"
	"fmt"
	"math"
)

const (
	B   = 1
	KB  = 1000
	KiB = 1024
	MB  = KB * 1000
	MiB = KiB * 1024
	GB  = MB * 1000
	GiB = MiB * 1024
	TB  = GB * 1000
	TiB = GiB * 1024
	PB  = TB * 1000
	PiB = TiB * 1024
	EB  = PB * 1000
	EiB = PiB * 1024
)

var unitConversion = map[string]int64{
	"":      B,
	"B":     B,
	"bytes": B,
	"KB":    KB,
	"K":     KiB,
	"KiB":   KiB,
	"MB":    MB,
	"M":     MiB,
	"MiB":   MiB,
	"GB":    GB,
	"G":     GiB,
	"GiB":   GiB,
	"TB":    TB,
	"T":     TiB,
	"TiB":   TiB,
	"PB":    PB,
	"P":     PiB,
	"PiB":   PiB,
	"EB":    EB,
	"E":     EiB,
	"EiB":   EiB,
}

var (
	ErrUnknownUnit = errors.New("unknown unit")
	ErrOverflow    = errors.New("int64 overflow")
)

// Convert value and unit to bytes.
// Return error for unknown units.
func Convert(value int64, unit string) (int64, error) {
	factor, ok := unitConversion[unit]
	if !ok {
		return 0, fmt.Errorf("%w: %q", ErrUnknownUnit, unit)
	}

	if value > math.MaxInt64/factor || value < math.MinInt64/factor {
		return 0, fmt.Errorf("%w: %d %s", ErrOverflow, value, unit)
	}

	return value * factor, nil
}
