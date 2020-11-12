package libvirt

import (
	"fmt"
	libvirt2 "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/mutexkv"
	libvirtc "github.com/libvirt/libvirt-go"
	"log"
	"net"
	"sync"
	"time"
)

// Config struct for the libvirt-provider
type Config struct {
	URI string
}

// Client libvirt
type Client struct {
	libvirt     *libvirt2.Libvirt
	libvirtc    *libvirtc.Connect
	poolMutexKV *mutexkv.MutexKV
	// define only one network at a time
	// https://gitlab.com/libvirt/libvirt/-/issues/78
	networkMutex sync.Mutex
}

// Client libvirt, generate libvirt client given URI
func (c *Config) Client() (*Client, error) {
	libvirtClient, err := libvirtc.NewConnect(c.URI)
	if err != nil {
		return nil, err
	}
	log.Println("[INFO] Created libvirt client")

	// TODO function url -> connection
	conn, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to dial libvirt: %w", err)
	}

	l := libvirt2.New(conn)
	if err := l.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	v, err := l.Version()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve libvirt version: %w", err)
	}
	log.Printf("[INFO] libvirt2 client libvirt version: %v\n", v)

	client := &Client{
		libvirt:     l,
		libvirtc:    libvirtClient,
		poolMutexKV: mutexkv.NewMutexKV(),
	}

	return client, nil
}
