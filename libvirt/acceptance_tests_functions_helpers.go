package libvirt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	libvirt "github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
)

// This file contain function helpers used for testsuite/testacc

// the following helpers are used in mostly all testacc.

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

// test in all testacc that resource is destroyed
func testaccCheckLibvirtDestroyResource(resourceName string, virConn libvirt.Connect) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceName {
				continue
			}
			if rs.Type == "libvirt_cloudinit_disk" {
				ciKey, errKey := getCloudInitVolumeKeyFromTerraformID(rs.Primary.ID)
				if errKey != nil {
					return errKey
				}
				_, err := virConn.LookupStorageVolByKey(ciKey)
				if err == nil {
					return fmt.Errorf(
						"Error waiting for CloudInit Volume (%s) to be destroyed: %s",
						ciKey, err)
				}
			}

			if rs.Type == "libvirt_ignition" {
				ignKey, errKey := getIgnitionVolumeKeyFromTerraformID(rs.Primary.ID)
				if errKey != nil {
					return errKey
				}
				_, err := virConn.LookupStorageVolByKey(ignKey)
				if err == nil {
					return fmt.Errorf(
						"Error waiting for IgnitionVolume (%s) to be destroyed: %s",
						ignKey, err)
				}
			}

			if rs.Type == "libvirt_domain" {
				_, err := virConn.LookupDomainByUUIDString(rs.Primary.ID)
				if err == nil {
					return fmt.Errorf("Error waiting for domain (%s) to be destroyed: %s", rs.Primary.ID, err)
				}
			}

			if rs.Type == "libvirt_network" {
				_, err := virConn.LookupNetworkByUUIDString(rs.Primary.ID)
				if err == nil {
					return fmt.Errorf(
						"Error waiting for network (%s) to be destroyed: %s",
						rs.Primary.ID, err)
				}
			}
			if rs.Type == "libvirt_volume" {
				_, err := virConn.LookupStorageVolByKey(rs.Primary.ID)
				if err == nil {
					return fmt.Errorf(
						"Error waiting for volume (%s) to be destroyed: %s",
						rs.Primary.ID, err)
				}
			}
		}
		return nil
	}
}

// ** resource specifics helpers **

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
