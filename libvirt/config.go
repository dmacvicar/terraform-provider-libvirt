package libvirt

import (
	"fmt"
	"log"
	"sync"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/dmacvicar/terraform-provider-libvirt/libvirt/helper/mutexkv"
	uri "github.com/dmacvicar/terraform-provider-libvirt/libvirt/uri"
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
	u, err := uri.Parse(c.URI)
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

	client := &Client{
		libvirt:     l,
		poolMutexKV: mutexkv.NewMutexKV(),
	}

	return client, nil
}
