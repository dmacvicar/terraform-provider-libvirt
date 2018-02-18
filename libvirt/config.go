package libvirt

import (
	"log"

	libvirt "github.com/libvirt/libvirt-go"
)

// Config struct for the libvirt-provider
type Config struct {
	URI string
}

// Client libvirt
type Client struct {
	libvirt *libvirt.Connect
}

// Client libvirt, generate libvirt client given URI
func (c *Config) Client() (*Client, error) {
	var err error

	if LibvirtClient == nil {
		LibvirtClient, err = libvirt.NewConnect(c.URI)
		if err != nil {
			return nil, err
		}
	}

	client := &Client{
		libvirt: LibvirtClient,
	}

	log.Println("[INFO] Created libvirt client")

	return client, nil
}
