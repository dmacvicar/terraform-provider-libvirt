package uri

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"strings"

	"github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	defaultSSHPort           = "22"
	defaultSSHKeyPath        = "${HOME}/.ssh/id_rsa"
	defaultSSHKnownHostsPath = "${HOME}/.ssh/known_hosts"
	defaultSSHConfigFile     = "${HOME}/.ssh/config"
	defaultSSHAuthMethods    = "agent,privkey"
)

func (u *ConnectionURI) parseAuthMethods() []ssh.AuthMethod {
	q := u.Query()

	authMethods := q.Get("sshauth")
	if authMethods == "" {
		authMethods = defaultSSHAuthMethods
	}

	sshKeyPath := q.Get("keyfile")
	if sshKeyPath == "" {
		sshKeyPath = defaultSSHKeyPath
	}

	auths := strings.Split(authMethods, ",")
	result := make([]ssh.AuthMethod, 0)
	for _, v := range auths {
		switch v {
		case "agent":
			socket := os.Getenv("SSH_AUTH_SOCK")
			if socket == "" {
				continue
			}
			conn, err := net.Dial("unix", socket)
			// Ignore error, we just fall back to another auth method
			if err != nil {
				log.Printf("[ERROR] Unable to connect to SSH agent: %v", err)
				continue
			}
			agentClient := agent.NewClient(conn)
			result = append(result, ssh.PublicKeysCallback(agentClient.Signers))
		case "privkey":
			sshKey, err := os.ReadFile(os.ExpandEnv(sshKeyPath))
			if err != nil {
				log.Printf("[ERROR] Failed to read ssh key: %v", err)
				continue
			}

			signer, err := ssh.ParsePrivateKey(sshKey)
			if err != nil {
				log.Printf("[ERROR] Failed to parse ssh key: %v", err)
			}
			result = append(result, ssh.PublicKeys(signer))
		case "ssh-password":
			if sshPassword, ok := u.User.Password(); ok {
				result = append(result, ssh.Password(sshPassword))
			} else {
				log.Printf("[ERROR] Missing password in userinfo of URI authority section")
			}
		default:
			// For future compatibility it's better to just warn and not error
			log.Printf("[WARN] Unsupported auth method: %s", v)
		}
	}

	return result
}

func (u *ConnectionURI) dialSSH() (net.Conn, error) {

	sshConfigFile, err := os.Open(os.ExpandEnv(defaultSSHConfigFile))
	if err != nil {
		log.Printf("[WARN] Failed to open ssh config file: %v", err)
	}

	sshcfg, err := ssh_config.Decode(sshConfigFile)
	if err != nil {
		log.Printf("[WARN] Failed to parse ssh config file: %v", err)
	}

	authMethods := u.parseAuthMethods()
	if len(authMethods) < 1 {
		return nil, fmt.Errorf("could not configure SSH authentication methods")
	}
	q := u.Query()

	knownHostsPath := q.Get("knownhosts")
	knownHostsVerify := q.Get("known_hosts_verify")
	doVerify := q.Get("no_verify") == ""

	if knownHostsVerify == "ignore" {
		doVerify = false
	}

	if knownHostsPath == "" {
		knownHostsPath = defaultSSHKnownHostsPath
	}

	hostKeyCallback := ssh.InsecureIgnoreHostKey()
	if doVerify {
		cb, err := knownhosts.New(os.ExpandEnv(knownHostsPath))
		if err != nil {
			return nil, fmt.Errorf("failed to read ssh known hosts: %w", err)
		}
		hostKeyCallback = cb
	}

	username := u.User.Username()
	if username == "" {
		sshu, err := sshcfg.Get(u.Host, "User")
		log.Printf("[DEBUG] SSH User: %v", sshu)
		if err != nil {
			log.Printf("[DEBUG] ssh user: system username")
			u, err := user.Current()
			if err != nil {
				return nil, fmt.Errorf("unable to get username: %w", err)
			}
			sshu = u.Username
		}
		username = sshu
	}

	cfg := ssh.ClientConfig{
		User:            username,
		HostKeyCallback: hostKeyCallback,
		Auth:            authMethods,
		Timeout:         dialTimeout,
	}

	port := u.Port()
	if port == "" {
		port = defaultSSHPort
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", u.Hostname(), port), &cfg)
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
