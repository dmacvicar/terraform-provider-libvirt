package uri

import (
	"fmt"
	"net"
	"time"
)

const (
	defaultTCPPort = "16509"
)

func (c *ConnectionURI) dialTCP() (net.Conn, error) {
	port := c.Port()
	if port == "" {
		port = defaultTCPPort
	}

	return net.DialTimeout("tcp", fmt.Sprintf("%s:%s", c.Hostname(), port), 2*time.Second)
}
