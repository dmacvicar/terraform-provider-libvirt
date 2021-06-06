package libvirt

type DomainMock struct {
	QemuAgentCommandResponse string
}

func (d DomainMock) QemuAgentCommand(command string, timeout int32, flags uint32) (string, error) {
	return d.QemuAgentCommandResponse, nil
}
