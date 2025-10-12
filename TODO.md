# TODO - Terraform Provider for Libvirt v2

## Technical Debt - Block to Nested Attribute Conversion

**Background:** Several fields were incorrectly implemented using blocks instead of nested attributes. Per [HashiCorp's guidance](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/blocks), blocks are primarily for SDK v2 migration compatibility. New providers should use nested attributes.

**Impact:** These currently work but violate framework best practices and create inconsistency in the schema design.

**Fields needing conversion:**

1. **`os` block → `os` nested attribute**
   - Status: ❌ Not started
   - Complexity: Medium (simple single-level structure)
   - Breaking change: Yes (users must change `os { }` to `os = { }`)

2. **`features` block → `features` nested attribute**
   - Status: ❌ Not started
   - Complexity: Medium (many boolean fields)
   - Breaking change: Yes

3. **`cpu` block → `cpu` nested attribute**
   - Status: ❌ Not started
   - Complexity: Low (few fields currently)
   - Breaking change: Yes

4. **`clock` block with `timer` nested blocks → `clock` nested attribute with `timers` list**
   - Status: ❌ Not started
   - Complexity: High (nested blocks within blocks, including `catchup` sub-block)
   - Breaking change: Yes (also renames `timer` to `timers`)
   - Example change: `clock { timer { catchup { } } }` → `clock = { timers = [{ catchup = { } }] }`

5. **`pm` block → `pm` nested attribute**
   - Status: ❌ Not started
   - Complexity: Low (only 2 fields)
   - Breaking change: Yes

6. **`create` block → `create` nested attribute**
   - Status: ❌ Not started
   - Complexity: Low (boolean flags only)
   - Breaking change: Yes

7. **`destroy` block → `destroy` nested attribute**
   - Status: ❌ Not started
   - Complexity: Low (2 fields)
   - Breaking change: Yes

**Migration strategy:**
- Since this is a v2 rewrite and pre-1.0, we can make breaking changes
- Consider doing all conversions in a single PR to minimize user disruption
- Update all examples and tests simultaneously
- Document the breaking changes clearly
- Priority: Should be done before 1.0 release, ideally sooner to reduce rework

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

## Priority 2: Old Provider Parity (Pure Libvirt Features)

These features are supported by libvirtxml, were present in the old provider, and are NOT provider additions on top of libvirt (no cloudinit, URL download, etc.)

### 4. Graphics Devices
**Status:** ✅ Completed

**Tasks:**
- [x] Add `graphics` nested attribute to devices
- [x] Support VNC type (socket, port, autoport, websocket, listen)
- [x] Support Spice type (port, tlsport, autoport, listen)
- [x] Model ↔ libvirtxml conversion
- [x] Acceptance test
- [x] Documentation

### 5. Domain Autostart
**Status:** ✅ Completed
**libvirtxml:** libvirt API (DomainSetAutostart/DomainGetAutostart)
**Old provider:** `autostart` boolean field

**Tasks:**
- [x] Add `autostart` boolean attribute to domain schema
- [x] Call DomainSetAutostart() in Create/Update
- [x] Call DomainGetAutostart() in Read
- [x] Add acceptance test
- [x] Update examples

### 6. Filesystem (virtio-9p)
**Status:** ✅ Completed
**libvirtxml:** `DomainFilesystem`
**Old provider:** `filesystem` block

**Tasks:**
- [x] Add filesystem nested attribute to devices
- [x] Fields: accessmode, source, target, readonly
- [x] Model ↔ libvirtxml.DomainFilesystem conversion
- [x] Add acceptance test (using t.TempDir())
- [x] Update examples

### 7. SCSI Disks with WWN
**Status:** ✅ Completed
**libvirtxml:** `DomainDisk` (Bus="scsi", WWN field)
**Old provider:** `disk.scsi` boolean + `disk.wwn`

**Tasks:**
- [x] Add `wwn` field to disk model (optional, computed)
- [x] Add WWN to XML ↔ Model conversion
- [x] Add acceptance test with SCSI disk

### 8. Block Device Disks
**Status:** ❌ Not started
**libvirtxml:** `DomainDisk` with type="block"
**Old provider:** `disk.block_device` field

**Tasks:**
- [ ] Support type="block" in disk source
- [ ] Add acceptance test with block device

### 9. Console/Serial Devices
**Status:** ✅ Completed
**libvirtxml:** `DomainConsole`, `DomainSerial`
**Old provider:** `console` block

**Tasks:**
- [x] Add console/serial nested attributes to devices
- [x] Fields: type, source_path, target_port, target_type (pty and file types supported)
- [x] Model ↔ libvirtxml conversion
- [x] Add acceptance test

### 10. Video Device
**Status:** ✅ Completed
**libvirtxml:** `DomainVideo`
**Old provider:** `video` block with type

**Tasks:**
- [x] Add video nested attribute to devices
- [x] Fields: type (model)
- [x] Model ↔ libvirtxml.DomainVideo conversion
- [x] Add acceptance test

### 11. Direct Network Types (macvtap, vepa, passthrough)
**Status:** ❌ Not started
**libvirtxml:** `DomainInterface` with type="direct"
**Old provider:** `network_interface` with vepa, macvtap, passthrough, private

**Tasks:**
- [ ] Add type="direct" support to interface
- [ ] Add source.dev and source.mode fields
- [ ] Update conversion for direct types
- [ ] Add acceptance test

### 12. NVRAM Template
**Status:** ⚠️ Partial (only file/path, not template)
**libvirtxml:** `DomainLoader.Template`
**Old provider:** `nvram.template` field

**Tasks:**
- [ ] Add nvram_template field to os block
- [ ] Update conversion to set loader template
- [ ] Add acceptance test with UEFI

### 13. TPM Device
**Status:** ❌ Not started
**libvirtxml:** `DomainTPM`
**Old provider:** `tpm` block

**Tasks:**
- [ ] Add tpm nested attribute to devices
- [ ] Fields: model, backend_type, backend_device_path, backend_encryption_secret, backend_version, backend_persistent_state
- [ ] Model ↔ libvirtxml.DomainTPM conversion
- [ ] Add acceptance test

### 14. Emulator Path
**Status:** ✅ Completed
**libvirtxml:** `DomainDeviceList.Emulator`
**Old provider:** `emulator` field

**Tasks:**
- [x] Add emulator string attribute to devices
- [x] Update conversion to set emulator path
- [x] Add acceptance test

### 15. Metadata (Custom XML)
**Status:** ❌ Not started
**libvirtxml:** `Domain.Metadata`
**Old provider:** `metadata` string field

**Tasks:**
- [ ] Add metadata string attribute to domain
- [ ] Update conversion to handle custom metadata XML
- [ ] Add acceptance test

## Priority 3: Additional Disk/Network Enhancements

### 16. Disk Driver Attributes (NEW - not in old provider)
**Status:** ❌ Not started
**libvirtxml:** `DomainDiskDriver` fields

**Tasks:**
- [ ] Add driver nested attribute (cache, io, discard, type) - **HIGH VALUE**
- [ ] Add readonly flag (for CD-ROMs)
- [ ] Add shareable flag (for shared storage)
- [ ] Add boot order attribute
- [ ] Support network disks (RBD, iSCSI)

### 17. RNG Device (virtio-rng)
**Status:** ❌ Not started
**libvirtxml:** `DomainRNG`
**Old provider:** Added by default

**Tasks:**
- [ ] Add RNG nested attribute to devices
- [ ] Fields: model, backend (random device path)
- [ ] Add acceptance test

### 18. Input Devices
**Status:** ❌ Not started
**libvirtxml:** `DomainInput`

**Tasks:**
- [ ] Add input devices (tablet for VNC)
- [ ] Add acceptance test

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

## Priority 3.5: Provider Infrastructure

### Machine Type Diff Logic
**Status:** ❌ Not started
**Issue:** Currently preserving user's machine value (e.g., "q35") in state to avoid diffs when libvirt expands it (e.g., to "pc-q35-10.1"). This works but storing unexpanded value isn't ideal.
**Better approach:** Implement sophisticated diff logic that understands libvirt's machine type expansion.

**Tasks:**
- [ ] Research Terraform Plugin Framework diff suppression patterns
- [ ] Implement machine type normalization/comparison logic
- [ ] Store actual libvirt value but suppress diff when it's just expansion
- [ ] Update tests

## Priority 4: Other Resources

### 11. Network Resource [OLD: ✅]
**Status:** ✅ Completed (basic modes: NAT and isolated)

**Tasks:**
- [x] Define libvirt_network resource
- [x] Schema (name, mode, bridge, addresses)
- [x] Network modes: nat, none (isolated)
- [x] IP address configuration (CIDR)
- [x] Autostart support
- [x] CRUD operations
- [x] Acceptance tests
- [x] Documentation
- [ ] Advanced modes: route, open, bridge (deferred)
- [ ] DHCP ranges and static hosts (deferred)
- [ ] DNS configuration (deferred)
- [ ] Routes (deferred)
- [ ] Dnsmasq options (deferred)

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
- [ ] URL download support - **SPEC READY**
- [ ] XML XSLT transforms (not implementing - against design principles)

**URL Download Spec (for libvirt_volume):**
- Detect URLs in existing `source` attribute (http://, https://)
- Require `capacity` and `format` when source is URL (no HTTP HEAD, no auto-detection)
- Stream HTTP GET → StorageVolUpload (works with remote libvirt)
- Simple retry logic: 3 attempts, 2s exponential backoff, retry on 5xx/network errors
- ForceNew on source changes (no re-download/update support)
- Implementation: ~50-80 lines in volume_resource.go
- No caching, no If-Modified-Since (Terraform state prevents unnecessary recreates)
- Example:
  ```hcl
  resource "libvirt_volume" "base" {
    name     = "ubuntu.qcow2"
    pool     = "default"
    source   = "https://cloud-images.ubuntu.com/.../ubuntu.img"
    capacity = 2361393152  # required for URLs
    format   = "qcow2"     # required for URLs
  }
  ```

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
