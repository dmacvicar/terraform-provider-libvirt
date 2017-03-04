package libvirt

// Interface used to expose a libvirt.VirDomain
// Used to allow testing
type LibVirtDomain interface {
	QemuAgentCommand(cmd string, timeout int, flags uint32) string
}
