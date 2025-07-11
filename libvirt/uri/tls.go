package uri

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

const (
	defaultTLSPort = "16514"

	caCertName     = "cacert.pem"
	clientCertName = "clientcert.pem"
	clientKeyName  = "clientkey.pem"

	defaultUserPKIPath = "${HOME}/.pki/libvirt"

	defaultGlobalCACertPath     = "/etc/pki/CA"
	defaultGlobalClientCertPath = "/etc/pki/libvirt"
	defaultGlobalClientKeyPath  = "/etc/pki/libvirt/private"
)

// find the first resource that exists in the given list of paths.
func findResource(name string, dirs ...string) (string, error) {
	for _, dir := range dirs {
		path := filepath.Join(os.ExpandEnv(dir), name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		} else if os.IsNotExist(err) {
			continue
		} else {
			return "", err
		}
	}
	return "", fmt.Errorf("can't locate resource '%s' in %v: %w", name, dirs, fs.ErrNotExist)
}

func amIRoot() (bool, error) {
	u, err := user.Current()
	if err != nil {
		// If user.Current() fails, try to get current UID directly
		// This handles cases where the user doesn't exist in /etc/passwd
		uid := os.Getuid()
		return uid == 0, nil
	}
	return u.Uid == "0", nil
}

func nonZero(s string) bool {
	n, err := strconv.Atoi(s)
	return (n != 0 && err == nil) || (len(s) > 0 && n == 0 && err != nil)
}

func (u *ConnectionURI) tlsConfig() (*tls.Config, error) {
	var caCertPath string
	var clientCertPath string
	var clientKeyPath string
	var err error

	caCertSearchPath := []string{defaultGlobalCACertPath}
	clientCertSearchPath := []string{defaultGlobalClientCertPath}
	clientKeySearchPath := []string{defaultGlobalClientKeyPath}

	q := u.Query()
	// if pkipath is provided, certs should all exist there
	if pkipath := q.Get("pkipath"); pkipath != "" {
		caCertSearchPath = []string{pkipath}
		clientCertSearchPath = []string{pkipath}
		clientKeySearchPath = []string{pkipath}
	} else {
		root, err := amIRoot()
		if err != nil {
			return nil, err
		}

		// non-root also looks in $HOME/.pki first
		if !root {
			caCertSearchPath = append([]string{os.ExpandEnv(defaultUserPKIPath)}, caCertSearchPath...)
			clientCertSearchPath = append([]string{os.ExpandEnv(defaultUserPKIPath)}, clientCertSearchPath...)
			clientKeySearchPath = append([]string{os.ExpandEnv(defaultUserPKIPath)}, clientKeySearchPath...)
		}
	}

	if caCertPath, err = findResource(caCertName, caCertSearchPath...); err != nil {
		return nil, err
	}
	if clientCertPath, err = findResource(clientCertName, clientCertSearchPath...); err != nil {
		return nil, err
	}
	if clientKeyPath, err = findResource(clientKeyName, clientKeySearchPath...); err != nil {
		return nil, err
	}

	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("can't read certificate '%s': %w", caCert, err)
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(caCert))
	if !ok {
		return nil, fmt.Errorf("failed to parse CA certificate '%s'", caCertPath)
	}

	clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		RootCAs:            roots,
		Certificates:       []tls.Certificate{clientCert},
		InsecureSkipVerify: nonZero(q.Get("no_verify")),
	}, nil
}

// TODO handle no_verify and pkipath URI options.
func (u *ConnectionURI) dialTLS() (net.Conn, error) {
	port := u.Port()
	if port == "" {
		port = defaultTLSPort
	}

	tlsConfig, err := u.tlsConfig()
	if err != nil {
		return nil, err
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", u.Hostname(), port), tlsConfig)
	if err != nil {
		return nil, err
	}

	// CRITICAL: Read server verification byte
	// When running over TLS, after connection libvirt writes a single byte to
	// the socket to indicate whether the server's check of the client's
	// certificate has succeeded.
	// See https://github.com/digitalocean/go-libvirt/issues/89#issuecomment-1607300636
	buf := make([]byte, 1)
	if n, err := conn.Read(buf); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read server verification byte: %w", err)
	} else if n != 1 || buf[0] != byte(1) {
		conn.Close()
		return nil, errors.New("server verification (of our certificate or IP address) failed")
	}
	
	return conn, nil
}
