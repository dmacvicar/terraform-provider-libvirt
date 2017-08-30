package libvirt

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	libvirt "github.com/libvirt/libvirt-go"
)

func TestAccLibvirtDomain_Basic(t *testing.T) {
	var domain libvirt.Domain
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
	var domain libvirt.Domain
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
	var domain libvirt.Domain
	var volume libvirt.StorageVol

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
	var domain libvirt.Domain
	var volume libvirt.StorageVol

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

func TestAccLibvirtDomain_ScsiDisk(t *testing.T) {
	var domain libvirt.Domain
	var configScsi = fmt.Sprintf(`
            resource "libvirt_volume" "acceptance-test-volume1" {
                    name = "terraform-test-vol1"
            }

            resource "libvirt_domain" "acceptance-test-domain" {
                    name = "terraform-test-domain"
                    disk {
                            volume_id = "${libvirt_volume.acceptance-test-volume1.id}"
                            scsi = "yes"
                            wwn = "000000123456789a"
                    }
            }`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configScsi,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					testAccCheckLibvirtScsiDisk("000000123456789a", &domain),
				),
			},
		},
	})

}

func TestAccLibvirtDomain_NetworkInterface(t *testing.T) {
	var domain libvirt.Domain

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
                            network_name = "default"
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

func TestAccLibvirtDomain_Graphics(t *testing.T) {
	var domain libvirt.Domain

	var config = fmt.Sprintf(`
            resource "libvirt_volume" "acceptance-test-graphics" {
                    name = "terraform-test"
            }

            resource "libvirt_domain" "acceptance-test-domain" {
                    name = "terraform-test"
                    graphics {
                            type = "spice"
                            autoport = "yes"
                            listen_type = "none"
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
						"libvirt_domain.acceptance-test-domain", "graphics.type", "spice"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "graphics.autoport", "yes"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "graphics.listen_type", "none"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_IgnitionObject(t *testing.T) {
	var domain libvirt.Domain
	var volume libvirt.StorageVol

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

	    resource "libvirt_ignition" "ignition" {
	    	name = "ignition"
	    	content = "${ignition_config.acceptance-test-config.rendered}"
	    }

	    resource "libvirt_domain" "acceptance-test-domain" {
		name = "terraform-test-domain"
		coreos_ignition = "${libvirt_ignition.ignition.id}"
	    }
	`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					testAccCheckIgnitionVolumeExists("libvirt_ignition.ignition", &volume),
					testAccCheckIgnitionXML(&domain, &volume),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_Cpu(t *testing.T) {
	var domain libvirt.Domain

	var config = fmt.Sprintf(`
            resource "libvirt_domain" "acceptance-test-domain" {
                    name = "terraform-test"
                    cpu {
                            mode = "custom"
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
						"libvirt_domain.acceptance-test-domain", "cpu.mode", "custom"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_Autostart(t *testing.T) {
	var domain libvirt.Domain

	var config = fmt.Sprintf(`
            resource "libvirt_domain" "acceptance-test-domain" {
                    name = "terraform-test"
                    autostart = true
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
					resource.TestCheckResourceAttr("libvirt_domain.acceptance-test-domain", "autostart", "true"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_Filesystems(t *testing.T) {
	var domain libvirt.Domain

	var config = fmt.Sprintf(`
            resource "libvirt_domain" "acceptance-test-domain" {
                    name = "terraform-test"
                    filesystem {
                      source = "/tmp"
                      target = "tmp"
                      readonly = false
                    }
                    filesystem {
                      source = "/proc"
                      target = "proc"
                      readonly = true
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
						"libvirt_domain.acceptance-test-domain", "filesystem.0.source", "/tmp"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "filesystem.0.target", "tmp"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "filesystem.0.readonly", "0"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "filesystem.1.source", "/proc"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "filesystem.1.target", "proc"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "filesystem.1.readonly", "1"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_Consoles(t *testing.T) {
	var domain libvirt.Domain

	var config = fmt.Sprintf(`
            resource "libvirt_domain" "acceptance-test-domain" {
                    name = "terraform-test"
                    console {
                      type = "pty"
                      target_port = "0"
                      source_path = "/dev/pts/1"
                    }
                    console {
                      type = "pty"
                      target_port = "0"
                      target_type = "virtio"
                      source_path = "/dev/pts/2"
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
						"libvirt_domain.acceptance-test-domain", "console.0.type", "pty"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "console.0.target_port", "0"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "console.0.source_path", "/dev/pts/1"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "console.1.type", "pty"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "console.1.target_port", "0"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "console.1.target_type", "virtio"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "console.1.source_path", "/dev/pts/2"),
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
		_, err := virtConn.LookupDomainByUUIDString(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf(
				"Error waiting for domain (%s) to be destroyed: %s",
				rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckLibvirtDomainExists(n string, domain *libvirt.Domain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No libvirt domain ID is set")
		}

		virConn := testAccProvider.Meta().(*Client).libvirt

		retrieveDomain, err := virConn.LookupDomainByUUIDString(rs.Primary.ID)

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

		*domain = *retrieveDomain

		return nil
	}
}

func testAccCheckIgnitionXML(domain *libvirt.Domain, volume *libvirt.StorageVol) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		xmlDesc, err := domain.GetXMLDesc(0)
		if err != nil {
			return fmt.Errorf("Error retrieving libvirt domain XML description: %s", err)
		}

		ignitionKey, err := volume.GetKey()
		if err != nil {
			return err
		}
		ignStr := fmt.Sprintf("name=opt/com.coreos/config,file=%s", ignitionKey)

		domainDef := newDomainDef()
		err = xml.Unmarshal([]byte(xmlDesc), &domainDef)
		if err != nil {
			return fmt.Errorf("Error reading libvirt domain XML description: %s", err)
		}

		cmdLine := domainDef.QEMUCommandline.Args
		for i, cmd := range cmdLine {
			if i == 1 && cmd.Value != ignStr {
				return fmt.Errorf("libvirt domain fw_cfg XML is incorrect %s", cmd.Value)
			}
		}
		return nil
	}
}

func TestHash(t *testing.T) {
	actual := hash("this is a test")
	expected := "2e99758548972a8e8822ad47fa1017ff72f06f3ff6a016851f45c398732bc50c"

	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func testAccCheckLibvirtScsiDisk(n string, domain *libvirt.Domain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		xmlDesc, err := domain.GetXMLDesc(0)
		if err != nil {
			return fmt.Errorf("Error retrieving libvirt domain XML description: %s", err)
		}

		domainDef := newDomainDef()
		err = xml.Unmarshal([]byte(xmlDesc), &domainDef)
		if err != nil {
			return fmt.Errorf("Error reading libvirt domain XML description: %s", err)
		}

		disks := domainDef.Devices.Disks
		for _, disk := range disks {
			if diskBus := disk.Target.Bus; diskBus != "scsi" {
				return fmt.Errorf("Disk bus is not scsi")
			}
			if wwn := disk.WWN; wwn != n {
				return fmt.Errorf("Disk wwn %s is not equal to %s", wwn, n)
			}
		}
		return nil
	}
}

func createNvramFile() (string, error) {
	// size of an accepted, valid, nvram backing store
	nvram_dummy_buffer := make([]byte, 131072)
	file, err := ioutil.TempFile("/tmp", "nvram")
	if err != nil {
		return "", err
	}
	file.Chmod(0777)
	_, err = file.Write(nvram_dummy_buffer)
	if err != nil {
		return "", err
	}
	if file.Close() != nil {
		return "", err
	}
	return file.Name(), nil
}

func TestAccLibvirtDomain_FirmwareNoTemplate(t *testing.T) {
	nvram_path, err := createNvramFile()
	if err != nil {
		t.Fatal(err)
	}

	var domain libvirt.Domain
	var config = fmt.Sprintf(`
            resource "libvirt_domain" "acceptance-test-domain" {
                name = "terraform-test-firmware-no-template"
	            firmware = "/usr/share/ovmf/OVMF.fd"
                nvram {
                    file = "%s"
  	            }
            }`, nvram_path)

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
						"libvirt_domain.acceptance-test-domain", "name", "terraform-test-firmware-no-template"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "nvram.file", nvram_path),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "firmware", "/usr/share/ovmf/OVMF.fd"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_FirmwareTemplate(t *testing.T) {
	nvram_path, err := createNvramFile()
	if err != nil {
		t.Fatal(err)
	}

	var domain libvirt.Domain
	var config = fmt.Sprintf(`
            resource "libvirt_domain" "acceptance-test-domain" {
                name = "terraform-test-firmware-with-template"
                firmware = "/usr/share/ovmf/OVMF.fd"
                nvram {
                	file = "%s"
                	template = "/usr/share/qemu/OVMF.fd"
                }
            }`, nvram_path)

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
						"libvirt_domain.acceptance-test-domain", "name", "terraform-test-firmware-with-template"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "nvram.file", nvram_path),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "nvram.template", "/usr/share/qemu/OVMF.fd"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "firmware", "/usr/share/ovmf/OVMF.fd"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_MachineType(t *testing.T) {
	var domain libvirt.Domain

	// Using machine type of pc as this is earliest QEMU target
	// and so most likely to be available
	var config = fmt.Sprintf(`
            resource "libvirt_domain" "acceptance-test-domain" {
                    name = "terraform-test"
                    machine = "pc"
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
					resource.TestCheckResourceAttr("libvirt_domain.acceptance-test-domain", "machine", "pc"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_ArchType(t *testing.T) {
	var domain libvirt.Domain

	// Using i686 as architecture in case anyone running tests on an i686 only host
	var config = fmt.Sprintf(`
            resource "libvirt_domain" "acceptance-test-domain" {
                    name = "terraform-test"
                    arch  = "i686"
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
					resource.TestCheckResourceAttr("libvirt_domain.acceptance-test-domain", "arch", "i686"),
				),
			},
		},
	})
}
