# wait_for_ip and Interface Address Querying Example

This example demonstrates the new pattern for waiting for and retrieving IP addresses from libvirt domains.

## Design Overview

This implementation uses **two separate features** to handle IP address management:

### 1. `wait_for_ip` - Creation-time Blocking Behavior

The `wait_for_ip` block in the interface configuration controls blocking behavior **during resource creation**:

```hcl
devices = {
  interfaces = [
    {
      type = "network"
      source = { network = "default" }
      wait_for_ip = {
        timeout = 300    # seconds
        source  = "any"  # "lease", "agent", or "any"
      }
    }
  ]
}
```

**Behavior:**
- ✅ Blocks during **Create** until IP is assigned (or timeout)
  - Wait happens as the **last step** after domain is defined and started
  - If timeout occurs, domain is **destroyed and undefined** (cleanup)
  - Resource only enters Terraform state if wait succeeds
- ❌ Does **NOT** block during **Update** (machines typically get IPs at boot)
- ❌ Does **NOT** block during **Read**
- ❌ Does **NOT** block during **Delete** (avoids the old provider's bug)
- ❌ Does **NOT** store IP in state (use data source for that)

**Source options:**
- `"lease"` - Query DHCP server leases (fast, no guest agent needed)
- `"agent"` - Query QEMU guest agent (requires qemu-guest-agent installed in guest)
- `"any"` - Try both sources (default)

### 2. `libvirt_domain_interface_addresses` Data Source - IP Querying

The data source provides a **query API** that mirrors libvirt's `virDomainInterfaceAddresses` API:

```hcl
data "libvirt_domain_interface_addresses" "vm" {
  domain = libvirt_domain.vm.id
  source = "lease"  # optional
}
```

**Benefits:**
- Query IPs at any time without blocking operations
- Doesn't suffer from blocking during Delete
- Can be refreshed independently of the domain resource
- Follows the principle: "Query APIs → Data Sources"

**Output structure:**
```hcl
interfaces = [
  {
    name   = "vnet0"               # interface name on host
    hwaddr = "52:54:00:12:34:56"  # MAC address
    addrs = [
      {
        type   = "ipv4"            # or "ipv6"
        addr   = "192.168.122.100"
        prefix = 24
      }
    ]
  }
]
```

## Why This Design?

### Separation of Concerns
- **wait_for_ip** = behavior control (when to block)
- **data source** = query mechanism (how to get IPs)

### Avoids Old Provider Bugs
The original provider's `wait_for_lease` had a critical bug: it blocked during **Delete** operations, causing hangs when domains couldn't be reached. This design avoids that by:
1. Never blocking during Delete
2. Not storing IPs in resource state
3. Using a separate data source for queries

### Allows Future Flexibility
We can later add computed IP fields to the interface state if needed, without changing the query mechanism.

## Running This Example

1. Ensure libvirt is running and the default network exists
2. Initialize Terraform:
   ```bash
   terraform init
   ```

3. Apply the configuration:
   ```bash
   terraform apply
   ```

4. View the IP addresses:
   ```bash
   terraform output vm_ip
   terraform output vm_all_ips
   ```

## Timeout and Error Handling

### What Happens on Timeout?

If `wait_for_ip` times out, the provider performs **automatic cleanup**:

1. Destroys the running domain (if `running = true`)
2. Undefines the domain from libvirt
3. Returns an error to Terraform
4. Resource does **not** enter Terraform state

This ensures no orphaned domains are left in libvirt.

### Troubleshooting Timeouts

Common causes of timeout:
- **Domain fails to boot** - Check console/logs with `virsh console` or `virsh dominfo`
- **DHCP not configured** - Verify network has DHCP enabled
- **Guest networking not configured** - Cloud-init may have failed
- **Guest agent not running** - Required when `source = "agent"`

**Debugging tips:**
```bash
# Check if domain booted (you'll need to run this before timeout)
virsh list --all

# View domain console
virsh console alpine-vm

# Check network configuration
virsh net-dhcp-leases default

# Check guest agent (if using source="agent")
virsh qemu-agent-command alpine-vm '{"execute":"guest-ping"}'
```

### Adjusting Timeout

Increase timeout for slow-booting guests:
```hcl
wait_for_ip = {
  timeout = 600  # 10 minutes
  source  = "any"
}
```

## Implementation Status

This example represents the **target design**. Implementation order:

1. **Phase 1:** Implement `libvirt_domain_interface_addresses` data source
2. **Phase 2:** Add `wait_for_ip` support to interface configuration
3. **Phase 3 (optional):** Add computed IP fields to interface state

Currently, this example serves as a **reference design** for implementing these features.
