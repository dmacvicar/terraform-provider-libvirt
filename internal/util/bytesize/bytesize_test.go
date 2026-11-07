package bytesize

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name    string
		value   int64
		unit    string
		want    int64
		wantErr error
	}{
		// Every unit: 1 -> its factor
		{"empty unit is bytes", 1, "", 1, nil},
		{"B", 1, "B", 1, nil},
		{"bytes", 1, "bytes", 1, nil},
		{"KB decimal", 1, "KB", 1_000, nil},
		{"K binary", 1, "K", 1 << 10, nil},
		{"KiB", 1, "KiB", 1 << 10, nil},
		{"MB decimal", 1, "MB", 1_000_000, nil},
		{"M binary", 1, "M", 1 << 20, nil},
		{"MiB", 1, "MiB", 1 << 20, nil},
		{"GB decimal", 1, "GB", 1_000_000_000, nil},
		{"G binary", 1, "G", 1 << 30, nil},
		{"GiB", 1, "GiB", 1 << 30, nil},
		{"TB decimal", 1, "TB", 1_000_000_000_000, nil},
		{"T binary", 1, "T", 1 << 40, nil},
		{"TiB", 1, "TiB", 1 << 40, nil},
		{"PB decimal", 1, "PB", 1_000_000_000_000_000, nil},
		{"P binary", 1, "P", 1 << 50, nil},
		{"PiB", 1, "PiB", 1 << 50, nil},
		{"EB decimal", 1, "EB", 1_000_000_000_000_000_000, nil},
		{"E binary", 1, "E", 1 << 60, nil},
		{"EiB", 1, "EiB", 1 << 60, nil},

		// Typical values
		{"multiple", 512, "MiB", 512 * MiB, nil},
		{"zero", 0, "EiB", 0, nil},
		{"zero bytes", 0, "", 0, nil},
		{"negative", -5, "KiB", -5 * KiB, nil},

		// Overflow boundaries
		{"max bytes", math.MaxInt64, "B", math.MaxInt64, nil},
		{"min bytes", math.MinInt64, "B", math.MinInt64, nil},
		{"7 EiB fits", 7, "EiB", 7 * EiB, nil},
		{"8 EiB overflows", 8, "EiB", 0, ErrOverflow},
		{"-8 EiB is exactly MinInt64", -8, "EiB", math.MinInt64, nil},
		{"-9 EiB overflows", -9, "EiB", 0, ErrOverflow},
		{"9 EB fits", 9, "EB", 9 * EB, nil},
		{"10 EB overflows", 10, "EB", 0, ErrOverflow},
		{"max in KB overflows", math.MaxInt64, "KB", 0, ErrOverflow},
		{"min in KB overflows", math.MinInt64, "KB", 0, ErrOverflow},

		// Unknown units
		{"unknown unit", 1, "ZB", 0, ErrUnknownUnit},
		{"lowercase", 1, "kb", 0, ErrUnknownUnit},
		{"lowercase binary", 1, "mib", 0, ErrUnknownUnit},
		{"bits-style casing", 1, "Mb", 0, ErrUnknownUnit},
		{"leading space", 1, " MiB", 0, ErrUnknownUnit},
		{"trailing space", 1, "MiB ", 0, ErrUnknownUnit},
		{"garbage", 1, "MBB", 0, ErrUnknownUnit},
		{"unknown unit with zero value", 0, "nope", 0, ErrUnknownUnit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert(tt.value, tt.unit)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Convert(%d, %q) error = %v, want %v", tt.value, tt.unit, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Convert(%d, %q) = %d, want %d", tt.value, tt.unit, got, tt.want)
			}
		})
	}
}

func TestConvertErrorsAreDistinct(t *testing.T) {
	_, err := Convert(1, "ZB")
	if errors.Is(err, ErrOverflow) {
		t.Error("unknown unit error must not match ErrOverflow")
	}

	_, err = Convert(math.MaxInt64, "KB")
	if errors.Is(err, ErrUnknownUnit) {
		t.Error("overflow error must not match ErrUnknownUnit")
	}
}

func TestConvertErrorMessages(t *testing.T) {
	_, err := Convert(1, "MBB")
	if err == nil || !strings.Contains(err.Error(), `"MBB"`) {
		t.Errorf("unknown unit error should quote the unit, got %v", err)
	}

	_, err = Convert(math.MaxInt64, "KB")
	if err == nil || !strings.Contains(err.Error(), "KB") {
		t.Errorf("overflow error should mention the unit, got %v", err)
	}
}

// For every unit, the largest/smallest value that fits converts exactly,
// and one step beyond returns ErrOverflow.
func TestConvertBoundaries(t *testing.T) {
	for unit, factor := range unitConversion {
		name := unit
		if name == "" {
			name = "empty"
		}

		t.Run(name, func(t *testing.T) {
			maxVal := math.MaxInt64 / factor
			minVal := math.MinInt64 / factor

			got, err := Convert(maxVal, unit)
			if err != nil || got != maxVal*factor {
				t.Errorf("max: Convert(%d, %q) = %d, %v; want %d, nil", maxVal, unit, got, err, maxVal*factor)
			}

			got, err = Convert(minVal, unit)
			if err != nil || got != minVal*factor {
				t.Errorf("min: Convert(%d, %q) = %d, %v; want %d, nil", minVal, unit, got, err, minVal*factor)
			}

			// For factor 1, maxVal+1 would wrap inside the test itself.
			if factor > 1 {
				if _, err := Convert(maxVal+1, unit); !errors.Is(err, ErrOverflow) {
					t.Errorf("Convert(%d, %q) error = %v, want ErrOverflow", maxVal+1, unit, err)
				}
				if _, err := Convert(minVal-1, unit); !errors.Is(err, ErrOverflow) {
					t.Errorf("Convert(%d, %q) error = %v, want ErrOverflow", minVal-1, unit, err)
				}
			}
		})
	}
}
