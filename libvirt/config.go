package libvirt

import (
	"fmt"
	"log"
	"sync"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/dmacvicar/terraform-provider-libvirt/libvirt/helper/mutexkv"
	uri "github.com/dmacvicar/terraform-provider-libvirt/libvirt/uri"
	"github.com/digitalocean/go-libvirt/socket/dialers"
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
	log.Printf("[DEBUG] Entering Client() for URI: %s", c.URI)

	// Parse the URI to extract connection details
	u, err := uri.Parse(c.URI)
	if err != nil {
		log.Printf("[ERROR] Failed to parse URI '%s': %v", c.URI, err)
		return nil, fmt.Errorf("failed to parse URI: %w", err)
	}
	log.Printf("[DEBUG] Parsed URI successfully: %s", u.String())

	// Extract remote hostname
	host := u.Hostname() // This ensures only the hostname is extracted
	if host == "" {
		return nil, fmt.Errorf("failed to extract valid host from URI: %s", c.URI)
	}
	log.Printf("[DEBUG] Using host for TLS connection: %s", host)

	// Create a new libvirt client with TLS dialer
	l := libvirt.NewWithDialer(dialers.NewTLS(host))

	// Establish the connection
	if err := l.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt at %s: %w", host, err)
	}
	log.Printf("[DEBUG] Successfully connected to libvirt on %s.", host)

	// Fetch and log the libvirt version
	v, err := l.Version()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve libvirt version: %w", err)
	}
	log.Printf("[INFO] libvirt client version: %v", v)

	// Create and return the Client instance
	client := &Client{
		libvirt:     l,
		poolMutexKV: mutexkv.NewMutexKV(),
	}

	return client, nil
}
