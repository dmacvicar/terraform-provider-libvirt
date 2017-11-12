package libvirt

import "github.com/libvirt/libvirt-go"

// Domain  Interface used to expose a libvirt.Domain
// Used to allow testing
type Domain interface {
	QemuAgentCommand(command string, timeout libvirt.DomainQemuAgentCommandTimeout, flags uint32) (string, error)
}

// LibVirtNetwork interface
type LibVirtNetwork interface {
	GetXMLDesc(flags libvirt.NetworkXMLFlags) (string, error)
}
