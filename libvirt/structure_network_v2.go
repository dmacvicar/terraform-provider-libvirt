package libvirt

import (
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

func flattenNetworkV2DNSForwarders(list []libvirtxml.NetworkDNSForwarder) []interface{} {
	forwarders := make([]interface{}, 0, len(list))
	for _, spec := range list {
		forwarders = append(forwarders,
			map[string]interface{}{"address": spec.Addr, "domain": spec.Domain})
	}
	return forwarders
}

func flattenNetworkV2DNSHosts(list []libvirtxml.NetworkDNSHost) []interface{} {
	result := make([]interface{}, 0, len(list))
	for _, host := range list {
		mHost := make(map[string]interface{})
		if host.IP != "" {
			mHost["ip"] = host.IP
		}
		// TODO we support a single hostname for now
		if len(host.Hostnames) > 0  {
			mHost["hostname"] = host.Hostnames[0]
		}
		result = append(result, mHost)
	}
	return result
}

func flattenNetworkV2IPs(list []libvirtxml.NetworkIP) []interface{} {
	result := make([]interface{}, 0, len(list))
	for _, ip := range list {
		mIP := make(map[string]interface{})
		if ip.Address != "" {
			mIP["address"] = ip.Address
		}
		if ip.Family != "" {
			mIP["family"] = ip.Family
		}

		if ip.Prefix != 0 {
			mIP["prefix"] = ip.Prefix
		}
		if ip.Netmask != "" {
			mIP["netmask"] = ip.Netmask
		}

		if ip.DHCP != nil {
			mIP["dhcp"] = flattenNetworkV2DHCP(ip.DHCP)
		}
		result = append(result, mIP)
	}
	return result
}

func flattenNetworkV2Domain(domain *libvirtxml.NetworkDomain) interface{} {
	mDomain := make(map[string]interface{})
	if domain.LocalOnly == "yes" {
		mDomain["local_only"] = true
	} else if domain.LocalOnly == "no" {
		mDomain["local_only"] = false
	} else {
		// no-op, we don't even set the value
	}

	if domain.Name != "" {
		mDomain["name"] = domain.Name
	}
	return mDomain
}

func flattenNetworkV2Bridge(bridge *libvirtxml.NetworkBridge) interface{} {
	mBridge := make(map[string]interface{})
	if bridge.Name != "" {
		mBridge["name"] = bridge.Name
	}
	return mBridge
}

func flattenNetworkV2DHCP(dhcp *libvirtxml.NetworkDHCP) interface{} {
	mDHCP := make(map[string]interface{})
	if dhcp.Ranges != nil {
		mDHCP["range"] = flattenNetworkV2DHCPRanges(dhcp.Ranges)
	}
	if dhcp.Ranges != nil {
		mDHCP["host"] = flattenNetworkV2DHCPHosts(dhcp.Hosts)
	}
	return mDHCP
}

func flattenNetworkV2DHCPHosts(list []libvirtxml.NetworkDHCPHost) []interface{} {
	result := make([]interface{}, 0, len(list))
	for _, host := range list {
		mHost := make(map[string]interface{})
		if host.IP != "" {
			mHost["ip"] = host.IP
		}
		if host.MAC != "" {
			mHost["mac"] = host.MAC
		}
		if host.ID != "" {
			mHost["id"] = host.ID
		}
		if host.Name != "" {
			mHost["name"] = host.Name
		}
		result = append(result, mHost)
	}
	return result
}

func flattenNetworkV2DHCPRanges(list []libvirtxml.NetworkDHCPRange) []interface{} {
	result := make([]interface{}, 0, len(list))
	for _, r := range list {
		mRange := make(map[string]interface{})
		if r.Start != "" {
			mRange["start"] = r.Start
		}
		if r.End != "" {
			mRange["end"] = r.End
		}
		result = append(result, mRange)
	}
	return result
}

func flattenNetworkV2Routes(list []libvirtxml.NetworkRoute) []interface{} {
	result := make([]interface{}, 0, len(list))
	for _, route := range list {
		mRoute := make(map[string]interface{})
		if route.Address != "" {
			mRoute["address"] = route.Address
		}
		if route.Prefix != 0 {
			mRoute["prefix"] = route.Prefix
		}
		if route.Netmask != "" {
			mRoute["netmask"] = route.Netmask
		}
		result = append(result, mRoute)
	}
	return result
}

func flattenNetworkV2DNS(spec *libvirtxml.NetworkDNS) []interface{} {
	mDNS := make(map[string]interface{})

	if len(spec.Forwarders) > 0 {
		mDNS["forwarder"] = flattenNetworkV2DNSForwarders(spec.Forwarders)
	}

	if spec.Enable == "yes" {
		mDNS["enable"] = true
	} else if spec.Enable == "no" {
		mDNS["enable"] = false
	} else {
		// no-op. Use the libvirt default, without setting it in the config
	}

	if len(spec.Host) > 0 {
		mDNS["host"] = flattenNetworkV2DNSHosts(spec.Host)
	}

	// TODO txt srv
	return flattenAsArray(mDNS)
}

func expandNetworkV2DHCPRanges(configured []interface{}) []libvirtxml.NetworkDHCPRange {
	ranges := make([]libvirtxml.NetworkDHCPRange, 0, len(configured))
	for _, raw := range configured {
		rang := libvirtxml.NetworkDHCPRange{}

		data := raw.(map[string]interface{})
		if v, ok := data["start"]; ok {
			rang.Start = v.(string)
		}
		if v, ok := data["end"]; ok {
			rang.End = v.(string)
		}

		ranges = append(ranges, rang)
	}
	return ranges
}

func expandNetworkV2DHCPHosts(configured []interface{}) []libvirtxml.NetworkDHCPHost {
	hosts := make([]libvirtxml.NetworkDHCPHost, 0, len(configured))
	for _, raw := range configured {
		host := libvirtxml.NetworkDHCPHost{}

		data := raw.(map[string]interface{})
		if v, ok := data["mac"]; ok {
			host.MAC = v.(string)
		}
		if v, ok := data["ip"]; ok {
			host.IP = v.(string)
		}
		if v, ok := data["name"]; ok {
			host.Name = v.(string)
		}

		hosts = append(hosts, host)
	}
	return hosts
}

func expandNetworkV2Routes(configured []interface{}) []libvirtxml.NetworkRoute {
	routes := make([]libvirtxml.NetworkRoute, 0, len(configured))
	for _, raw := range configured {
		route := libvirtxml.NetworkRoute{}

		data := raw.(map[string]interface{})
		if v, ok := data["adress"]; ok {
			route.Address = v.(string)
		}
		if v, ok := data["prefix"]; ok {
			route.Prefix = v.(uint)
		}
		if v, ok := data["gateway"]; ok {
			route.Gateway = v.(string)
		}
		if v, ok := data["metric"]; ok {
			route.Metric = v.(string)
		}

		routes = append(routes, route)
	}
	return routes
}

func expandNetworkV2Bridge(configured []interface{}) *libvirtxml.NetworkBridge {
	if len(configured) == 0 {
		return nil
	}
	data := configured[0].(map[string]interface{})

	bridge := libvirtxml.NetworkBridge{}
	if v, ok := data["name"]; ok {
		bridge.Name = v.(string)
	}
	return &bridge
}

func expandNetworkV2Domain(configured []interface{}) *libvirtxml.NetworkDomain {
	if len(configured) == 0 {
		return nil
	}
	data := configured[0].(map[string]interface{})

	domain := libvirtxml.NetworkDomain{}
	if v, ok := data["name"]; ok {
		domain.Name = v.(string)
	}
	if v, ok := data["local_only"]; ok {
		if v.(bool) {
			domain.LocalOnly = "yes"
		} else {
			domain.LocalOnly = "no"
		}
	}
	return &domain
}

func expandNetworkV2DNS(configured []interface{}) *libvirtxml.NetworkDNS {
	if len(configured) == 0 {
		return nil
	}
	data := configured[0].(map[string]interface{})

	dns := libvirtxml.NetworkDNS{}
	if v, ok := data["enable"]; ok {
		if v.(bool) {
			dns.Enable = "yes"
		} else {
			dns.Enable = "no"
		}
	}

	if v, ok := data["forwarder"]; ok {
		dns.Forwarders = expandNetworkV2DNSForwarders(v.([]interface{}))

	}
	return &dns
}

func expandNetworkV2DNSForwarders(configured []interface{}) []libvirtxml.NetworkDNSForwarder {
	forwarders := make([]libvirtxml.NetworkDNSForwarder, 0, len(configured))

	for _, raw := range configured {
		data := raw.(map[string]interface{})
		forwarder := libvirtxml.NetworkDNSForwarder{}
		if v, ok := data["addr"]; ok {
			forwarder.Addr = v.(string)
		}
		if v, ok := data["domain"]; ok {
			forwarder.Domain = v.(string)
		}
		forwarders = append(forwarders, forwarder)
	}
	return forwarders
}

func expandNetworkV2Forward(configured []interface{}) *libvirtxml.NetworkForward {

	data := configured[0].(map[string]interface{})

	forward := libvirtxml.NetworkForward{}
	if v, ok := data["mode"]; ok {
		forward.Mode = v.(string)
	}
	return &forward
}

func expandNetworkV2DHCP(configured []interface{}) *libvirtxml.NetworkDHCP {
	if len(configured) == 0 {
		return nil
	}
	// while dhcp block is an array, there is only one
	data := configured[0].(map[string]interface{})

	dhcp := libvirtxml.NetworkDHCP{}
	if v, ok := data["range"]; ok {
		dhcp.Ranges = expandNetworkV2DHCPRanges(v.([]interface{}))
	}
	if v, ok := data["host"]; ok {
		dhcp.Hosts = expandNetworkV2DHCPHosts(v.([]interface{}))
	}

	return &dhcp
}

func expandNetworkV2IPs(configured []interface{}) []libvirtxml.NetworkIP {
	ips := make([]libvirtxml.NetworkIP, 0, len(configured))

	for _, raw := range configured {
		data := raw.(map[string]interface{})
		ip := libvirtxml.NetworkIP{}
		if v, ok := data["address"]; ok {
			ip.Address = v.(string)
		}
		// TODO autodetect and remove family
		if v, ok := data["family"]; ok {
			ip.Family = v.(string)
		}
		if v, ok := data["netmask"]; ok {
			ip.Netmask = v.(string)
		}
		if v, ok := data["prefix"]; ok {
			ip.Prefix = uint(v.(int))
		}
		if v, ok := data["dhcp"]; ok {
			va := v.([]interface{})
			ip.DHCP = expandNetworkV2DHCP(va)
		}

		ips = append(ips, ip)
	}
	return ips
}

func networkAddressStateFunc(v interface{}) string {
	return net.ParseIP(v.(string)).String()
}

func networkAddressDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	oldIP := net.ParseIP(old)
	newIP := net.ParseIP(new)
	if (oldIP.String() == newIP.String()) && oldIP != nil && newIP != nil {
		return true
	}
	return false
}
