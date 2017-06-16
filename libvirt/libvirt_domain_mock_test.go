package libvirt

import "github.com/libvirt/libvirt-go"

type LibVirtDomainMock struct {
	QemuAgentCommandResponse string
}

func (d LibVirtDomainMock) QemuAgentCommand(command string, timeout libvirt.DomainQemuAgentCommandTimeout, flags uint32) (string, error) {
	return d.QemuAgentCommandResponse, nil
}
