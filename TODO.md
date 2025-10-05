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
**Status:** ❌ Not started
**Why:** VMs should start automatically by default

**Tasks:**
- [ ] Add `running` boolean attribute (default true)
- [ ] Start domain on create if running=true
- [ ] Handle state transitions on update
- [ ] Add helper methods for domain lifecycle
- [ ] Test running/stopped states

## Priority 2: Common Use Cases

### 4. Graphics Devices
**Status:** ❌ Not started
**Why:** Interactive VMs need display

**Tasks:**
- [ ] Add `graphics` block
- [ ] Support VNC type
- [ ] Support Spice type
- [ ] Add listen address, port, password
- [ ] Test with virt-viewer

### 5. Expand Disk Support
**Status:** ❌ Partially done (basic only)

**Tasks:**
- [ ] Add driver attributes (type, cache, io, discard)
- [ ] Add readonly, shareable flags
- [ ] Add boot order
- [ ] Support network disks (RBD, iSCSI)
- [ ] Add disk driver test

### 6. Essential Devices
**Status:** ❌ Not started

**Tasks:**
- [ ] Console/Serial devices
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

### 11. Network Resource
**Status:** ❌ Not started

**Tasks:**
- [ ] Define libvirt_network resource
- [ ] Schema (name, mode, bridge, addresses, dhcp)
- [ ] CRUD operations
- [ ] Tests

### 12. Pool Resource
**Status:** ❌ Not started

**Tasks:**
- [ ] Define libvirt_pool resource
- [ ] Schema (name, type, target path)
- [ ] CRUD operations
- [ ] Tests

### 13. Volume Resource
**Status:** ❌ Not started

**Tasks:**
- [ ] Define libvirt_volume resource
- [ ] Schema (name, pool, size, format)
- [ ] CRUD operations
- [ ] Integration with disk devices
- [ ] Tests

## Priority 5: Advanced Features

### 14. Security Labels
**Status:** ❌ Not started

**Tasks:**
- [ ] Add seclabel block (SELinux, AppArmor)

### 15. Host Device Passthrough
**Status:** ❌ Not started

**Tasks:**
- [ ] Add hostdev block (PCI, USB passthrough)

### 16. Advanced Tuning
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
