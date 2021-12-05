package uri

import (
	"net"
	"time"
)

const (
	defaultUnixSock = "/var/run/libvirt/libvirt-sock"
)

func (u *ConnectionURI) dialUNIX() (net.Conn, error) {

	q := u.Query()
	address := q.Get("socket")
	if address == "" {
		address = defaultUnixSock
	}

	return net.DialTimeout("unix", address, 2*time.Second)
}
