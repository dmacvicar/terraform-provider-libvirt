package libvirt

import (
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"
	"sort"

	"github.com/hashicorp/terraform/helper/resource"
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

func waitForNetworkActive(network libvirt.Network) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		active, err := network.IsActive()
		if err != nil {
			return nil, "", err
		}
		if active {
			return network, "ACTIVE", nil
		}
		return network, "BUILD", err
	}
}

// wait for network to be up and timeout after 5 minutes.
func waitForNetworkDestroyed(virConn *libvirt.Connect, uuid string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("Waiting for network %s to be destroyed", uuid)
		network, err := virConn.LookupNetworkByUUIDString(uuid)
		if err.(libvirt.Error).Code == libvirt.ERR_NO_NETWORK {
			return virConn, "NOT-EXISTS", nil
		}
		defer network.Free()
		return virConn, "ACTIVE", err
	}
}

func setDhcpByCIDRAdressesSubnets(d *schema.ResourceData, networkDef *libvirtxml.Network) error {
	if addresses, ok := d.GetOk("addresses"); ok {
		ipsPtrsLst := []libvirtxml.NetworkIP{}
		for _, addressI := range addresses.([]interface{}) {
			// get the IP address entry for this subnet (with a guessed DHCP range)
			dni, dhcp, err := setNetworkIP(addressI.(string))
			if err != nil {
				return err
			}
			if d.Get("dhcp.0.enabled").(bool) {
				// prepare host static ip assignement if present
				var staticHosts []libvirtxml.NetworkDHCPHost
				if hostsCount, ok := d.GetOk("dhcp.0.hosts.#"); ok {
					for i := 0; i < hostsCount.(int); i++ {
						ip := d.Get(fmt.Sprintf("dhcp.0.hosts.%d.ip", i)).(string)
						if net.ParseIP(ip) == nil {
							return fmt.Errorf("Could not parse address '%s'", ip)
						}
						mac := d.Get(fmt.Sprintf("dhcp.0.hosts.%d.mac", i)).(string)
						name := d.Get(fmt.Sprintf("dhcp.0.hosts.%d.name", i)).(string)
						staticHosts = append(staticHosts, libvirtxml.NetworkDHCPHost{
							IP:   ip,
							MAC:  mac,
							Name: name,
						})
					}
					dhcp.Hosts = staticHosts
				}
				dni.DHCP = dhcp
			} else {
				// if a network exist with enabled but an user want to disable
				// dhcp, we need to set dhcp struct to nil.
				dni.DHCP = nil
			}

			ipsPtrsLst = append(ipsPtrsLst, *dni)
		}
		networkDef.IPs = ipsPtrsLst
	}
	return nil
}

func setNetworkIP(address string) (*libvirtxml.NetworkIP, *libvirtxml.NetworkDHCP, error) {
	_, ipNet, err := net.ParseCIDR(address)
	if err != nil {
		return nil, nil, fmt.Errorf("Error parsing addresses definition '%s': %s", address, err)
	}
	ones, bits := ipNet.Mask.Size()
	family := "ipv4"
	if bits == (net.IPv6len * 8) {
		family = "ipv6"
	}
	ipsRange := 2 ^ bits - 2 ^ ones
	if ipsRange < 4 {
		return nil, nil, fmt.Errorf("Netmask seems to be too strict: only %d IPs available (%s)", ipsRange-3, family)
	}

	// we should calculate the range served by DHCP. For example, for
	// 192.168.121.0/24 we will serve 192.168.121.2 - 192.168.121.254
	start, end := networkRange(ipNet)

	// skip the .0, (for the network),
	start[len(start)-1]++

	// assign the .1 to the host interface
	dni := &libvirtxml.NetworkIP{
		Address: start.String(),
		Prefix:  uint(ones),
		Family:  family,
	}

	start[len(start)-1]++ // then skip the .1
	end[len(end)-1]--     // and skip the .255 (for broadcast)

	dhcp := &libvirtxml.NetworkDHCP{
		Ranges: []libvirtxml.NetworkDHCPRange{
			{
				Start: start.String(),
				End:   end.String(),
			},
		},
	}

	return dni, dhcp, nil
}
