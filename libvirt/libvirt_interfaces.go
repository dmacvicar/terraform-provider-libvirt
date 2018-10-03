package libvirt

import "github.com/libvirt/libvirt-go"

// Network interface used to expose a libvirt.Network
type Network interface {
	GetXMLDesc(flags libvirt.NetworkXMLFlags) (string, error)
}
