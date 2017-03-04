package libvirt

type LibVirtDomainMock struct {
	QemuAgentCommandResponse string
}

func (d LibVirtDomainMock) QemuAgentCommand(cmd string, timeout int, flags uint32) string {
	return d.QemuAgentCommandResponse
}
