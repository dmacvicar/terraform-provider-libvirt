package libvirt

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	libvirt "github.com/libvirt/libvirt-go"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
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
			{
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
		name   = "terraform-test"
		memory = 384
		vcpu   = 2
	}`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
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
			{
				Config: configVolAttached,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					testAccCheckLibvirtVolumeExists("libvirt_volume.acceptance-test-volume", &volume),
				),
			},
			{
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
			{
				Config: configVolAttached,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					testAccCheckLibvirtVolumeExists("libvirt_volume.acceptance-test-volume1", &volume),
					testAccCheckLibvirtVolumeExists("libvirt_volume.acceptance-test-volume2", &volume),
				),
			},
			{
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

// tests that disk driver is set correctly for the volume format
func TestAccLibvirtDomain_VolumeDriver(t *testing.T) {
	var domain libvirt.Domain
	var volumeRaw libvirt.StorageVol
	var volumeQCOW2 libvirt.StorageVol

	var config = fmt.Sprintf(`
	resource "libvirt_volume" "acceptance-test-volume-raw" {
		name = "terraform-test-raw"
        format = "raw"
	}

	resource "libvirt_volume" "acceptance-test-volume-qcow2" {
		name = "terraform-test-qcow2"
        format = "qcow2"
	}

	resource "libvirt_domain" "acceptance-test-domain" {
		name = "terraform-test-domain"
		disk {
			volume_id = "${libvirt_volume.acceptance-test-volume-raw.id}"
		}

		disk {
			volume_id = "${libvirt_volume.acceptance-test-volume-qcow2.id}"
		}
	}`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					testAccCheckLibvirtVolumeExists("libvirt_volume.acceptance-test-volume-raw", &volumeRaw),
					testAccCheckLibvirtVolumeExists("libvirt_volume.acceptance-test-volume-qcow2", &volumeQCOW2),
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
	var configScsi = fmt.Sprintf(`
	resource "libvirt_volume" "acceptance-test-volume1" {
		name = "terraform-test-vol1"
	}

	resource "libvirt_domain" "acceptance-test-domain" {
		name = "terraform-test-domain"
		disk {
			volume_id = "${libvirt_volume.acceptance-test-volume1.id}"
			scsi      = "true"
			wwn       = "000000123456789a"
		}
	}`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: configScsi,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					testAccCheckLibvirtScsiDisk("000000123456789a", &domain),
				),
			},
		},
	})

}

func TestAccLibvirtDomainURLDisk(t *testing.T) {
	var domain libvirt.Domain

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
	resource "libvirt_domain" "acceptance-test-domain" {
		name = "terraform-test-domain"
		disk {
			url = "%s"
		}
	}`, url.String())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: configURL,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					testAccCheckLibvirtURLDisk(url, &domain),
				),
			},
		},
	})

}

func TestAccLibvirtDomainKernelInitrdCmdline(t *testing.T) {
	var domain libvirt.Domain
	var kernel libvirt.StorageVol
	var initrd libvirt.StorageVol

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

	resource "libvirt_domain" "acceptance-test-domain" {
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
	}`)

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
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					testAccCheckLibvirtDomainKernelInitrdCmdline(&domain, &kernel, &initrd),
				),
			},
		},
	})

}

func TestAccLibvirtDomain_NetworkInterface(t *testing.T) {
	var domain libvirt.Domain

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	var config = fmt.Sprintf(`
	resource "libvirt_network" "acceptance-test-network" {
		name      = "terraform-test"
		addresses = ["10.17.3.0/24"]
	}

	resource "libvirt_domain" "acceptance-test-domain" {
		name              = "terraform-test"
		network_interface = {
			network_name = "default"
		}
		network_interface = {
			network_name   = "default"
			mac            = "52:54:00:A9:F5:17"
			wait_for_lease = 1
		}
		disk {
			file = "%s/testdata/tcl.iso"
		}
	}`, currentDir)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
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

func TestAccLibvirtDomain_CheckDHCPEntries(t *testing.T) {
	var domain libvirt.Domain
	var network libvirt.Network

	var configWithDomain = fmt.Sprintf(`
	    resource "libvirt_network" "acceptance-test-network" {
		    name = "acceptance-test-network"
		    mode = "nat"
		    domain = "acceptance-test-network-local"
		    addresses = ["192.0.0.0/24"]
	    }

            resource "libvirt_domain" "acceptance-test-domain" {
                    name = "terraform-test"
                    network_interface {
                            network_id = "${libvirt_network.acceptance-test-network.id}"
                            hostname = "terraform-test"
                            addresses = ["192.0.0.2"]
                    }
            }`)

	var configWithoutDomain = fmt.Sprintf(`
	    resource "libvirt_network" "acceptance-test-network" {
		    name = "acceptance-test-network"
		    mode = "nat"
		    domain = "acceptance-test-network-local"
		    addresses = ["192.0.0.0/24"]
	    }`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config:             configWithDomain,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					testAccCheckLibvirtNetworkExists("libvirt_network.acceptance-test-network", &network),
				),
			},
			resource.TestStep{
				Config: configWithoutDomain,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDestroyLeavesIPs("libvirt_network.acceptance-test-network",
						"192.0.0.2", &network),
				),
			},
			resource.TestStep{
				Config:             configWithDomain,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
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
			type        = "spice"
			autoport    = "true"
			listen_type = "none"
		}
	}`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "graphics.0.type", "spice"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "graphics.0.autoport", "true"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "graphics.0.listen_type", "none"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_IgnitionObject(t *testing.T) {
	var domain libvirt.Domain
	var volume libvirt.StorageVol

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

	resource "libvirt_ignition" "ignition" {
		name    = "ignition"
		content = "${data.ignition_config.acceptance-test-config.rendered}"
	}

	resource "libvirt_domain" "acceptance-test-domain" {
		name            = "terraform-test-domain"
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
			{
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

	var autostartTrue = fmt.Sprintf(`
	resource "libvirt_domain" "acceptance-test-domain" {
		name      = "terraform-test"
		autostart = true
	}`)

	var autostartFalse = fmt.Sprintf(`
	resource "libvirt_domain" "acceptance-test-domain" {
		name      = "terraform-test"
		autostart = false
	}`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: autostartTrue,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					resource.TestCheckResourceAttr("libvirt_domain.acceptance-test-domain", "autostart", "true"),
				),
			},
			{
				Config: autostartFalse,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					resource.TestCheckResourceAttr("libvirt_domain.acceptance-test-domain", "autostart", "false"),
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
			source   = "/tmp"
			target   = "tmp"
			readonly = false
		}
		filesystem {
			source   = "/proc"
			target   = "proc"
			readonly = true
		}
	}`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "filesystem.0.source", "/tmp"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "filesystem.0.target", "tmp"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "filesystem.0.readonly", "false"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "filesystem.1.source", "/proc"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "filesystem.1.target", "proc"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "filesystem.1.readonly", "true"),
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
	}`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
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
		xmlDesc, err := domain.GetXMLDesc(0)
		if err != nil {
			return fmt.Errorf("Error retrieving libvirt domain XML description: %s", err)
		}

		domainDef := newDomainDef()
		err = xml.Unmarshal([]byte(xmlDesc), &domainDef)
		if err != nil {
			return fmt.Errorf("Error reading libvirt domain XML description: %s", err)
		}

		return checkFunc(domainDef)
	}
}

func testAccCheckLibvirtURLDisk(u *url.URL, domain *libvirt.Domain) resource.TestCheckFunc {
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

func testAccCheckLibvirtDestroyLeavesIPs(n string, ip string, network *libvirt.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No libvirt network ID is set")
		}

		virConn := testAccProvider.Meta().(*Client).libvirt

		retrieveNetwork, err := virConn.LookupNetworkByUUIDString(rs.Primary.ID)

		if err != nil {
			return err
		}

		networkDef, err := newDefNetworkfromLibvirt(retrieveNetwork)

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
		xmlDesc, err := domain.GetXMLDesc(0)
		if err != nil {
			return fmt.Errorf("Error retrieving libvirt domain XML description: %s", err)
		}

		domainDef := newDomainDef()
		err = xml.Unmarshal([]byte(xmlDesc), &domainDef)
		if err != nil {
			return fmt.Errorf("Error reading libvirt domain XML description: %s", err)
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
	var config = fmt.Sprintf(`
	resource "libvirt_domain" "acceptance-test-domain" {
		name     = "terraform-test-firmware-no-template"
		firmware = "%s"
		nvram {
			file = "%s"
		}
	}`, firmware, NVRAMPath)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "name", "terraform-test-firmware-no-template"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "nvram.0.file", NVRAMPath),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "firmware", firmware),
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

	var domain libvirt.Domain
	var config = fmt.Sprintf(`
	resource "libvirt_domain" "acceptance-test-domain" {
		name     = "terraform-test-firmware-with-template"
		firmware = "%s"
		nvram {
			file     = "%s"
			template = "%s"
		}
	}`, firmware, NVRAMPath, template)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "name", "terraform-test-firmware-with-template"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "nvram.0.file", NVRAMPath),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "nvram.0.template", template),
					resource.TestCheckResourceAttr(
						"libvirt_domain.acceptance-test-domain", "firmware", firmware),
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
		name    = "terraform-test"
		machine = "pc"
	}`)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
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
		name  = "terraform-test"
		arch  = "i686"
	}`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					resource.TestCheckResourceAttr("libvirt_domain.acceptance-test-domain", "arch", "i686"),
				),
			},
		},
	})
}

func testAccCheckLibvirtNetworkExists(n string, network *libvirt.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No libvirt network ID is set")
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
func TestShutoffDomain(t *testing.T) {
	var domain libvirt.Domain
	var volume libvirt.StorageVol
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
			    resource "libvirt_volume" "acceptance-test-volume" {
			    	name = "terraform-test"
			    }
			    resource "libvirt_domain" "acceptance-test-domain" {
			    	name = "terraform-test"
					running = false
			    	disk {
			    		volume_id = "${libvirt_volume.acceptance-test-volume.id}"
			    		}
			    }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.acceptance-test-domain", &domain),
					testAccCheckLibvirtVolumeExists("libvirt_volume.acceptance-test-volume", &volume),
					testAccCheckLibvirtDomainStateEqual("libvirt_domain.acceptance-test-domain", &domain, "shutoff"),
				),
			},
		},
	})
}
func TestShutoffMultiDomainsRunning(t *testing.T) {
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

func testAccCheckLibvirtDomainStateEqual(n string, domain *libvirt.Domain, exptectedState string) resource.TestCheckFunc {
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
	var config = fmt.Sprintf(`
	resource "libvirt_domain" "acceptance-test-domain-2" {
		name   = "terraform-test"
		memory = 384
		vcpu   = 2
	}`)

	resourceName := "libvirt_domain.acceptance-test-domain-2"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: config,
			},
			resource.TestStep{
				ResourceName: resourceName,
				ImportState:  true,
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
