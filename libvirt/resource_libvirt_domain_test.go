package libvirt

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	//"gopkg.in/alexzorin/libvirt-go.v2"
	libvirt "github.com/dmacvicar/libvirt-go"
	"strconv"
	"testing"
)

func TestAccLibvirtDomain_Basic(t *testing.T) {
	var domain libvirt.VirDomain

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckLibvirtDomainConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.terraform-acceptance-test-1", &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain.terraform-acceptance-test-1", "name", "terraform-test"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.terraform-acceptance-test-1", "memory", "512"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.terraform-acceptance-test-1", "vcpu", "1"),
				),
			},
		},
	})
}

func TestAccLibvirtDomain_Detailed(t *testing.T) {
	var domain libvirt.VirDomain

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLibvirtDomainDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckLibvirtDomainConfig_detailed,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLibvirtDomainExists("libvirt_domain.terraform-acceptance-test-2", &domain),
					resource.TestCheckResourceAttr(
						"libvirt_domain.terraform-acceptance-test-2", "name", "terraform-test"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.terraform-acceptance-test-2", "memory", "384"),
					resource.TestCheckResourceAttr(
						"libvirt_domain.terraform-acceptance-test-2", "vcpu", "2"),
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

		domainId, _ := strconv.Atoi(rs.Primary.ID)

		// Try to find the server
		_, err := virtConn.LookupDomainById(uint32(domainId))
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

		domainId, err := strconv.Atoi(rs.Primary.ID)

		if err != nil {
			return err
		}

		virConn := testAccProvider.Meta().(*Client).libvirt

		retrieveDomain, err := virConn.LookupDomainById(uint32(domainId))

		if err != nil {
			return err
		}

		fmt.Printf("The ID is %d", domainId)

		realId, err := retrieveDomain.GetID()
		if err != nil {
			return err
		}

		if realId != uint(domainId) {
			return fmt.Errorf("Libvirt domain not found")
		}

		*domain = retrieveDomain

		return nil
	}
}

const testAccCheckLibvirtDomainConfig_basic = `
resource "libvirt_domain" "terraform-acceptance-test-1" {
    name = "terraform-test"
}
`
const testAccCheckLibvirtDomainConfig_detailed = `
resource "libvirt_domain" "terraform-acceptance-test-2" {
    name = "terraform-test"
    memory = 384
    vcpu = 2
}
`
