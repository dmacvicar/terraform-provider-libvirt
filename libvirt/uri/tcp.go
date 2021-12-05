package uri

import (
	"fmt"
	"net"
)

const (
	defaultTCPPort = "16509"
)

func (u *ConnectionURI) dialTCP() (net.Conn, error) {
	port := u.Port()
	if port == "" {
		port = defaultTCPPort
	}

	return net.DialTimeout("tcp", fmt.Sprintf("%s:%s", u.Hostname(), port), dialTimeout)
}
