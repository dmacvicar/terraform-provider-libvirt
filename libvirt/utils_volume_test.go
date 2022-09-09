package libvirt

import (
	"testing"
	"time"
)

func TestTimeFromEpoch(t *testing.T) {
	if ts := timeFromEpoch(""); ts.UnixNano() > 0 {
		t.Fatalf("expected timestamp '0.0', got %v.%v", ts.Unix(), ts.Nanosecond())
	}

	if ts := timeFromEpoch("abc"); ts.UnixNano() > 0 {
		t.Fatalf("expected timestamp '0.0', got %v.%v", ts.Unix(), ts.Nanosecond())
	}

	if ts := timeFromEpoch("123"); ts.UnixNano() != time.Unix(123, 0).UnixNano() {
		t.Fatalf("expected timestamp '123.0', got %v.%v", ts.Unix(), ts.Nanosecond())
	}

	if ts := timeFromEpoch("123.456"); ts.UnixNano() != time.Unix(123, 456).UnixNano() {
		t.Fatalf("expected timestamp '123.456', got %v.%v", ts.Unix(), ts.Nanosecond())
	}
}

func TestSizeFromString(t *testing.T) {
	if size := sizeFromString(""); size != 0 {
		t.Fatalf("expected size '0', got %v", size)
	}

	if size := sizeFromString("abc"); size != 0 {
		t.Fatalf("expected size '0', got %v", size)
	}

	if size := sizeFromString("1"); size != 1 {
		t.Fatalf("expected size '1', got %v", size)
	}

	if size := sizeFromString("1B"); size != 1 {
		t.Fatalf("expected size '1', got %v", size)
	}

	if size := sizeFromString("1KB"); size != 1000 {
		t.Fatalf("expected size '1000', got %v", size)
	}

	if size := sizeFromString("1k"); size != 1000 {
		t.Fatalf("expected size '1000', got %v", size)
	}

	if size := sizeFromString("1m"); size != 1000000 {
		t.Fatalf("expected size '1000000', got %v", size)
	}

	if size := sizeFromString("1g"); size != 1000000000 {
		t.Fatalf("expected size '1000000000', got %v", size)
	}

	if size := sizeFromString("1t"); size != 1000000000000 {
		t.Fatalf("expected size '1000000000000', got %v", size)
	}

	if size := sizeFromString("1000000000000"); size != 1000000000000 {
		t.Fatalf("expected size '1000000000000', got %v", size)
	}

	if size := sizeFromString("1KiB"); size != 1024 {
		t.Fatalf("expected size '1024', got %v", size)
	}

	if size := sizeFromString("1ki"); size != 1024 {
		t.Fatalf("expected size '1024', got %v", size)
	}

	if size := sizeFromString("1mi"); size != 1048576 {
		t.Fatalf("expected size '1048576', got %v", size)
	}

	if size := sizeFromString("1gi"); size != 1073741824 {
		t.Fatalf("expected size '1073741824', got %v", size)
	}

	if size := sizeFromString("1ti"); size != 1099511627776 {
		t.Fatalf("expected size '1099511627776', got %v", size)
	}

	if size := sizeFromString("1099511627776"); size != 1099511627776 {
		t.Fatalf("expected size '1099511627776', got %v", size)
	}
}
