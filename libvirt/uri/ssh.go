package uri

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	defaultSSHPort           = "22"
	defaultSSHKeyPath        = "${HOME}/.ssh/id_rsa"
	defaultSSHKnownHostsPath = "${HOME}/.ssh/known_hosts"
)

// TODO handle known_hosts_verify, no_verify and sshauth URI options
func (curi *ConnectionURI) dialSSH() (net.Conn, error) {
	q := curi.Query()

	knownHostsPath := q.Get("knownhosts")
	if knownHostsPath == "" {
		knownHostsPath = defaultSSHKnownHostsPath
	}
	hostKeyCallback, err := knownhosts.New(os.ExpandEnv(knownHostsPath))
	if err != nil {
		return nil, fmt.Errorf("failed to read ssh known hosts: %w", err)
	}

	sshKeyPath := q.Get("keyfile")
	if sshKeyPath == "" {
		sshKeyPath = defaultSSHKeyPath
	}
	sshKey, err := ioutil.ReadFile(os.ExpandEnv(sshKeyPath))
	if err != nil {
		return nil, fmt.Errorf("failed to read ssh key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(sshKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ssh key: %w", err)
	}

	username := curi.User.Username()
	if username == "" {
		u, err := user.Current()
		if err != nil {
			return nil, err
		}
		username = u.Name
	}

	cfg := ssh.ClientConfig{
		User:            username,
		HostKeyCallback: hostKeyCallback,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		Timeout:         2 * time.Second,
	}

	port := curi.Port()
	if port == "" {
		port = defaultSSHPort
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", curi.Hostname(), port), &cfg)
	if err != nil {
		return nil, err
	}

	address := q.Get("socket")
	if address == "" {
		address = defaultUnixSock
	}

	c, err := sshClient.Dial("unix", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt on the remote host: %w", err)
	}

	return c, nil
}
