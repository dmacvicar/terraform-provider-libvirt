package libvirt

import (
	"fmt"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
	"net"
	"strings"
)

func flattenNetworkDNSForwarders(list []libvirtxml.NetworkDNSForwarder) []interface{} {
	forwarders := make([]interface{}, 0, len(list))
	for _, spec := range list {
		forwarders = append(forwarders,
			map[string]interface{}{"address": spec.Addr, "domain": spec.Domain})
	}
	return forwarders
}

func flattenNetworkDNS(spec *libvirtxml.NetworkDNS, specDomain *libvirtxml.NetworkDomain) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mDNS := map[string]interface{}{}
	mDNS["forwarders"] = flattenNetworkDNSForwarders(spec.Forwarders)

	if specDomain != nil {
		mDNS["local_only"] = (strings.ToLower(specDomain.LocalOnly) == "yes")
	}

	return flattenAsArray(mDNS)
}

// old way of specifying addresses
func flattenNetworkAddresses(ips []libvirtxml.NetworkIP) ([]string, error) {
	if ips == nil {
		return []string{}, nil
	}

	addresses := []string{}
	for _, address := range ips {
		// we get the host interface IP (ie, 10.10.8.1) but we want the network CIDR (ie, 10.10.8.0/24)
		// so we need some transformations...
		addr := net.ParseIP(address.Address)
		if addr == nil {
			return nil, fmt.Errorf("Error parsing IP '%s'", address.Address)
		}
		bits := net.IPv6len * 8
		if addr.To4() != nil {
			bits = net.IPv4len * 8
		}

		mask := net.CIDRMask(int(address.Prefix), bits)
		network := addr.Mask(mask)
		addresses = append(addresses, fmt.Sprintf("%s/%d", network, address.Prefix))
	}
	return addresses, nil
}

func flattenNetworkDHCP(ips []libvirtxml.NetworkIP) []map[string]bool {
	if ips == nil {
		return []map[string]bool{{"enabled": false}}
	}

	for _, address := range ips {
		if address.DHCP != nil {
			return []map[string]bool{{"enabled": true}}

		}
	}
	return []map[string]bool{{"enabled": false}}
}

func flattenNetworkRoutes(list []libvirtxml.NetworkRoute) []interface{} {
	routes := make([]interface{}, 0, len(list))
	for _, spec := range list {
		routes = append(routes,
			map[string]interface{}{"gateway": spec.Gateway,
				"cidr": fmt.Sprintf("%s/%d", spec.Address, spec.Prefix)})
	}
	return routes
}
