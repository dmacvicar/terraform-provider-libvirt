package uri

import (
	"net"
	"time"
)

const (
	defaultUnixSock = "/var/run/libvirt/libvirt-sock"
)

func (c *ConnectionURI) dialUNIX() (net.Conn, error) {

	q := c.Query()
	address := q.Get("socket")
	if address == "" {
		address = defaultUnixSock
	}

	return net.DialTimeout("unix", address, 2*time.Second)
}
