package libvirt

import (
	"fmt"
	"net"
	"reflect"
	"sort"
	"strconv"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"libvirt.org/go/libvirtxml"
)

// updateDNSHosts detects changes in the DNS hosts entries
// updating the network definition accordingly.
func updateDNSHosts(d *schema.ResourceData, meta interface{}, network libvirt.Network) error {
	virConn := meta.(*Client).libvirt

	hostsKey := dnsPrefix + ".hosts"
	if d.HasChange(hostsKey) {
		oldInterface, newInterface := d.GetChange(hostsKey)

		oldEntries, err := parseNetworkDNSHostsChange(oldInterface)
		if err != nil {
			return fmt.Errorf("parse old %s: %w", hostsKey, err)
		}

		newEntries, err := parseNetworkDNSHostsChange(newInterface)
		if err != nil {
			return fmt.Errorf("parse new %s: %w", hostsKey, err)
		}

		// process all the old DNS entries that must be removed
		for _, oldEntry := range oldEntries {
			found := false
			for _, newEntry := range newEntries {
				if reflect.DeepEqual(newEntry, oldEntry) {
					found = true
					break
				}
			}
			if found {
				continue
			}

			data, err := xmlMarshallIndented(libvirtxml.NetworkDNSHost{IP: oldEntry.IP})
			if err != nil {
				return fmt.Errorf("serialize update: %w", err)
			}

			err = virConn.NetworkUpdateCompat(network, libvirt.NetworkUpdateCommandDelete,
				libvirt.NetworkSectionDNSHost, -1, data, libvirt.NetworkUpdateAffectLive|libvirt.NetworkUpdateAffectConfig)
			if err != nil {
				return fmt.Errorf("delete %s: %w", oldEntry.IP, err)
			}
		}

		// process all the new DNS entries that must be added
		for _, newEntry := range newEntries {
			found := false
			for _, oldEntry := range oldEntries {
				if reflect.DeepEqual(oldEntry, newEntry) {
					found = true
					break
				}
			}
			if found {
				continue
			}

			data, err := xmlMarshallIndented(newEntry)
			if err != nil {
				return fmt.Errorf("serialize update: %w", err)
			}

			err = virConn.NetworkUpdateCompat(network, libvirt.NetworkUpdateCommandAddLast,
				libvirt.NetworkSectionDNSHost, -1, data, libvirt.NetworkUpdateAffectLive|libvirt.NetworkUpdateAffectConfig)
			if err != nil {
				return fmt.Errorf("add %v: %w", newEntry, err)
			}
		}
	}

	return nil
}

func parseNetworkDNSHostsChange(change interface{}) (entries []libvirtxml.NetworkDNSHost, err error) {
	slice, ok := change.(*schema.Set)
	if !ok {
		return entries, fmt.Errorf("not set %v", change)
	}

	mapEntries := map[string][]string{}
	for i, entryInterface := range slice.List() {
		entryMap, ok := entryInterface.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("entry %d is not a map", i)
		}

		ipInterface, ok := entryMap["ip"]
		if !ok {
			return nil, fmt.Errorf("entry %d.ip is missing", i)
		}

		ip, ok := ipInterface.(string)
		if !ok {
			return nil, fmt.Errorf("entry %d.ip is not a string", i)
		}

		hostnameInterface, ok := entryMap["hostname"]
		if !ok {
			return nil, fmt.Errorf("entry %d.hostname is missing", i)
		}

		hostname, ok := hostnameInterface.(string)
		if !ok {
			return nil, fmt.Errorf("entry %d.hostname is not a string", i)
		}

		_, ok = mapEntries[ip]
		if ok {
			mapEntries[ip] = append(mapEntries[ip], hostname)
		} else {
			mapEntries[ip] = []string{hostname}
		}
	}

	entries = make([]libvirtxml.NetworkDNSHost, 0, len(mapEntries))
	for ip, hostnames := range mapEntries {
		sort.Strings(hostnames)
		xmlHostnames := make([]libvirtxml.NetworkDNSHostHostname, 0, len(hostnames))
		for _, hostname := range hostnames {
			xmlHostnames = append(xmlHostnames, libvirtxml.NetworkDNSHostHostname{
				Hostname: hostname,
			})
		}
		entries = append(entries, libvirtxml.NetworkDNSHost{
			IP:        ip,
			Hostnames: xmlHostnames,
		})
	}

	return entries, nil
}

// getDNSHostsFromResource returns a list of libvirt's DNS hosts
// from the network definition.
func getDNSHostsFromResource(d *schema.ResourceData) ([]libvirtxml.NetworkDNSHost, error) {
	dnsHostsMap := map[string][]string{}
	if dnsHosts, ok := d.GetOk(dnsPrefix + ".hosts"); ok {
		for _, hostEntry := range dnsHosts.(*schema.Set).List() {
			hostEntryMap := hostEntry.(map[string]interface{})
			address := hostEntryMap["ip"].(string)
			if net.ParseIP(address) == nil {
				return nil, fmt.Errorf("could not parse address '%s'", address)
			}
			dnsHostsMap[address] = append(dnsHostsMap[address], hostEntryMap["hostname"].(string))
		}
	}

	var dnsHosts []libvirtxml.NetworkDNSHost

	for ip, hostnames := range dnsHostsMap {
		dnsHostnames := []libvirtxml.NetworkDNSHostHostname{}
		for _, hostname := range hostnames {
			dnsHostnames = append(dnsHostnames, libvirtxml.NetworkDNSHostHostname{Hostname: hostname})
		}
		dnsHosts = append(dnsHosts, libvirtxml.NetworkDNSHost{
			IP:        ip,
			Hostnames: dnsHostnames,
		})
	}

	return dnsHosts, nil
}

// getDNSForwardersFromResource returns the list of libvirt's DNS forwarders
// in the network definition.
func getDNSForwardersFromResource(d *schema.ResourceData) ([]libvirtxml.NetworkDNSForwarder, error) {
	var dnsForwarders []libvirtxml.NetworkDNSForwarder
	if dnsForwardCount, ok := d.GetOk(dnsPrefix + ".forwarders.#"); ok {
		for i := 0; i < dnsForwardCount.(int); i++ {
			forward := libvirtxml.NetworkDNSForwarder{}
			forwardPrefix := fmt.Sprintf(dnsPrefix+".forwarders.%d", i)
			if address, ok := d.GetOk(forwardPrefix + ".address"); ok {
				ip := net.ParseIP(address.(string))
				if ip == nil {
					return nil, fmt.Errorf("could not parse address '%s'", address)
				}
				forward.Addr = ip.String()
			}
			if domain, ok := d.GetOk(forwardPrefix + ".domain"); ok {
				forward.Domain = domain.(string)
			}
			dnsForwarders = append(dnsForwarders, forward)
		}
	}

	return dnsForwarders, nil
}

// getDNSEnableFromResource returns string to enable ("yes") or disable ("no") dns
// in the network definition.
func getDNSEnableFromResource(d *schema.ResourceData) string {
	if dnsEnabled, ok := d.GetOk(dnsPrefix + ".enabled"); ok {
		if dnsEnabled.(bool) {
			return "yes" // this "boolean" must be "yes"|"no"
		}
		return "no"
	}
	return "no"
}

// getDNSSRVFromResource returns a list of libvirt's DNS SRVs
// in the network definition.
func getDNSSRVFromResource(d *schema.ResourceData) ([]libvirtxml.NetworkDNSSRV, error) {
	var dnsSRVs []libvirtxml.NetworkDNSSRV

	if dnsSRVCount, ok := d.GetOk(dnsPrefix + ".srvs.#"); ok {
		for i := 0; i < dnsSRVCount.(int); i++ {
			srv := libvirtxml.NetworkDNSSRV{}
			srvPrefix := fmt.Sprintf(dnsPrefix+".srvs.%d", i)
			if service, ok := d.GetOk(srvPrefix + ".service"); ok {
				srv.Service = service.(string)
			}
			if protocol, ok := d.GetOk(srvPrefix + ".protocol"); ok {
				srv.Protocol = protocol.(string)
			}
			if domain, ok := d.GetOk(srvPrefix + ".domain"); ok {
				srv.Domain = domain.(string)
			}
			if target, ok := d.GetOk(srvPrefix + ".target"); ok {
				srv.Target = target.(string)
			}
			if port, ok := d.GetOk(srvPrefix + ".port"); ok {
				p, err := strconv.Atoi(port.(string))
				if err != nil {
					return nil, fmt.Errorf("could not convert port '%s' to int", port)
				}
				srv.Port = uint(p)
			}
			if weight, ok := d.GetOk(srvPrefix + ".weight"); ok {
				w, err := strconv.Atoi(weight.(string))
				if err != nil {
					return nil, fmt.Errorf("could not convert weight '%s' to int", weight)
				}
				srv.Weight = uint(w)
			}
			if priority, ok := d.GetOk(srvPrefix + ".priority"); ok {
				w, err := strconv.Atoi(priority.(string))
				if err != nil {
					return nil, fmt.Errorf("could not convert priority '%s' to int", priority)
				}
				srv.Priority = uint(w)
			}
			dnsSRVs = append(dnsSRVs, srv)
		}
	}

	return dnsSRVs, nil
}
