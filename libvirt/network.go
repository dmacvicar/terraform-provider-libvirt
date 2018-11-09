package libvirt

import (
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/hashicorp/terraform/helper/schema"
	libvirt "github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
)

func resourceLibvirtNetworkUpdateDNSHosts(d *schema.ResourceData, network *libvirt.Network) error {
	hostsKey := dnsPrefix + ".hosts"
	if d.HasChange(hostsKey) {
		oldInterface, newInterface := d.GetChange(hostsKey)

		oldEntries, err := parseNetworkDNSHostsChange(oldInterface)
		if err != nil {
			return fmt.Errorf("parse old %s: %s", hostsKey, err)
		}

		newEntries, err := parseNetworkDNSHostsChange(newInterface)
		if err != nil {
			return fmt.Errorf("parse new %s: %s", hostsKey, err)
		}

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
				return fmt.Errorf("serialize update: %s", err)
			}

			err = network.Update(libvirt.NETWORK_UPDATE_COMMAND_DELETE, libvirt.NETWORK_SECTION_DNS_HOST, -1, data, libvirt.NETWORK_UPDATE_AFFECT_LIVE|libvirt.NETWORK_UPDATE_AFFECT_CONFIG)
			if err != nil {
				return fmt.Errorf("delete %s: %s", oldEntry.IP, err)
			}
		}

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
				return fmt.Errorf("serialize update: %s", err)
			}

			err = network.Update(libvirt.NETWORK_UPDATE_COMMAND_ADD_LAST, libvirt.NETWORK_SECTION_DNS_HOST, -1, data, libvirt.NETWORK_UPDATE_AFFECT_LIVE|libvirt.NETWORK_UPDATE_AFFECT_CONFIG)
			if err != nil {
				return fmt.Errorf("add %v: %s", newEntry, err)
			}
		}

		d.SetPartial(hostsKey)
	}

	return nil
}

func parseNetworkDNSHostsChange(change interface{}) (entries []libvirtxml.NetworkDNSHost, err error) {
	slice, ok := change.([]interface{})
	if !ok {
		return entries, errors.New("not slice")
	}

	mapEntries := map[string][]string{}
	for i, entryInterface := range slice {
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
