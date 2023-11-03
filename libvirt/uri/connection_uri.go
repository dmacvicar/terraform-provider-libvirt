// Deals with https://libvirt.org/uri.html
// go-libvirt needs a working transport to talk rpc to libvirt.
// This module deals with setting up those transports
package uri

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

const (
	dialTimeout = 2 * time.Second
)

type ConnectionURI struct {
	*url.URL
}

func Parse(uriStr string) (*ConnectionURI, error) {
	url, err := url.Parse(uriStr)
	if err != nil {
		return nil, err
	}
	return &ConnectionURI{URL: url}, nil
}

// According to https://libvirt.org/uri.html
// The name passed to the remote virConnectOpen function is formed by removing
// transport, hostname, port number, username and extra parameters from the remote URI
// unless the name option is specified.
func (u ConnectionURI) RemoteName() string {
	q := u.Query()
	if name := q.Get("name"); name != "" {
		return name
	}

	newURI := *u.URL
	newURI.Scheme = u.driver()
	newURI.User = nil
	newURI.Host = ""
	newURI.RawQuery = ""

	return newURI.String()
}

func (u *ConnectionURI) transport() string {
	parts := strings.Split(u.Scheme, "+")
	if len(parts) > 1 {
		return parts[1]
	}

	if u.Host != "" {
		return "tls"
	}
	return "unix"
}

func (u *ConnectionURI) driver() string {
	return strings.Split(u.Scheme, "+")[0]
}

// Dial implements go-libvirt Dialer interface, which is used
// to retrieve connections to talk via RPC to libvirtd.
//
// For example, a qemu+ssh:/// uri would return a SSH connection
// to localhost, and a new URI to qemu+unix:///system
// dials the transport for this connection URI.
func (u *ConnectionURI) Dial() (net.Conn, error) {
	t := u.transport()
	switch t {
	case "tcp":
		return u.dialTCP()
	case "tls":
		return u.dialTLS()
	case "unix":
		return u.dialUNIX()
	case "ssh":
		return u.dialSSH()
	}
	return nil, fmt.Errorf("transport '%s' not implemented", t)
}
