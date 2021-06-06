package libvirt

import (
	libvirt "github.com/digitalocean/go-libvirt"
)

type NetworkMock struct {
	GetXMLDescReply string
	GetXMLDescError error
}

func (n NetworkMock) GetXMLDesc(flags libvirt.NetworkXMLFlags) (string, error) {
	return n.GetXMLDescReply, n.GetXMLDescError
}
