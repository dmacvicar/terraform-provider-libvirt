package libvirt

import (
	"fmt"
	"log"
	"net"

	"net/url"
	"sync"
	"time"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/mutexkv"
)

const (
	defaultUnixSock = "/var/run/libvirt/libvirt-sock"
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

	u, err := url.Parse(c.URI)
	if err != nil {
		return nil, err
	}

	network := "unix"
	address := ""
	if u.Host != "" {
		network = "tcp"
		address = u.Host
		if u.Port() != "" {
			address = fmt.Sprintf("%s:%s", address, u.Port())
		}
	} else {
		q := u.Query()
		address = q.Get("socket")
		if address == "" {
			address = defaultUnixSock
		}
	}

	conn, err := net.DialTimeout(network, address, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to dial libvirt: %w", err)
	}
	log.Printf("[INFO] Set up libvirt transport: %v\n", conn)

	l := libvirt.New(conn)
	if err := l.ConnectToURI(libvirt.ConnectURI(c.URI)); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	v, err := l.ConnectGetLibVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve libvirt version: %w", err)
	}
	log.Printf("[INFO] libvirt2 client libvirt version: %v\n", v)

	client := &Client{
		libvirt:     l,
		poolMutexKV: mutexkv.NewMutexKV(),
	}

	return client, nil
}
