package libvirt

import "github.com/libvirt/libvirt-go"

// Interface used to expose a libvirt.Domain
// Used to allow testing
type LibVirtDomain interface {
	QemuAgentCommand(command string, timeout libvirt.DomainQemuAgentCommandTimeout, flags uint32) (string, error)
}

type LibVirtNetwork interface {
	GetXMLDesc(flags libvirt.NetworkXMLFlags) (string, error)
}
