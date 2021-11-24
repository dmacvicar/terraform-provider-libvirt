package libvirt

import (
	"fmt"
	"log"
	"math/rand"
	"sync"

	libvirt "github.com/digitalocean/go-libvirt"
	uri "github.com/dmacvicar/terraform-provider-libvirt/libvirt/uri"
	"github.com/hashicorp/terraform-plugin-sdk/helper/mutexkv"
)

// ConfigHost struct for the libvirt-provider
type ConfigHost struct {
	Region string
	AZ     string
	URI    string
}

// Config struct for the libvirt-provider
type Config struct {
	URI   string
	Hosts []ConfigHost
}

// Client libvirt
type Client struct {
	Region      string
	AZ          string
	libvirt     *libvirt.Libvirt
	poolMutexKV *mutexkv.MutexKV
	// define only one network at a time
	// https://gitlab.com/libvirt/libvirt/-/issues/78
	networkMutex sync.Mutex
}

// NewClient libvirt, generate libvirt client given URI
func NewClient(url, region, az string) (*Client, error) {

	u, err := uri.Parse(url)
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
		Region:      region,
		AZ:          az,
		libvirt:     l,
		poolMutexKV: mutexkv.NewMutexKV(),
	}

	return client, nil
}

// Clients libvirt
type Clients struct {
	Clients []*Client
}

// Clients libvirt, generate libvirt clients list
func (c *Config) Clients() (*Clients, error) {

	clients := &Clients{}

	if c.URI != "" {
		if _, ok := globalClientMap[c.URI]; ok {
			log.Printf("[DEBUG] Reusing client for uri: '%s'", c.URI)
		} else {
			log.Printf("[DEBUG] Configuring provider for '%s'", c.URI)
			client, err := NewClient(c.URI, "default", "default")
			if err != nil {
				return clients, err
			}
			clients.Clients = append(clients.Clients, client)
			globalClientMap[c.URI] = client
		}
	}

	for _, h := range c.Hosts {
		if _, ok := globalClientMap[h.URI]; ok {
			log.Printf("[DEBUG] Reusing client for uri: '%s'", h.URI)
		} else {
			log.Printf("[DEBUG] Configuring provider for '%s'", h.URI)
			client, err := NewClient(h.URI, h.Region, h.AZ)
			if err != nil {
				return clients, err
			}
			clients.Clients = append(clients.Clients, client)
			globalClientMap[c.URI] = client
		}
	}

	return clients, nil
}

// Find libvirt client, find the appropriate endpoint
func (clients *Clients) Find(region, az string) (*Client, error) {

	var client *Client
	var compatibleClients []*Client

	// filter all known clients which map the requested region and AZ
	for _, c := range clients.Clients {
		if c.Region == region && c.AZ == az {
			compatibleClients = append(compatibleClients, c)
		}
	}

	if len(compatibleClients) == 0 {
		return nil, fmt.Errorf("no provider seems to be avaiable for the requested Region/AZ tuple: (%s/%s)", region, az)
	}

	// if only one matches, go for it
	if len(compatibleClients) == 1 {
		client = compatibleClients[0]
	} else {
		// otherwise, pick and choose a random host matching requested criterias
		rd := rand.Intn(len(compatibleClients) - 1)
		if rd < 0 {
			rd = 0
		}
		client = compatibleClients[rd]
	}

	return client, nil
}
