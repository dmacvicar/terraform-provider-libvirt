package stringutil

import "testing"

func TestSnakeCaseHypervisorPrefixes(t *testing.T) {
	testCases := map[string]string{
		"QEMUCommandline":      "qemu_commandline",
		"LXCNamespace":         "lxc_namespace",
		"BHyveCommandline":     "bhyve_commandline",
		"VMWareDataCenterPath": "vmware_data_center_path",
	}

	for input, want := range testCases {
		if got := SnakeCase(input); got != want {
			t.Fatalf("SnakeCase(%q) = %q, want %q", input, got, want)
		}
	}
}
