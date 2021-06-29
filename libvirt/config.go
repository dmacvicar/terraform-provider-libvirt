package libvirt

import (
	"fmt"
	"log"
	"sync"

	libvirt "github.com/digitalocean/go-libvirt"
	uri "github.com/dmacvicar/terraform-provider-libvirt/libvirt/uri"
	"github.com/hashicorp/terraform-plugin-sdk/helper/mutexkv"
)

// Config struct for the libvirt-provider
type Config struct {
	URI string
}

// Client libvirt
type Client struct {
	libvirt     *libvirt.Libvirt
	poolMutexKV *mutexkv.MutexKV
	// define only one network at a time
	// https://gitlab.com/libvirt/libvirt/-/issues/78
	networkMutex sync.Mutex
}

// Client libvirt, generate libvirt client given URI
func (c *Config) Client() (*Client, error) {

	u, err := uri.Parse(c.URI)
	if err != nil {
		return nil, err
	}

	conn, err := u.DialTransport()
	if err != nil {
		return nil, fmt.Errorf("failed to dial libvirt: %w", err)
	}
	log.Printf("[INFO] Set up libvirt transport: %v\n", conn)

	l := libvirt.New(conn)
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
