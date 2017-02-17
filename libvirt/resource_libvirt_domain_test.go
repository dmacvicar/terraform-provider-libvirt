package libvirt

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	//"gopkg.in/alexzorin/libvirt-go.v2"
	"encoding/xml"
	libvirt "github.com/dmacvicar/libvirt-go"
)

func TestAccLibvirtDomain_Basic(t *testing.T) {
	var domain libvirt.VirDomain
	var config = fmt.Sprintf(`
            resource "libvirt_domain" "acceptance-test-domain-1" {
                    name = "terraform-test"
            }`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain-1", &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain-1", "name", "terraform-test"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain-1", "memory", "512"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain-1", "vcpu", "1"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_Detailed(t *testing.T) {
	var domain libvirt.VirDomain
	var config = fmt.Sprintf(`
            resource "libvirt_domain" "acceptance-test-domain-2" {
                    name = "terraform-test"
                    memory = 384
                    vcpu = 2
            }`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain-2", &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain-2", "name", "terraform-test"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain-2", "memory", "384"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain-2", "vcpu", "2"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_Volume(t *testing.T) {
	var domain libvirt.VirDomain
	var volume libvirt.VirStorageVol

	var configVolAttached = fmt.Sprintf(`
            resource "libvirt_volume" "acceptance-test-volume" {
                    name = "terraform-test"
            }

            resource "libvirt_domain" "acceptance-test-domain" {
                    name = "terraform-test"
                    disk {
                            volume_id = "${libvirt_volume.acceptance-test-volume.id}"
                    }
            }`)

	var configVolDettached = fmt.Sprintf(`
            resource "libvirt_domain" "acceptance-test-domain" {
                    name = "terraform-test"
            }`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configVolAttached,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					testAccCheckLibvirtVolumeExists("libvirt_volume.acceptance-test-volume", &volume),
				),
			},
			resource.TestStep{
				Config: configVolDettached,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					testAccCheckLibvirtVolumeDoesNotExists("libvirt_volume.acceptance-test-volume", &volume),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_VolumeTwoDisks(t *testing.T) {
	var domain libvirt.VirDomain
	var volume libvirt.VirStorageVol

	var configVolAttached = fmt.Sprintf(`
            resource "libvirt_volume" "acceptance-test-volume1" {
                    name = "terraform-test-vol1"
            }

            resource "libvirt_volume" "acceptance-test-volume2" {
                    name = "terraform-test-vol2"
            }

            resource "libvirt_domain" "acceptance-test-domain" {
                    name = "terraform-test-domain"
                    disk {
                            volume_id = "${libvirt_volume.acceptance-test-volume1.id}"
                    }

                    disk {
                            volume_id = "${libvirt_volume.acceptance-test-volume2.id}"
                    }
            }`)

	var configVolDettached = fmt.Sprintf(`
            resource "libvirt_domain" "acceptance-test-domain" {
                    name = "terraform-test-domain"
            }`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configVolAttached,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					testAccCheckLibvirtVolumeExists("libvirt_volume.acceptance-test-volume1", &volume),
					testAccCheckLibvirtVolumeExists("libvirt_volume.acceptance-test-volume2", &volume),
				),
			},
			resource.TestStep{
				Config: configVolDettached,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					testAccCheckLibvirtVolumeDoesNotExists("libvirt_volume.acceptance-test-volume1", &volume),
					testAccCheckLibvirtVolumeDoesNotExists("libvirt_volume.acceptance-test-volume2", &volume),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_NetworkInterface(t *testing.T) {
	var domain libvirt.VirDomain

	var config = fmt.Sprintf(`
            resource "libvirt_volume" "acceptance-test-volume" {
                    name = "terraform-test"
            }

            resource "libvirt_domain" "acceptance-test-domain" {
                    name = "terraform-test"
                    network_interface = {
                            network_name = "default"
                    }
                    network_interface = {
                            bridge = "br0"
                            mac = "52:54:00:A9:F5:17"
                    }
                    disk {
                            volume_id = "${libvirt_volume.acceptance-test-volume.id}"
                    }
            }`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "network_interface.0.network_name", "default"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "network_interface.1.mac", "52:54:00:A9:F5:17"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_IgnitionObject(t *testing.T) {
	var domain libvirt.VirDomain

	var config = fmt.Sprintf(`
	    resource "ignition_systemd_unit" "acceptance-test-systemd" {
    		name = "example.service"
    		content = "[Service]\nType=oneshot\nExecStart=/usr/bin/echo Hello World\n\n[Install]\nWantedBy=multi-user.target"
	    }

	    resource "ignition_config" "acceptance-test-config" {
    		systemd = [
        		"${ignition_systemd_unit.acceptance-test-systemd.id}",
    		]
	    }

	    resource "libvirt_domain" "acceptance-test-domain" {
		name = "terraform-test-domain"
		coreos_ignition {
		        content = "${ignition_config.acceptance-test-config.rendered}"
		}
	    }
	`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: config,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					testAccCheckIgnitionFileNameExists(&domain),
				),
			},
		},
	})
}

func testAccCheckLibvirtDomainDestroy(s *terraform.State) error {
	virtConn := testAccProvider.Meta().(*Client).libvirt

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "libvirt_domain" {
			continue
		}

		// Try to find the server
		_, err := virtConn.LookupByUUIDString(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf(
				"Error waiting for domain (%s) to be destroyed: %s",
				rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckLibvirtDomainExists(n string, domain *libvirt.VirDomain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No libvirt domain ID is set")
		}

		virConn := testAccProvider.Meta().(*Client).libvirt

		retrieveDomain, err := virConn.LookupByUUIDString(rs.Primary.ID)

		if err != nil {
			return err
		}

		log.Printf("The ID is %s", rs.Primary.ID)

		realId, err := retrieveDomain.GetUUIDString()
		if err != nil {
			return err
		}

		if realId != rs.Primary.ID {
			return fmt.Errorf("Libvirt domain not found")
		}

		*domain = retrieveDomain

		return nil
	}
}

func testAccCheckIgnitionFileNameExists(domain *libvirt.VirDomain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var ignStr string
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "libvirt_domain" {
				continue
			}
			ignStr = rs.Primary.Attributes["coreos_ignition.content"]
		}

		xmlDesc, err := domain.GetXMLDesc(0)
		if err != nil {
			return fmt.Errorf("Error retrieving libvirt domain XML description: %s", err)
		}

		domainDef := newDomainDef()
		err = xml.Unmarshal([]byte(xmlDesc), &domainDef)
		if err != nil {
			return fmt.Errorf("Error reading libvirt domain XML description: %s", err)
		}

		ignitionFile := domainDef.Metadata.TerraformLibvirt.IgnitionFile
		if ignitionFile == "" {
			return fmt.Errorf("No ignition file meta-data")
		}

		hashStr := hash(ignStr)
		hashFile := fmt.Sprint("/tmp/", hashStr, ".ign")
		if ignitionFile != hashFile {
			return fmt.Errorf("Igntion file metadata incorrect %s %s", ignitionFile, hashFile)
		}
		return nil
	}
}
