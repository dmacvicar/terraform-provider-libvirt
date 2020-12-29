package libvirt

import (
	"github.com/digitalocean/go-libvirt"
	libvirtc "github.com/libvirt/libvirt-go"
)

// Domain Interface used to expose a libvirtc.Domain
// Used to allow testing
type Domain interface {
	QemuAgentCommand(command string, timeout libvirtc.DomainQemuAgentCommandTimeout, flags uint32) (string, error)
}

// Network interface used to expose a libvirt.Network
type Network interface {
	GetXMLDesc(flags libvirt.NetworkXMLFlags) (string, error)
}
