package libvirt

import (
	"log"

	libvirt "github.com/dmacvicar/libvirt-go"
)

type Config struct {
	Uri string
}

type Client struct {
	libvirt *libvirt.VirConnection
}

func (c *Config) Client() (*Client, error) {
	conn, err := libvirt.NewVirConnection(c.Uri)
	if err != nil {
		return nil, err
	}

	client := &Client{
		libvirt: &conn,
	}

	log.Println("[INFO] Created libvirt client")

	return client, nil
}
