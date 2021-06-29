package libvirt

import (
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"

	"github.com/hashicorp/terraform/helper/schema"
	libvirt "github.com/libvirt/libvirt-go"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

// from the network definition
func getDHCPHostsFromResource(d *schema.ResourceData, num int) ([]libvirtxml.NetworkDHCPHost, error) {
	var dhcpHosts []libvirtxml.NetworkDHCPHost
	prefix := fmt.Sprintf(dhcpPrefix+".%d.host", num)
	log.Printf("using prefix: %s\n", prefix)
	if dhcpHostCount, ok := d.GetOk(prefix + ".#"); ok {
		for i := 0; i < dhcpHostCount.(int); i++ {
			hostPrefix := fmt.Sprintf(prefix+".%d", i)

			address := d.Get(hostPrefix + ".ip").(string)
			if net.ParseIP(address) == nil {
				return nil, fmt.Errorf("Could not parse address '%s'", address)
			}
			name := d.Get(hostPrefix + ".name").(string)
			mac := d.Get(hostPrefix + ".mac").(string)
			//id := d.Get(hostPrefix + ".id").(string)

			dhcpHosts = append(dhcpHosts, libvirtxml.NetworkDHCPHost{
				IP:   address,
				Name: name,
				MAC:  mac,
				//	ID:   id,
			})
		}
	}

	return dhcpHosts, nil
}

// updateDHCPHosts detects changes in the DHCP host entries
// updating the network definition accordingly
func updateDHCPHosts(d *schema.ResourceData, num int, network *libvirt.Network) error {
	prefix := fmt.Sprintf(dhcpPrefix+".%d.host", num)
	if d.HasChange(prefix) {
		oldInterface, newInterface := d.GetChange(prefix)

		oldEntries, err := parseNetworkDhcpHostsChange(oldInterface)
		if err != nil {
			return fmt.Errorf("parse old %s: %s", prefix, err)
		}

		newEntries, err := parseNetworkDhcpHostsChange(newInterface)
		if err != nil {
			return fmt.Errorf("parse new %s: %s", prefix, err)
		}

		// process all the old DHCP entries that must be removed
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

			data, err := xmlMarshallIndented(libvirtxml.NetworkDHCPHost{IP: oldEntry.IP})
			if err != nil {
				return fmt.Errorf("serialize update: %s", err)
			}

			log.Printf("[DEBUG] genetated delete dhcp host xml: %s\n", data)

			err = network.Update(libvirt.NETWORK_UPDATE_COMMAND_DELETE, libvirt.NETWORK_SECTION_IP_DHCP_HOST, -1, data, libvirt.NETWORK_UPDATE_AFFECT_LIVE|libvirt.NETWORK_UPDATE_AFFECT_CONFIG)
			if err != nil {
				return fmt.Errorf("delete %s: %s", oldEntry.IP, err)
			}
		}

		// process all the new DHCP entries that must be added
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

			log.Printf("[DEBUG] genetated add host xml: %s\n", data)
			err = network.Update(libvirt.NETWORK_UPDATE_COMMAND_ADD_LAST, libvirt.NETWORK_SECTION_IP_DHCP_HOST, -1, data, libvirt.NETWORK_UPDATE_AFFECT_LIVE|libvirt.NETWORK_UPDATE_AFFECT_CONFIG)
			if err != nil {
				return fmt.Errorf("add %v: %s", newEntry, err)
			}
		}

		d.SetPartial(prefix)
	}

	return nil
}

func parseNetworkDhcpHostsChange(change interface{}) (entries []libvirtxml.NetworkDHCPHost, err error) {
	slice, ok := change.([]interface{})
	if !ok {
		return entries, errors.New("not slice")
	}

	mapEntries := map[string][]string{}

	entries = make([]libvirtxml.NetworkDHCPHost, 0, len(mapEntries))
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

		hostnameInterface, ok := entryMap["name"]
		if !ok {
			return nil, fmt.Errorf("entry %d.name is missing", i)
		}

		name, ok := hostnameInterface.(string)
		if !ok {
			return nil, fmt.Errorf("entry %d.hostname is not a string", i)
		}

		macInterface, ok := entryMap["mac"]
		if !ok {
			return nil, fmt.Errorf("entry %d.mac is missing", i)
		}

		mac, ok := macInterface.(string)
		if !ok {
			return nil, fmt.Errorf("entry %d.mac is not a string", i)
		}

		entries = append(entries, libvirtxml.NetworkDHCPHost{
			IP:   ip,
			Name: name,
			MAC:  mac,
		})
	}

	return entries, nil
}
