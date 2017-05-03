package libvirt

type LibVirtNetworkMock struct {
	GetXMLDescReply     string
	GetXMLDescError     error
	UpdateXMLDescError  error
	UpdateXMLDescCalled bool
}

func (n *LibVirtNetworkMock) GetXMLDesc(flags uint32) (string, error) {
	return n.GetXMLDescReply, n.GetXMLDescError
}

func (n *LibVirtNetworkMock) UpdateXMLDesc(xmldesc string, command, section int) error {
	n.UpdateXMLDescCalled = true
	return n.UpdateXMLDescError
}
