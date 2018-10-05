package libvirt

import (
	"log"

	"github.com/hashicorp/terraform/helper/mutexkv"
	libvirt "github.com/libvirt/libvirt-go"
)

// Config struct for the libvirt-provider
type Config struct {
	URI string
}

// Client libvirt
type Client struct {
	libvirt     *libvirt.Connect
	poolMutexKV *mutexkv.MutexKV
}

func eventloop() {
	for {
		libvirt.EventRunDefaultImpl()
	}
}

// Client libvirt, generate libvirt client given URI
func (c *Config) Client() (*Client, error) {
	libvirt.EventRegisterDefaultImpl()
	libvirtClient, err := libvirt.NewConnect(c.URI)
	if err != nil {
		return nil, err
	}
	log.Println("[INFO] Created libvirt client")

	client := &Client{
		libvirt:     libvirtClient,
		poolMutexKV: mutexkv.NewMutexKV(),
	}

	go eventloop()

	return client, nil
}
