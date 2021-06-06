package libvirt

import (
	libvirt "github.com/digitalocean/go-libvirt"
)

// Domain Interface used to expose a libvirt.Domain
// Used to allow testing
type Domain interface {
	QemuAgentCommand(command string, timeout int32, flags uint32) (string, error)
}

// Network interface used to expose a libvirt.Network
type Network interface {
	GetXMLDesc(flags libvirt.NetworkXMLFlags) (string, error)
}
