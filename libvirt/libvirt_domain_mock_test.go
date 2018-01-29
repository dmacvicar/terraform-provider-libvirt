package libvirt

import "github.com/libvirt/libvirt-go"

type DomainMock struct {
	QemuAgentCommandResponse string
}

func (d DomainMock) QemuAgentCommand(command string, timeout libvirt.DomainQemuAgentCommandTimeout, flags uint32) (string, error) {
	return d.QemuAgentCommandResponse, nil
}
