package libvirt

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
	"github.com/terraform-providers/terraform-provider-ignition/ignition"
)

// This file contain function helpers used for testsuite/testacc

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"libvirt":  testAccProvider,
		"ignition": ignition.Provider(),
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("LIBVIRT_DEFAULT_URI"); v == "" {
		t.Fatal("LIBVIRT_DEFAULT_URI must be set for acceptance tests")
	}
}

func testAccEnabled() bool {
	v := os.Getenv("TF_ACC")
	return v == "1" || strings.ToLower(v) == "true"
}

func skipIfPrivilegedDisabled(t *testing.T) {
	if os.Getenv("TF_LIBVIRT_DISABLE_PRIVILEGED_TESTS") != "" {
		t.Skip("skipping test; Enviornemnt variable `TF_LIBVIRT_DISABLE_PRIVILEGED_TESTS` is set")
	}
}

// //////////////////////////////////////////////////////////////////
// general
// //////////////////////////////////////////////////////////////////

// getResourceFromTerraformState get aresource by name
// from terraform states produced during testacc
// and return the resource
func getResourceFromTerraformState(resourceName string, state *terraform.State) (*terraform.ResourceState, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("Not found: %s", resourceName)
	}

	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("No libvirt resource key ID is set")
	}
	return rs, nil
}

// ** resource specifics helpers **

// getPoolFromTerraformState lookup pool by name and return the libvirt pool from a terraform state
func getPoolFromTerraformState(name string, state *terraform.State, virConn libvirt.Connect) (*libvirt.StoragePool, error) {
	rs, err := getResourceFromTerraformState(name, state)
	if err != nil {
		return nil, err
	}

	pool, err := virConn.LookupStoragePoolByUUIDString(rs.Primary.ID)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG]:The ID is %s", rs.Primary.ID)
	return pool, nil
}

// getVolumeFromTerraformState lookup volume by name and return the libvirt volume from a terraform state
func getVolumeFromTerraformState(name string, state *terraform.State, virConn libvirt.Connect) (*libvirt.StorageVol, error) {
	rs, err := getResourceFromTerraformState(name, state)
	if err != nil {
		return nil, err
	}

	vol, err := virConn.LookupStorageVolByKey(rs.Primary.ID)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG]:The ID is %s", rs.Primary.ID)
	return vol, nil
}

// helper used in network tests for retrieve xml network definition.
func getNetworkDef(state *terraform.State, name string, virConn libvirt.Connect) (*libvirtxml.Network, error) {
	var network *libvirt.Network
	rs, err := getResourceFromTerraformState(name, state)
	if err != nil {
		return nil, err
	}
	network, err = virConn.LookupNetworkByUUIDString(rs.Primary.ID)
	if err != nil {
		return nil, err
	}
	networkDef, err := getXMLNetworkDefFromLibvirt(network)
	if err != nil {
		return nil, fmt.Errorf("Error reading libvirt network XML description: %s", err)
	}
	return &networkDef, nil
}

// //////////////////////////////////////////////////////////////////
// network
// //////////////////////////////////////////////////////////////////

// testAccCheckNetworkExists checks that the network exists
func testAccCheckNetworkExists(name string, network *libvirt.Network) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		rs, err := getResourceFromTerraformState(name, state)
		if err != nil {
			return err
		}

		virConn := testAccProvider.Meta().(*Client).libvirt
		networkRetrived, err := virConn.LookupNetworkByUUIDString(rs.Primary.ID)
		if err != nil {
			return err
		}

		realID, err := networkRetrived.GetUUIDString()
		if err != nil {
			return err
		}

		if realID != rs.Primary.ID {
			return fmt.Errorf("Libvirt network not found")
		}

		*network = *networkRetrived

		return nil
	}
}

// testAccCheckLibvirtNetworkDestroy checks that the network has been destroyed
func testAccCheckLibvirtNetworkDestroy(s *terraform.State) error {
	virtConn := testAccProvider.Meta().(*Client).libvirt
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "libvirt_network" {
			continue
		}
		_, err := virtConn.LookupNetworkByUUIDString(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf(
				"Error waiting for network (%s) to be destroyed: %s",
				rs.Primary.ID, err)
		}
	}
	return nil
}

// testAccCheckDNSHosts checks the expected DNS hosts in a network
func testAccCheckDNSHosts(name string, expected []libvirtxml.NetworkDNSHost) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		virConn := testAccProvider.Meta().(*Client).libvirt
		networkDef, err := getNetworkDef(s, name, *virConn)
		if err != nil {
			return err
		}
		if networkDef.DNS == nil {
			return fmt.Errorf("DNS block not found in networkDef")
		}
		actual := networkDef.DNS.Host
		if len(expected) != len(actual) {
			return fmt.Errorf("len(expected): %d != len(actual): %d", len(expected), len(actual))
		}
		for _, e := range expected {
			found := false
			for _, a := range actual {
				if reflect.DeepEqual(a.IP, e.IP) && reflect.DeepEqual(a.Hostnames, e.Hostnames) {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("Unable to find:%v in: %v", e, actual)
			}
		}
		return nil
	}
}

// testAccCheckLibvirtNetworkDhcpStatus checks the expected DHCP status
func testAccCheckLibvirtNetworkDhcpStatus(name string, expectedDhcpStatus string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt
		networkDef, err := getNetworkDef(s, name, *virConn)
		if err != nil {
			return err
		}
		if expectedDhcpStatus == "disabled" {
			for _, ips := range networkDef.IPs {
				// &libvirtxml.NetworkDHCP{..} should be nil when dhcp is disabled
				if ips.DHCP != nil {
					fmt.Printf("%#v", ips.DHCP)
					return fmt.Errorf("the network should have DHCP disabled")
				}
			}
		}
		if expectedDhcpStatus == "enabled" {
			for _, ips := range networkDef.IPs {
				if ips.DHCP == nil {
					return fmt.Errorf("the network should have DHCP enabled")
				}
			}
		}
		return nil
	}
}

// testAccCheckLibvirtNetworkBridge checks the bridge exists and has the expected properties
func testAccCheckLibvirtNetworkBridge(resourceName string, bridgeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		virConn := testAccProvider.Meta().(*Client).libvirt
		networkDef, err := getNetworkDef(s, resourceName, *virConn)
		if err != nil {
			return err
		}

		if networkDef.Bridge == nil {
			return fmt.Errorf("Bridge type of network should be not nil")
		}

		if networkDef.Bridge.Name != bridgeName {
			fmt.Printf("%#v", networkDef)
			return fmt.Errorf("fail: network brigde property were not set correctly")
		}

		return nil
	}
}

// testAccCheckLibvirtNetworkDNSForwarders checks the DNS forwarders in the libvirt network
func testAccCheckLibvirtNetworkDNSForwarders(name string, expected []libvirtxml.NetworkDNSForwarder) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		virConn := testAccProvider.Meta().(*Client).libvirt

		networkDef, err := getNetworkDef(s, name, *virConn)
		if err != nil {
			return err
		}
		if networkDef.DNS == nil {
			return fmt.Errorf("DNS block not found in networkDef")
		}
		actual := networkDef.DNS.Forwarders
		if len(expected) != len(actual) {
			return fmt.Errorf("len(expected): %d != len(actual): %d", len(expected), len(actual))
		}
		for _, e := range expected {
			found := false
			for _, a := range actual {
				if reflect.DeepEqual(a, e) {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("Unable to find %v in %v", e, actual)
			}
		}
		return nil
	}
}

// testAccCheckLibvirtNetworkLocalOnly checks the local-only property of the Domain
func testAccCheckLibvirtNetworkLocalOnly(name string, expectLocalOnly bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		virConn := testAccProvider.Meta().(*Client).libvirt

		networkDef, err := getNetworkDef(s, name, *virConn)
		if err != nil {
			return err
		}
		if expectLocalOnly {
			if networkDef.Domain == nil || networkDef.Domain.LocalOnly != "yes" {
				return fmt.Errorf("networkDef.Domain.LocalOnly is not true")
			}
		} else {
			if networkDef.Domain != nil && networkDef.Domain.LocalOnly != "no" {
				return fmt.Errorf("networkDef.Domain.LocalOnly is true")
			}
		}
		return nil
	}
}

// testAccCheckLibvirtNetworkDNSEnable checks the dns-enable property of the Domain
func testAccCheckLibvirtNetworkDNSEnableOrDisable(name string, expectDNS bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		virConn := testAccProvider.Meta().(*Client).libvirt

		networkDef, err := getNetworkDef(s, name, *virConn)
		if err != nil {
			return err
		}
		if expectDNS {
			if networkDef.DNS == nil || networkDef.DNS.Enable != "yes" {
				return fmt.Errorf("networkDef.DNS.Enable is not true")
			}
		}
		if !expectDNS {
			if networkDef.DNS != nil && networkDef.DNS.Enable != "no" {
				return fmt.Errorf("networkDef.DNS.Enable is true")
			}
		}
		return nil
	}
}

// testAccCheckDnsmasqOptions checks the expected Dnsmasq options in a network
func testAccCheckDnsmasqOptions(name string, expected []libvirtxml.NetworkDnsmasqOption) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		virConn := testAccProvider.Meta().(*Client).libvirt
		networkDef, err := getNetworkDef(s, name, *virConn)
		if err != nil {
			return err
		}
		if networkDef.DnsmasqOptions == nil {
			return fmt.Errorf("DnsmasqOptions block not found in networkDef")
		}
		actual := networkDef.DnsmasqOptions.Option
		if len(expected) != len(actual) {
			return fmt.Errorf("len(expected): %d != len(actual): %d", len(expected), len(actual))
		}
		for _, e := range expected {
			found := false
			for _, a := range actual {
				if reflect.DeepEqual(a.Value, e.Value) {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("Unable to find:%v in: %v", e, actual)
			}
		}
		return nil
	}
}
