package libvirt

import "github.com/libvirt/libvirt-go"

type NetworkMock struct {
	GetXMLDescReply string
	GetXMLDescError error
}

func (n NetworkMock) GetXMLDesc(flags libvirt.NetworkXMLFlags) (string, error) {
	return n.GetXMLDescReply, n.GetXMLDescError
}
