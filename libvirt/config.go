package libvirt

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/dmacvicar/terraform-provider-libvirt/libvirt/dialers"
	"github.com/dmacvicar/terraform-provider-libvirt/libvirt/helper/mutexkv"
	luri "github.com/dmacvicar/terraform-provider-libvirt/libvirt/uri"
)

// Config struct for the libvirt-provider.
type Config struct {
	URI string
}

// Client libvirt.
type Client struct {
	libvirt     *libvirt.Libvirt
	poolMutexKV *mutexkv.MutexKV
	// define only one network at a time
	// https://gitlab.com/libvirt/libvirt/-/issues/78
	networkMutex sync.Mutex
}

// Client libvirt, returns a libvirt client for a config.
func (c *Config) Client() (*Client, error) {
	uri, err := url.Parse(c.URI)
	if err != nil {
		return nil, err
	}

	var l *libvirt.Libvirt

	// Check if we should use the command-line SSH tool
	useSSHCmd := uri.Query().Has("use_ssh_cmd")

	// We only use our custom SSH command dialer if:
	// 1. The use_ssh_cmd parameter is present
	// 2. The URI scheme contains +ssh (like qemu+ssh)
	if useSSHCmd && strings.Contains(uri.Scheme, "+ssh") {
		// Remove the special param to not interfere with other URI processing
		q := uri.Query()
		q.Del("use_ssh_cmd")
		uri.RawQuery = q.Encode()

		// Create a dialer using the SSH command-line tool
		sshDialer := dialers.NewSSHCmdDialer(uri)

		// Use NewWithDialer to create a connection with the custom dialer
		l = libvirt.NewWithDialer(sshDialer)

		// Connect to the remote URI
		remoteName := libvirt.RemoteURI(uri)
		if err := l.ConnectToURI(remoteName); err != nil {
			return nil, fmt.Errorf("failed to connect to libvirt with SSH command: %w", err)
		}

		log.Printf("[INFO] Connected to libvirt using SSH command-line tool")
	} else if strings.Contains(uri.Scheme, "+ssh") {
		u, err := luri.Parse(c.URI)
		if err != nil {
			return nil, err
		}

		l := libvirt.NewWithDialer(u)

		if err := l.ConnectToURI(libvirt.ConnectURI(u.RemoteName())); err != nil {
			return nil, fmt.Errorf("failed to connect: %w", err)
		}
	} else {
		// Use the default connection method
		l, err = libvirt.ConnectToURI(uri)
		if err != nil {
			return nil, fmt.Errorf("failed to connect: %w", err)
		}
	}

	v, err := l.ConnectGetLibVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve libvirt version: %w", err)
	}
	log.Printf("[INFO] libvirt client libvirt version: %v\n", v)

	client := &Client{
		libvirt:     l,
		poolMutexKV: mutexkv.NewMutexKV(),
	}

	return client, nil
}
