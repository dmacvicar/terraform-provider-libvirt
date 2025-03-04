package libvirt

import (
	"fmt"
	"sync"
	"net/url"
	"strings"

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
	networkMutex sync.Mutex
}

// Client returns a libvirt client for a config.
func (c *Config) Client() (*Client, error) {
	var l *libvirt.Libvirt

	if strings.HasPrefix(c.URI, "qemu+ssh://") && strings.Contains(c.URI, "+ssh://") {
		u, err := uri.Parse(c.URI)
		if err != nil {
			return nil, fmt.Errorf("failed to parse SSH URI: %w", err)
		}
		l = libvirt.NewWithDialer(u)
		if err := l.ConnectToURI(libvirt.ConnectURI(u.RemoteName())); err != nil {
			return nil, fmt.Errorf("failed to connect: %w", err)
		}
	} else {
		parsedURL, err := url.Parse(c.URI)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URI: %w", err)
		}
		l, err = libvirt.ConnectToURI(parsedURL)
		if err != nil {
			return nil, fmt.Errorf("failed to connect: %w", err)
		}
	}

	client := &Client{
		libvirt:     l,
		poolMutexKV: mutexkv.NewMutexKV(),
	}

	return client, nil
}

