package libvirt

import (
	"testing"

	libvirtxml "github.com/libvirt/libvirt-go-xml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlattenNetworkV2DNS(t *testing.T) {

	cases := []struct {
		Input libvirtxml.NetworkDNS
		// this is unusual, as we change the DNS data depending on a
		// domain configuration
		InputDomain    libvirtxml.NetworkDomain
		ExpectedOutput []interface{}
	}{
		{
			libvirtxml.NetworkDNS{
				Forwarders: []libvirtxml.NetworkDNSForwarder{{Addr: "1.2.3.4", Domain: "foo.com"}},
			},
			libvirtxml.NetworkDomain{LocalOnly: "yes"},
			[]interface{}{
				map[string]interface{}{
					"local_only": true,
					"forwarders": []interface{}{
						map[string]interface{}{"address": "1.2.3.4", "domain": "foo.com"}}}},
		},
	}

	for _, tc := range cases {
		output := flattenNetworkDNS(&tc.Input, &tc.InputDomain)
		assert.Equal(t, output, tc.ExpectedOutput)
	}
}

func TestFlattenNetworkV2Addresses(t *testing.T) {

	cases := []struct {
		Input          []libvirtxml.NetworkIP
		ExpectedOutput []string
	}{
		{
			[]libvirtxml.NetworkIP{
				{Address: "10.10.8.1", Prefix: 24},
			},
			[]string{"10.10.8.0/24"},
		},
	}

	for _, tc := range cases {
		output, err := flattenNetworkAddresses(tc.Input)
		require.NoError(t, err)
		assert.Equal(t, tc.ExpectedOutput, output)
	}
}

func TestFlattenNetworkV2DHCP(t *testing.T) {

	cases := []struct {
		Input          []libvirtxml.NetworkIP
		ExpectedOutput []map[string]bool
	}{
		{
			[]libvirtxml.NetworkIP{
				{
					Address: "10.10.8.1",
					Prefix:  24,
					DHCP:    &libvirtxml.NetworkDHCP{},
				},
			},
			[]map[string]bool{{"enabled": true}},
		},
	}

	for _, tc := range cases {
		output := flattenNetworkDHCP(tc.Input)
		assert.Equal(t, tc.ExpectedOutput, output)
	}
}

func TestFlattenNetworkV2Routes(t *testing.T) {
	cases := []struct {
		Input          []libvirtxml.NetworkRoute
		ExpectedOutput []interface{}
	}{
		{
			[]libvirtxml.NetworkRoute{
				{
					Gateway: "192.168.178.1",
					Address: "192.168.178.0",
					Prefix:  24,
				},
			},
			[]interface{}{map[string]interface{}{
				"gateway": "192.168.178.1",
				"cidr":    "192.168.178.0/24"}},
		},
	}

	for _, tc := range cases {
		output := flattenNetworkV2Routes(tc.Input)
		assert.Equal(t, tc.ExpectedOutput, output)
	}
}
