package libvirt

type LibVirtNetworkMock struct {
	GetXMLDescReply string
	GetXMLDescError error
}

func (n LibVirtNetworkMock) GetXMLDesc(flags uint32) (string, error) {
	return n.GetXMLDescReply, n.GetXMLDescError
}
