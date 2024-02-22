package libvirt

import (
	"fmt"
	"log"
	"sync"
	"errors"
	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/dmacvicar/terraform-provider-libvirt/libvirt/helper/mutexkv"
	uri "github.com/dmacvicar/terraform-provider-libvirt/libvirt/uri"
)

// Client libvirt.
type Client struct {
	defaultURI  string
	connections map[string]*libvirt.Libvirt //libvirt     *libvirt.Libvirt
	poolMutexKV *mutexkv.MutexKV

	// define only one network at a time
	// https://gitlab.com/libvirt/libvirt/-/issues/78
	// note: this issue has been resolved in 2021 https://gitlab.com/libvirt/libvirt/-/commit/ea0cfa11
	networkMutex sync.Mutex
}

// obtain a connection for this provider, if target is empty, use the default provider connection
func (c *Client) Connection(target *string) (*libvirt.Libvirt , error) {

	URI := c.defaultURI
	if target != nil && *target != "" {
		URI = *target
	}

	if URI == "" {
		return nil, errors.New("either the provider-wide default `uri` or the resource block `host` must be specified")
	}

	log.Printf("[DEBUG] Configuring connection for target host '%s'", URI)

	if conn, ok := c.connections[URI] ; ok {
		log.Printf("[DEBUG] Found existing connection for target host: '%s'", URI)
		return conn, nil
	}

	u, err := uri.Parse(URI)
	if err != nil {
		return nil, err
	}

	l := libvirt.NewWithDialer(u)

	if err := l.ConnectToURI(libvirt.ConnectURI(u.RemoteName())); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	v, err := l.ConnectGetLibVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve libvirt version: %w", err)
	}
	log.Printf("[INFO] libvirt client libvirt version: %v\n", v)

	c.connections[URI] = l
	return l, nil
}

// Close all connections associated with this provider instance
func (c *Client) Close() {

	for uri, connection := range c.connections {
		log.Printf("[DEBUG] cleaning up connection for URI: %s", uri)
		// TODO: Confirm appropriate IsAlive() validation
		err := connection.ConnectClose()
		if err != nil {
			log.Printf("[ERROR] cannot close libvirt connection: %v", err)
		}
	}
}
