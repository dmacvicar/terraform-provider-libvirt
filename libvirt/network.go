package libvirt

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/libvirt/libvirt-go"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

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

// waitForNetworkDestroyed waits for a network to destroyed
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

// getNetModeFromResource returns the network mode fromm a network definition
func getNetModeFromResource(d *schema.ResourceData) string {
	return strings.ToLower(d.Get("mode").(string))
}

// getIPsFromResource gets the IPs configurations from the resource definition
func getIPsFromResource(d *schema.ResourceData) ([]libvirtxml.NetworkIP, error) {
	addresses, ok := d.GetOk("addresses")
	if !ok {
		return []libvirtxml.NetworkIP{}, nil
	}

	// check if DHCP must be enabled by default
	var dhcpEnabled bool
	netMode := getNetModeFromResource(d)
	if netMode == netModeIsolated || netMode == netModeNat || netMode == netModeRoute {
		dhcpEnabled = true
	}

	ipsPtrsLst := []libvirtxml.NetworkIP{}
	for num, addressI := range addresses.([]interface{}) {
		// get the IP address entry for this subnet (with a guessed DHCP range)
		dni, dhcp, err := getNetworkIPConfig(addressI.(string))
		if err != nil {
			return nil, err
		}

		dhcpKey := fmt.Sprintf("dhcp.%d.enabled", num)
		dhcpEnabledByUser, ok := d.GetOkExists(dhcpKey)
		if ok {
			dhcpEnabled = dhcpEnabledByUser.(bool)
		}

		if dhcpEnabled {
			dni.DHCP = dhcp
		} else {
			// if a network exist with enabled but an user want to disable it
			// we need to set DHCP struct to nil.
			dni.DHCP = nil
		}

		ipsPtrsLst = append(ipsPtrsLst, *dni)
	}

	return ipsPtrsLst, nil
}

func getNetworkIPConfig(address string) (*libvirtxml.NetworkIP, *libvirtxml.NetworkDHCP, error) {
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

// getBridgeFromResource returns a libvirt's NetworkBridge
// from the ResourceData provided.
func getBridgeFromResource(d *schema.ResourceData) *libvirtxml.NetworkBridge {
	// use a bridge provided by the user, or create one otherwise (libvirt will assign on automatically when empty)
	bridgeName := ""
	if b, ok := d.GetOk("bridge"); ok {
		bridgeName = b.(string)
	}

	bridge := &libvirtxml.NetworkBridge{
		Name: bridgeName,
		STP:  "on",
	}

	return bridge
}

// getDomainFromResource returns a libvirt's NetworkDomain
// from the ResourceData provided.
func getDomainFromResource(d *schema.ResourceData) *libvirtxml.NetworkDomain {
	domainName, ok := d.GetOk("domain")
	if !ok {
		return nil
	}

	domain := &libvirtxml.NetworkDomain{
		Name: domainName.(string),
	}

	if dnsLocalOnly, ok := d.GetOk(dnsPrefix + ".local_only"); ok {
		if dnsLocalOnly.(bool) {
			domain.LocalOnly = "yes" // this "boolean" must be "yes"|"no"
		}
	}

	return domain
}

func getMTUFromResource(d *schema.ResourceData) *libvirtxml.NetworkMTU {
	if mtu, ok := d.GetOk("mtu"); ok {
		return &libvirtxml.NetworkMTU{Size: uint(mtu.(int))}
	}

	return nil
}

// getDNSMasqOptionFromResource returns a list of dnsmasq options
// from the network definition
func getDNSMasqOptionFromResource(d *schema.ResourceData) ([]libvirtxml.NetworkDnsmasqOption, error) {
	var dnsmasqOption []libvirtxml.NetworkDnsmasqOption
	dnsmasqOptionPrefix := "dnsmasq_options.0"
	if dnsmasqOptionCount, ok := d.GetOk(dnsmasqOptionPrefix + ".options.#"); ok {
		for i := 0; i < dnsmasqOptionCount.(int); i++ {
			dnsmasqOptionsPrefix := fmt.Sprintf(dnsmasqOptionPrefix+".options.%d", i)

			optionName := d.Get(dnsmasqOptionsPrefix + ".option_name").(string)
			optionValue := d.Get(dnsmasqOptionsPrefix + ".option_value").(string)
			dnsmasqOption = append(dnsmasqOption, libvirtxml.NetworkDnsmasqOption{
				Value: optionName + "=" + optionValue,
			})
		}
	}

	return dnsmasqOption, nil
}
