package libvirt

import (
	libvirtc "github.com/libvirt/libvirt-go"
)

type NetworkMock struct {
	GetXMLDescReply string
	GetXMLDescError error
}

func (n NetworkMock) GetXMLDesc(flags libvirtc.NetworkXMLFlags) (string, error) {
	return n.GetXMLDescReply, n.GetXMLDescError
}
