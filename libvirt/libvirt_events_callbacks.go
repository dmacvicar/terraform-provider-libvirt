package libvirt

import (
	"log"

	libvirt "github.com/libvirt/libvirt-go"
)

func rebootCallBack(c *libvirt.Connect, d *libvirt.Domain) {
	log.Printf("REBOOT EVENT!")
	// Domain rebooted so we assume installation was fine.

	//  once we know that domain rebooted we do following operations:
	// 1) Shutdown domain
	// 2) Remove kernel and initrd if they are present otherwise skip
	// 3) start the domain again ( in this way user can use the installed OS)
}
