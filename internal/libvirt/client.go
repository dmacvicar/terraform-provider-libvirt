// Package libvirt provides a client wrapper for libvirt connections.
package libvirt

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Client wraps the libvirt connection and provides helper methods
type Client struct {
	conn *libvirt.Libvirt
	uri  string
}

// NewClient creates a new libvirt client from a connection URI
func NewClient(ctx context.Context, uri string) (*Client, error) {
	tflog.Debug(ctx, "Creating new libvirt client", map[string]any{
		"uri": uri,
	})

	parsedURI, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid libvirt URI: %w", err)
	}

	// For now, we only support qemu:///system (Unix socket connection)
	// We'll add support for other transports later
	if parsedURI.Scheme != "qemu" {
		return nil, fmt.Errorf("unsupported URI scheme: %s (only qemu:// is currently supported)", parsedURI.Scheme)
	}

	if parsedURI.Host != "" {
		return nil, fmt.Errorf("remote connections not yet supported (URI contains host: %s)", parsedURI.Host)
	}

	// Determine the socket path based on the URI path
	socketPath := "/var/run/libvirt/libvirt-sock"
	if parsedURI.Path == "/session" {
		socketPath = fmt.Sprintf("/run/user/%d/libvirt/libvirt-sock", 1000) // TODO: Get actual UID
	}

	// Connect to libvirt via Unix socket
	conn, err := net.DialTimeout("unix", socketPath, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to dial libvirt socket %s: %w", socketPath, err)
	}

	// Create libvirt client
	//nolint:staticcheck // NewWithDialer is too complex for our simple use case
	l := libvirt.New(conn)
	if err := l.Connect(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}

	tflog.Info(ctx, "Successfully connected to libvirt", map[string]any{
		"uri":    uri,
		"socket": socketPath,
	})

	return &Client{
		conn: l,
		uri:  uri,
	}, nil
}

// Close closes the libvirt connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Disconnect()
	}
	return nil
}

// Libvirt returns the underlying go-libvirt client for direct API access
func (c *Client) Libvirt() *libvirt.Libvirt {
	return c.conn
}

// URI returns the connection URI
func (c *Client) URI() string {
	return c.uri
}

// Ping verifies the connection is still alive
func (c *Client) Ping(ctx context.Context) error {
	// ConnectGetLibVersion is a simple API call to verify connectivity
	_, err := c.conn.ConnectGetLibVersion()
	if err != nil {
		tflog.Error(ctx, "Libvirt connection ping failed", map[string]any{
			"error": err.Error(),
		})
		return fmt.Errorf("libvirt connection failed: %w", err)
	}
	return nil
}

// LookupDomainByUUID looks up a domain by its UUID string
func (c *Client) LookupDomainByUUID(uuidStr string) (libvirt.Domain, error) {
	uuid, err := parseUUID(uuidStr)
	if err != nil {
		return libvirt.Domain{}, err
	}

	// Look up the domain
	domain, err := c.conn.DomainLookupByUUID(uuid)
	if err != nil {
		return libvirt.Domain{}, fmt.Errorf("domain not found: %w", err)
	}

	return domain, nil
}

// LookupPoolByUUID looks up a storage pool by its UUID string
func (c *Client) LookupPoolByUUID(uuidStr string) (libvirt.StoragePool, error) {
	uuid, err := parseUUID(uuidStr)
	if err != nil {
		return libvirt.StoragePool{}, err
	}

	// Look up the pool
	pool, err := c.conn.StoragePoolLookupByUUID(uuid)
	if err != nil {
		return libvirt.StoragePool{}, fmt.Errorf("storage pool not found: %w", err)
	}

	return pool, nil
}

// LookupNetworkByUUID looks up a network by its UUID string
func (c *Client) LookupNetworkByUUID(uuidStr string) (libvirt.Network, error) {
	uuid, err := parseUUID(uuidStr)
	if err != nil {
		return libvirt.Network{}, err
	}

	// Look up the network
	network, err := c.conn.NetworkLookupByUUID(uuid)
	if err != nil {
		return libvirt.Network{}, fmt.Errorf("network not found: %w", err)
	}

	return network, nil
}

// parseUUID converts a UUID string to libvirt.UUID type
func parseUUID(uuidStr string) (libvirt.UUID, error) {
	// Remove hyphens from UUID string
	uuidStr = strings.ReplaceAll(uuidStr, "-", "")

	// Decode hex string to bytes
	uuidBytes, err := hex.DecodeString(uuidStr)
	if err != nil {
		return libvirt.UUID{}, fmt.Errorf("invalid UUID: %w", err)
	}

	if len(uuidBytes) != 16 {
		return libvirt.UUID{}, fmt.Errorf("invalid UUID length: expected 16 bytes, got %d", len(uuidBytes))
	}

	// Convert to libvirt.UUID
	var uuid libvirt.UUID
	copy(uuid[:], uuidBytes)

	return uuid, nil
}

// UUIDString converts a libvirt.UUID to a hyphenated string representation
func UUIDString(uuid libvirt.UUID) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}
