# Libvirt Connection Transports

The Terraform libvirt provider supports multiple transport mechanisms for connecting to libvirt daemons. This allows you to manage virtual machines on local systems, remote hosts via SSH, or over encrypted TLS connections.

## Supported Transports

### Local Unix Socket

Connect to a local libvirt daemon using Unix domain sockets.

**System daemon:**
```hcl
provider "libvirt" {
  uri = "qemu:///system"
}
```

**User session daemon:**
```hcl
provider "libvirt" {
  uri = "qemu:///session"
}
```

**Custom socket path:**
```hcl
provider "libvirt" {
  uri = "qemu:///system?socket=/var/run/libvirt/libvirt-sock-ro"
}
```

### SSH Transport (Go Library)

Connect to a remote libvirt daemon over SSH using Go's SSH library. This transport is efficient and handles authentication programmatically.

**Basic SSH connection:**
```hcl
provider "libvirt" {
  uri = "qemu+ssh://user@example.com/system"
}
```

**With custom port and key file:**
```hcl
provider "libvirt" {
  uri = "qemu+ssh://user@example.com:2222/system?keyfile=/home/user/.ssh/id_ed25519"
}
```

**With known hosts verification:**
```hcl
provider "libvirt" {
  uri = "qemu+ssh://user@example.com/system?knownhosts=/home/user/.ssh/known_hosts"
}
```

**Skip host key verification (not recommended for production):**
```hcl
provider "libvirt" {
  uri = "qemu+ssh://user@example.com/system?no_verify=1"
}
```

**Custom remote socket path:**
```hcl
provider "libvirt" {
  uri = "qemu+ssh://user@example.com/system?socket=/var/run/libvirt/libvirt-sock"
}
```

### SSH Transport (Native Command)

Connect to a remote libvirt daemon using the native SSH command-line tool. This transport respects your `~/.ssh/config` settings including ProxyJump, ControlMaster, and other SSH configuration options.

**Basic SSH command connection:**
```hcl
provider "libvirt" {
  uri = "qemu+sshcmd://user@example.com/system"
}
```

**With SSH config support:**
```hcl
# ~/.ssh/config:
# Host libvirt-prod
#   HostName example.com
#   User terraform
#   IdentityFile ~/.ssh/prod_key
#   ProxyJump bastion.example.com

provider "libvirt" {
  uri = "qemu+sshcmd://libvirt-prod/system"
}
```

**With custom proxy mode:**
```hcl
# Auto mode: tries virt-ssh-helper, falls back to netcat (default)
provider "libvirt" {
  uri = "qemu+sshcmd://user@example.com/system?proxy=auto"
}

# Native mode: uses virt-ssh-helper
provider "libvirt" {
  uri = "qemu+sshcmd://user@example.com/system?proxy=native"
}

# Netcat mode: uses netcat to connect to Unix socket
provider "libvirt" {
  uri = "qemu+sshcmd://user@example.com/system?proxy=netcat"
}
```

**With custom authentication methods:**
```hcl
# Use SSH agent and private key authentication
provider "libvirt" {
  uri = "qemu+sshcmd://user@example.com/system?sshauth=agent,privkey"
}

# Allow interactive password authentication
provider "libvirt" {
  uri = "qemu+sshcmd://user@example.com/system?sshauth=password"
}
```

**All SSH parameters:**
```hcl
provider "libvirt" {
  uri = "qemu+sshcmd://user@example.com:2222/system?keyfile=/home/user/.ssh/id_rsa&knownhosts=/home/user/.ssh/known_hosts&socket=/var/run/libvirt/libvirt-sock&proxy=auto&netcat=/usr/bin/nc"
}
```

### TCP Transport

Connect to a remote libvirt daemon over an unencrypted TCP connection. Only use this on trusted networks.

**Basic TCP connection:**
```hcl
provider "libvirt" {
  uri = "qemu+tcp://example.com/system"
}
```

**With custom port:**
```hcl
provider "libvirt" {
  uri = "qemu+tcp://example.com:16509/system"
}
```

### TLS Transport

Connect to a remote libvirt daemon over an encrypted TLS connection with certificate authentication.

**Basic TLS connection:**
```hcl
provider "libvirt" {
  uri = "qemu+tls://example.com/system"
}
```

**With custom PKI path:**
```hcl
provider "libvirt" {
  uri = "qemu+tls://example.com/system?pkipath=/etc/pki/libvirt"
}
```

**With custom port:**
```hcl
provider "libvirt" {
  uri = "qemu+tls://example.com:16514/system"
}
```

**Skip certificate verification (not recommended for production):**
```hcl
provider "libvirt" {
  uri = "qemu+tls://example.com/system?no_verify=1"
}
```

## URI Query Parameters

### Common Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `socket` | Path to Unix socket on remote system | `socket=/var/run/libvirt/libvirt-sock` |
| `no_verify` | Skip host key/certificate verification | `no_verify=1` |

### SSH-Specific Parameters (ssh and sshcmd)

| Parameter | Description | Transport | Example |
|-----------|-------------|-----------|---------|
| `keyfile` | Path to SSH private key | Both | `keyfile=~/.ssh/id_ed25519` |
| `knownhosts` | Path to known_hosts file | Both | `knownhosts=~/.ssh/known_hosts` |
| `known_hosts_verify` | Host key verification mode | Both | `known_hosts_verify=ignore` |
| `sshauth` | Authentication methods (comma-separated) | sshcmd | `sshauth=agent,privkey` |
| `proxy` | Proxy mode (auto/native/netcat) | sshcmd | `proxy=native` |
| `netcat` | Netcat binary path | sshcmd | `netcat=/usr/bin/nc` |
| `command` | SSH binary path | sshcmd | `command=/usr/bin/ssh` |

### TLS-Specific Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `pkipath` | Path to PKI certificates directory | `pkipath=/etc/pki/libvirt` |

## Choosing Between SSH Transports

The provider offers two SSH transport options:

### qemu+ssh:// (Go SSH Library)
- **Pros:**
  - Pure Go implementation, no external dependencies
  - Consistent behavior across platforms
  - Programmatic authentication handling
- **Cons:**
  - Does not respect `~/.ssh/config` settings
  - Limited to features implemented in Go's SSH library
  - No support for ProxyJump or advanced SSH features

### qemu+sshcmd:// (Native SSH Command)
- **Pros:**
  - Respects all `~/.ssh/config` settings (ProxyJump, ControlMaster, etc.)
  - Supports all SSH features available on your system
  - Familiar behavior for users accustomed to SSH command
- **Cons:**
  - Requires `ssh` binary to be installed
  - Requires `nc` (netcat) or `virt-ssh-helper` on remote system
  - Slightly more overhead due to process spawning

**Recommendation:** Use `qemu+sshcmd://` if you:
- Have complex SSH configurations (bastion hosts, proxy jumps)
- Need to respect existing SSH config files
- Use SSH agent forwarding or advanced SSH features

Use `qemu+ssh://` if you:
- Want a pure Go solution without external dependencies
- Have simple, straightforward SSH authentication
- Need maximum portability

## Environment Variables

The provider respects standard libvirt environment variables:

```bash
# Default URI if not specified in provider configuration
export LIBVIRT_DEFAULT_URI="qemu+ssh://user@example.com/system"

# SSH authentication socket (for SSH agent)
export SSH_AUTH_SOCK="/run/user/1000/keyring/ssh"
```

## Examples

### Multi-Region Infrastructure

Manage VMs across multiple hosts using different providers:

```hcl
# Local development environment
provider "libvirt" {
  alias = "local"
  uri   = "qemu:///system"
}

# Production environment via SSH
provider "libvirt" {
  alias = "prod"
  uri   = "qemu+sshcmd://terraform@prod-host.example.com/system"
}

# Staging environment via SSH with custom key
provider "libvirt" {
  alias = "staging"
  uri   = "qemu+ssh://terraform@staging-host.example.com/system?keyfile=~/.ssh/staging_key"
}

resource "libvirt_domain" "local_vm" {
  provider = libvirt.local
  name     = "dev-vm"
  # ...
}

resource "libvirt_domain" "prod_vm" {
  provider = libvirt.prod
  name     = "prod-vm"
  # ...
}
```

### SSH with Bastion Host

Using native SSH command with ProxyJump:

```hcl
# ~/.ssh/config:
# Host prod-libvirt
#   HostName 10.0.1.100
#   User libvirt-admin
#   ProxyJump bastion.example.com
#   IdentityFile ~/.ssh/prod_key

provider "libvirt" {
  uri = "qemu+sshcmd://prod-libvirt/system"
}
```

### TLS with Custom Certificates

```hcl
provider "libvirt" {
  uri = "qemu+tls://libvirt.example.com/system?pkipath=/etc/pki/libvirt-custom"
}
```

## Troubleshooting

### SSH Connection Issues

**"Permission denied" errors:**
- Verify SSH key permissions: `chmod 600 ~/.ssh/id_rsa`
- Test SSH connection manually: `ssh user@host`
- Check SSH agent: `ssh-add -l`

**"Host key verification failed":**
- Add host to known_hosts: `ssh-keyscan example.com >> ~/.ssh/known_hosts`
- Or use `no_verify=1` (not recommended for production)

**sshcmd "netcat not found":**
- Install netcat on remote system: `apt install netcat-openbsd` or `yum install nmap-ncat`
- Or install virt-ssh-helper (included in libvirt-client package)
- Or specify netcat path: `netcat=/bin/nc`

### TLS Connection Issues

**Certificate verification errors:**
- Ensure PKI certificates are properly configured
- Check certificate paths: `/etc/pki/libvirt/clientcert.pem`
- Verify CA trust: `/etc/pki/CA/cacert.pem`

### Connection Timeout

**Slow or timeout connections:**
- Check firewall rules on remote host
- Verify libvirtd is running: `systemctl status libvirtd`
- Test connectivity: `nc -zv example.com 16509` (TCP) or `16514` (TLS)

## Security Considerations

1. **Never use `no_verify` in production** - Always verify host keys and TLS certificates
2. **Use TLS for public networks** - Encrypt traffic with `qemu+tls://` instead of `qemu+tcp://`
3. **Rotate SSH keys regularly** - Use short-lived certificates or rotate keys periodically
4. **Use SSH agent forwarding carefully** - Only forward agent to trusted hosts
5. **Restrict libvirt socket access** - Ensure Unix socket permissions are properly configured
6. **Use bastion hosts** - Access production systems through hardened jump hosts

## See Also

- [Libvirt URI Documentation](https://libvirt.org/uri.html)
- [SSH Config Documentation](https://man.openbsd.org/ssh_config)
- [Provider Configuration](./index.md)
