// Package dialers provides connection dialers for libvirt transports.
//
// This package uses the dialers from github.com/digitalocean/go-libvirt/socket/dialers
// for most transports (local, remote, SSH via Go library, TLS) and adds custom dialers
// like SSHCmd which uses the native SSH command-line tool.
package dialers

import (
	"net"
)

// Dialer is the interface for establishing connections to libvirt.
// This matches the interface used by go-libvirt dialers.
type Dialer interface {
	// Dial establishes a connection to libvirt and returns a net.Conn.
	Dial() (net.Conn, error)
}
