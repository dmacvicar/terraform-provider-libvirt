package libvirt

import "github.com/libvirt/libvirt-go"

type LibVirtNetworkMock struct {
	GetXMLDescReply string
	GetXMLDescError error
}

func (n LibVirtNetworkMock) GetXMLDesc(flags libvirt.NetworkXMLFlags) (string, error) {
	return n.GetXMLDescReply, n.GetXMLDescError
}
