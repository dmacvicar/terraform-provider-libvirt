package libvirt

import (
	"fmt"
	"syscall"
	"testing"
	"unsafe"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	libvirt "github.com/libvirt/libvirt-go"
)

func createPts() (string, error) {

	var ptsNumber int

	fd, err := syscall.Open("/dev/ptmx", int(syscall.O_RDWR), 0)
	if err != nil {
		fmt.Printf("Error creating pts %v", err)
	}

	_, _, ep := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TIOCGPTN), uintptr(unsafe.Pointer(&ptsNumber)))

	if ep != 0 {
		return "", syscall.Errno(ep)
	}

	return fmt.Sprintf("/dev/pts/%d", ptsNumber), nil

}

func TestAccLibvirtDomainConsoles(t *testing.T) {
	skipIfPrivilegedDisabled(t)

	var domain libvirt.Domain
	randomDomainName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	pts1, err := createPts()
	if err != nil {
		t.Errorf("error creating pts %v", err)
	}

	pts2, err := createPts()
	if err != nil {
		t.Errorf("error creating pts %v", err)
	}

	var config = fmt.Sprintf(`
	resource "libvirt_domain" "%s" {
		name = "%s"
		console {
			type        = "pty"
			target_port = "0"
			source_path = "%s"
		}
		console {
			type        = "pty"
			target_port = "0"
			target_type = "virtio"
			source_path = "%s"
		}
		console {
			type        = "tcp"
			target_port = "0"
			target_type = "virtio"
			source_host = "127.0.1.1"
			source_service = "cisco-sccp"
		}
	}`, randomDomainName, randomDomainName, pts1, pts2)

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
						"libvirt_domain."+randomDomainName, "console.0.source_path", pts1),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "console.1.type", "pty"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "console.1.target_port", "0"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "console.1.target_type", "virtio"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "console.1.source_path", pts2),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "console.2.type", "tcp"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "console.2.target_port", "0"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "console.2.target_type", "virtio"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "console.2.source_host", "127.0.1.1"),
					resource.TestCheckResourceAttr(
						"libvirt_domain."+randomDomainName, "console.2.source_service", "cisco-sccp"),
				),
			},
		},
	})
}
