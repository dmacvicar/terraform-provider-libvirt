package libvirt

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform/helper/resource"
	libvirt "github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
)

const skipQemuEnvVar = "TF_SKIP_QEMU_AGENT"

const domWaitLeaseStillWaiting = "waiting-addresses"
const domWaitLeaseDone = "all-addresses-obtained"

var errDomainInvalidState = errors.New("invalid state for domain")

func domainWaitForLeases(domain *libvirt.Domain, waitForLeases []*libvirtxml.DomainInterface,
	timeout time.Duration, domainDef libvirtxml.Domain, virConn *libvirt.Connect) error {
	waitFunc := func() (interface{}, string, error) {

		state, err := domainGetState(*domain)
		if err != nil {
			return false, "", err
		}

		for _, fatalState := range []string{"crashed", "shutoff", "shutdown", "pmsuspended"} {
			if state == fatalState {
				return false, "", errDomainInvalidState
			}
		}

		if state != "running" {
			return false, domWaitLeaseStillWaiting, nil
		}

		// check we have IPs for all the interfaces we are waiting for
		for _, iface := range waitForLeases {
			found, ignore, err := domainIfaceHasAddress(*domain, *iface, domainDef, virConn)
			if err != nil {
				return false, "", err
			}
			if ignore {
				log.Printf("[DEBUG] we don't care about the IP address for %+v", iface)
				continue
			}
			if !found {
				log.Printf("[DEBUG] IP address not found for iface=%+v: will try in a while", strings.ToUpper(iface.MAC.Address))
				return false, domWaitLeaseStillWaiting, nil
			}
		}

		log.Printf("[DEBUG] all the %d IP addresses obtained for the domain", len(waitForLeases))
		return true, domWaitLeaseDone, nil
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{domWaitLeaseStillWaiting},
		Target:     []string{domWaitLeaseDone},
		Refresh:    waitFunc,
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      5 * time.Second,
	}

	_, err := stateConf.WaitForState()
	log.Print("[DEBUG] wait-for-leases was successful")
	return err
}

func domainIfaceHasAddress(domain libvirt.Domain, iface libvirtxml.DomainInterface,
	domainDef libvirtxml.Domain, virConn *libvirt.Connect) (found bool, ignore bool, err error) {

	mac := strings.ToUpper(iface.MAC.Address)
	if mac == "" {
		log.Printf("[DEBUG] Can't wait without a MAC address: ignoring interface %+v.\n", iface)
		// we can't get the ip without a mac address
		return false, true, nil
	}

	log.Printf("[DEBUG] waiting for network address for iface=%s\n", mac)
	ifacesWithAddr, err := domainGetIfacesInfo(domain, domainDef, virConn)
	if err != nil {
		return false, false, fmt.Errorf("Error retrieving interface addresses: %s", err)
	}
	log.Printf("[DEBUG] ifaces with addresses: %+v\n", ifacesWithAddr)

	for _, ifaceWithAddr := range ifacesWithAddr {
		if mac == strings.ToUpper(ifaceWithAddr.Hwaddr) {
			log.Printf("[DEBUG] found IPs for MAC=%+v: %+v\n", mac, ifaceWithAddr.Addrs)
			return true, false, nil
		}
	}

	log.Printf("[DEBUG] %+v doesn't have IP address(es) yet...\n", mac)
	return false, false, nil
}

func domainGetState(domain libvirt.Domain) (string, error) {
	state, _, err := domain.GetState()
	if err != nil {
		return "", err
	}

	var stateStr string

	switch state {
	case libvirt.DOMAIN_NOSTATE:
		stateStr = "nostate"
	case libvirt.DOMAIN_RUNNING:
		stateStr = "running"
	case libvirt.DOMAIN_BLOCKED:
		stateStr = "blocked"
	case libvirt.DOMAIN_PAUSED:
		stateStr = "paused"
	case libvirt.DOMAIN_SHUTDOWN:
		stateStr = "shutdown"
	case libvirt.DOMAIN_CRASHED:
		stateStr = "crashed"
	case libvirt.DOMAIN_PMSUSPENDED:
		stateStr = "pmsuspended"
	case libvirt.DOMAIN_SHUTOFF:
		stateStr = "shutoff"
	default:
		stateStr = fmt.Sprintf("unknown: %v", state)
	}

	return stateStr, nil
}

func domainIsRunning(domain libvirt.Domain) (bool, error) {
	state, _, err := domain.GetState()
	if err != nil {
		return false, fmt.Errorf("Couldn't get state of domain: %s", err)
	}

	return state == libvirt.DOMAIN_RUNNING, nil
}

func domainGetIfacesInfo(domain libvirt.Domain, domainDef libvirtxml.Domain,
	virConn *libvirt.Connect) ([]libvirt.DomainInterface, error) {

	_, found := os.LookupEnv(skipQemuEnvVar)
	if found {
		log.Printf("[DEBUG] %s defined in environment: skipping qemu-agent", skipQemuEnvVar)
	} else {
		// get all the interfaces using the qemu-agent, this includes also
		// interfaces that are not attached to networks managed by libvirt
		// (eg. bridges, macvtap,...)
		log.Print("[DEBUG] fetching networking interfaces using qemu-agent")
		interfaces := qemuAgentGetInterfacesInfo(&domain, true)
		if len(interfaces) > 0 {
			// the agent will always return all the interfaces, both the
			// ones managed by libvirt and the ones attached to bridge interfaces
			// or macvtap. Hence it has the highest priority
			return interfaces, nil
		}
	}

	log.Print("[DEBUG] getting domain addresses from networks")
	interfaces := getDomainInterfacesFromNetworks(domainDef, virConn)
	if len(interfaces) > 0 {
		return interfaces, nil
	}

	// get all the interfaces attached to libvirt networks
	log.Print("[DEBUG] no interfaces could be obtained with qemu-agent: falling back to the libvirt API")

	domainRunningNow, err := domainIsRunning(domain)
	if err != nil {
		return interfaces, err
	}
	if !domainRunningNow {
		log.Print("[DEBUG] no interfaces could be obtained with libvirt API: domain not running")
		return interfaces, nil
	}

	interfaces, err = domain.ListAllInterfaceAddresses(libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
	if err != nil {
		switch err.(type) {
		default:
			return interfaces, fmt.Errorf("Error retrieving interface addresses: %s", err)
		case libvirt.Error:
			virErr := err.(libvirt.Error)
			if virErr.Code != libvirt.ERR_OPERATION_INVALID || virErr.Domain != libvirt.FROM_QEMU {
				return interfaces, fmt.Errorf("Error retrieving interface addresses: %s", err)
			}
		}
	}
	log.Printf("[DEBUG] Interfaces info obtained with libvirt API:\n%s\n", spew.Sdump(interfaces))

	return interfaces, nil
}

func getDomainInterfacesFromNetworks(domain libvirtxml.Domain,
	virConn *libvirt.Connect) []libvirt.DomainInterface {

	var ifacesList []libvirt.DomainInterface
	var networkNames []string
	var macAddresses []string

	for _, networkInterface := range domain.Devices.Interfaces {
		networkNames = append(networkNames, networkInterface.Source.Network.Network)
		macAddresses = append(macAddresses, strings.ToUpper(networkInterface.MAC.Address))
	}

	networkMacAddresses := make(map[string]map[string][]string)
	for _, networkName := range networkNames {
		network, err := virConn.LookupNetworkByName(networkName)
		if err != nil {
			log.Printf("[ERROR] Error retrieving libvirt network: %s", err)
			return ifacesList
		}
		defer network.Free()

		networkDef, err := newDefNetworkfromLibvirt(network)
		macAddresses := make(map[string][]string)
		for _, ips := range networkDef.IPs {
			for _, dhcpHost := range ips.DHCP.Hosts {
				macAddresses[dhcpHost.MAC] = append(macAddresses[dhcpHost.MAC], dhcpHost.IP)
			}
		}
		networkMacAddresses[networkName] = macAddresses
	}

	for name, networkMacAddress := range networkMacAddresses {
		for mac, ips := range networkMacAddress {
			for _, domMac := range macAddresses {
				if mac == domMac {
					virDom := libvirt.DomainInterface{}
					virDom.Name = name
					virDom.Hwaddr = mac
					for _, ip := range ips {
						virDomIP := libvirt.DomainIPAddress{}
						virDomIP.Addr = ip
						virDom.Addrs = append(virDom.Addrs, virDomIP)
					}
					ifacesList = append(ifacesList, virDom)
				}
			}
		}
	}

	log.Printf("[DEBUG] Interfaces: %s", spew.Sdump(ifacesList))
	return ifacesList
}
