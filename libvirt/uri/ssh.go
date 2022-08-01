package uri

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/user"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	defaultSSHPort           = "22"
	defaultSSHKeyPath        = "${HOME}/.ssh/id_rsa"
	defaultSSHKnownHostsPath = "${HOME}/.ssh/known_hosts"
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
			sshKey, err := ioutil.ReadFile(os.ExpandEnv(sshKeyPath))
			if err != nil {
				log.Printf("[ERROR] Failed to read ssh key: %v", err)
				continue
			}

			signer, err := ssh.ParsePrivateKey(sshKey)
			if err != nil {
				log.Printf("[ERROR] Failed to parse ssh key: %v", err)
			}

			// Use SSH certificate if any
			sshCertPath := sshKeyPath + "-cert.pub"
			sshCert, err := ioutil.ReadFile(os.ExpandEnv(sshCertPath))
			if err != nil {
				// Log and ignore error if no certificate is found: we just continue using private key only
				log.Printf("[ERROR] Failed to read ssh certificate: %v", err)
				result = append(result, ssh.PublicKeys(signer))
				continue
			} else {
				publiKeyCert, _, _, _, err := ssh.ParseAuthorizedKey(sshCert)
				if err != nil {
					// Log and ignore error, if not a parseable certificate: we just continue using private key only
					log.Printf("[ERROR] Failed to parse ssh certificate: %v", err)
					result = append(result, ssh.PublicKeys(signer))
					continue
				} else {
					certSigner, err := ssh.NewCertSigner(publiKeyCert.(*ssh.Certificate), signer)
					if err != nil {
						log.Printf("[ERROR] Failed to use ssh certificate: %v", err)
					} else {
						result = append(result, ssh.PublicKeys(certSigner))
					}
				}
			}
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
		u, err := user.Current()
		if err != nil {
			return nil, err
		}
		username = u.Username
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
