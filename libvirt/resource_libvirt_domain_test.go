package libvirt

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	libvirt "github.com/libvirt/libvirt-go"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

func TestAccLibvirtDomain_Basic(t *testing.T) {
	var domain libvirt.Domain
	randomResourceName := acctest.RandString(10)
	randomDomainName := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_domain" "%s" {
					name = "%s"
				}`, randomResourceName, randomDomainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomResourceName, &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomResourceName, "name", randomDomainName),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomResourceName, "memory", "512"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomResourceName, "vcpu", "1"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_Detailed(t *testing.T) {
	var domain libvirt.Domain
	randomResourceName := acctest.RandString(10)
	randomDomainName := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_domain" "%s" {
					name   = "%s"
					memory = 384
					vcpu   = 2
				}`, randomResourceName, randomDomainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomResourceName, &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomResourceName, "name", randomDomainName),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomResourceName, "memory", "384"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomResourceName, "vcpu", "2"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_Volume(t *testing.T) {
	var domain libvirt.Domain
	var volume libvirt.StorageVol
	randomVolumeName := acctest.RandString(10)
	randomDomainName := acctest.RandString(10)
	var configVolAttached = fmt.Sprintf(`
	resource "libvirt_volume" "%s" {
		name = "%s"
	}

	resource "libvirt_domain" "%s" {
		name = "%s"
		disk {
			volume_id = "${libvirt_volume.%s.id}"
		}
	}`, randomVolumeName, randomVolumeName, randomDomainName, randomDomainName, randomVolumeName)

	var configVolDettached = fmt.Sprintf(`
	resource "libvirt_domain" "%s" {
		name = "%s"
	}`, randomDomainName, randomDomainName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: configVolAttached,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeName, &volume),
				),
			},
			{
				Config: configVolDettached,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					testAccCheckLibvirtVolumeDoesNotExists("libvirt_volume."+randomVolumeName, &volume),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_VolumeTwoDisks(t *testing.T) {
	var domain libvirt.Domain
	var volume libvirt.StorageVol
	randomVolumeName := acctest.RandString(10)
	randomVolumeName2 := acctest.RandString(9)
	randomDomainName := acctest.RandString(10)

	var configVolAttached = fmt.Sprintf(`
	resource "libvirt_volume" "%s" {
		name = "%s"
	}

	resource "libvirt_volume" "%s" {
		name = "%s"
	}

	resource "libvirt_domain" "%s" {
		name = "%s"
		disk {
			volume_id = "${libvirt_volume.%s.id}"
		}

		disk {
			volume_id = "${libvirt_volume.%s.id}"
		}
	}`, randomVolumeName, randomVolumeName, randomVolumeName2, randomVolumeName2, randomDomainName, randomDomainName, randomVolumeName, randomVolumeName2)

	var configVolDettached = fmt.Sprintf(`
	resource "libvirt_domain" "%s" {
		name = "%s"
	}`, randomDomainName, randomDomainName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: configVolAttached,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeName, &volume),
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeName2, &volume),
				),
			},
			{
				Config: configVolDettached,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					testAccCheckLibvirtVolumeDoesNotExists("libvirt_volume."+randomVolumeName, &volume),
					testAccCheckLibvirtVolumeDoesNotExists("libvirt_volume."+randomVolumeName2, &volume),
				),
			},
		},
	})
}

// tests that disk driver is set correctly for the volume format
func TestAccLibvirtDomain_VolumeDriver(t *testing.T) {
	var domain libvirt.Domain
	var volumeRaw libvirt.StorageVol
	var volumeQCOW2 libvirt.StorageVol
	randomVolumeQCOW2 := acctest.RandString(10)
	randomVolumeRaw := acctest.RandString(9)
	randomDomainName := acctest.RandString(10)

	var config = fmt.Sprintf(`
	resource "libvirt_volume" "%s" {
		name = "%s"
        format = "raw"
	}

	resource "libvirt_volume" "%s" {
		name = "%s"
        format = "qcow2"
	}

	resource "libvirt_domain" "%s" {
		name = "%s"
		disk {
			volume_id = "${libvirt_volume.%s.id}"
		}

		disk {
			volume_id = "${libvirt_volume.%s.id}"
		}
	}`, randomVolumeRaw, randomVolumeRaw, randomVolumeQCOW2, randomVolumeQCOW2, randomDomainName, randomDomainName, randomVolumeRaw, randomVolumeQCOW2)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeRaw, &volumeRaw),
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeQCOW2, &volumeQCOW2),
					// Check that each disk has the appropriate driver
					testAccCheckLibvirtDomainDescription(&domain, func(domainDef libvirtxml.Domain) error {
						if domainDef.Devices.Disks[0].Driver.Type != "raw" {
							return fmt.Errorf("Expected disk to have RAW driver")
						}
						if domainDef.Devices.Disks[1].Driver.Type != "qcow2" {
							return fmt.Errorf("Expected disk to have QCOW2 driver")
						}
						return nil
					})),
			},
		},
	})
}

func TestAccLibvirtDomain_ScsiDisk(t *testing.T) {
	var domain libvirt.Domain
	randomVolumeName := acctest.RandString(10)
	randomDomainName := acctest.RandString(10)
	var configScsi = fmt.Sprintf(`
	resource "libvirt_volume" "%s" {
		name = "%s"
	}

	resource "libvirt_domain" "%s" {
		name = "%s"
		disk {
			volume_id = "${libvirt_volume.%s.id}"
			scsi      = "true"
			wwn       = "000000123456789a"
		}
	}`, randomVolumeName, randomVolumeName, randomDomainName, randomDomainName, randomVolumeName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: configScsi,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					testAccCheckLibvirtScsiDisk("000000123456789a", &domain),
				),
			},
		},
	})

}

func TestAccLibvirtDomain_URLDisk(t *testing.T) {
	var domain libvirt.Domain
	randomDomainName := acctest.RandString(10)
	fws := fileWebServer{}
	if err := fws.Start(); err != nil {
		t.Fatal(err)
	}
	defer fws.Stop()

	isoPath, err := filepath.Abs("testdata/tcl.iso")
	if err != nil {
		t.Fatal(err)
	}

	u, err := fws.AddFile(isoPath)
	if err != nil {
		t.Error(err)
	}

	url, err := url.Parse(u)
	if err != nil {
		t.Error(err)
	}

	var configURL = fmt.Sprintf(`
	resource "libvirt_domain" "%s" {
		name = "%s"
		disk {
			url = "%s"
		}
	}`, randomDomainName, randomDomainName, url.String())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: configURL,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					testAccCheckLibvirtURLDisk(url, &domain),
				),
			},
		},
	})

}

func TestAccLibvirtDomain_KernelInitrdCmdline(t *testing.T) {
	var domain libvirt.Domain
	var kernel libvirt.StorageVol
	var initrd libvirt.StorageVol
	randomDomainName := acctest.RandString(10)

	var config = fmt.Sprintf(`
	resource "libvirt_volume" "kernel" {
		source = "testdata/tetris.elf"
		name   = "kernel"
		pool   = "default"
		format = "raw"
	}

	resource "libvirt_volume" "initrd" {
		source = "testdata/initrd.img"
		name   = "initrd"
		pool   = "default"
		format = "raw"
	}

	resource "libvirt_domain" "%s" {
		name   = "terraform-test-domain"
		kernel = "${libvirt_volume.kernel.id}"
		initrd = "${libvirt_volume.initrd.id}"
		cmdline {
			foo = 1
			bar = "bye"
		}
		cmdline {
			foo = 2
		}
	}`, randomDomainName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtVolumeExists("libvirt_volume.kernel", &kernel),
					testAccCheckLibvirtVolumeExists("libvirt_volume.initrd", &initrd),
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					testAccCheckLibvirtDomainKernelInitrdCmdline(&domain, &kernel, &initrd),
				),
			},
		},
	})

}

func TestAccLibvirtDomain_NetworkInterface(t *testing.T) {
	var domain libvirt.Domain
	randomDomainName := acctest.RandString(10)
	randomNetworkName := acctest.RandString(10)

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	var config = fmt.Sprintf(`
	resource "libvirt_network" "%s" {
		name      = "%s"
		addresses = ["10.17.3.0/24"]
	}

	resource "libvirt_domain" "%s" {
		name              = "%s"
		network_interface = {
			network_name = "default"
		}
		network_interface = {
			network_name   = "default"
			mac            = "52:54:00:A9:F5:17"
			wait_for_lease = true
		}
		network_interface = {
			model = "e1000"
		}
		disk {
			file = "%s/testdata/tcl.iso"
		}
	}`, randomNetworkName, randomNetworkName, randomDomainName, randomDomainName, currentDir)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "network_interface.0.network_name", "default"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "network_interface.1.mac", "52:54:00:A9:F5:17"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "network_interface.2.model", "e1000"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_CheckDHCPEntries(t *testing.T) {
	var domain libvirt.Domain
	var network libvirt.Network
	randomDomainName := acctest.RandString(10)
	randomNetworkName := acctest.RandString(10)

	var configWithDomain = fmt.Sprintf(`
	    resource "libvirt_network" "%s" {
		    name = "%s"
		    mode = "nat"
		    domain = "%s"
		    addresses = ["192.0.0.0/24"]
	    }

            resource "libvirt_domain" "%s" {
                    name = "%s"
                    network_interface {
                            network_id = "${libvirt_network.%s.id}"
                            hostname = "terraform-test"
                            addresses = ["192.0.0.2"]
                    }
            }`, randomNetworkName, randomNetworkName, randomDomainName, randomDomainName, randomDomainName, randomNetworkName)

	var configWithoutDomain = fmt.Sprintf(`
	    resource "libvirt_network" "%s" {
		    name = "%s"
		    mode = "nat"
		    domain = "%s"
		    addresses = ["192.0.0.0/24"]
	    }`, randomNetworkName, randomNetworkName, randomDomainName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config:             configWithDomain,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					testAccCheckLibvirtNetworkExists("libvirt_network."+randomNetworkName, &network),
				),
			},
			{
				Config: configWithoutDomain,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDestroyLeavesIPs("libvirt_network."+randomNetworkName,
						"192.0.0.2", &network),
				),
			},
			{
				Config:             configWithDomain,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_Graphics(t *testing.T) {
	var domain libvirt.Domain

	randomDomainName := acctest.RandString(10)
	randomVolumeName := acctest.RandString(10)
	var config = fmt.Sprintf(`
	resource "libvirt_volume" "%s" {
		name = "%s"
	}

	resource "libvirt_domain" "%s" {
		name = "%s"
		graphics {
			type        = "spice"
			autoport    = "true"
			listen_type = "none"
		}
	}`, randomVolumeName, randomVolumeName, randomDomainName, randomDomainName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "graphics.0.type", "spice"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "graphics.0.autoport", "true"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "graphics.0.listen_type", "none"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_IgnitionObject(t *testing.T) {
	var domain libvirt.Domain
	var volume libvirt.StorageVol
	randomDomainName := acctest.RandString(10)
	randomIgnitionName := acctest.RandString(10)
	var config = fmt.Sprintf(`
	data "ignition_systemd_unit" "acceptance-test-systemd" {
		name    = "example.service"
		content = "[Service]\nType=oneshot\nExecStart=/usr/bin/echo Hello World\n\n[Install]\nWantedBy=multi-user.target"
	}

	data "ignition_config" "acceptance-test-config" {
		systemd = [
		"${data.ignition_systemd_unit.acceptance-test-systemd.id}",
		]
	}

	resource "libvirt_ignition" "%s" {
		name    = "ignition"
		content = "${data.ignition_config.acceptance-test-config.rendered}"
	}

	resource "libvirt_domain" "%s" {
		name            = "terraform-test-domain"
		coreos_ignition = "${libvirt_ignition.%s.id}"
	}
	`, randomIgnitionName, randomDomainName, randomIgnitionName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					testAccCheckIgnitionVolumeExists("libvirt_ignition."+randomIgnitionName, &volume),
					testAccCheckIgnitionXML(&domain, &volume),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_Cpu(t *testing.T) {
	var domain libvirt.Domain
	randomDomainName := acctest.RandString(10)

	var config = fmt.Sprintf(`
	resource "libvirt_domain" "%s" {
		name = "%s"
		cpu {
			mode = "custom"
		}
	}`, randomDomainName, randomDomainName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "cpu.mode", "custom"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_Autostart(t *testing.T) {
	var domain libvirt.Domain
	randomDomainName := acctest.RandString(10)
	var autostartTrue = fmt.Sprintf(`
	resource "libvirt_domain" "%s" {
		name      = "%s"
		autostart = true
	}`, randomDomainName, randomDomainName)

	var autostartFalse = fmt.Sprintf(`
	resource "libvirt_domain" "%s" {
		name      = "%s"
		autostart = false
	}`, randomDomainName, randomDomainName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: autostartTrue,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					resource.TestCheckResourceAttr("libvirt_domain."+randomDomainName, "autostart", "true"),
				),
			},
			{
				Config: autostartFalse,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					resource.TestCheckResourceAttr("libvirt_domain."+randomDomainName, "autostart", "false"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_Filesystems(t *testing.T) {
	var domain libvirt.Domain
	randomDomainName := acctest.RandString(10)

	var config = fmt.Sprintf(`
	resource "libvirt_domain" "%s" {
		name = "%s"
		filesystem {
			source   = "/tmp"
			target   = "tmp"
			readonly = false
		}
		filesystem {
			source   = "/proc"
			target   = "proc"
			readonly = true
		}
	}`, randomDomainName, randomDomainName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "filesystem.0.source", "/tmp"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "filesystem.0.target", "tmp"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "filesystem.0.readonly", "false"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "filesystem.1.source", "/proc"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "filesystem.1.target", "proc"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "filesystem.1.readonly", "true"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_Consoles(t *testing.T) {
	var domain libvirt.Domain
	randomDomainName := acctest.RandString(10)
	var config = fmt.Sprintf(`
	resource "libvirt_domain" "%s" {
		name = "%s"
		console {
			type        = "pty"
			target_port = "0"
			source_path = "/dev/pts/1"
		}
		console {
			type        = "pty"
			target_port = "0"
			target_type = "virtio"
			source_path = "/dev/pts/2"
		}
	}`, randomDomainName, randomDomainName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "console.0.type", "pty"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "console.0.target_port", "0"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "console.0.source_path", "/dev/pts/1"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "console.1.type", "pty"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "console.1.target_port", "0"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "console.1.target_type", "virtio"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "console.1.source_path", "/dev/pts/2"),
				),
			},
		},
	})
}

func testAccCheckLibvirtDomainExists(name string, domain *libvirt.Domain) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		rs, err := getResourceFromTerraformState(name, state)
		if err != nil {
			return err
		}

		virConn := testAccProvider.Meta().(*Client).libvirt

		retrieveDomain, err := virConn.LookupDomainByUUIDString(rs.Primary.ID)

		if err != nil {
			return err
		}

		log.Printf("The ID is %s", rs.Primary.ID)

		realID, err := retrieveDomain.GetUUIDString()
		if err != nil {
			return err
		}

		if realID != rs.Primary.ID {
			return fmt.Errorf("Libvirt domain not found")
		}

		*domain = *retrieveDomain

		return nil
	}
}

func testAccCheckIgnitionXML(domain *libvirt.Domain, volume *libvirt.StorageVol) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		domainDef, err := getXMLDomainDefFromLibvirt(domain)
		if err != nil {
			return fmt.Errorf("Error retrieving libvirt domain XML description from existing domain: %s", err)
		}

		ignitionKey, err := volume.GetKey()
		if err != nil {
			return err
		}
		ignStr := fmt.Sprintf("name=opt/com.coreos/config,file=%s", ignitionKey)

		cmdLine := domainDef.QEMUCommandline.Args
		for i, cmd := range cmdLine {
			if i == 1 && cmd.Value != ignStr {
				return fmt.Errorf("libvirt domain fw_cfg XML is incorrect %s", cmd.Value)
			}
		}
		return nil
	}
}

func testAccCheckLibvirtScsiDisk(n string, domain *libvirt.Domain) resource.TestCheckFunc {
	return testAccCheckLibvirtDomainDescription(domain, func(domainDef libvirtxml.Domain) error {
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
	})
}

func testAccCheckLibvirtDomainDescription(domain *libvirt.Domain, checkFunc func(libvirtxml.Domain) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		domainDef, err := getXMLDomainDefFromLibvirt(domain)
		if err != nil {
			return fmt.Errorf("Error retrieving libvirt domain XML description from existing domain: %s", err)
		}
		return checkFunc(domainDef)
	}
}

func testAccCheckLibvirtURLDisk(u *url.URL, domain *libvirt.Domain) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		domainDef, err := getXMLDomainDefFromLibvirt(domain)
		if err != nil {
			return fmt.Errorf("Error getting libvirt XML defintion from existing libvirt domain: %s", err)
		}

		disks := domainDef.Devices.Disks
		for _, disk := range disks {
			if disk.Source.Network == nil {
				return fmt.Errorf("Disk type is not network")
			}
			if disk.Source.Network.Protocol != u.Scheme {
				return fmt.Errorf("Disk protocol is not %s", u.Scheme)
			}
			if disk.Source.Network.Name != u.Path {
				return fmt.Errorf("Disk name is not %s", u.Path)
			}
			if len(disk.Source.Network.Hosts) < 1 {
				return fmt.Errorf("Disk has no hosts defined")
			}
			if disk.Source.Network.Hosts[0].Name != u.Hostname() {
				return fmt.Errorf("Disk hostname is not %s", u.Hostname())
			}
		}
		return nil
	}
}

func testAccCheckLibvirtDestroyLeavesIPs(name string, ip string, network *libvirt.Network) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		virConn := testAccProvider.Meta().(*Client).libvirt
		networkDef, err := getNetworkDef(state, name, *virConn)
		if err != nil {
			return err
		}
		for _, ips := range networkDef.IPs {
			for _, dhcpHost := range ips.DHCP.Hosts {
				if dhcpHost.IP == ip {
					return nil
				}
			}
		}
		return fmt.Errorf("Hostname with ip '%s' does not have a dhcp entry in network", ip)
	}
}

func testAccCheckLibvirtDomainKernelInitrdCmdline(domain *libvirt.Domain, kernel *libvirt.StorageVol, initrd *libvirt.StorageVol) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		domainDef, err := getXMLDomainDefFromLibvirt(domain)
		if err != nil {
			return fmt.Errorf("Error retrieving libvirt domain XML description from existing domain: %s", err)
		}

		key, err := kernel.GetKey()
		if err != nil {
			return fmt.Errorf("Can't get kernel volume id")
		}
		if domainDef.OS.Kernel != key {
			return fmt.Errorf("Kernel is not set correctly: '%s' vs '%s'", domainDef.OS.Kernel, key)
		}

		key, err = initrd.GetKey()
		if err != nil {
			return fmt.Errorf("Can't get initrd volume id")
		}
		if domainDef.OS.Initrd != key {
			return fmt.Errorf("Initrd is not set correctly: '%s' vs '%s'", domainDef.OS.Initrd, key)
		}
		if domainDef.OS.Cmdline != "bar=bye foo=1 foo=2" {
			return fmt.Errorf("Kernel args not set correctly: '%s'", domainDef.OS.Cmdline)
		}
		return nil
	}
}

func createNvramFile() (string, error) {
	// size of an accepted, valid, nvram backing store
	NVRAMDummyBuffer := make([]byte, 131072)
	file, err := ioutil.TempFile("/tmp", "nvram")
	if err != nil {
		return "", err
	}
	file.Chmod(0777)
	_, err = file.Write(NVRAMDummyBuffer)
	if err != nil {
		return "", err
	}
	if file.Close() != nil {
		return "", err
	}
	return file.Name(), nil
}

func TestAccLibvirtDomainFirmware(t *testing.T) {
	NVRAMPath, err := createNvramFile()
	if err != nil {
		t.Fatal(err)
	}

	firmware := fmt.Sprintf("/usr/share/qemu/ovmf-x86_64-code.bin")
	if _, err := os.Stat(firmware); os.IsNotExist(err) {
		firmware = "/usr/share/ovmf/OVMF.fd"
		if _, err := os.Stat(firmware); os.IsNotExist(err) {
			t.Skipf("Can't test domain custom firmware: OVMF firmware not found: %s", err)
		}
	}

	template := fmt.Sprintf("/usr/share/qemu/ovmf-x86_64-vars.bin")
	if _, err := os.Stat(template); os.IsNotExist(err) {
		template = "/usr/share/qemu/OVMF.fd"
		if _, err := os.Stat(template); os.IsNotExist(err) {
			t.Skipf("Can't test domain custom firmware template: OVMF template not found: %s", err)
		}
	}

	t.Run("No Template", func(t *testing.T) {
		subtestAccLibvirtDomainFirmwareNoTemplate(t, NVRAMPath, firmware)
	})
	t.Run("With Template", func(t *testing.T) {
		subtestAccLibvirtDomainFirmwareTemplate(t, NVRAMPath, firmware, template)
	})
}

func subtestAccLibvirtDomainFirmwareNoTemplate(t *testing.T, NVRAMPath string, firmware string) {
	var domain libvirt.Domain
	randomDomainName := acctest.RandString(10)
	var config = fmt.Sprintf(`
	resource "libvirt_domain" "%s" {
		name     = "terraform-test-firmware-no-template"
		firmware = "%s"
		nvram {
			file = "%s"
		}
	}`, randomDomainName, firmware, NVRAMPath)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "name", "terraform-test-firmware-no-template"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "nvram.0.file", NVRAMPath),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "firmware", firmware),
				),
			},
		},
	})
}

func subtestAccLibvirtDomainFirmwareTemplate(t *testing.T, NVRAMPath string, firmware string, template string) {
	NVRAMPath, err := createNvramFile()
	if err != nil {
		t.Fatal(err)
	}
	randomDomainName := acctest.RandString(10)
	var domain libvirt.Domain
	var config = fmt.Sprintf(`
	resource "libvirt_domain" "%s" {
		name     = "terraform-test-firmware-with-template"
		firmware = "%s"
		nvram {
			file     = "%s"
			template = "%s"
		}
	}`, randomDomainName, firmware, NVRAMPath, template)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "name", "terraform-test-firmware-with-template"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "nvram.0.file", NVRAMPath),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "nvram.0.template", template),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "firmware", firmware),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_MachineType(t *testing.T) {
	var domain libvirt.Domain
	randomDomainName := acctest.RandString(10)
	// Using machine type of pc as this is earliest QEMU target
	// and so most likely to be available
	var config = fmt.Sprintf(`
	resource "libvirt_domain" "%s" {
		name    = "%s"
		machine = "pc"
	}`, randomDomainName, randomDomainName)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					resource.TestCheckResourceAttr("libvirt_domain."+randomDomainName, "machine", "pc"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_ArchType(t *testing.T) {
	var domain libvirt.Domain
	randomDomainName := acctest.RandString(10)
	// Using i686 as architecture in case anyone running tests on an i686 only host
	var config = fmt.Sprintf(`
	resource "libvirt_domain" "%s" {
		name  = "%s"
		arch  = "i686"
	}`, randomDomainName, randomDomainName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					resource.TestCheckResourceAttr("libvirt_domain."+randomDomainName, "arch", "i686"),
				),
			},
		},
	})
}

func testAccCheckLibvirtNetworkExists(name string, network *libvirt.Network) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		rs, err := getResourceFromTerraformState(name, state)
		if err != nil {
			return err
		}

		virConn := testAccProvider.Meta().(*Client).libvirt

		retrieveNetwork, err := virConn.LookupNetworkByUUIDString(rs.Primary.ID)

		if err != nil {
			return err
		}

		log.Printf("The ID is %s", rs.Primary.ID)

		realID, err := retrieveNetwork.GetUUIDString()
		if err != nil {
			return err
		}

		if realID != rs.Primary.ID {
			return fmt.Errorf("Libvirt network not found")
		}

		network = retrieveNetwork

		return nil
	}
}

// we want to destroy (shutdown volume after creation)
func TestAccLibvirtDomain_ShutoffDomain(t *testing.T) {
	var domain libvirt.Domain
	var volume libvirt.StorageVol
	randomDomainName := acctest.RandString(10)
	randomVolumeName := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
			    resource "libvirt_volume" "%s" {
			    	name = "%s"
			    }
			    resource "libvirt_domain" "%s" {
			    	name = "%s"
					running = false
			    	disk {
			    		volume_id = "${libvirt_volume.%s.id}"
			    		}
			    }`, randomVolumeName, randomVolumeName, randomDomainName, randomDomainName, randomVolumeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					testAccCheckLibvirtVolumeExists("libvirt_volume."+randomVolumeName, &volume),
					testAccCheckLibvirtDomainStateEqual("libvirt_domain."+randomDomainName, &domain, "shutoff"),
				),
			},
		},
	})
}
func TestAccLibvirtDomain_ShutoffMultiDomainsRunning(t *testing.T) {
	var domain libvirt.Domain
	var domain2 libvirt.Domain
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
   				resource "libvirt_domain" "domainoff" {
					name = "domainfalse"
					vcpu = 1
					running = false
				}
				resource "libvirt_domain" "domainok" {
					name = "domaintrue"
					vcpu = 1
					running = true
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainStateEqual("libvirt_domain.domainoff", &domain, "shutoff"),
					testAccCheckLibvirtDomainStateEqual("libvirt_domain.domainok", &domain2, "running"),
				),
			},
		},
	})
}

func testAccCheckLibvirtDomainStateEqual(name string, domain *libvirt.Domain, exptectedState string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, err := getResourceFromTerraformState(name, s)
		if err != nil {
			return err
		}

		virConn := testAccProvider.Meta().(*Client).libvirt

		retrieveDomain, err := virConn.LookupDomainByUUIDString(rs.Primary.ID)

		if err != nil {
			return err
		}

		realID, err := retrieveDomain.GetUUIDString()
		if err != nil {
			return err
		}

		if realID != rs.Primary.ID {
			return fmt.Errorf("Libvirt domain not found")
		}

		*domain = *retrieveDomain
		state, err := domainGetState(*domain)
		if err != nil {
			return fmt.Errorf("could not get domain state: %s", err)
		}

		if state != exptectedState {
			return fmt.Errorf("Domain state should be == %s, but is %s", exptectedState, state)
		}

		return nil
	}
}

func TestAccLibvirtDomain_Import(t *testing.T) {
	var domain libvirt.Domain
	randomDomainName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "libvirt_domain" "%s" {
					name   = "%s"
					memory = 384
					vcpu   = 2
				}`, randomDomainName, randomDomainName),
			},
			{
				ResourceName: "libvirt_domain." + randomDomainName,
				ImportState:  true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.%s-2", &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain.%s-2", "name", randomDomainName),
					resource.TestCheckResourceAttr(
						"libvirt_domain.%s-2", "memory", "384"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.%s-2", "vcpu", "2"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_XSLT_UnsupportedAttribute(t *testing.T) {
	var domain libvirt.Domain
	randomDomainName := acctest.RandString(10)
	randomNetworkName := acctest.RandString(10)

	var config = fmt.Sprintf(`
	resource "libvirt_network" "%s" {
	  name      = "%s"
  	  addresses = ["10.17.3.0/24"]
    }

	resource "libvirt_domain" "%s" {
	  name = "%s"
	  network_interface = {
	    network_name = "default"
	  }
      xml {
        xslt = <<EOF
		<?xml version="1.0" ?>
		<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
		  <xsl:output omit-xml-declaration="yes" indent="yes"/>
		  <xsl:template match="node()|@*">
			 <xsl:copy>
			   <xsl:apply-templates select="node()|@*"/>
			 </xsl:copy>
		  </xsl:template>

		  <xsl:template match="/domain/devices/interface[@type='network']/model/@type">
			<xsl:attribute name="type">
			  <xsl:value-of select="'e1000'"/>
			</xsl:attribute>
		  </xsl:template>

		</xsl:stylesheet>
EOF
      }
	}`, randomNetworkName, randomNetworkName, randomDomainName, randomDomainName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					testAccCheckLibvirtDomainDescription(&domain, func(domainDef libvirtxml.Domain) error {
						if domainDef.Devices.Interfaces[0].Model.Type != "e1000" {
							return fmt.Errorf("Expecting XSLT to tranform network model to e1000")
						}
						return nil
					}),
				),
			},
		},
	})
}

// If using XSLT to transform a supported attribute by the terraform
// provider schema, the provider will try to change it back to the
// known state.
// Therefore we explicitly advise against using it with existing
// schema attributes
func TestAccLibvirtDomain_XSLT_SupportedAttribute(t *testing.T) {
	var domain libvirt.Domain
	randomDomainName := acctest.RandString(10)
	randomNetworkName := acctest.RandString(10)

	var config = fmt.Sprintf(`
	resource "libvirt_network" "%s" {
	  name      = "%s"
  	  addresses = ["10.17.3.0/24"]
    }

	resource "libvirt_domain" "%s" {
	  name = "%s"
	  network_interface = {
	    network_name = "default"
	  }
      xml {
        xslt = <<EOF
        <?xml version="1.0" ?>
          <xsl:stylesheet version="1.0"
                xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
            <xsl:output omit-xml-declaration="yes" indent="yes"/>
              <xsl:template match="node()|@*">
              <xsl:copy>
                <xsl:apply-templates select="node()|@*"/>
              </xsl:copy>
            </xsl:template>

            <xsl:template match="/domain/devices/interface[@type='network']/source/@network">
              <xsl:attribute name="network">
                <xsl:value-of select="'%s'"/>
              </xsl:attribute>
            </xsl:template>
          </xsl:stylesheet>
EOF
      }
	}`, randomNetworkName, randomNetworkName, randomDomainName, randomDomainName, randomNetworkName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config:             config,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "network_interface.0.network_name", randomNetworkName),
				),
			},
		},
	})
}

// changed whitespace in the xslt should create an empty plan
// as the supress diff function should take care of seeing they are equivalent
func TestAccLibvirtDomain_XSLT_Whitespace(t *testing.T) {
	var domain libvirt.Domain
	randomDomainName := acctest.RandString(10)
	randomNetworkName := acctest.RandString(10)

	var config = fmt.Sprintf(`
	resource "libvirt_network" "%s" {
	  name      = "%s"
  	  addresses = ["10.17.3.0/24"]
    }

	resource "libvirt_domain" "%s" {
	  name = "%s"
	  network_interface = {
	    network_name = "default"
	  }
      xml {
        xslt = <<EOF
		<?xml version="1.0" ?>
		<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
		  <xsl:output omit-xml-declaration="yes" indent="yes"/>
		  <xsl:template match="node()|@*">
			 <xsl:copy>
			   <xsl:apply-templates select="node()|@*"/>
			 </xsl:copy>
		  </xsl:template>

		  <xsl:template match="/domain/devices/interface[@type='network']/model/@type">
			<xsl:attribute name="type">
			  <xsl:value-of select="'e1000'"/>
			</xsl:attribute>
		  </xsl:template>

		</xsl:stylesheet>
EOF
      }
	}`, randomNetworkName, randomNetworkName, randomDomainName, randomDomainName)

	var configAfter = fmt.Sprintf(`
	resource "libvirt_network" "%s" {
	  name      = "%s"
  	  addresses = ["10.17.3.0/24"]
    }

	resource "libvirt_domain" "%s" {
	  name = "%s"
	  network_interface = {
	    network_name = "default"
	  }
      xml {
        xslt = <<EOF
		<?xml version="1.0" ?>
		<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
		  <xsl:output omit-xml-declaration="yes" indent="yes"/>
		  <xsl:template match="node()|@*">
			 <xsl:copy><xsl:apply-templates select="node()|@*"/></xsl:copy>
		  </xsl:template>
		  <xsl:template match="/domain/devices/interface[@type='network']/model/@type">
			<xsl:attribute name="type"><xsl:value-of select="'e1000'"/></xsl:attribute>
		  </xsl:template>
		</xsl:stylesheet>
EOF
      }
	}`, randomNetworkName, randomNetworkName, randomDomainName, randomDomainName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					testAccCheckLibvirtDomainDescription(&domain, func(domainDef libvirtxml.Domain) error {
						if domainDef.Devices.Interfaces[0].Model.Type != "e1000" {
							return fmt.Errorf("Expecting XSLT to tranform network model to e1000")
						}
						return nil
					}),
				),
			},
			{
				Config:             configAfter,
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain."+randomDomainName, &domain),
					testAccCheckLibvirtDomainDescription(&domain, func(domainDef libvirtxml.Domain) error {
						if domainDef.Devices.Interfaces[0].Model.Type != "e1000" {
							return fmt.Errorf("Expecting XSLT to tranform network model to e1000")
						}
						return nil
					}),
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
