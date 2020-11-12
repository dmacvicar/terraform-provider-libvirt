package libvirt

import (
	libvirtc "github.com/libvirt/libvirt-go"
)

// Domain Interface used to expose a libvirtc.Domain
// Used to allow testing
type Domain interface {
	QemuAgentCommand(command string, timeout libvirtc.DomainQemuAgentCommandTimeout, flags uint32) (string, error)
}

// Network interface used to expose a libvirtc.Network
type Network interface {
	GetXMLDesc(flags libvirtc.NetworkXMLFlags) (string, error)
}
