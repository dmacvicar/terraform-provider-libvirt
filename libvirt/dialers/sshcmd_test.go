package dialers

import (
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestNewSSHCmdDialer(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected *SSHCmdDialer
	}{
		{
			name: "basic ssh uri",
			uri:  "qemu+ssh://user@example.com/system",
			expected: &SSHCmdDialer{
				hostname:        "example.com",
				port:            "",
				username:        "user",
				socket:          defaultUnixSock,
				remoteURI:       "qemu:///system",
				proxyMode:       ProxyAuto,
				netcatBin:       defaultNetcatBin,
				sshBin:          defaultSSHBin,
				strictHostCheck: true,
				batchMode:       true,
				forwardAgent:    false,
				authMethods:     []string{},
				keyFiles:        []string{},
			},
		},
		{
			name: "ssh uri with port",
			uri:  "qemu+ssh://user@example.com:2222/system",
			expected: &SSHCmdDialer{
				hostname:        "example.com",
				port:            "2222",
				username:        "user",
				socket:          defaultUnixSock,
				remoteURI:       "qemu:///system",
				proxyMode:       ProxyAuto,
				netcatBin:       defaultNetcatBin,
				sshBin:          defaultSSHBin,
				strictHostCheck: true,
				batchMode:       true,
				forwardAgent:    false,
				authMethods:     []string{},
				keyFiles:        []string{},
			},
		},
		{
			name: "ssh uri with custom socket",
			uri:  "qemu+ssh://user@example.com/system?socket=/tmp/libvirt.sock",
			expected: &SSHCmdDialer{
				hostname:        "example.com",
				port:            "",
				username:        "user",
				socket:          "/tmp/libvirt.sock",
				remoteURI:       "qemu:///system",
				proxyMode:       ProxyAuto,
				netcatBin:       defaultNetcatBin,
				sshBin:          defaultSSHBin,
				strictHostCheck: true,
				batchMode:       true,
				forwardAgent:    false,
				authMethods:     []string{},
				keyFiles:        []string{},
			},
		},
		{
			name: "ssh uri with proxy mode",
			uri:  "qemu+ssh://user@example.com/system?proxy=netcat",
			expected: &SSHCmdDialer{
				hostname:        "example.com",
				port:            "",
				username:        "user",
				socket:          defaultUnixSock,
				remoteURI:       "qemu:///system",
				proxyMode:       ProxyNetcat,
				netcatBin:       defaultNetcatBin,
				sshBin:          defaultSSHBin,
				strictHostCheck: true,
				batchMode:       true,
				forwardAgent:    false,
				authMethods:     []string{},
				keyFiles:        []string{},
			},
		},
		{
			name: "ssh uri with keyfile",
			uri:  "qemu+ssh://user@example.com/system?keyfile=/home/user/.ssh/id_rsa",
			expected: &SSHCmdDialer{
				hostname:        "example.com",
				port:            "",
				username:        "user",
				socket:          defaultUnixSock,
				remoteURI:       "qemu:///system",
				proxyMode:       ProxyAuto,
				netcatBin:       defaultNetcatBin,
				sshBin:          defaultSSHBin,
				strictHostCheck: true,
				batchMode:       true,
				forwardAgent:    false,
				authMethods:     []string{},
				keyFiles:        []string{"/home/user/.ssh/id_rsa"},
			},
		},
		{
			name: "ssh uri with known hosts settings",
			uri:  "qemu+ssh://user@example.com/system?knownhosts=/home/user/.ssh/known_hosts&known_hosts_verify=ignore",
			expected: &SSHCmdDialer{
				hostname:        "example.com",
				port:            "",
				username:        "user",
				socket:          defaultUnixSock,
				remoteURI:       "qemu:///system",
				proxyMode:       ProxyAuto,
				netcatBin:       defaultNetcatBin,
				sshBin:          defaultSSHBin,
				strictHostCheck: false,
				batchMode:       true,
				forwardAgent:    false,
				authMethods:     []string{},
				keyFiles:        []string{},
				knownHostsFile:  "/home/user/.ssh/known_hosts",
			},
		},
		{
			name: "ssh uri with auth methods",
			uri:  "qemu+ssh://user@example.com/system?sshauth=agent,password",
			expected: &SSHCmdDialer{
				hostname:        "example.com",
				port:            "",
				username:        "user",
				socket:          defaultUnixSock,
				remoteURI:       "qemu:///system",
				proxyMode:       ProxyAuto,
				netcatBin:       defaultNetcatBin,
				sshBin:          defaultSSHBin,
				strictHostCheck: true,
				batchMode:       false, // Disabled for password auth
				forwardAgent:    true,  // Enabled for agent auth
				authMethods:     []string{"agent", "password"},
				keyFiles:        []string{},
			},
		},
	}

	// Save original environment to restore later
	origEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range origEnv {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the URI
			parsedURI, err := url.Parse(tt.uri)
			if err != nil {
				t.Fatalf("Failed to parse URI: %v", err)
			}

			// Create the dialer
			dialer := NewSSHCmdDialer(parsedURI)

			// Compare fields that we care about
			if dialer.hostname != tt.expected.hostname {
				t.Errorf("hostname = %v, want %v", dialer.hostname, tt.expected.hostname)
			}
			if dialer.port != tt.expected.port {
				t.Errorf("port = %v, want %v", dialer.port, tt.expected.port)
			}
			if dialer.username != tt.expected.username {
				t.Errorf("username = %v, want %v", dialer.username, tt.expected.username)
			}
			if dialer.socket != tt.expected.socket {
				t.Errorf("socket = %v, want %v", dialer.socket, tt.expected.socket)
			}
			if dialer.remoteURI != tt.expected.remoteURI {
				t.Errorf("remoteURI = %v, want %v", dialer.remoteURI, tt.expected.remoteURI)
			}
			if dialer.proxyMode != tt.expected.proxyMode {
				t.Errorf("proxyMode = %v, want %v", dialer.proxyMode, tt.expected.proxyMode)
			}
			if dialer.netcatBin != tt.expected.netcatBin {
				t.Errorf("netcatBin = %v, want %v", dialer.netcatBin, tt.expected.netcatBin)
			}
			if dialer.sshBin != tt.expected.sshBin {
				t.Errorf("sshBin = %v, want %v", dialer.sshBin, tt.expected.sshBin)
			}
			if dialer.strictHostCheck != tt.expected.strictHostCheck {
				t.Errorf("strictHostCheck = %v, want %v", dialer.strictHostCheck, tt.expected.strictHostCheck)
			}
			if dialer.batchMode != tt.expected.batchMode {
				t.Errorf("batchMode = %v, want %v", dialer.batchMode, tt.expected.batchMode)
			}
			if dialer.forwardAgent != tt.expected.forwardAgent {
				t.Errorf("forwardAgent = %v, want %v", dialer.forwardAgent, tt.expected.forwardAgent)
			}
			if !reflect.DeepEqual(dialer.authMethods, tt.expected.authMethods) {
				t.Errorf("authMethods = %v, want %v", dialer.authMethods, tt.expected.authMethods)
			}
			if !reflect.DeepEqual(dialer.keyFiles, tt.expected.keyFiles) {
				t.Errorf("keyFiles = %v, want %v", dialer.keyFiles, tt.expected.keyFiles)
			}
		})
	}
}

func TestBuildSSHArgs(t *testing.T) {
	tests := []struct {
		name     string
		dialer   *SSHCmdDialer
		expected []string // Complete expected command arguments
	}{
		{
			name: "basic args",
			dialer: &SSHCmdDialer{
				hostname:  "example.com",
				username:  "user",
				proxyMode: ProxyAuto,
				remoteURI: "qemu:///system",
			},
			expected: []string{
				"-T",
				"-o", "ControlPath=none",
				"-e", "none",
				"-o", "StrictHostKeyChecking=no",
				"--",
				"user@example.com",
				"sh -c 'which virt-ssh-helper 1>/dev/null 2>&1; if test $? = 0; then virt-ssh-helper \"qemu:///system\"; else if \"\" -q 2>&1 | grep \"requires an argument\" >/dev/null 2>&1; then ARG=-q0; else ARG=; fi; \"\" $ARG -U ; fi'",
			},
		},
		{
			name: "with port",
			dialer: &SSHCmdDialer{
				hostname:  "example.com",
				port:      "2222",
				username:  "user",
				proxyMode: ProxyAuto,
				remoteURI: "qemu:///system",
			},
			expected: []string{
				"-p", "2222",
				"-T",
				"-o", "ControlPath=none",
				"-e", "none",
				"-o", "StrictHostKeyChecking=no",
				"--",
				"user@example.com",
				"sh -c 'which virt-ssh-helper 1>/dev/null 2>&1; if test $? = 0; then virt-ssh-helper \"qemu:///system\"; else if \"\" -q 2>&1 | grep \"requires an argument\" >/dev/null 2>&1; then ARG=-q0; else ARG=; fi; \"\" $ARG -U ; fi'",
			},
		},
		{
			name: "with keyfile",
			dialer: &SSHCmdDialer{
				hostname:  "example.com",
				username:  "user",
				proxyMode: ProxyAuto,
				remoteURI: "qemu:///system",
				keyFiles:  []string{"/home/user/.ssh/id_rsa"},
			},
			expected: []string{
				"-T",
				"-o", "ControlPath=none",
				"-e", "none",
				"-i", "/home/user/.ssh/id_rsa",
				"-o", "StrictHostKeyChecking=no",
				"--",
				"user@example.com",
				"sh -c 'which virt-ssh-helper 1>/dev/null 2>&1; if test $? = 0; then virt-ssh-helper \"qemu:///system\"; else if \"\" -q 2>&1 | grep \"requires an argument\" >/dev/null 2>&1; then ARG=-q0; else ARG=; fi; \"\" $ARG -U ; fi'",
			},
		},
		{
			name: "with known hosts settings",
			dialer: &SSHCmdDialer{
				hostname:        "example.com",
				username:        "user",
				proxyMode:       ProxyAuto,
				remoteURI:       "qemu:///system",
				knownHostsFile:  "/home/user/.ssh/known_hosts",
				strictHostCheck: false,
			},
			expected: []string{
				"-T",
				"-o", "ControlPath=none",
				"-e", "none",
				"-o", "UserKnownHostsFile=/home/user/.ssh/known_hosts",
				"-o", "StrictHostKeyChecking=no",
				"--",
				"user@example.com",
				"sh -c 'which virt-ssh-helper 1>/dev/null 2>&1; if test $? = 0; then virt-ssh-helper \"qemu:///system\"; else if \"\" -q 2>&1 | grep \"requires an argument\" >/dev/null 2>&1; then ARG=-q0; else ARG=; fi; \"\" $ARG -U ; fi'",
			},
		},
		{
			name: "with batch mode disabled",
			dialer: &SSHCmdDialer{
				hostname:  "example.com",
				username:  "user",
				proxyMode: ProxyAuto,
				remoteURI: "qemu:///system",
				batchMode: false,
			},
			expected: []string{
				"-T",
				"-o", "ControlPath=none",
				"-e", "none",
				"-o", "StrictHostKeyChecking=no",
				"--",
				"user@example.com",
				"sh -c 'which virt-ssh-helper 1>/dev/null 2>&1; if test $? = 0; then virt-ssh-helper \"qemu:///system\"; else if \"\" -q 2>&1 | grep \"requires an argument\" >/dev/null 2>&1; then ARG=-q0; else ARG=; fi; \"\" $ARG -U ; fi'",
			},
		},
		{
			name: "with agent forwarding",
			dialer: &SSHCmdDialer{
				hostname:     "example.com",
				username:     "user",
				proxyMode:    ProxyAuto,
				remoteURI:    "qemu:///system",
				forwardAgent: true,
			},
			expected: []string{
				"-T",
				"-o", "ControlPath=none",
				"-e", "none",
				"-o", "StrictHostKeyChecking=no",
				"-o", "ForwardAgent=yes",
				"--",
				"user@example.com",
				"sh -c 'which virt-ssh-helper 1>/dev/null 2>&1; if test $? = 0; then virt-ssh-helper \"qemu:///system\"; else if \"\" -q 2>&1 | grep \"requires an argument\" >/dev/null 2>&1; then ARG=-q0; else ARG=; fi; \"\" $ARG -U ; fi'",
			},
		},
		{
			name: "with netcat proxy mode",
			dialer: &SSHCmdDialer{
				hostname:  "example.com",
				username:  "user",
				proxyMode: ProxyNetcat,
				remoteURI: "qemu:///system",
				socket:    "/var/run/libvirt/libvirt-sock",
				netcatBin: "nc",
			},
			expected: []string{
				"-T",
				"-o", "ControlPath=none",
				"-e", "none",
				"-o", "StrictHostKeyChecking=no",
				"--",
				"user@example.com",
				"sh -c 'if \"nc\" -q 2>&1 | grep \"requires an argument\" >/dev/null 2>&1; then ARG=-q0; else ARG=; fi; \"nc\" $ARG -U /var/run/libvirt/libvirt-sock'",
			},
		},
		{
			name: "with native proxy mode",
			dialer: &SSHCmdDialer{
				hostname:  "example.com",
				username:  "user",
				proxyMode: ProxyNative,
				remoteURI: "qemu:///system",
			},
			expected: []string{
				"-T",
				"-o", "ControlPath=none",
				"-e", "none",
				"-o", "StrictHostKeyChecking=no",
				"--",
				"user@example.com",
				"sh -c 'virt-ssh-helper \"qemu:///system\"'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.dialer.buildSSHArgs()

			// Check if the number of arguments matches
			if len(args) != len(tt.expected) {
				t.Errorf("Expected %d arguments, got %d", len(tt.expected), len(args))
				t.Errorf("Expected: %v", tt.expected)
				t.Errorf("Got: %v", args)
				return
			}

			// Check each argument
			for i, arg := range tt.expected {
				if i >= len(args) {
					t.Errorf("Missing argument at position %d, expected '%s'", i, arg)
					continue
				}

				if args[i] != arg {
					t.Errorf("Argument mismatch at position %d: expected '%s', got '%s'", i, arg, args[i])
				}
			}
		})
	}
}
