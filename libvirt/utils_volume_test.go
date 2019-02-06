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
