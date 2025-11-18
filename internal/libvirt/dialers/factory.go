package dialers

import (
	"fmt"
	"net/url"
	"os"
	"os/user"
	"strings"

	"github.com/digitalocean/go-libvirt/socket/dialers"
)

const (
	// DefaultSystemSocket is the default libvirt system socket path
	DefaultSystemSocket = "/var/run/libvirt/libvirt-sock"
)

// NewDialerFromURI creates the appropriate Dialer based on the libvirt URI.
// It uses upstream go-libvirt dialers for most transports and custom dialers
// for special cases like SSHCmd.
func NewDialerFromURI(uriStr string) (Dialer, error) {
	parsedURI, err := url.Parse(uriStr)
	if err != nil {
		return nil, fmt.Errorf("invalid libvirt URI: %w", err)
	}

	// Parse the scheme to extract driver and transport
	// Format: driver[+transport]://[host]/path
	// Examples: qemu:///system, qemu+ssh://host/system, qemu+sshcmd://host/system
	schemeParts := strings.Split(parsedURI.Scheme, "+")
	driver := schemeParts[0]
	transport := ""
	if len(schemeParts) > 1 {
		transport = schemeParts[1]
	}

	// Validate driver
	switch driver {
	case "qemu", "lxc", "xen", "vbox", "test":
		// Valid drivers
	default:
		return nil, fmt.Errorf("unsupported libvirt driver: %s", driver)
	}

	// Local connection (no transport specified and no host)
	if transport == "" && parsedURI.Host == "" {
		return newLocalDialer(parsedURI)
	}

	// Remote connections
	switch transport {
	case "ssh":
		// Use Go SSH library (upstream dialer)
		return newGoSSHDialer(parsedURI)
	case "sshcmd":
		// Use native SSH command (custom dialer)
		return NewSSHCmd(parsedURI), nil
	case "tcp":
		// Plain TCP connection (upstream dialer)
		return newRemoteDialer(parsedURI)
	case "tls":
		// TLS connection (upstream dialer)
		return newTLSDialer(parsedURI)
	case "":
		// No transport but has host - assume SSH
		return newGoSSHDialer(parsedURI)
	default:
		return nil, fmt.Errorf("unsupported transport: %s", transport)
	}
}

// newLocalDialer creates a Local dialer for Unix socket connections
func newLocalDialer(parsedURI *url.URL) (Dialer, error) {
	query := parsedURI.Query()
	socketPath := query.Get("socket")

	if socketPath == "" {
		// Determine socket based on path
		if parsedURI.Path == "/session" {
			// Session socket - use current user's runtime directory
			currentUser, err := user.Current()
			if err != nil {
				return nil, fmt.Errorf("failed to get current user for session socket: %w", err)
			}
			socketPath = fmt.Sprintf("/run/user/%s/libvirt/libvirt-sock", currentUser.Uid)
		} else {
			// System socket
			socketPath = DefaultSystemSocket
		}
	}

	return dialers.NewLocal(
		dialers.WithSocket(socketPath),
	), nil
}

// newGoSSHDialer creates an SSH dialer using the Go SSH library (upstream)
func newGoSSHDialer(parsedURI *url.URL) (Dialer, error) {
	hostname := parsedURI.Hostname()
	if hostname == "" {
		return nil, fmt.Errorf("SSH transport requires a hostname")
	}

	currUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	query := parsedURI.Query()
	opts := []dialers.SSHOption{
		dialers.WithSystemSSHDefaults(currUser), // Use system defaults
	}

	// Port
	if port := parsedURI.Port(); port != "" {
		opts = append(opts, dialers.UseSSHPort(port))
	}

	// Username
	if parsedURI.User != nil {
		username := parsedURI.User.Username()
		opts = append(opts, dialers.UseSSHUsername(username))

		// Password (if provided in URI)
		if password, ok := parsedURI.User.Password(); ok {
			opts = append(opts, dialers.UseSSHPassword(password))
		}
	}

	// Key file
	if keyFile := query.Get("keyfile"); keyFile != "" {
		keyFile = os.ExpandEnv(keyFile)
		if strings.HasPrefix(keyFile, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				keyFile = strings.Replace(keyFile, "~", home, 1)
			}
		}
		opts = append(opts, dialers.UseKeyFile(keyFile))
	}

	// Known hosts
	if knownHosts := query.Get("knownhosts"); knownHosts != "" {
		knownHosts = os.ExpandEnv(knownHosts)
		if strings.HasPrefix(knownHosts, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				knownHosts = strings.Replace(knownHosts, "~", home, 1)
			}
		}
		opts = append(opts, dialers.UseKnownHostsFile(knownHosts))
	}

	// Known hosts verification
	knownHostsVerify := query.Get("known_hosts_verify")
	if knownHostsVerify == "ignore" || query.Has("no_verify") {
		opts = append(opts, dialers.WithInsecureIgnoreHostKey())
	}

	// Remote socket path
	if socket := query.Get("socket"); socket != "" {
		opts = append(opts, dialers.WithRemoteSocket(socket))
	}

	return dialers.NewSSH(hostname, opts...), nil
}

// newRemoteDialer creates a Remote (TCP) dialer
func newRemoteDialer(parsedURI *url.URL) (Dialer, error) {
	hostname := parsedURI.Hostname()
	if hostname == "" {
		return nil, fmt.Errorf("TCP transport requires a hostname")
	}

	opts := []dialers.RemoteOption{}

	// Port
	if port := parsedURI.Port(); port != "" {
		opts = append(opts, dialers.UsePort(port))
	}

	return dialers.NewRemote(hostname, opts...), nil
}

// newTLSDialer creates a TLS dialer
func newTLSDialer(parsedURI *url.URL) (Dialer, error) {
	hostname := parsedURI.Hostname()
	if hostname == "" {
		return nil, fmt.Errorf("TLS transport requires a hostname")
	}

	query := parsedURI.Query()
	opts := []dialers.TLSOption{}

	// Port
	if port := parsedURI.Port(); port != "" {
		opts = append(opts, dialers.UseTLSPort(port))
	}

	// PKI path
	if pkiPath := query.Get("pkipath"); pkiPath != "" {
		opts = append(opts, dialers.UsePKIPath(pkiPath))
	}

	// No verify
	if query.Has("no_verify") {
		opts = append(opts, dialers.WithInsecureNoVerify())
	}

	return dialers.NewTLS(hostname, opts...), nil
}
