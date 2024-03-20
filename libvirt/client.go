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

type Connection struct {
	connection *libvirt.Libvirt
	poolMutexKV *mutexkv.MutexKV

	// define only one network at a time
	// https://gitlab.com/libvirt/libvirt/-/issues/78
	// note: this issue has been resolved in 2021 https://gitlab.com/libvirt/libvirt/-/commit/ea0cfa11
	networkMutex sync.Mutex
}

// Client libvirt.
type Client struct {
	defaultURI  string
	connections map[string]*Connection
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
		return conn.connection, nil
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

	conn := &Connection {
		connection: l,
			poolMutexKV: mutexkv.NewMutexKV(),
		}
	c.connections[URI] = conn
	return conn.connection, nil
}

// obtain a Lock for this client, don't create if not found since the network connection must exist anyways to get this far
func (c *Client) GetLock(target *string) (*mutexkv.MutexKV) {

	URI := c.defaultURI
	if target != nil && *target != "" {
		URI = *target
	}

	if URI == "" {
		// if we get here, somehow the calling code was able top get a live connection but the
		// connection string doesn't exist. This breaks a code invariant and is indicative of a bug.
		log.Printf("[WARN] unable to find mutex for '%v' - this is not expected behaviour", target)
		return nil
	}

	if conn, ok := c.connections[URI] ; ok {
		return conn.poolMutexKV
	}

	return nil
}

// obtain a Lock for this client, don't create if not found since the network connection must exist anyways to get this far
func (c *Client) GetMutex(target *string) (*sync.Mutex) {

	URI := c.defaultURI
	if target != nil && *target != "" {
		URI = *target
	}

	if URI == "" {
		// if we get here, somehow the calling code was able top get a live connection but the
		// connection string doesn't exist. This breaks a code invariant and is indicative of a bug.
		log.Printf("[WARN] unable to find mutex for '%v' - this is not expected behaviour", target)
		return nil
	}

	if conn, ok := c.connections[URI] ; ok {
		return &conn.networkMutex
	}

	return nil
}

// Close all connections associated with this provider instance
func (c *Client) Close() {

	for uri, c := range c.connections {
		log.Printf("[DEBUG] cleaning up connection for URI: %s", uri)
		// TODO: Confirm appropriate IsAlive() validation
		err := c.connection.ConnectClose()
		if err != nil {
			log.Printf("[ERROR] cannot close libvirt connection: %v", err)
		}
	}
}
