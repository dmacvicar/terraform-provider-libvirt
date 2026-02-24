package provider

import (
	"testing"

	golibvirt "github.com/digitalocean/go-libvirt"
)

func TestDomainUndefineFlagsForUpdate(t *testing.T) {
	testCases := []struct {
		name           string
		libvirtVersion uint64
		expected       golibvirt.DomainUndefineFlagsValues
	}{
		{
			name:           "before keep nvram support",
			libvirtVersion: 2_002_999,
			expected:       0,
		},
		{
			name:           "keep nvram only",
			libvirtVersion: 2_003_000,
			expected:       golibvirt.DomainUndefineKeepNvram,
		},
		{
			name:           "before keep tpm support",
			libvirtVersion: 8_008_999,
			expected:       golibvirt.DomainUndefineKeepNvram,
		},
		{
			name:           "keep nvram and keep tpm",
			libvirtVersion: 8_009_000,
			expected:       golibvirt.DomainUndefineKeepNvram | golibvirt.DomainUndefineKeepTpm,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := domainUndefineFlagsForUpdate(tc.libvirtVersion)
			if actual != tc.expected {
				t.Fatalf("expected update flags %v, got %v", tc.expected, actual)
			}
		})
	}
}

func TestDomainUndefineFlagsForDelete(t *testing.T) {
	testCases := []struct {
		name           string
		libvirtVersion uint64
		expected       golibvirt.DomainUndefineFlagsValues
	}{
		{
			name:           "before nvram support",
			libvirtVersion: 1_002_008,
			expected:       0,
		},
		{
			name:           "nvram only",
			libvirtVersion: 1_002_009,
			expected:       golibvirt.DomainUndefineNvram,
		},
		{
			name:           "before tpm support",
			libvirtVersion: 8_008_999,
			expected:       golibvirt.DomainUndefineNvram,
		},
		{
			name:           "nvram and tpm",
			libvirtVersion: 8_009_000,
			expected:       golibvirt.DomainUndefineNvram | golibvirt.DomainUndefineTpm,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := domainUndefineFlagsForDelete(tc.libvirtVersion)
			if actual != tc.expected {
				t.Fatalf("expected delete flags %v, got %v", tc.expected, actual)
			}
		})
	}
}
