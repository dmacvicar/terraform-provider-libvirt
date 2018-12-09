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
	libvirtClient, err := libvirt.NewConnect(c.URI)
	if err != nil {
		return nil, err
	}
	log.Println("[INFO] Created libvirt client")

	client := &Client{
		libvirt: libvirtClient,
	}

	return client, nil
}
