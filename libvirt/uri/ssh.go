package uri

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	maxHostHops              = 10
	defaultSSHPort           = "22"
	defaultSSHKeyPath        = "${HOME}/.ssh/id_rsa"
	defaultSSHKnownHostsPath = "${HOME}/.ssh/known_hosts"
	defaultSSHConfigFile     = "${HOME}/.ssh/config"
	defaultSSHAuthMethods    = "agent,privkey"
)

func (u *ConnectionURI) parseAuthMethods(target string, sshcfg *ssh_config.Config) []ssh.AuthMethod {
	q := u.Query()

	authMethods := q.Get("sshauth")
	if authMethods == "" {
		authMethods = defaultSSHAuthMethods
	}

	log.Printf("[DEBUG] auth methods for %v: %v", target, authMethods)

	// keyfile order of precedence:
	//  1. load uri encoded keyfile
	//  2. load override as specified in ssh config
	//  3. load default ssh keyfile path
	sshKeyPaths := []string {}
	sshKeyPath := q.Get("keyfile")
	if sshKeyPath != "" {
		sshKeyPaths = append(sshKeyPaths, sshKeyPath)
	}

	keyPaths, err := sshcfg.GetAll(target, "IdentityFile")
	if err != nil {
		log.Printf("[WARN] unable to get IdentityFile values - ignoring")
	} else {
		sshKeyPaths = append(sshKeyPaths, keyPaths...)
	}

	if len(keyPaths) == 0 {
		log.Printf("[DEBUG] found no ssh keys, using default keypath")
		sshKeyPaths = []string{defaultSSHKeyPath}
	}

	log.Printf("[DEBUG] ssh identity files for host '%s': %s", target, sshKeyPaths);

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
			for _, keypath := range sshKeyPaths {
				log.Printf("[DEBUG] Reading ssh key '%s'", keypath)
				path := os.ExpandEnv(keypath)
				if strings.HasPrefix(path, "~/") {
					home, err := os.UserHomeDir()
					if err == nil {
						path = filepath.Join(home, path[2:])
					}
				}
				sshKey, err := os.ReadFile(path)
				if err != nil {
					log.Printf("[ERROR] Failed to read ssh key '%s': %v", keypath, err)
					continue
				}

				signer, err := ssh.ParsePrivateKey(sshKey)
				if err != nil {
					log.Printf("[ERROR] Failed to parse ssh key: %v", err)
				}
				result = append(result, ssh.PublicKeys(signer))
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

// construct the whole ssh connection, which can consist of multiple hops if using proxy jumps,
// the ssh configuration file is loaded once and passed along to each host connection.
func (u *ConnectionURI) dialSSH() (net.Conn, error) {
	sshConfigFile, err := os.Open(os.ExpandEnv(defaultSSHConfigFile))
	if err != nil {
		log.Printf("[WARN] Failed to open ssh config file: %v", err)
	}

	sshcfg, err := ssh_config.Decode(sshConfigFile)
	if err != nil {
		log.Printf("[WARN] Failed to parse ssh config file: %v", err)
	}

	// configuration loaded, build tunnel
	sshClient, err := u.dialHost(u.Host, sshcfg, 0)
	if err != nil {
		return nil, err
	}

	// tunnel established, connect to the libvirt unix socket to communicate
	// e.g. /var/run/libvirt/libvirt-sock
	address := u.Query().Get("socket")
	if address == "" {
		address = defaultUnixSock
	}

	c, err := sshClient.Dial("unix", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt on the remote host: %w", err)
	}

	return c, nil
}

func (u *ConnectionURI) dialHost(target string, sshcfg *ssh_config.Config, depth int) (*ssh.Client, error) {

	if depth > maxHostHops {
		return nil, fmt.Errorf("[ERROR] dialHost failed: max tunnel depth of 10 reached")
	}

	log.Printf("[INFO] establishing ssh connection to '%s'", target);

	q := u.Query()

	port := u.Port()
	if port == "" {
		port = defaultSSHPort
	} else {
		log.Printf("[DEBUG] ssh Port is overridden to: '%s'", port);
	}

	hostName, err := sshcfg.Get(target, "HostName")
	if err == nil {
		if hostName == "" {
			hostName = target;
		} else {
			log.Printf("[DEBUG] HostName is overridden to: '%s'", hostName);
		}
	}

	// we must check for knownhosts and verification for each host we connect to.
	// the query string values have higher precedence to local configs
	knownHostsPath := q.Get("knownhosts")
	knownHostsVerify := q.Get("known_hosts_verify")
	skipVerify := q.Has("no_verify")

	if knownHostsVerify == "ignore" {
		skipVerify = true
	} else {
		strictCheck, err := sshcfg.Get(target, "StrictHostKeyChecking")
		if err != nil && strictCheck == "yes" {
			skipVerify = false
		}
	}

	if knownHostsPath == "" {
		knownHosts, err := sshcfg.Get(target, "UserKnownHostsFile")
		if err == nil && knownHosts != "" {
			knownHostsPath = knownHosts
		} else {
			knownHostsPath = defaultSSHKnownHostsPath
		}
	}

	hostKeyCallback := ssh.InsecureIgnoreHostKey()
	hostKeyAlgorithms := []string{ // https://github.com/golang/go/issues/29286
		// this can be solved using https://github.com/skeema/knownhosts/tree/main
		// there is an open issue requiring attention
		ssh.KeyAlgoED25519,
		ssh.KeyAlgoRSA,
		ssh.KeyAlgoRSASHA256,
		ssh.KeyAlgoRSASHA512,
		ssh.KeyAlgoSKECDSA256,
		ssh.KeyAlgoSKED25519,
		ssh.KeyAlgoECDSA256,
		ssh.KeyAlgoECDSA384,
		ssh.KeyAlgoECDSA521,
	}
	if !skipVerify {
		kh, err := knownhosts.New(os.ExpandEnv(knownHostsPath))
		if err != nil {
			return nil, fmt.Errorf("failed to read ssh known hosts: %w", err)
		}
		log.Printf("[DEBUG] Using known hosts file '%s' for target '%s'", os.ExpandEnv(knownHostsPath), target)

		hostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			err := kh(net.JoinHostPort(hostName, port), remote, key)
			if err != nil {
				log.Printf("Host key verification failed for host '%s' (%s) %v: %v", hostName, remote, key, err)
			}
			return err
		}

		keyAlgs, err := sshcfg.Get(target, "HostKeyAlgorithms")
		if err == nil && keyAlgs != "" {
			log.Printf("Got host key algorithms '%s'", keyAlgs)
			hostKeyAlgorithms = strings.Split(keyAlgs, ",")
		}

	}

	cfg := ssh.ClientConfig{
		User:            u.User.Username(),
		HostKeyCallback: hostKeyCallback,
		HostKeyAlgorithms: hostKeyAlgorithms,
		Timeout:         dialTimeout,
	}

	proxy, err := sshcfg.Get(target, "ProxyCommand")
	if err == nil && proxy != "" {
		log.Printf("[WARNING] unsupported ssh ProxyCommand '%v'", proxy)
	}

	proxy, err = sshcfg.Get(target, "ProxyJump")
	var bastion *ssh.Client
	if err == nil && proxy != "" {
		log.Printf("[DEBUG] found ProxyJump '%v'", proxy)

		// this is a proxy jump: we recurse into that proxy
		bastion, err = u.dialHost(proxy, sshcfg, depth + 1)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to bastion host '%v': %w", proxy, err)
		}
	}

	if cfg.User == "" {
		sshu, err := sshcfg.Get(target, "User")
		log.Printf("[DEBUG] SSH User for target '%v' is '%v'", target, sshu)
		if err != nil {
			log.Printf("[DEBUG] ssh user: using current login")
			u, err := user.Current()
			if err != nil {
				return nil, fmt.Errorf("unable to get username: %w", err)
			}
			sshu = u.Username
		}
		cfg.User = sshu
	}

	cfg.Auth = u.parseAuthMethods(target, sshcfg)
	if len(cfg.Auth) < 1 {
		return nil, fmt.Errorf("could not configure SSH authentication methods")
	}

	if (bastion != nil) {
		// if this is a proxied connection, we want to dial through the bastion host
		log.Printf("[INFO] SSH connecting to '%v' (%v) through bastion host '%v'", target, hostName, proxy)
		// Dial a connection to the service host, from the bastion
		conn, err := bastion.Dial("tcp", net.JoinHostPort(hostName, port))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to remote host '%v': %w", target, err)
		}

		ncc, chans, reqs, err := ssh.NewClientConn(conn, target, &cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to remote host '%v': %w", target, err)
		}

		sClient := ssh.NewClient(ncc, chans, reqs)
		return sClient, nil

	} else {
		// this is a direct connection to the target host
		log.Printf("[INFO] SSH connecting to '%v' (%v)", target, hostName)
		conn,err := ssh.Dial("tcp", net.JoinHostPort(hostName, port), &cfg)

		if err != nil {
			return nil, fmt.Errorf("failed to connect to remote host '%v': %w", target, err)
		}
		return conn, nil
	}
}
