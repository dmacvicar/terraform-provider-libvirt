package dialers

import (
	"bufio"
	"container/ring"
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultUnixSock  = "/var/run/libvirt/libvirt-sock"
	defaultNetcatBin = "nc"
	defaultSSHBin    = "ssh"
)

// ProxyMode defines the SSH proxy mode for connecting to libvirt.
// See https://libvirt.org/uri.html#proxy-parameter
type ProxyMode string

const (
	// ProxyAuto tries virt-ssh-helper first, falls back to netcat
	ProxyAuto ProxyMode = "auto"
	// ProxyNative uses virt-ssh-helper directly
	ProxyNative ProxyMode = "native"
	// ProxyNetcat uses netcat to connect to Unix socket
	ProxyNetcat ProxyMode = "netcat"
)

// SSHCmd implements the Dialer interface using the command-line ssh tool.
// This dialer automatically respects OpenSSH config settings in ~/.ssh/config.
type SSHCmd struct {
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

// NewSSHCmd creates a new SSHCmd dialer from a libvirt URI.
func NewSSHCmd(uri *url.URL) *SSHCmd {
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

	dialer := &SSHCmd{
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
		}
	}

	if netcatParam := query.Get("netcat"); netcatParam != "" {
		dialer.netcatBin = netcatParam
	}

	dialer.applyURIOptions(uri)

	return dialer
}

// applyURIOptions applies URI query parameters to configure the SSH connection.
// See: https://libvirt.org/uri.html#ssh-transport
func (d *SSHCmd) applyURIOptions(uri *url.URL) {
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
}

// Dial implements the Dialer interface by establishing an SSH connection.
func (d *SSHCmd) Dial() (net.Conn, error) {
	args := d.buildSSHArgs()

	//nolint:gosec
	cmd := exec.Command(d.sshBin, args...)

	var err error

	var stdout io.ReadCloser
	if stdout, err = cmd.StdoutPipe(); err != nil {
		return nil, fmt.Errorf("failed to acquire stdout pipe: %w", err)
	}

	var stdin io.WriteCloser
	if stdin, err = cmd.StdinPipe(); err != nil {
		return nil, fmt.Errorf("failed to acquire stdin pipe: %w", err)
	}

	var stderr io.ReadCloser
	if stderr, err = cmd.StderrPipe(); err != nil {
		return nil, fmt.Errorf("failed to acquire stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ssh command: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// custom net.Conn implementation that communicates with the ssh process
	conn := &sshCmdConn{
		cmd:             cmd,
		stdin:           stdin,
		stdout:          stdout,
		stderr:          stderr,
		cancel:          cancel,
		hostAndPort:     d.hostname,
		remoteSocket:    d.socket,
		lastStdErrLines: ring.New(5),
	}

	go func() {
		err := cmd.Wait()
		if err != nil {
			tflog.Error(ctx, "SSH command exited unexpectedly", map[string]any{"error": err.Error()})
		}
		cancel() // Ensure cleanup is triggered
	}()

	// Monitor the process in a goroutine
	go func() {
		defer cancel()
		<-ctx.Done()
		if cmd.Process != nil {
			if err := cmd.Process.Kill(); err != nil {
				tflog.Error(ctx, "Failed to kill ssh command", map[string]any{"error": err.Error()})
			}
		}
	}()

	// collect stderr to give context to any errors later
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			conn.appendStderrLine(scanner.Text())
			tflog.Warn(ctx, "SSH stderr", map[string]any{"message": scanner.Text()})
		}
	}()

	// Wait for initial connection (give ssh some time to establish the connection)
	//nolint:mnd
	time.Sleep(100 * time.Millisecond)
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		return nil, fmt.Errorf("ssh command terminated prematurely with exit code %d:\n%s",
			cmd.ProcessState.ExitCode(), strings.Join(conn.lastStderrLines(), "\n"))
	}

	return conn, nil
}

func (d *SSHCmd) buildSSHArgs() []string {
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

	case ProxyNetcat:
		// Netcat mode - detect proper flags for netcat
		//nolint:lll
		shellCmd = fmt.Sprintf("sh -c 'if \"%s\" -q 2>&1 | grep \"requires an argument\" >/dev/null 2>&1; then ARG=-q0; else ARG=; fi; \"%s\" $ARG -U %s'",
			d.netcatBin, d.netcatBin, d.socket)

	case ProxyAuto:
		// Auto mode - try virt-ssh-helper first, then fall back to netcat
		//nolint:lll
		shellCmd = fmt.Sprintf("sh -c 'which virt-ssh-helper 1>/dev/null 2>&1; if test $? = 0; then virt-ssh-helper \"%s\"; else if \"%s\" -q 2>&1 | grep \"requires an argument\" >/dev/null 2>&1; then ARG=-q0; else ARG=; fi; \"%s\" $ARG -U %s; fi'",
			d.remoteURI, d.netcatBin, d.netcatBin, d.socket)

	default:
		// This shouldn't happen, but use netcat as a safe fallback
		shellCmd = fmt.Sprintf("sh -c '\"%s\" -U %s'", d.netcatBin, d.socket)
	}

	args = append(args, "--", destination, shellCmd)

	return args
}

// sshCmdConn implements net.Conn to communicate with the ssh process.
type sshCmdConn struct {
	cmd          *exec.Cmd
	stdin        io.WriteCloser
	stdout       io.ReadCloser
	stderr       io.ReadCloser
	cancel       context.CancelFunc
	hostAndPort  string
	remoteSocket string

	lastStdErrLines *ring.Ring
	stderrRingMu    sync.Mutex
}

func (c *sshCmdConn) Read(b []byte) (int, error) {
	n, err := c.stdout.Read(b)
	if err != nil {
		return n, fmt.Errorf("ssh: %s", strings.Join(c.lastStderrLines(), "\n"))
	}
	return n, nil
}

func (c *sshCmdConn) Write(b []byte) (int, error) {
	n, err := c.stdin.Write(b)
	if err != nil {
		return n, fmt.Errorf("ssh: %s", strings.Join(c.lastStderrLines(), "\n"))
	}
	return n, nil
}

func (c *sshCmdConn) Close() error {
	c.cancel()
	_ = c.stdin.Close()
	_ = c.stdout.Close()
	_ = c.stderr.Close()
	return nil
}

func (c *sshCmdConn) lastStderrLines() []string {
	c.stderrRingMu.Lock()
	defer c.stderrRingMu.Unlock()

	var lines []string
	c.lastStdErrLines.Do(func(el any) {
		if el == nil {
			return
		}
		if str, ok := el.(string); ok {
			lines = append(lines, str)
		}
	})

	return lines
}

func (c *sshCmdConn) appendStderrLine(line string) {
	c.stderrRingMu.Lock()
	defer c.stderrRingMu.Unlock()

	c.lastStdErrLines.Value = line
	c.lastStdErrLines = c.lastStdErrLines.Next()
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
