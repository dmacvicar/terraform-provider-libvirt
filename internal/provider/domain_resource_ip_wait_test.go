package provider

import (
	"testing"

	golibvirt "github.com/digitalocean/go-libvirt"
)

func TestMatchesNetwork(t *testing.T) {
	tests := []struct {
		name    string
		ifaces  []golibvirt.DomainInterface
		mac     string
		network string
		want    bool
	}{
		{
			name: "ipv4-only rejects link-local ipv6 (user repro)",
			ifaces: []golibvirt.DomainInterface{
				{Name: "vnet0", Addrs: []golibvirt.DomainIPAddr{{Addr: "fe80::1"}}},
			},
			network: "0.0.0.0/0",
			want:    false,
		},
		{
			name: "ipv4-only accepts private ipv4",
			ifaces: []golibvirt.DomainInterface{
				{Name: "vnet0", Addrs: []golibvirt.DomainIPAddr{{Addr: "192.168.122.10"}}},
			},
			network: "0.0.0.0/0",
			want:    true,
		},
		{
			name: "ipv6-only accepts global ipv6",
			ifaces: []golibvirt.DomainInterface{
				{Name: "vnet0", Addrs: []golibvirt.DomainIPAddr{{Addr: "2001:db8::1"}}},
			},
			network: "::/0",
			want:    true,
		},
		{
			name: "ipv6-only rejects ipv4",
			ifaces: []golibvirt.DomainInterface{
				{Name: "vnet0", Addrs: []golibvirt.DomainIPAddr{{Addr: "192.168.122.10"}}},
			},
			network: "::/0",
			want:    false,
		},
		{
			name: "subnet boundary in",
			ifaces: []golibvirt.DomainInterface{
				{Name: "vnet0", Addrs: []golibvirt.DomainIPAddr{{Addr: "10.10.0.5"}}},
			},
			network: "10.10.0.0/26",
			want:    true,
		},
		{
			name: "subnet boundary out",
			ifaces: []golibvirt.DomainInterface{
				{Name: "vnet0", Addrs: []golibvirt.DomainIPAddr{{Addr: "10.10.0.70"}}},
			},
			network: "10.10.0.0/26",
			want:    false,
		},
		{
			name: "no filter any address",
			ifaces: []golibvirt.DomainInterface{
				{Name: "vnet0", Addrs: []golibvirt.DomainIPAddr{{Addr: "fe80::1"}}},
			},
			network: "",
			want:    true,
		},
		{
			name: "mac filter misses",
			ifaces: []golibvirt.DomainInterface{
				{Name: "vnet0", Hwaddr: golibvirt.OptString{"52:54:00:aa:bb:cc"}, Addrs: []golibvirt.DomainIPAddr{{Addr: "192.168.122.10"}}},
			},
			mac:     "52:54:00:00:00:00",
			network: "0.0.0.0/0",
			want:    false,
		},
		{
			name: "mac filter hits",
			ifaces: []golibvirt.DomainInterface{
				{Name: "vnet0", Hwaddr: golibvirt.OptString{"52:54:00:aa:bb:cc"}, Addrs: []golibvirt.DomainIPAddr{{Addr: "192.168.122.10"}}},
			},
			mac:     "52:54:00:aa:bb:cc",
			network: "0.0.0.0/0",
			want:    true,
		},
		{
			name: "invalid network never matches",
			ifaces: []golibvirt.DomainInterface{
				{Name: "vnet0", Addrs: []golibvirt.DomainIPAddr{{Addr: "192.168.122.10"}}},
			},
			network: "10.10/26",
			want:    false,
		},
		{
			name:    "empty no filter",
			ifaces:  []golibvirt.DomainInterface{},
			network: "",
			want:    false,
		},
		{
			name:    "empty with network",
			ifaces:  []golibvirt.DomainInterface{},
			network: "0.0.0.0/0",
			want:    false,
		},
		{
			name: "loopback matches ipv4 any",
			ifaces: []golibvirt.DomainInterface{
				{Name: "lo", Addrs: []golibvirt.DomainIPAddr{{Addr: "127.0.0.1"}}},
			},
			network: "0.0.0.0/0",
			want:    true,
		},
		{
			name: "ipv4-mapped ipv6 not in ipv4",
			ifaces: []golibvirt.DomainInterface{
				{Name: "vnet0", Addrs: []golibvirt.DomainIPAddr{{Addr: "::ffff:192.168.1.1"}}},
			},
			network: "0.0.0.0/0",
			want:    false,
		},
		{
			name: "unparseable plus valid",
			ifaces: []golibvirt.DomainInterface{
				{Name: "vnet0", Addrs: []golibvirt.DomainIPAddr{{Addr: "garbage"}, {Addr: "192.168.122.10"}}},
			},
			network: "0.0.0.0/0",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesNetwork(tt.ifaces, tt.mac, tt.network); got != tt.want {
				t.Fatalf("matchesNetwork(...) = %v, want %v", got, tt.want)
			}
		})
	}
}
