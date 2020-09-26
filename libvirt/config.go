package libvirt

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/mutexkv"
	libvirt "github.com/libvirt/libvirt-go"
	"log"
	"sync"
)

// Config struct for the libvirt-provider
type Config struct {
	URI string
}

// Client libvirt
type Client struct {
	libvirt     *libvirt.Connect
	poolMutexKV *mutexkv.MutexKV
	// define only one network at a time
	// https://gitlab.com/libvirt/libvirt/-/issues/78
	networkMutex sync.Mutex
}

// Client libvirt, generate libvirt client given URI
func (c *Config) Client() (*Client, error) {
	libvirtClient, err := libvirt.NewConnect(c.URI)
	if err != nil {
		return nil, err
	}
	log.Println("[INFO] Created libvirt client")

	client := &Client{
		libvirt:     libvirtClient,
		poolMutexKV: mutexkv.NewMutexKV(),
	}

	return client, nil
}
