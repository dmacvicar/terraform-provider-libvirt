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
	defaultHostKeyAlgorithm  = ssh.KeyAlgoRSA
)

func (u *ConnectionURI) parseAuthMethods(target string, sshcfg *ssh_config.Config) []ssh.AuthMethod {
	q := u.Query()

	authMethods := q.Get("sshauth")
	if authMethods == "" {
		authMethods = defaultSSHAuthMethods
	}

	// keyfile order of precedence:
	//  1. load uri encoded keyfile
	//  2. load override as specified in ssh config
	//  3. load default ssh keyfile path
	sshKeyPath := q.Get("keyfile")
	if sshKeyPath == "" {

		keyPaths, err := sshcfg.GetAll(target, "IdentityFile")
		if err == nil && len(keyPaths) != 0 {
			sshKeyPath = keyPaths[len(keyPaths) - 1]
		} else {
			sshKeyPath = defaultSSHKeyPath
		}
	}
	log.Printf("[DEBUG] ssh identity file for host '%s': %s", target, sshKeyPath);

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
			path := os.ExpandEnv(sshKeyPath)
			if strings.HasPrefix(path, "~/") {
				home, err := os.UserHomeDir()
				if err == nil {
					path = strings.Replace(path, "~", home, 1)
				}
			}
			sshKey, err := os.ReadFile(path)
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

	cfg := ssh.ClientConfig{
		User:            u.User.Username(),
		HostKeyCallback: hostKeyCallback,
		Timeout:         dialTimeout,
	}

	sshClient, err := u.dialHost(u.Host, sshcfg, cfg, 0)
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

func (u *ConnectionURI) dialHost(target string, sshcfg *ssh_config.Config, cfg ssh.ClientConfig, depth int) (*ssh.Client, error) {
	if depth > 10 {
		return nil, fmt.Errorf("[ERROR] dialHost failed: max tunnel depth of 10 reached")
	}

	log.Printf("[DEBUG] Connecting to target: %s", target);

	proxy, err := sshcfg.Get(target, "ProxyCommand")
	if err == nil && proxy != "" {
		log.Printf("[WARNING] unsupported ssh ProxyCommand '%v'", proxy)
	}

	proxy, err = sshcfg.Get(target, "ProxyJump")
	var bastion *ssh.Client
	if err == nil && proxy != "" {
		log.Printf("[DEBUG] SSH ProxyJump '%v'", proxy)

		// if this is a proxy jump, we recurse into that proxy
		bastion, err = u.dialHost(proxy, sshcfg, cfg, depth + 1)
	}

	if cfg.User == "" {
		sshu, err := sshcfg.Get(target, "User")
		log.Printf("[DEBUG] SSH User for target '%v': %v", target, sshu)
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

	port := u.Port()
	if port == "" {
		port = defaultSSHPort
	} else {
		log.Printf("[DEBUG] SSH Port is overriden to: '%s'", port);
	}

	hostName, err := sshcfg.Get(target, "HostName")
	if err == nil {
		if hostName == "" {
			hostName = target;
		} else {
			log.Printf("[DEBUG] HostName is overriden to: '%s'", hostName);
		}
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
			log.Fatal(err)
			return nil, fmt.Errorf("failed to connect to remote host '%v': %w", target, err)
		}

		ncc, chans, reqs, err := ssh.NewClientConn(conn, target, &cfg)
		if err != nil {
			log.Fatal(err)
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
