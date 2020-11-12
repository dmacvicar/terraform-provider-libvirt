package libvirt

import (
	libvirtc "github.com/libvirt/libvirt-go"
)

type DomainMock struct {
	QemuAgentCommandResponse string
}

func (d DomainMock) QemuAgentCommand(command string, timeout libvirtc.DomainQemuAgentCommandTimeout, flags uint32) (string, error) {
	return d.QemuAgentCommandResponse, nil
}
