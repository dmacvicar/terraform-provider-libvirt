package dialers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	defaultUnixSock  = "/var/run/libvirt/libvirt-sock"
	defaultNetcatBin = "nc"
	defaultHelperBin = "virt-ssh-helper"
	defaultSSHBin    = "ssh"
)

// https://libvirt.org/uri.html#proxy-parameter
type ProxyMode string

const (
	ProxyAuto   ProxyMode = "auto"
	ProxyNative ProxyMode = "native"
	ProxyNetcat ProxyMode = "netcat"
)

// SSHCmdDialer implements socket.Dialer interface for go-libvirt
// It uses the command-line ssh tool for communication, which automatically
// respects OpenSSH config settings in ~/.ssh/config
type SSHCmdDialer struct {
	// Connection details
	hostname  string
	port      string
	username  string
	socket    string
	remoteURI string // Remote URI to pass to virt-ssh-helper

	// Proxy configuration
	proxyMode ProxyMode
	netcatBin string // Netcat binary to use when needed

	// SSH options
	sshBin          string
	keyFiles        []string
	knownHostsFile  string
	strictHostCheck bool
	batchMode       bool
	forwardAgent    bool
	authMethods     []string // Authentication methods to use
	extraArgs       []string // Any additional SSH arguments
}

func NewSSHCmdDialer(uri *url.URL) *SSHCmdDialer {
	hostname := uri.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}

	query := uri.Query()

	// Extract the driver part of the URI (qemu, lxc, etc.)
	driver := strings.Split(uri.Scheme, "+")[0]

	// Construct the remote URI (e.g., "qemu:///system")
	remoteName := driver + ":///system"
	if uri.Path != "" && uri.Path != "/" {
		remoteName = driver + "://" + uri.Path
	}

	dialer := &SSHCmdDialer{
		// Connection details
		hostname:  hostname,
		port:      uri.Port(),
		socket:    defaultUnixSock,
		remoteURI: remoteName,

		// Proxy configuration
		proxyMode: ProxyAuto, // default like upstream libvirt
		netcatBin: defaultNetcatBin,

		// SSH options
		sshBin:          defaultSSHBin,
		strictHostCheck: true,
		batchMode:       true,
		forwardAgent:    false,
		authMethods:     []string{},
		keyFiles:        []string{},
	}

	if uri.User != nil {
		dialer.username = uri.User.Username()
	}

	if socketParam := query.Get("socket"); socketParam != "" {
		dialer.socket = socketParam
	}

	if proxyParam := query.Get("proxy"); proxyParam != "" {
		switch ProxyMode(proxyParam) {
		case ProxyAuto, ProxyNative, ProxyNetcat:
			dialer.proxyMode = ProxyMode(proxyParam)
		default:
			log.Printf("[WARN] Unknown proxy mode: %s, using 'auto'", proxyParam)
		}
	}

	if netcatParam := query.Get("netcat"); netcatParam != "" {
		dialer.netcatBin = netcatParam
	}

	dialer.applyURIOptions(uri)

	return dialer
}

// see: https://libvirt.org/uri.html#ssh-transport
func (d *SSHCmdDialer) applyURIOptions(uri *url.URL) {
	query := uri.Query()

	if keyFile := query.Get("keyfile"); keyFile != "" {
		keyFile = os.ExpandEnv(keyFile)
		if strings.HasPrefix(keyFile, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				keyFile = strings.Replace(keyFile, "~", home, 1)
			}
		}
		d.keyFiles = append(d.keyFiles, keyFile)
	}

	if knownHosts := query.Get("knownhosts"); knownHosts != "" {
		knownHosts = os.ExpandEnv(knownHosts)
		if strings.HasPrefix(knownHosts, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				knownHosts = strings.Replace(knownHosts, "~", home, 1)
			}
		}
		d.knownHostsFile = knownHosts
	}

	knownHostsVerify := query.Get("known_hosts_verify")
	if knownHostsVerify == "ignore" || query.Has("no_verify") {
		d.strictHostCheck = false
	}

	if command := query.Get("command"); command != "" {
		d.sshBin = command
	}

	if proxy := query.Get("proxy"); proxy != "" {
		switch proxy {
		case string(ProxyAuto):
			d.proxyMode = ProxyAuto
		case string(ProxyNative):
			d.proxyMode = ProxyNative
		case string(ProxyNetcat):
			d.proxyMode = ProxyNetcat
		default:
			log.Printf("[WARN] Unknown proxy mode: %s, using 'auto'", proxy)

		}
	}

	if sshAuth := query.Get("sshauth"); sshAuth != "" {
		authMethods := strings.Split(sshAuth, ",")
		d.authMethods = authMethods

		// Set specific options based on auth methods
		for _, auth := range authMethods {
			switch auth {
			case "agent":
				d.forwardAgent = true
			case "password", "keyboard-interactive":
				d.batchMode = false // Disable batch mode for interactive auth
			}
		}
	}

	// TODO mode parameter
}

// Dial implements the socket.Dialer interface to enable using this dialer with go-libvirt
func (d *SSHCmdDialer) Dial() (net.Conn, error) {
	args := d.buildSSHArgs()

	log.Printf("[INFO] SSH command dialer connecting to %s with args: %v", d.hostname, args)

	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	cmd := exec.Command(d.sshBin, args...)
	cmd.Stdin = stdinReader
	cmd.Stdout = stdoutWriter
	cmd.Stderr = os.Stderr // Log error output to stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ssh command: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		err := cmd.Wait()
		if err != nil {
			log.Printf("[ERROR] SSH command exited unexpectedly: %v", err)
		}
		cancel() // Ensure cleanup is triggered
	}()

	// custom net.Conn implementation that communicates with the ssh process
	conn := &sshCmdConn{
		cmd:          cmd,
		stdin:        stdinWriter,
		stdout:       stdoutReader,
		cancel:       cancel,
		hostAndPort:  d.hostname,
		remoteSocket: d.socket,
	}

	// Monitor the process in a goroutine
	go func() {
		defer cancel()
		<-ctx.Done()
		// Process monitoring is done, clean up
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// Wait for initial connection (give ssh some time to establish the connection)
	time.Sleep(100 * time.Millisecond)
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		return nil, fmt.Errorf("ssh command terminated prematurely with exit code %d", cmd.ProcessState.ExitCode())
	}

	return conn, nil
}

func (d *SSHCmdDialer) buildSSHArgs() []string {
	args := []string{}

	if d.port != "" {
		args = append(args, "-p", d.port)
	}

	// Standard arguments for libvirt connections
	args = append(args,
		"-T",                     // Disable pseudo-terminal allocation
		"-o", "ControlPath=none", // Don't use multiplexing
		"-e", "none", // Disable escape character
	)

	for _, keyFile := range d.keyFiles {
		args = append(args, "-i", keyFile)
	}

	if d.knownHostsFile != "" {
		args = append(args, "-o", "UserKnownHostsFile="+d.knownHostsFile)
	}

	if !d.strictHostCheck {
		args = append(args, "-o", "StrictHostKeyChecking=no")
	}

	if d.batchMode {
		args = append(args, "-o", "BatchMode=yes")
	}

	if d.forwardAgent {
		args = append(args, "-o", "ForwardAgent=yes")
	}

	for _, auth := range d.authMethods {
		switch auth {
		case "privkey":
			args = append(args, "-o", "PreferredAuthentications=publickey")
		case "password":
			args = append(args, "-o", "PreferredAuthentications=password")
		case "keyboard-interactive":
			args = append(args, "-o", "PreferredAuthentications=keyboard-interactive")
		}
	}

	args = append(args, d.extraArgs...)

	// Build the destination string (user@host)
	destination := d.hostname
	if d.username != "" {
		destination = d.username + "@" + destination
	}

	// Use the remote URI that was constructed during initialization
	var shellCmd string

	switch d.proxyMode {
	case ProxyNative:
		// Native mode uses virt-ssh-helper directly
		shellCmd = fmt.Sprintf("sh -c 'virt-ssh-helper \"%s\"'", d.remoteURI)
		log.Printf("[DEBUG] Using native virt-ssh-helper with URI: %s", d.remoteURI)

	case ProxyNetcat:
		// Netcat mode - detect proper flags for netcat
		shellCmd = fmt.Sprintf("sh -c 'if \"%s\" -q 2>&1 | grep \"requires an argument\" >/dev/null 2>&1; then ARG=-q0; else ARG=; fi; \"%s\" $ARG -U %s'",
			d.netcatBin, d.netcatBin, d.socket)
		log.Printf("[DEBUG] Using netcat %s for socket connection to %s", d.netcatBin, d.socket)

	case ProxyAuto:
		// Auto mode - try virt-ssh-helper first, then fall back to netcat
		shellCmd = fmt.Sprintf("sh -c 'which virt-ssh-helper 1>/dev/null 2>&1; if test $? = 0; then virt-ssh-helper \"%s\"; else if \"%s\" -q 2>&1 | grep \"requires an argument\" >/dev/null 2>&1; then ARG=-q0; else ARG=; fi; \"%s\" $ARG -U %s; fi'",
			d.remoteURI, d.netcatBin, d.netcatBin, d.socket)
		log.Printf("[DEBUG] Using auto proxy mode with URI: %s", d.remoteURI)

	default:
		// This shouldn't happen, but use netcat as a safe fallback
		shellCmd = fmt.Sprintf("sh -c '\"%s\" -U %s'", d.netcatBin, d.socket)
		log.Printf("[WARN] Unknown proxy mode, falling back to netcat")
	}

	args = append(args, "--", destination, shellCmd)

	return args
}

// sshCmdConn implements net.Conn to communicate with the ssh process
type sshCmdConn struct {
	cmd          *exec.Cmd
	stdin        io.WriteCloser
	stdout       io.ReadCloser
	cancel       context.CancelFunc
	hostAndPort  string
	remoteSocket string
}

func (c *sshCmdConn) Read(b []byte) (n int, err error) {
	return c.stdout.Read(b)
}

func (c *sshCmdConn) Write(b []byte) (n int, err error) {
	return c.stdin.Write(b)
}

func (c *sshCmdConn) Close() error {
	c.cancel()
	c.stdin.Close()
	c.stdout.Close()
	return nil
}

func (c *sshCmdConn) LocalAddr() net.Addr {
	return &net.UnixAddr{Name: "local", Net: "unix"}
}

func (c *sshCmdConn) RemoteAddr() net.Addr {
	return &net.UnixAddr{Name: c.remoteSocket, Net: "unix"}
}

func (c *sshCmdConn) SetDeadline(t time.Time) error {
	return fmt.Errorf("SetDeadline not implemented for SSH command connection")
}

func (c *sshCmdConn) SetReadDeadline(t time.Time) error {
	return fmt.Errorf("SetReadDeadline not implemented for SSH command connection")
}

func (c *sshCmdConn) SetWriteDeadline(t time.Time) error {
	return fmt.Errorf("SetWriteDeadline not implemented for SSH command connection")
}
