# TODO - Terraform Provider for Libvirt v2

## Current Status

### ✅ Completed - Domain Resource (Basic)

**Top-level fields:**
- Basic metadata: name, uuid, hwuuid, type, title, description
- Memory: memory, unit, current_memory, max_memory, max_memory_slots
- CPU: vcpu, iothreads
- Lifecycle: on_poweroff, on_reboot, on_crash
- Bootloader: bootloader, bootloader_args

**Blocks:**
- `os` - OS configuration (type, arch, machine, firmware, boot devices, kernel boot, UEFI loader)
- `features` - 20+ features (PAE, ACPI, APIC, HAP, PMU, VMPort, PVSpinlock, VMCoreInfo, HTM, NestedHV, CCFAssist, RAS, PS2, Viridian, PrivNet, IOAPIC, GIC)
- `cpu` - Basic CPU (mode, match, check, migratable, model, vendor)
- `clock` - Clock (offset, basis, adjustment, timezone) with **nested timer blocks** (name, track, tickpolicy, frequency, mode, present, catchup)
- `pm` - Power management (suspend_to_mem, suspend_to_disk)
- `disk` - Basic file-based disks (device, source, target, bus)
- `interface` - Network interfaces (network, bridge, user types with type-dependent source block)
- `create` - Domain creation flags (paused, autodestroy, bypass_cache, force_boot, validate, reset_nvram)
- `destroy` - Domain destruction options (graceful, timeout)

**State Management:**
- `running` - Boolean attribute to control whether domain should be running
- Proper state transitions during create, update, and delete
- Uses golibvirt constants for all state and flag operations

**Tests:**
- 12 acceptance tests covering feature groups
- Test that verifies VMs can actually boot
- Test that verifies update with running domain works
- Tests for complex nested blocks (clock timers with catchup, network interfaces)

## Priority 1: Critical for Usability

### 1. Update Method Implementation
**Status:** ✅ Completed

**Tasks:**
- [x] Implement Update method in domain_resource.go
- [x] Add state tracking (running/stopped)
- [x] Add graceful shutdown with force fallback before updates
- [x] Test update scenarios (including running domain updates)

### 2. Network Interfaces (Devices)
**Status:** ✅ Completed

**Tasks:**
- [x] Add `interface` block to domain schema
- [x] Implement basic types: network, bridge, user
- [x] Add MAC address, model fields
- [x] Conversion: model ↔ libvirtxml.DomainInterface
- [x] Add acceptance test
- [x] Add example configuration

### 3. State Management
**Status:** ✅ Completed

**Tasks:**
- [x] Add `running` boolean attribute
- [x] Add `create` block with flags (paused, autodestroy, bypass_cache, force_boot, validate, reset_nvram)
- [x] Add `destroy` block with graceful and timeout options
- [x] Start domain on create if running=true
- [x] Handle state transitions on update
- [x] Add helper methods for domain lifecycle (waitForDomainState)
- [x] Test running/stopped states
- [x] Use proper golibvirt constants instead of magic numbers

## Priority 2: Common Use Cases

### 4. Graphics Devices [OLD: ✅]
**Status:** ❌ Not started
**Why:** Interactive VMs need display

**Tasks:**
- [ ] Add `graphics` block
- [ ] Support VNC type
- [ ] Support Spice type
- [ ] Add listen address, port, password
- [ ] Test with virt-viewer

### 5. Expand Disk Support [OLD: ✅]
**Status:** ❌ Partially done (basic only)

**Tasks:**
- [ ] Add volume_id support (reference to libvirt_volume)
- [ ] Add URL source support
- [ ] Add block device support
- [ ] Add SCSI support (scsi flag, WWN)
- [ ] Add driver attributes (type, cache, io, discard)
- [ ] Add readonly, shareable flags
- [ ] Add boot order
- [ ] Support network disks (RBD, iSCSI)
- [ ] Add disk driver test

### 6. Essential Devices [OLD: ✅]
**Status:** ❌ Not started

**Tasks:**
- [ ] Console/Serial devices
- [ ] Video device
- [ ] RNG device (virtio-rng)
- [ ] Input devices (tablet for VNC)

## Priority 3: Advanced Domain Fields

### 7. CPU Enhancements
**Status:** ❌ Partially done (basic only)

**Tasks:**
- [ ] Add CPU topology (sockets, cores, threads)
- [ ] Add CPU features (enable/disable specific features)
- [ ] Add NUMA cell configuration
- [ ] Test topology scenarios

### 8. Memory Enhancements
**Status:** ❌ Partially done (basic only)

**Tasks:**
- [ ] Memory backing (hugepages, locked, source)
- [ ] Memory tune (hard_limit, soft_limit)
- [ ] Test hugepages

### 9. Features Enhancements
**Status:** ❌ Partially done (simple features only)

**Tasks:**
- [ ] Add HyperV feature block (for Windows)
- [ ] Add KVM feature block
- [ ] Add SMM with TSeg

### 10. Clock Timers
**Status:** ✅ Completed

**Tasks:**
- [x] Add timer blocks (rtc, pit, hpet, kvmclock, hypervclock, etc.)
- [x] Add nested catchup configuration
- [x] Test timer configurations

## Priority 4: Other Resources

### 11. Network Resource [OLD: ✅]
**Status:** ❌ Not started

**Tasks:**
- [ ] Define libvirt_network resource
- [ ] Schema (name, mode, bridge, addresses)
- [ ] Network modes (nat, isolated, route, open, bridge)
- [ ] DHCP configuration
- [ ] DNS configuration (hosts, forwarders, SRV records)
- [ ] Routes
- [ ] Dnsmasq options
- [ ] Autostart
- [ ] CRUD operations
- [ ] Tests

### 12. Pool Resource [OLD: ✅]
**Status:** ✅ Completed (~95% parity)

**Tasks:**
- [x] Define libvirt_pool resource
- [x] Schema (name, type, target path)
- [x] Pool types (dir, logical/LVM)
- [x] Target permissions (owner, group, mode, label)
- [x] Source (name, device for LVM)
- [x] CRUD operations
- [x] Tests
- [ ] XML XSLT transforms (not implementing - against design principles)

### 13. Volume Resource [OLD: ✅]
**Status:** ✅ Completed (~95% parity)

**Tasks:**
- [x] Define libvirt_volume resource
- [x] Schema (name, pool, capacity, format)
- [x] Type attribute (file, block, dir, network, netdir)
- [x] Format support (qcow2, raw)
- [x] Backing volumes (backing_store)
- [x] Permissions (owner, group, mode, label)
- [x] CRUD operations
- [x] Integration with disk devices (volume_id)
- [x] Tests
- [ ] URL download support (deferred - needs elegant redesign)
- [ ] XML XSLT transforms (not implementing - against design principles)

## Priority 5: Advanced Features

### 14. Additional Domain Features [OLD: ✅]
**Status:** ❌ Not started

**Tasks:**
- [ ] Metadata - Custom metadata XML
- [ ] Autostart - Domain autostart
- [ ] TPM device
- [ ] Filesystem (9p) - virtio-9p host directory sharing
- [ ] NVRAM template support
- [ ] QEMU agent integration
- [ ] XML XSLT transforms
- [ ] Network wait_for_lease
- [ ] Network macvtap/vepa/passthrough

### 15. Cloud-init Resources [OLD: ✅]
**Status:** ❌ Not started

**Tasks:**
- [ ] libvirt_cloudinit_disk resource
- [ ] libvirt_ignition resource (CoreOS)
- [ ] libvirt_combustion resource

### 16. Data Sources [OLD: ✅]
**Status:** ❌ Not started

**Tasks:**
- [ ] libvirt_node_info datasource
- [ ] libvirt_node_device_info datasource
- [ ] libvirt_node_devices datasource
- [ ] libvirt_network_dns_host_template datasource
- [ ] libvirt_network_dns_srv_template datasource
- [ ] libvirt_network_dnsmasq_options_template datasource

### 17. Provider Configuration [OLD: ✅]
**Status:** ❌ Not started

**Tasks:**
- [ ] SSH transport support (qemu+ssh://)
- [ ] Other URI schemes support

### 18. Security Labels
**Status:** ❌ Not started

**Tasks:**
- [ ] Add seclabel block (SELinux, AppArmor)

### 19. Test Suite Configuration
**Status:** ❌ Not started

**Tasks:**
- [ ] Centralize libvirt URI configuration for acceptance tests (avoid hardcoding `qemu:///system`; allow overriding when running the suite)
- [ ] Replace hardcoded `/tmp/...` paths in tests with suite-managed temporary directories

### 20. Host Device Passthrough
**Status:** ❌ Not started

**Tasks:**
- [ ] Add hostdev block (PCI, USB passthrough)

### 20. Advanced Tuning
**Status:** ❌ Not started

**Tasks:**
- [ ] CPUTune (vcpu pinning, shares, quota)
- [ ] NUMATune
- [ ] BlockIOTune
- [ ] IOThreadIDs with pinning

## Known Gaps (Low Priority)

See DOMAIN_COVERAGE_ANALYSIS.md for complete details.

**Domain fields not yet implemented:**
- GenID, Metadata (custom XML), Resource (cgroup partition)
- Perf events, KeyWrap, LaunchSecurity
- SysInfo (SMBIOS), IDMap (containers), ThrottleGroups
- Hypervisor namespaces (QEMU commandline, etc.)

**Device types not yet implemented:**
- Controllers, Leases, Filesystems, Smartcards
- Parallel/Serial ports, Channels, TPMs
- Sounds, Audios, Videos, Watchdogs
- MemBalloon, Panics, Shmems, Memorydevs
- IOMMU, VSock, Crypto, PStore

## Documentation Tasks

- [ ] Update README.md with current status
- [ ] Add examples for all implemented features
- [ ] Add migration guide from old provider
- [ ] Document breaking changes vs old provider

## Notes

- Incremental development: Add → Test → Commit → Repeat
- Follow libvirt XML schema closely (per AGENTS.md)
- Only support what libvirtxml supports
- User intent preservation for optional fields
- Keep commits small and focused
