# TODO - Terraform Provider for Libvirt v2

## Executive Summary

### ✅ Core Resources - ALL COMPLETE
- **libvirt_domain** - Virtual machines (full feature parity with old provider for pure libvirt features)
- **libvirt_volume** - Storage volumes (100% feature parity with old provider)
- **libvirt_pool** - Storage pools (95% parity)
- **libvirt_network** - Networks (basic NAT/isolated modes complete, DHCP/DNS/routes deferred)

### 🟡 Pure Libvirt Features Still Missing
- **Data Sources**: node_info, node_device_info, node_devices (Priority 2)
- **Advanced Domain**: CPU topology/features, memory backing/hugepages, HyperV/KVM features, input devices
- **Advanced Network**: DHCP ranges, DNS config, static routes, network modes (route/open/bridge)
- **Advanced Disk**: driver options (cache, io, discard), readonly/shareable flags, boot order, network disks

### 🔴 Provider Conveniences (Deferred - "On Top" of Libvirt)
- **Cloud-init Resources**: libvirt_cloudinit_disk, libvirt_ignition, libvirt_combustion
- **Template Data Sources**: dns_host_template, dns_srv_template, dnsmasq_options_template
- **XSLT Transforms**: Custom XML transformations
- **Provider Config**: SSH transport (qemu+ssh://)

### 📋 Next Priority Actions
1. Implement native libvirt data sources (node_info, node_device_info, node_devices)
2. Advanced domain features (CPU topology, memory backing, HyperV features)
3. Advanced disk features (driver options, readonly/shareable)
4. Advanced network features (DHCP, DNS, routes)

---

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
**Status:** ✅ Completed
**libvirtxml:** `DomainDisk` with type="block"
**Old provider:** `disk.block_device` field

**Tasks:**
- [x] Support type="block" in disk source
- [x] Add acceptance test with block device

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
**Status:** ✅ Completed
**libvirtxml:** `DomainInterface` with type="direct"
**Old provider:** `network_interface` with vepa, macvtap, passthrough, private

**Tasks:**
- [x] Add type="direct" support to interface
- [x] Add source.dev and source.mode fields
- [x] Update conversion for direct types
- [x] Add acceptance test

### 12. NVRAM Template
**Status:** ✅ Completed
**libvirtxml:** `DomainNVRam.Template`
**Old provider:** `nvram.template` field

**Tasks:**
- [x] Add nvram template field to os block (path, template, format, template_format)
- [x] Update conversion functions to handle NVRAM template properly
- [x] Add acceptance test with UEFI
- [x] Remove unsupported address and alias fields (not in libvirtxml)

### 13. TPM Device
**Status:** ✅ Completed
**libvirtxml:** `DomainTPM`
**Old provider:** `tpm` block

**Tasks:**
- [x] Add tpm nested attribute to devices
- [x] Fields: model, backend_type, backend_device_path, backend_encryption_secret, backend_version, backend_persistent_state
- [x] Model ↔ libvirtxml.DomainTPM conversion
- [x] Add acceptance test (skipped - requires swtpm)

### 14. Emulator Path
**Status:** ✅ Completed
**libvirtxml:** `DomainDeviceList.Emulator`
**Old provider:** `emulator` field

**Tasks:**
- [x] Add emulator string attribute to devices
- [x] Update conversion to set emulator path
- [x] Add acceptance test

### 15. Metadata (Custom XML)
**Status:** ✅ Completed
**libvirtxml:** `Domain.Metadata`
**Old provider:** `metadata` string field

**Tasks:**
- [x] Add metadata string attribute to domain
- [x] Update conversion to handle custom metadata XML
- [x] Add acceptance test

### 16. RNG Device (virtio-rng)
**Status:** ✅ Completed
**libvirtxml:** `DomainRNG`
**Old provider:** Added by default

**Tasks:**
- [x] Add RNG nested attribute to devices
- [x] Fields: model, backend (random device path)
- [x] Add acceptance test

### 17. Native Libvirt Data Sources
**Status:** ❌ Not started
**Priority:** Should be implemented next for full old provider parity

**libvirt_node_info** - Host system information
- Uses: `virNodeGetInfo` API
- Fields: cores, cpus, memory, mhz, model, nodes, sockets, threads
- Old provider: data_source_libvirt_node_info.go

**libvirt_node_device_info** - Device details
- Uses: `virNodeDeviceGetXMLDesc` API
- Fields: name, path, parent, driver, devnode, capability (with PCI/USB details)
- Old provider: data_source_libvirt_node_device_info.go

**libvirt_node_devices** - Device enumeration
- Uses: `virNodeListDevices` API
- Fields: capability filter
- Returns: list of device names
- Old provider: data_source_libvirt_node_devices.go

**Tasks:**
- [ ] Implement libvirt_node_info data source
- [ ] Implement libvirt_node_device_info data source
- [ ] Implement libvirt_node_devices data source
- [ ] Add acceptance tests for each
- [ ] Add documentation

## Priority 3: Additional Disk/Network Enhancements

### 18. Advanced Disk Features (NEW - not in old provider)
**Status:** ❌ Not started
**libvirtxml:** Additional `DomainDiskDriver` and `DomainDisk` fields
**Note:** Old provider only had basic driver.name/type auto-detection, no user-configurable driver attributes

**Tasks:**
- [ ] Add driver nested attribute (cache, io, discard) - **NEW FEATURES**
- [ ] Add readonly flag (for CD-ROMs)
- [ ] Add shareable flag (for shared storage)
- [ ] Add boot order attribute
- [ ] Support network disks (RBD, iSCSI)

### 19. Input Devices
**Status:** ❌ Not started
**libvirtxml:** `DomainInput`

**Tasks:**
- [ ] Add input devices (tablet for VNC)
- [ ] Add acceptance test

## Priority 3: Advanced Domain Fields

### 20. CPU Enhancements
**Status:** ❌ Partially done (basic only)

**Tasks:**
- [ ] Add CPU topology (sockets, cores, threads)
- [ ] Add CPU features (enable/disable specific features)
- [ ] Add NUMA cell configuration
- [ ] Test topology scenarios

### 21. Memory Enhancements
**Status:** ❌ Partially done (basic only)

**Tasks:**
- [ ] Memory backing (hugepages, locked, source)
- [ ] Memory tune (hard_limit, soft_limit)
- [ ] Test hugepages

### 22. Features Enhancements
**Status:** ❌ Partially done (simple features only)

**Tasks:**
- [ ] Add HyperV feature block (for Windows)
- [ ] Add KVM feature block
- [ ] Add SMM with TSeg

## Priority 3.5: Provider Infrastructure

### 23. Machine Type Diff Logic
**Status:** ❌ Not started
**Issue:** Currently preserving user's machine value (e.g., "q35") in state to avoid diffs when libvirt expands it (e.g., to "pc-q35-10.1"). This works but storing unexpanded value isn't ideal.
**Better approach:** Implement sophisticated diff logic that understands libvirt's machine type expansion.

**Tasks:**
- [ ] Research Terraform Plugin Framework diff suppression patterns
- [ ] Implement machine type normalization/comparison logic
- [ ] Store actual libvirt value but suppress diff when it's just expansion
- [ ] Update tests

### 24. Test Suite Configuration
**Status:** ❌ Not started

**Tasks:**
- [ ] Centralize libvirt URI configuration for acceptance tests (avoid hardcoding `qemu:///system`; allow overriding when running the suite)
- [ ] Replace hardcoded `/tmp/...` paths in tests with suite-managed temporary directories

## Priority 4: Advanced Network Features

### 25. Network Resource Enhancements
**Status:** Basic modes complete (NAT and isolated), advanced features deferred

**Advanced Network Modes:**
- [ ] Route mode (forward mode="route")
- [ ] Open mode (forward mode="open")
- [ ] Bridge mode (forward mode="bridge")

**DHCP Configuration:**
- [ ] DHCP ranges (start/end addresses)
- [ ] Static host mappings (MAC to IP)
- [ ] DHCP options

**DNS Configuration:**
- [ ] DNS hosts (hostname to IP)
- [ ] DNS forwarders
- [ ] DNS SRV records

**Other:**
- [ ] Static routes
- [ ] Dnsmasq options
- [ ] Network port groups

## Priority 5: Advanced Libvirt Features

### 26. Security Labels
**Status:** ❌ Not started
**libvirtxml:** `DomainSecLabel`

**Tasks:**
- [ ] Add seclabel block (SELinux, AppArmor)
- [ ] Support dynamic, static, and none labels
- [ ] Add acceptance tests

### 27. Host Device Passthrough
**Status:** ❌ Not started
**libvirtxml:** `DomainHostdev`

**Tasks:**
- [ ] Add hostdev block (PCI, USB passthrough)
- [ ] Support managed/unmanaged modes
- [ ] Add acceptance tests

### 28. Advanced Tuning
**Status:** ❌ Not started
**libvirtxml:** `DomainCPUTune`, `DomainNUMATune`, `DomainBlockIOTune`

**Tasks:**
- [ ] CPUTune (vcpu pinning, shares, quota)
- [ ] NUMATune (memory node affinity)
- [ ] BlockIOTune (I/O throttling)
- [ ] IOThreadIDs with pinning

## Priority 6: Provider Conveniences (Deferred - "On Top" of Libvirt)

These features are **provider-specific additions** that build on top of libvirt's native functionality. They do not directly correspond to libvirt APIs and are lower priority than pure libvirt features.

### 29. Cloud-init Resources
**Status:** ❌ Not started - Provider convenience, not native libvirt
**Old provider:** libvirt_cloudinit_disk, libvirt_ignition, libvirt_combustion

These resources generate ISO images with cloud-init/ignition/combustion data and attach them as CD-ROM devices. They are abstractions on top of libvirt's disk functionality.

**Tasks:**
- [ ] libvirt_cloudinit_disk resource (generates cloud-init ISO)
- [ ] libvirt_ignition resource (CoreOS Ignition ISO)
- [ ] libvirt_combustion resource (openSUSE Combustion ISO)

### 30. Template Data Sources
**Status:** ❌ Not started - Provider convenience, not native libvirt
**Old provider:** dns_host_template, dns_srv_template, dnsmasq_options_template

These are simple template formatters with no libvirt API calls. They format data for use with the network resource.

**Tasks:**
- [ ] libvirt_network_dns_host_template datasource
- [ ] libvirt_network_dns_srv_template datasource
- [ ] libvirt_network_dnsmasq_options_template datasource

### 31. Volume URL Download
**Status:** ✅ Completed
**Old provider:** volume `source` attribute with HTTP(S) URLs

This feature downloads images from URLs and uploads them to libvirt storage. It's a convenience wrapper around HTTP + StorageVolUpload.

**Implementation:**
- Uses `create.content.url` block instead of `source` attribute
- Supports HTTPS URLs and local file paths (file:// or absolute paths)
- Capacity is computed from Content-Length header (HTTPS) or file size (local)
- Format must be explicitly specified by user
- Streams directly to StorageVolUpload (no temp files)
- ForceNew on URL change (no re-download support)
- ~200 lines total (volume_upload.go utility + volume_resource.go integration)

Example:
```hcl
resource "libvirt_volume" "base" {
  name   = "ubuntu.qcow2"
  pool   = "default"
  format = "qcow2"  # required

  create = {
    content = {
      url = "https://cloud-images.ubuntu.com/.../ubuntu.img"
    }
  }
  # capacity automatically computed
}
```

### 32. XML XSLT Transforms
**Status:** ❌ Not implementing - Against design principles
**Old provider:** xslt attribute on resources

XSLT transforms allow arbitrary XML manipulation, which conflicts with the design principle of closely modeling the libvirt API. Users should use the provider's schema instead.

### 33. Provider Configuration Enhancements
**Status:** ❌ Not started
**Old provider:** SSH transport support

**Tasks:**
- [ ] SSH transport support (qemu+ssh://, qemu+ssh://user@host/system)
- [ ] Other URI schemes support (test, tcp, etc.)
- [ ] Connection pooling/reuse improvements

### 34. Additional Provider Features
**Status:** ❌ Not started - Provider conveniences

**Tasks:**
- [ ] QEMU agent integration (wait for IP address, execute commands)
- [ ] Network wait_for_lease (poll DHCP lease tables)
- [ ] Domain wait_for_guest_agent (wait for agent to come online)

## Known Gaps (Low Priority)

See DOMAIN_COVERAGE_ANALYSIS.md for complete details. These are libvirtxml-supported fields that haven't been prioritized yet.

**Domain fields not yet implemented:**
- GenID, Resource (cgroup partition)
- Perf events, KeyWrap, LaunchSecurity
- SysInfo (SMBIOS), IDMap (containers), ThrottleGroups
- Hypervisor namespaces (QEMU commandline, etc.)

**Device types not yet implemented:**
- Controllers (except implicit ones)
- Channels, Smartcards
- Sound/Audio devices
- Watchdog, MemBalloon, Panic devices
- Advanced devices: Shmems, Memorydevs, IOMMU, VSock, Crypto, PStore

## Documentation & Migration

### Documentation Tasks
- [ ] Update README.md status table to reflect completed features
- [ ] Add comprehensive examples for all implemented features
- [ ] Write migration guide from old provider (v1 → v2)
- [ ] Document breaking changes vs old provider
- [ ] Add architecture decision records (ADRs)

### Example Gaps
- [ ] Complete domain example with all device types
- [ ] Network configuration examples (NAT, isolated, etc.)
- [ ] Storage pool examples (directory, LVM)
- [ ] Volume examples (qcow2, raw, backing stores)
- [ ] UEFI boot examples
- [ ] Multi-VM examples with shared storage

## Development Notes

**Workflow:**
- Incremental development: Add → Test → Commit → Repeat
- Keep commits small and focused (1-3 related features max)
- Always run `make lint` before committing
- Run acceptance tests for affected resources

**Design Principles:**
- Follow libvirt XML schema closely (per AGENTS.md)
- Only support what libvirtxml supports (no custom XML structs)
- Preserve user intent for optional fields (avoid unnecessary diffs)
- Use nested attributes, not blocks (per Terraform Plugin Framework best practices)
- Always use golibvirt constants, never magic numbers
