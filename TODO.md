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
- **Template Data Sources**: dns_host_template, dns_srv_template, dnsmasq_options_template, libvirt_network (data source)
- **XSLT Transforms**: Custom XML transformations

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

These are pure libvirt API features - no provider abstractions or conveniences.

---

#### libvirt_node_info - Host System Information
**Complexity:** Low (1-2 hours)
**go-libvirt API:** `NodeGetInfo()`
**libvirtxml:** Not needed (direct API return values)
**Old provider:** data_source_libvirt_node_info.go

**API Surface:**
```go
func (l *Libvirt) NodeGetInfo() (
    rModel [32]int8,      // CPU model name
    rMemory uint64,       // Total memory in KB
    rCpus int32,          // Total logical CPUs
    rMhz int32,           // CPU frequency
    rNodes int32,         // NUMA nodes
    rSockets int32,       // CPU sockets
    rCores int32,         // Cores per socket
    rThreads int32,       // Threads per core
    err error
)
```

**Schema:**
- `cpu_model` (string, computed) - CPU model name
- `memory_total_kb` (int64, computed) - Total host memory in KB
- `cpu_cores_total` (int32, computed) - Total logical CPUs
- `numa_nodes` (int32, computed) - Number of NUMA nodes
- `cpu_sockets` (int32, computed) - Number of CPU sockets
- `cpu_cores_per_socket` (int32, computed) - Cores per socket
- `cpu_threads_per_core` (int32, computed) - Threads per core

**Implementation Notes:**
- Single API call, no XML parsing required
- Need helper to convert `[32]int8` model array to string (null-terminated)
- All fields are computed-only
- Use hash of all values for data source ID

**Use Cases:**
- Determining host capabilities before provisioning VMs
- Validating CPU/memory requirements
- Dynamic resource allocation based on host specs

---

#### libvirt_node_devices - Device Enumeration
**Complexity:** Low (1-2 hours)
**go-libvirt API:** `NodeNumOfDevices()` + `NodeListDevices()`
**libvirtxml:** Not needed (returns string array)
**Old provider:** data_source_libvirt_node_devices.go

**API Surface:**
```go
func (l *Libvirt) NodeNumOfDevices(Cap OptString, Flags uint32) (rNum int32, err error)
func (l *Libvirt) NodeListDevices(Cap OptString, Maxnames int32, Flags uint32) (rNames []string, err error)
```

**Schema:**
- `capability` (string, optional) - Filter by device capability type
  - Valid values: "system", "pci", "usb_device", "usb", "net", "scsi_host", "scsi", "storage", "drm", "mdev", "ccw", "css", "ap_queue", "ap_card", "ap_matrix", "ccw_group"
  - Empty/null = return all devices
- `devices` (set of strings, computed) - List of device names (sorted)

**Implementation Notes:**
- Two API calls: get count first, then list devices
- Device names are libvirt-generated strings (e.g., "pci_0000_00_1f_2", "net_eth0_00_11_22_33_44_55")
- Sort the device list for consistency
- Capability filter uses `libvirt.OptString` type (optional string)

**Use Cases:**
- Finding PCI devices for passthrough (capability="pci")
- Listing network interfaces (capability="net")
- Discovering storage devices (capability="storage")
- Building dynamic device lists for domain configuration

---

#### libvirt_node_device_info - Device Details
**Complexity:** High (8-12 hours)
**go-libvirt API:** `NodeDeviceLookupByName()` + `NodeDeviceGetXMLDesc()`
**libvirtxml:** ✅ **FULLY SUPPORTED** via `libvirtxml.NodeDevice`
**Old provider:** data_source_libvirt_node_device_info.go (with custom XML structs)

**API Surface:**
```go
func (l *Libvirt) NodeDeviceLookupByName(Name string) (rDev NodeDevice, err error)
func (l *Libvirt) NodeDeviceGetXMLDesc(Name string, Flags uint32) (rXML string, err error)
```

**libvirtxml.NodeDevice Structure:**
```go
type NodeDevice struct {
    Name       string                   // Device name
    Path       string                   // Sysfs path
    DevNodes   []NodeDeviceDevNode      // Device nodes (for DRM)
    Parent     string                   // Parent device
    Driver     *NodeDeviceDriver        // Driver info (optional)
    Capability NodeDeviceCapability     // Device-specific capability
}

// NodeDeviceCapability contains pointers to all device types
// Only one will be non-nil based on device type
type NodeDeviceCapability struct {
    System      *NodeDeviceSystemCapability      // system: host info
    PCI         *NodeDevicePCICapability         // pci: PCI devices
    USB         *NodeDeviceUSBCapability         // usb: USB host controllers
    USBDevice   *NodeDeviceUSBDeviceCapability   // usb_device: USB devices
    Net         *NodeDeviceNetCapability         // net: network interfaces
    SCSIHost    *NodeDeviceSCSIHostCapability    // scsi_host: SCSI hosts
    SCSITarget  *NodeDeviceSCSITargetCapability  // scsi_target: SCSI targets
    SCSI        *NodeDeviceSCSICapability        // scsi: SCSI devices
    Storage     *NodeDeviceStorageCapability     // storage: storage devices
    DRM         *NodeDeviceDRMCapability         // drm: DRM devices
    CCW         *NodeDeviceCCWCapability         // ccw: s390 CCW
    MDev        *NodeDeviceMDevCapability        // mdev: mediated devices
    CSS         *NodeDeviceCSSCapability         // css: s390 CSS
    APQueue     *NodeDeviceAPQueueCapability     // ap_queue: AP queue
    APCard      *NodeDeviceAPCardCapability      // ap_card: AP card
    APMatrix    *NodeDeviceAPMatrixCapability    // ap_matrix: AP matrix
    CCWGroup    *NodeDeviceCCWGroupCapability    // ccw_group: CCW group
}
```

**Schema:**
- `name` (string, required) - Device name from libvirt_node_devices
- `path` (string, computed) - Sysfs path to device
- `parent` (string, computed) - Parent device name
- `xml` (string, computed) - Raw XML description (for debugging)
- `devnode` (list of objects, computed) - Device nodes (for DRM devices)
  - `type` (string) - Node type
  - `path` (string) - Device node path
- `capability` (single nested, computed) - Device capability details
  - `type` (string) - Capability type ("pci", "usb_device", etc.)
  - All device-specific fields (domain, bus, slot, function for PCI; vendor, product for USB, etc.)

**Important Design Decision:**
The old provider created separate XML structs for each device type (DevicePCI, DeviceUSB, etc.), which violates our design principle. We MUST use `libvirtxml.NodeDevice` exclusively.

**Recommended Implementation Approach:**
1. Use `libvirtxml.NodeDevice.Unmarshal()` to parse device XML
2. Create a flat capability schema with ALL possible fields from all device types
3. Based on which capability pointer is non-nil, populate only relevant fields
4. All device-specific fields should be optional+computed in schema
5. Type discrimination uses the non-nil capability pointer

**Schema Example (Flat Capability with All Fields):**
```hcl
capability = {
  type = "pci"

  # PCI-specific fields (populated)
  domain   = 0
  bus      = 1
  slot     = 0
  function = 0
  class    = "0x030000"
  product  = { id = "0x1234", name = "GPU Model" }
  vendor   = { id = "0x10de", name = "NVIDIA" }
  iommu_group = { number = 15, addresses = [...] }

  # USB-specific fields (null/unset)
  # Storage-specific fields (null/unset)
  # ... other device types (null/unset)
}
```

**Conversion Pattern:**
```go
func convertNodeDeviceCapability(cap libvirtxml.NodeDeviceCapability) types.Object {
    attrs := map[string]attr.Value{}

    // Populate all fields as null first
    attrs["domain"] = types.Int64Null()
    attrs["bus"] = types.Int64Null()
    // ... all other fields ...

    // Then populate based on which capability is present
    switch {
    case cap.PCI != nil:
        attrs["type"] = types.StringValue("pci")
        attrs["domain"] = types.Int64Value(int64(cap.PCI.Domain))
        attrs["bus"] = types.Int64Value(int64(cap.PCI.Bus))
        attrs["slot"] = types.Int64Value(int64(cap.PCI.Slot))
        attrs["function"] = types.Int64Value(int64(cap.PCI.Function))
        // ... other PCI fields

    case cap.USBDevice != nil:
        attrs["type"] = types.StringValue("usb_device")
        // ... USB device fields

    case cap.Net != nil:
        attrs["type"] = types.StringValue("net")
        // ... network interface fields

    case cap.Storage != nil:
        attrs["type"] = types.StringValue("storage")
        // ... storage device fields

    // ... handle all 16 device capability types
    }

    return types.ObjectValueMust(capabilityAttrTypes, attrs)
}
```

**Device Types to Support (16 total):**
1. **system** - Host system info (vendor, model, serial, UUID)
2. **pci** - PCI devices (domain, bus, slot, function, vendor, product, IOMMU group)
3. **usb** - USB host controllers (number, class, subclass, protocol)
4. **usb_device** - USB devices (bus, device, vendor, product)
5. **net** - Network interfaces (interface name, address, link state, features)
6. **scsi_host** - SCSI host controllers (host number, unique_id)
7. **scsi_target** - SCSI targets
8. **scsi** - SCSI devices (host, bus, target, lun, type)
9. **storage** - Storage devices (block, drive_type, model, serial, size)
10. **drm** - DRM devices (type, devnodes)
11. **ccw** - s390 CCW devices
12. **mdev** - Mediated devices
13. **css** - s390 CSS devices
14. **ap_queue** - AP queue devices
15. **ap_card** - AP card devices
16. **ap_matrix** - AP matrix devices
17. **ccw_group** - CCW group devices

**Testing Priority:**
Must test with at least these common device types:
- PCI devices (most common for GPU passthrough)
- USB devices
- Network interfaces
- Storage devices

**Use Cases:**
- Getting PCI device details for GPU passthrough (vendor/product IDs, IOMMU group)
- Finding USB device information for device passthrough
- Checking storage device capabilities and serial numbers
- Network interface details (MAC address, link state, speed)
- Validating IOMMU group membership for PCI passthrough

**Effort Breakdown:**
- Schema definition: 2-3 hours (many fields across 16+ device types)
- Conversion logic: 3-4 hours (switch/case for each device type)
- Testing: 3-5 hours (need real hardware or mocked device XML)

---

**Tasks:**
- [ ] Implement libvirt_node_info data source (Low complexity, 1-2 hours)
- [ ] Implement libvirt_node_devices data source (Low complexity, 1-2 hours)
- [ ] Implement libvirt_node_device_info data source (High complexity, 8-12 hours)
- [ ] Add acceptance tests for each data source
- [ ] Add documentation and examples
- [ ] Test with real PCI/USB/network/storage devices

**Total Estimated Effort:** 15-20 hours for complete native data source parity

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

### 30. Network Data Source
**Status:** ❌ Deferred - Unclear use case vs resource
**Old provider:** data_source_libvirt_network.go

**Rationale for Deferring:**
- The data source appears to look up existing networks by name and return their configuration
- This duplicates the resource Read functionality
- Unclear what API surface it provides beyond what's available in the resource
- Users can reference networks created outside Terraform using resource imports or explicit dependencies
- Low priority compared to node device data sources which provide unique functionality

**If Implemented Later:**
- Use `NetworkLookupByName()` + `NetworkGetXMLDesc()`
- Reuse existing network resource schema and conversion functions
- Schema: `name` (required) + all network resource fields (computed)

---

### 31. Template Data Sources
**Status:** ❌ Not started - Provider convenience, not native libvirt
**Old provider:** dns_host_template, dns_srv_template, dnsmasq_options_template

These are simple template formatters with no libvirt API calls. They format data for use with the network resource.

**Rationale for Deferring:**
- These don't interact with libvirt at all - they just format maps/objects
- They were workarounds for Terraform's limited dynamic expression capabilities in older versions
- Modern Terraform (0.12+) has `for` expressions and dynamic blocks that make these unnecessary
- Against design principle of closely modeling libvirt API
- Low value compared to native libvirt data sources

**Modern Alternative:**
Users can format these directly in HCL:
```hcl
# Old way (template data source):
data "libvirt_network_dns_host_template" "host" {
  ip       = "10.0.0.10"
  hostname = "server1"
}

# Modern way (direct HCL):
locals {
  dns_host = {
    ip       = "10.0.0.10"
    hostname = "server1"
  }
}
```

**Tasks:**
- [ ] libvirt_network_dns_host_template datasource (if ever needed)
- [ ] libvirt_network_dns_srv_template datasource (if ever needed)
- [ ] libvirt_network_dnsmasq_options_template datasource (if ever needed)

### 32. Volume URL Download
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

### 33. XML XSLT Transforms
**Status:** ❌ Not implementing - Against design principles
**Old provider:** xslt attribute on resources

XSLT transforms allow arbitrary XML manipulation, which conflicts with the design principle of closely modeling the libvirt API. Users should use the provider's schema instead.

### 34. Provider Configuration Enhancements
**Status:** ✅ Completed (Transport support)
**Old provider:** SSH transport support

**Tasks:**
- [x] SSH transport support (qemu+ssh://, qemu+sshcmd://user@host/system)
- [x] TCP and TLS transport support (qemu+tcp://, qemu+tls://)
- [x] Local socket support (qemu:///system, qemu:///session)
- [x] Transport documentation with examples
- [ ] Connection pooling/reuse improvements
- [ ] SSH dialer strategy decision (see below)
- [ ] Transport acceptance tests (see below)

**SSH Dialer Strategy Discussion:**

Currently, the provider supports both SSH transports:
- `qemu+ssh://` - Uses Go SSH library (upstream go-libvirt dialer)
- `qemu+sshcmd://` - Uses native SSH command (ported from old provider)

**Decision Points:**
1. **Default behavior:** Should we make `qemu+ssh://` default to one or the other?
   - Option A: Keep current behavior (ssh = Go library, sshcmd = native command)
   - Option B: Make sshcmd the default for `qemu+ssh://`, add flag to use Go library
   - Option C: Auto-detect based on presence of ~/.ssh/config or complexity of URI parameters

2. **Trade-offs:**
   - Go library (`qemu+ssh://`):
     - ✅ Pure Go, no external dependencies
     - ✅ Consistent cross-platform behavior
     - ❌ Doesn't respect ~/.ssh/config (ProxyJump, ControlMaster, etc.)
     - ❌ Limited SSH feature support

   - Native command (`qemu+sshcmd://`):
     - ✅ Full SSH feature support (ProxyJump, agent forwarding, etc.)
     - ✅ Respects ~/.ssh/config settings
     - ✅ Familiar behavior for SSH users
     - ❌ Requires ssh binary on PATH
     - ❌ Requires nc or virt-ssh-helper on remote system
     - ❌ Process spawning overhead

3. **User Feedback:**
   - Old provider used native SSH command approach
   - Users migrating may expect SSH config to be respected
   - But pure Go solution is more portable for CI/CD environments

**Recommendation:** Keep current explicit distinction (ssh vs sshcmd) for now. This gives users clear control and allows both use cases. Could revisit based on user feedback.

**Transport Acceptance Tests:**

Need to determine approach for testing remote transports:

**Challenges:**
- SSH tests require SSH server setup
- TLS tests require certificate infrastructure
- Tests need to be reproducible in CI/CD

**Options:**
1. **Mock/Stub tests:** Test dialer creation and URI parsing without actual connections
2. **Docker-based tests:** Spin up containers with SSH/libvirt for integration tests
3. **Conditional tests:** Skip remote transport tests unless specific environment variables set
4. **Manual tests:** Document manual testing procedures, rely on real-world usage

**Current Testing:**
- Local transport tests work (qemu:///system)
- URI parsing and dialer factory tested via builds
- No automated remote transport tests yet

**Recommended Approach:**
- Phase 1: Add unit tests for URI parsing and dialer creation (no network I/O)
- Phase 2: Add Docker-based integration tests for SSH transport (optional, CI can skip)
- Phase 3: Document manual testing procedures for TLS and other transports
- Phase 4: Consider test fixtures or recording/playback for remote transport tests

### 35. Additional Provider Features
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
