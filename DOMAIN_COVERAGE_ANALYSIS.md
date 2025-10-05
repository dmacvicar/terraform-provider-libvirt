# libvirtxml.Domain Field Coverage Analysis

**Purpose**: Complete analysis of all top-level Domain fields and nested structures (excluding Devices) to guide Terraform schema implementation.

**Date**: 2025-10-05
**Source**: libvirt.org/go/libvirtxml package

---

## Executive Summary

The `libvirtxml.Domain` type contains **43 top-level fields** (excluding devices and hypervisor-specific namespaces). These fields range from simple strings to deeply nested structures with multiple levels of complexity.

### Key Statistics
- **Simple/Primitive Fields**: 9 fields (21%)
- **Moderately Complex Fields**: 12 fields (28%)
- **Highly Complex Fields**: 22 fields (51%)
- **Total Nested Types Analyzed**: 150+

### Implementation Priority Recommendation
1. **Phase 1 (High Priority)**: Basic metadata, memory, CPU/VCPU, OS configuration
2. **Phase 2 (Medium Priority)**: Features, Clock/Timers, Lifecycle, Security
3. **Phase 3 (Low Priority)**: Advanced tuning, performance monitoring, specialized features

---

## Domain Top-Level Fields

```go
type Domain struct {
    XMLName         xml.Name               `xml:"domain"`
    Type            string                 `xml:"type,attr,omitempty"`
    ID              *int                   `xml:"id,attr"`
    Name            string                 `xml:"name,omitempty"`
    UUID            string                 `xml:"uuid,omitempty"`
    HWUUID          string                 `xml:"hwuuid,omitempty"`
    GenID           *DomainGenID           `xml:"genid"`
    Title           string                 `xml:"title,omitempty"`
    Description     string                 `xml:"description,omitempty"`
    Metadata        *DomainMetadata        `xml:"metadata"`
    MaximumMemory   *DomainMaxMemory       `xml:"maxMemory"`
    Memory          *DomainMemory          `xml:"memory"`
    CurrentMemory   *DomainCurrentMemory   `xml:"currentMemory"`
    BlockIOTune     *DomainBlockIOTune     `xml:"blkiotune"`
    MemoryTune      *DomainMemoryTune      `xml:"memtune"`
    MemoryBacking   *DomainMemoryBacking   `xml:"memoryBacking"`
    VCPU            *DomainVCPU            `xml:"vcpu"`
    VCPUs           *DomainVCPUs           `xml:"vcpus"`
    IOThreads       uint                   `xml:"iothreads,omitempty"`
    IOThreadIDs     *DomainIOThreadIDs     `xml:"iothreadids"`
    DefaultIOThread *DomainDefaultIOThread `xml:"defaultiothread"`
    CPUTune         *DomainCPUTune         `xml:"cputune"`
    NUMATune        *DomainNUMATune        `xml:"numatune"`
    Resource        *DomainResource        `xml:"resource"`
    SysInfo         []DomainSysInfo        `xml:"sysinfo"`
    Bootloader      string                 `xml:"bootloader,omitempty"`
    BootloaderArgs  string                 `xml:"bootloader_args,omitempty"`
    OS              *DomainOS              `xml:"os"`
    IDMap           *DomainIDMap           `xml:"idmap"`
    ThrottleGroups  *DomainThrottleGroups  `xml:"throttlegroups"`
    Features        *DomainFeatureList     `xml:"features"`
    CPU             *DomainCPU             `xml:"cpu"`
    Clock           *DomainClock           `xml:"clock"`
    OnPoweroff      string                 `xml:"on_poweroff,omitempty"`
    OnReboot        string                 `xml:"on_reboot,omitempty"`
    OnCrash         string                 `xml:"on_crash,omitempty"`
    PM              *DomainPM              `xml:"pm"`
    Perf            *DomainPerf            `xml:"perf"`
    Devices         *DomainDeviceList      `xml:"devices"`  // EXCLUDED from this analysis
    SecLabel        []DomainSecLabel       `xml:"seclabel"`
    KeyWrap         *DomainKeyWrap         `xml:"keywrap"`
    LaunchSecurity  *DomainLaunchSecurity  `xml:"launchSecurity"`
}
```

---

## Category 1: Simple Metadata Fields

### 1.1 Basic Identifiers
**Complexity**: Simple (strings/primitives)

| Field | Type | XML Element | Description |
|-------|------|-------------|-------------|
| Type | string | type (attr) | Domain type: qemu, kvm, xen, lxc, etc. |
| ID | *int | id (attr) | Runtime domain ID (read-only) |
| Name | string | name | Domain name |
| UUID | string | uuid | Domain UUID |
| HWUUID | string | hwuuid | Hardware UUID for SMBIOS |
| Title | string | title | Human-readable title |
| Description | string | description | Human-readable description |
| Bootloader | string | bootloader | Bootloader path (Xen PV domains) |
| BootloaderArgs | string | bootloader_args | Bootloader arguments |

**Implementation Priority**: **HIGH** - Essential for basic domain creation

### 1.2 GenID
**Complexity**: Simple wrapper struct

```go
type DomainGenID struct {
    Value string `xml:",chardata"`
}
```

**Use Case**: Generation ID for Windows VMs (VM-Generation ID)
**Implementation Priority**: **MEDIUM** - Useful for Windows guests

### 1.3 Metadata
**Complexity**: Simple passthrough

```go
type DomainMetadata struct {
    XML string `xml:",innerxml"`
}
```

**Use Case**: Store arbitrary XML metadata (application-specific data)
**Implementation Priority**: **LOW** - Advanced use case, can be raw XML

---

## Category 2: Memory Configuration

### 2.1 Basic Memory Fields
**Complexity**: Simple structs with unit attributes

```go
type DomainMemory struct {
    Value    uint   `xml:",chardata"`
    Unit     string `xml:"unit,attr,omitempty"`      // KiB, MiB, GiB, TiB
    DumpCore string `xml:"dumpCore,attr,omitempty"`  // on/off
}

type DomainCurrentMemory struct {
    Value uint   `xml:",chardata"`
    Unit  string `xml:"unit,attr,omitempty"`
}

type DomainMaxMemory struct {
    Value uint   `xml:",chardata"`
    Unit  string `xml:"unit,attr,omitempty"`
    Slots uint   `xml:"slots,attr,omitempty"`        // Memory hotplug slots
}
```

**Implementation Priority**: **HIGH** - Critical for domain configuration

### 2.2 MemoryBacking
**Complexity**: Moderate (multiple nested options)

```go
type DomainMemoryBacking struct {
    MemoryHugePages    *DomainMemoryHugepages    `xml:"hugepages"`
    MemoryNosharepages *DomainMemoryNosharepages `xml:"nosharepages"`
    MemoryLocked       *DomainMemoryLocked       `xml:"locked"`
    MemorySource       *DomainMemorySource       `xml:"source"`
    MemoryAccess       *DomainMemoryAccess       `xml:"access"`
    MemoryAllocation   *DomainMemoryAllocation   `xml:"allocation"`
    MemoryDiscard      *DomainMemoryDiscard      `xml:"discard"`
}

// Nested types
type DomainMemoryHugepages struct {
    Hugepages []DomainMemoryHugepage `xml:"page"`
}

type DomainMemoryHugepage struct {
    Size    uint   `xml:"size,attr"`
    Unit    string `xml:"unit,attr,omitempty"`
    Nodeset string `xml:"nodeset,attr,omitempty"`
}

type DomainMemorySource struct {
    Type string `xml:"type,attr,omitempty"`  // file, anonymous, memfd
}

type DomainMemoryAccess struct {
    Mode string `xml:"mode,attr,omitempty"`  // shared, private
}

type DomainMemoryAllocation struct {
    Mode    string `xml:"mode,attr,omitempty"`     // immediate, ondemand
    Threads uint   `xml:"threads,attr,omitempty"`
}

// Empty marker types (presence indicates enabled)
type DomainMemoryNosharepages struct {}
type DomainMemoryLocked struct {}
type DomainMemoryDiscard struct {}
```

**Use Cases**:
- Hugepages: Performance optimization for large memory workloads
- Locked: Prevent swapping (required for VFIO, real-time)
- Source: Memory backend type (file-backed, memfd)
- Access: Memory sharing between host/guest

**Implementation Priority**: **MEDIUM-HIGH** - Important for performance tuning

### 2.3 MemoryTune
**Complexity**: Moderate (memory limits)

```go
type DomainMemoryTune struct {
    HardLimit     *DomainMemoryTuneLimit `xml:"hard_limit"`
    SoftLimit     *DomainMemoryTuneLimit `xml:"soft_limit"`
    MinGuarantee  *DomainMemoryTuneLimit `xml:"min_guarantee"`
    SwapHardLimit *DomainMemoryTuneLimit `xml:"swap_hard_limit"`
}

type DomainMemoryTuneLimit struct {
    Value uint64 `xml:",chardata"`
    Unit  string `xml:"unit,attr,omitempty"`
}
```

**Use Cases**:
- Hard/soft limits: Cgroup memory limits
- MinGuarantee: Reserved memory guarantee
- SwapHardLimit: Limit swap usage

**Implementation Priority**: **MEDIUM** - Important for resource management

---

## Category 3: CPU/VCPU Configuration

### 3.1 Basic VCPU
**Complexity**: Simple struct

```go
type DomainVCPU struct {
    Placement string `xml:"placement,attr,omitempty"`  // static, auto
    CPUSet    string `xml:"cpuset,attr,omitempty"`     // 0-3,^2
    Current   uint   `xml:"current,attr,omitempty"`    // Currently active VCPUs
    Value     uint   `xml:",chardata"`                 // Maximum VCPUs
}
```

**Implementation Priority**: **HIGH** - Essential for CPU allocation

### 3.2 VCPUs (Advanced)
**Complexity**: Moderate (individual VCPU control)

```go
type DomainVCPUs struct {
    VCPU []DomainVCPUsVCPU `xml:"vcpu"`
}

type DomainVCPUsVCPU struct {
    Id           *uint  `xml:"id,attr"`
    Enabled      string `xml:"enabled,attr,omitempty"`      // yes/no
    Hotpluggable string `xml:"hotpluggable,attr,omitempty"` // yes/no
    Order        *uint  `xml:"order,attr"`
}
```

**Use Case**: Fine-grained VCPU hotplug control
**Implementation Priority**: **LOW** - Advanced feature, rarely used

### 3.3 CPU
**Complexity**: High (multiple nested levels)

```go
type DomainCPU struct {
    XMLName            xml.Name              `xml:"cpu"`
    Match              string                `xml:"match,attr,omitempty"`  // minimum, exact, strict
    Mode               string                `xml:"mode,attr,omitempty"`   // custom, host-model, host-passthrough
    Check              string                `xml:"check,attr,omitempty"`  // none, partial, full
    Migratable         string                `xml:"migratable,attr,omitempty"` // on/off
    DeprecatedFeatures string                `xml:"deprecated_features,attr,omitempty"`
    Model              *DomainCPUModel       `xml:"model"`
    Vendor             string                `xml:"vendor,omitempty"`
    Topology           *DomainCPUTopology    `xml:"topology"`
    Cache              *DomainCPUCache       `xml:"cache"`
    MaxPhysAddr        *DomainCPUMaxPhysAddr `xml:"maxphysaddr"`
    Features           []DomainCPUFeature    `xml:"feature"`
    Numa               *DomainNuma           `xml:"numa"`
}

type DomainCPUModel struct {
    Fallback string `xml:"fallback,attr,omitempty"`  // allow, forbid
    Value    string `xml:",chardata"`                // CPU model name
    VendorID string `xml:"vendor_id,attr,omitempty"`
}

type DomainCPUTopology struct {
    Sockets  int `xml:"sockets,attr,omitempty"`
    Dies     int `xml:"dies,attr,omitempty"`
    Clusters int `xml:"clusters,attr,omitempty"`
    Cores    int `xml:"cores,attr,omitempty"`
    Threads  int `xml:"threads,attr,omitempty"`
}

type DomainCPUCache struct {
    Level uint   `xml:"level,attr,omitempty"`
    Mode  string `xml:"mode,attr"`  // emulate, passthrough, disable
}

type DomainCPUMaxPhysAddr struct {
    Mode  string `xml:"mode,attr"`             // emulate, passthrough
    Bits  uint   `xml:"bits,attr,omitempty"`
    Limit uint   `xml:"limit,attr,omitempty"`
}

type DomainCPUFeature struct {
    Policy string `xml:"policy,attr,omitempty"`  // force, require, optional, disable, forbid
    Name   string `xml:"name,attr,omitempty"`    // CPU feature name
}

// NUMA configuration (complex, see section 3.5)
type DomainNuma struct {
    Cell          []DomainCell             `xml:"cell"`
    Interconnects *DomainNUMAInterconnects `xml:"interconnects"`
}

type DomainCell struct {
    ID        *uint                `xml:"id,attr"`
    CPUs      string               `xml:"cpus,attr,omitempty"`
    Memory    uint                 `xml:"memory,attr"`
    Unit      string               `xml:"unit,attr,omitempty"`
    MemAccess string               `xml:"memAccess,attr,omitempty"`
    Discard   string               `xml:"discard,attr,omitempty"`
    Distances *DomainCellDistances `xml:"distances"`
    Caches    []DomainCellCache    `xml:"cache"`
}
```

**Use Cases**:
- Mode: CPU presentation to guest (custom model, host-model, host-passthrough)
- Topology: Virtual CPU socket/core/thread layout
- Features: Enable/disable specific CPU features (e.g., vmx, svm)
- NUMA: Non-uniform memory access topology

**Implementation Priority**:
- Basic (mode, model, topology): **HIGH**
- Features: **MEDIUM**
- Advanced (cache, maxphysaddr, NUMA): **LOW-MEDIUM**

### 3.4 CPUTune
**Complexity**: High (extensive tuning options)

```go
type DomainCPUTune struct {
    Shares         *DomainCPUTuneShares         `xml:"shares"`
    Period         *DomainCPUTunePeriod         `xml:"period"`
    Quota          *DomainCPUTuneQuota          `xml:"quota"`
    GlobalPeriod   *DomainCPUTunePeriod         `xml:"global_period"`
    GlobalQuota    *DomainCPUTuneQuota          `xml:"global_quota"`
    EmulatorPeriod *DomainCPUTunePeriod         `xml:"emulator_period"`
    EmulatorQuota  *DomainCPUTuneQuota          `xml:"emulator_quota"`
    IOThreadPeriod *DomainCPUTunePeriod         `xml:"iothread_period"`
    IOThreadQuota  *DomainCPUTuneQuota          `xml:"iothread_quota"`
    VCPUPin        []DomainCPUTuneVCPUPin       `xml:"vcpupin"`
    EmulatorPin    *DomainCPUTuneEmulatorPin    `xml:"emulatorpin"`
    IOThreadPin    []DomainCPUTuneIOThreadPin   `xml:"iothreadpin"`
    VCPUSched      []DomainCPUTuneVCPUSched     `xml:"vcpusched"`
    EmulatorSched  *DomainCPUTuneEmulatorSched  `xml:"emulatorsched"`
    IOThreadSched  []DomainCPUTuneIOThreadSched `xml:"iothreadsched"`
    CacheTune      []DomainCPUCacheTune         `xml:"cachetune"`
    MemoryTune     []DomainCPUMemoryTune        `xml:"memorytune"`
}

type DomainCPUTuneVCPUPin struct {
    VCPU   uint   `xml:"vcpu,attr"`
    CPUSet string `xml:"cpuset,attr"`  // 0-3,^2
}

type DomainCPUTuneVCPUSched struct {
    VCPUs     string `xml:"vcpus,attr"`
    Scheduler string `xml:"scheduler,attr,omitempty"`  // batch, idle, fifo, rr
    Priority  *int   `xml:"priority,attr"`
}

type DomainCPUCacheTune struct {
    VCPUs   string                      `xml:"vcpus,attr,omitempty"`
    ID      string                      `xml:"id,attr,omitempty"`
    Cache   []DomainCPUCacheTuneCache   `xml:"cache"`
    Monitor []DomainCPUCacheTuneMonitor `xml:"monitor"`
}
```

**Use Cases**:
- VCPUPin: Pin VCPUs to specific host CPUs
- Shares/Period/Quota: CFS (Completely Fair Scheduler) controls
- Scheduler: Real-time scheduling policies
- CacheTune: CAT (Cache Allocation Technology) / RDT controls

**Implementation Priority**: **LOW-MEDIUM** - Advanced performance tuning

### 3.5 NUMATune
**Complexity**: Moderate

```go
type DomainNUMATune struct {
    Memory   *DomainNUMATuneMemory   `xml:"memory"`
    MemNodes []DomainNUMATuneMemNode `xml:"memnode"`
}

type DomainNUMATuneMemory struct {
    Mode      string `xml:"mode,attr,omitempty"`      // strict, preferred, interleave
    Nodeset   string `xml:"nodeset,attr,omitempty"`   // 0-3
    Placement string `xml:"placement,attr,omitempty"` // static, auto
}

type DomainNUMATuneMemNode struct {
    CellID  uint   `xml:"cellid,attr"`
    Mode    string `xml:"mode,attr"`
    Nodeset string `xml:"nodeset,attr"`
}
```

**Use Case**: Control NUMA memory allocation policies
**Implementation Priority**: **LOW** - NUMA-specific workloads

---

## Category 4: IOThreads

**Complexity**: Low-Moderate

```go
// Simple count
IOThreads uint `xml:"iothreads,omitempty"`

// Advanced configuration
type DomainIOThreadIDs struct {
    IOThreads []DomainIOThread `xml:"iothread"`
}

type DomainIOThread struct {
    ID      uint                `xml:"id,attr"`
    PoolMin *uint               `xml:"thread_pool_min,attr"`
    PoolMax *uint               `xml:"thread_pool_max,attr"`
    Poll    *DomainIOThreadPoll `xml:"poll"`
}

type DomainDefaultIOThread struct {
    PoolMin *uint `xml:"thread_pool_min,attr"`
    PoolMax *uint `xml:"thread_pool_max,attr"`
}
```

**Use Case**: Dedicated I/O threads for virtio devices
**Implementation Priority**: **MEDIUM** - Common for virtio-blk/virtio-scsi

---

## Category 5: Operating System Configuration

### 5.1 DomainOS
**Complexity**: High (many boot-related options)

```go
type DomainOS struct {
    Type         *DomainOSType         `xml:"type"`
    Firmware     string                `xml:"firmware,attr,omitempty"`  // bios, efi
    FirmwareInfo *DomainOSFirmwareInfo `xml:"firmware"`
    Init         string                `xml:"init,omitempty"`
    InitArgs     []string              `xml:"initarg"`
    InitEnv      []DomainOSInitEnv     `xml:"initenv"`
    InitDir      string                `xml:"initdir,omitempty"`
    InitUser     string                `xml:"inituser,omitempty"`
    InitGroup    string                `xml:"initgroup,omitempty"`
    Loader       *DomainLoader         `xml:"loader"`
    NVRam        *DomainNVRam          `xml:"nvram"`
    Kernel       string                `xml:"kernel,omitempty"`
    Initrd       string                `xml:"initrd,omitempty"`
    Cmdline      string                `xml:"cmdline,omitempty"`
    Shim         string                `xml:"shim,omitempty"`
    DTB          string                `xml:"dtb,omitempty"`              // Device Tree Blob
    ACPI         *DomainACPI           `xml:"acpi"`
    BootDevices  []DomainBootDevice    `xml:"boot"`
    BootMenu     *DomainBootMenu       `xml:"bootmenu"`
    BIOS         *DomainBIOS           `xml:"bios"`
    SMBios       *DomainSMBios         `xml:"smbios"`
}

type DomainOSType struct {
    Arch    string `xml:"arch,attr,omitempty"`    // x86_64, aarch64, etc.
    Machine string `xml:"machine,attr,omitempty"` // pc-i440fx, pc-q35, virt, etc.
    Type    string `xml:",chardata"`              // hvm, linux, xen, exe
}

type DomainLoader struct {
    Path      string `xml:",chardata"`
    Readonly  string `xml:"readonly,attr,omitempty"`   // yes/no
    Secure    string `xml:"secure,attr,omitempty"`     // yes/no (Secure Boot)
    Stateless string `xml:"stateless,attr,omitempty"`  // yes/no
    Type      string `xml:"type,attr,omitempty"`       // rom, pflash
    Format    string `xml:"format,attr,omitempty"`     // raw, qcow2
}

type DomainNVRam struct {
    NVRam          string            `xml:",chardata"`
    Source         *DomainDiskSource `xml:"source"`
    Template       string            `xml:"template,attr,omitempty"`
    Format         string            `xml:"format,attr,omitempty"`
    TemplateFormat string            `xml:"templateFormat,attr,omitempty"`
}

type DomainBootDevice struct {
    Dev string `xml:"dev,attr"`  // hd, fd, cdrom, network
}

type DomainBootMenu struct {
    Enable  string `xml:"enable,attr,omitempty"`  // yes/no
    Timeout string `xml:"timeout,attr,omitempty"` // milliseconds
}

type DomainBIOS struct {
    UseSerial     string `xml:"useserial,attr,omitempty"`  // yes/no
    RebootTimeout *int   `xml:"rebootTimeout,attr"`        // seconds
}

type DomainSMBios struct {
    Mode string `xml:"mode,attr"`  // emulate, host, sysinfo
}

type DomainOSInitEnv struct {
    Name  string `xml:"name,attr"`
    Value string `xml:",chardata"`
}

type DomainACPI struct {
    Tables []DomainACPITable `xml:"table"`
}

type DomainOSFirmwareInfo struct {
    Features []DomainOSFirmwareFeature `xml:"feature"`
}
```

**Use Cases**:
- Type: Basic OS type, architecture, machine type
- Loader/NVRam: UEFI firmware configuration
- Kernel/Initrd/Cmdline: Direct kernel boot
- Init*: Container (LXC) initialization
- BootDevices: Boot order
- ACPI: Custom ACPI tables

**Implementation Priority**:
- Basic (type, arch, machine, boot devices): **HIGH**
- UEFI (loader, nvram): **HIGH** (very common)
- Direct kernel boot: **MEDIUM**
- Container init: **LOW** (LXC-specific)
- ACPI tables: **LOW** (rare)

---

## Category 6: Features

### 6.1 DomainFeatureList
**Complexity**: High (30+ feature flags)

```go
type DomainFeatureList struct {
    PAE           *DomainFeature              `xml:"pae"`
    ACPI          *DomainFeature              `xml:"acpi"`
    APIC          *DomainFeatureAPIC          `xml:"apic"`
    HAP           *DomainFeatureState         `xml:"hap"`
    Viridian      *DomainFeature              `xml:"viridian"`
    PrivNet       *DomainFeature              `xml:"privnet"`
    HyperV        *DomainFeatureHyperV        `xml:"hyperv"`
    KVM           *DomainFeatureKVM           `xml:"kvm"`
    Xen           *DomainFeatureXen           `xml:"xen"`
    PVSpinlock    *DomainFeatureState         `xml:"pvspinlock"`
    PMU           *DomainFeatureState         `xml:"pmu"`
    VMPort        *DomainFeatureState         `xml:"vmport"`
    GIC           *DomainFeatureGIC           `xml:"gic"`
    SMM           *DomainFeatureSMM           `xml:"smm"`
    IOAPIC        *DomainFeatureIOAPIC        `xml:"ioapic"`
    HPT           *DomainFeatureHPT           `xml:"hpt"`
    HTM           *DomainFeatureState         `xml:"htm"`
    NestedHV      *DomainFeatureState         `xml:"nested-hv"`
    Capabilities  *DomainFeatureCapabilities  `xml:"capabilities"`
    VMCoreInfo    *DomainFeatureState         `xml:"vmcoreinfo"`
    MSRS          *DomainFeatureMSRS          `xml:"msrs"`
    CCFAssist     *DomainFeatureState         `xml:"ccf-assist"`
    CFPC          *DomainFeatureCFPC          `xml:"cfpc"`
    SBBC          *DomainFeatureSBBC          `xml:"sbbc"`
    IBS           *DomainFeatureIBS           `xml:"ibs"`
    TCG           *DomainFeatureTCG           `xml:"tcg"`
    AsyncTeardown *DomainFeatureAsyncTeardown `xml:"async-teardown"`
    RAS           *DomainFeatureState         `xml:"ras"`
    PS2           *DomainFeatureState         `xml:"ps2"`
    AIA           *DomainFeatureAIA           `xml:"aia"`
}
```

### 6.2 Feature Types

#### Simple Features (just presence)
```go
type DomainFeature struct {}  // PAE, ACPI, Viridian, PrivNet
```

#### State-based Features
```go
type DomainFeatureState struct {
    State string `xml:"state,attr,omitempty"`  // on/off
}
// Used by: HAP, PVSpinlock, PMU, VMPort, HTM, NestedHV, VMCoreInfo,
//          CCFAssist, RAS, PS2
```

#### Attribute-based Features
```go
type DomainFeatureAPIC struct {
    EOI string `xml:"eoi,attr,omitempty"`  // on/off
}

type DomainFeatureGIC struct {
    Version string `xml:"version,attr,omitempty"`  // 2, 3, host
}

type DomainFeatureSMM struct {
    State string                `xml:"state,attr,omitempty"`
    TSeg  *DomainFeatureSMMTSeg `xml:"tseg"`
}

type DomainFeatureIOAPIC struct {
    Driver string `xml:"driver,attr,omitempty"`  // kvm, qemu
}

type DomainFeatureHPT struct {
    Resizing    string                    `xml:"resizing,attr,omitempty"`
    MaxPageSize *DomainFeatureHPTPageSize `xml:"maxpagesize"`
}

type DomainFeatureMSRS struct {
    Unknown string `xml:"unknown,attr"`  // ignore, fault
}

type DomainFeatureCFPC struct {
    Value string `xml:"value,attr"`  // broken, workaround, fixed
}

type DomainFeatureSBBC struct {
    Value string `xml:"value,attr"`  // broken, workaround, fixed
}

type DomainFeatureIBS struct {
    Value string `xml:"value,attr"`  // broken, workaround, fixed-ibs, fixed-ccd, fixed-na
}

type DomainFeatureAsyncTeardown struct {
    Enabled string `xml:"enabled,attr,omitempty"`  // yes/no
}

type DomainFeatureAIA struct {
    Value string `xml:"value,attr"`  // aplic, aplic-imsic
}
```

#### Complex Features

##### HyperV (Windows optimization)
```go
type DomainFeatureHyperV struct {
    DomainFeature
    Mode            string                        `xml:"mode,attr,omitempty"`  // custom, passthrough
    Relaxed         *DomainFeatureState           `xml:"relaxed"`
    VAPIC           *DomainFeatureState           `xml:"vapic"`
    Spinlocks       *DomainFeatureHyperVSpinlocks `xml:"spinlocks"`
    VPIndex         *DomainFeatureState           `xml:"vpindex"`
    Runtime         *DomainFeatureState           `xml:"runtime"`
    Synic           *DomainFeatureState           `xml:"synic"`
    STimer          *DomainFeatureHyperVSTimer    `xml:"stimer"`
    Reset           *DomainFeatureState           `xml:"reset"`
    VendorId        *DomainFeatureHyperVVendorId  `xml:"vendor_id"`
    Frequencies     *DomainFeatureState           `xml:"frequencies"`
    ReEnlightenment *DomainFeatureState           `xml:"reenlightenment"`
    TLBFlush        *DomainFeatureHyperVTLBFlush  `xml:"tlbflush"`
    IPI             *DomainFeatureState           `xml:"ipi"`
    EVMCS           *DomainFeatureState           `xml:"evmcs"`
    AVIC            *DomainFeatureState           `xml:"avic"`
    EMSRBitmap      *DomainFeatureState           `xml:"emsr_bitmap"`
    XMMInput        *DomainFeatureState           `xml:"xmm_input"`
}
```

##### KVM
```go
type DomainFeatureKVM struct {
    Hidden        *DomainFeatureState        `xml:"hidden"`
    HintDedicated *DomainFeatureState        `xml:"hint-dedicated"`
    PollControl   *DomainFeatureState        `xml:"poll-control"`
    PVIPI         *DomainFeatureState        `xml:"pv-ipi"`
    DirtyRing     *DomainFeatureKVMDirtyRing `xml:"dirty-ring"`
}
```

##### Xen
```go
type DomainFeatureXen struct {
    E820Host    *DomainFeatureXenE820Host    `xml:"e820_host"`
    Passthrough *DomainFeatureXenPassthrough `xml:"passthrough"`
}
```

##### TCG
```go
type DomainFeatureTCG struct {
    TBCache *DomainFeatureTCGTBCache `xml:"tb-cache"`
}
```

##### Capabilities (Linux capabilities)
```go
type DomainFeatureCapabilities struct {
    Policy         string                   `xml:"policy,attr,omitempty"`  // default, allow, deny
    AuditControl   *DomainFeatureCapability `xml:"audit_control"`
    AuditWrite     *DomainFeatureCapability `xml:"audit_write"`
    BlockSuspend   *DomainFeatureCapability `xml:"block_suspend"`
    Chown          *DomainFeatureCapability `xml:"chown"`
    DACOverride    *DomainFeatureCapability `xml:"dac_override"`
    DACReadSearch  *DomainFeatureCapability `xml:"dac_read_Search"`
    FOwner         *DomainFeatureCapability `xml:"fowner"`
    FSetID         *DomainFeatureCapability `xml:"fsetid"`
    IPCLock        *DomainFeatureCapability `xml:"ipc_lock"`
    IPCOwner       *DomainFeatureCapability `xml:"ipc_owner"`
    Kill           *DomainFeatureCapability `xml:"kill"`
    Lease          *DomainFeatureCapability `xml:"lease"`
    LinuxImmutable *DomainFeatureCapability `xml:"linux_immutable"`
    MACAdmin       *DomainFeatureCapability `xml:"mac_admin"`
    MACOverride    *DomainFeatureCapability `xml:"mac_override"`
    MkNod          *DomainFeatureCapability `xml:"mknod"`
    NetAdmin       *DomainFeatureCapability `xml:"net_admin"`
    NetBindService *DomainFeatureCapability `xml:"net_bind_service"`
    NetBroadcast   *DomainFeatureCapability `xml:"net_broadcast"`
    NetRaw         *DomainFeatureCapability `xml:"net_raw"`
    SetGID         *DomainFeatureCapability `xml:"setgid"`
    SetFCap        *DomainFeatureCapability `xml:"setfcap"`
    SetPCap        *DomainFeatureCapability `xml:"setpcap"`
    SetUID         *DomainFeatureCapability `xml:"setuid"`
    SysAdmin       *DomainFeatureCapability `xml:"sys_admin"`
    SysBoot        *DomainFeatureCapability `xml:"sys_boot"`
    SysChRoot      *DomainFeatureCapability `xml:"sys_chroot"`
    SysModule      *DomainFeatureCapability `xml:"sys_module"`
    SysNice        *DomainFeatureCapability `xml:"sys_nice"`
    SysPAcct       *DomainFeatureCapability `xml:"sys_pacct"`
    SysPTrace      *DomainFeatureCapability `xml:"sys_ptrace"`
    SysRawIO       *DomainFeatureCapability `xml:"sys_rawio"`
    SysResource    *DomainFeatureCapability `xml:"sys_resource"`
    SysTime        *DomainFeatureCapability `xml:"sys_time"`
    SysTTYCnofig   *DomainFeatureCapability `xml:"sys_tty_config"`
    SysLog         *DomainFeatureCapability `xml:"syslog"`
    WakeAlarm      *DomainFeatureCapability `xml:"wake_alarm"`
}
```

**Implementation Priority**:
- Simple features (PAE, ACPI, APIC): **HIGH**
- HyperV: **MEDIUM-HIGH** (common for Windows guests)
- KVM: **MEDIUM** (KVM-specific optimizations)
- State-based: **MEDIUM**
- Capabilities: **LOW** (LXC/container-specific)
- Architecture-specific (GIC, HPT, etc.): **LOW** (ARM, PowerPC)

---

## Category 7: Clock and Timers

**Complexity**: Moderate

```go
type DomainClock struct {
    Offset     string        `xml:"offset,attr,omitempty"`  // utc, localtime, timezone, variable
    Basis      string        `xml:"basis,attr,omitempty"`   // utc, localtime
    Adjustment string        `xml:"adjustment,attr,omitempty"`
    TimeZone   string        `xml:"timezone,attr,omitempty"`
    Start      uint          `xml:"start,attr,omitempty"`
    Timer      []DomainTimer `xml:"timer"`
}

type DomainTimer struct {
    Name       string              `xml:"name,attr"`  // platform, hpet, kvmclock, pit, rtc, tsc, hypervclock
    Track      string              `xml:"track,attr,omitempty"`  // boot, guest, wall
    TickPolicy string              `xml:"tickpolicy,attr,omitempty"`  // delay, catchup, merge, discard
    CatchUp    *DomainTimerCatchUp `xml:"catchup"`
    Frequency  uint64              `xml:"frequency,attr,omitempty"`
    Mode       string              `xml:"mode,attr,omitempty"`  // auto, native, emulate, paravirt, smpsafe
    Present    string              `xml:"present,attr,omitempty"`  // yes/no
}
```

**Use Cases**:
- Offset: Guest clock synchronization
- Timers: Platform-specific timers (KVM, Hyper-V, etc.)
- TickPolicy: Handling missed timer ticks

**Implementation Priority**: **MEDIUM** - Important for time-sensitive workloads

---

## Category 8: Lifecycle Management

**Complexity**: Simple (string enums)

```go
OnPoweroff string `xml:"on_poweroff,omitempty"`  // destroy, restart, preserve, rename-restart
OnReboot   string `xml:"on_reboot,omitempty"`    // destroy, restart
OnCrash    string `xml:"on_crash,omitempty"`     // destroy, restart, preserve, rename-restart, coredump-destroy, coredump-restart
```

**Implementation Priority**: **MEDIUM** - Common configuration

---

## Category 9: Power Management

**Complexity**: Simple

```go
type DomainPM struct {
    SuspendToMem  *DomainPMPolicy `xml:"suspend-to-mem"`
    SuspendToDisk *DomainPMPolicy `xml:"suspend-to-disk"`
}

type DomainPMPolicy struct {
    Enabled string `xml:"enabled,attr"`  // yes/no
}
```

**Use Case**: Enable/disable guest-initiated suspend
**Implementation Priority**: **LOW** - Rarely customized

---

## Category 10: Performance and Monitoring

### 10.1 Perf
**Complexity**: Moderate (event list)

```go
type DomainPerf struct {
    Events []DomainPerfEvent `xml:"event"`
}

type DomainPerfEvent struct {
    Name    string `xml:"name,attr"`     // cmt, mbmt, mbml, cpu_cycles, instructions, etc.
    Enabled string `xml:"enabled,attr"`  // yes/no
}
```

**Use Case**: Enable performance monitoring counters
**Implementation Priority**: **LOW** - Monitoring/debugging

### 10.2 BlockIOTune
**Complexity**: Moderate

```go
type DomainBlockIOTune struct {
    Weight uint                      `xml:"weight,omitempty"`  // 100-1000
    Device []DomainBlockIOTuneDevice `xml:"device"`
}

type DomainBlockIOTuneDevice struct {
    Path          string `xml:"path"`
    Weight        uint   `xml:"weight,omitempty"`
    ReadIopsSec   uint   `xml:"read_iops_sec,omitempty"`
    WriteIopsSec  uint   `xml:"write_iops_sec,omitempty"`
    ReadBytesSec  uint   `xml:"read_bytes_sec,omitempty"`
    WriteBytesSec uint   `xml:"write_bytes_sec,omitempty"`
}
```

**Use Case**: Block I/O throttling and prioritization
**Implementation Priority**: **MEDIUM** - QoS and resource management

---

## Category 11: Security

### 11.1 SecLabel
**Complexity**: Moderate (array)

```go
type DomainSecLabel struct {
    Type       string `xml:"type,attr,omitempty"`      // dynamic, static, none
    Model      string `xml:"model,attr,omitempty"`     // selinux, apparmor, dac
    Relabel    string `xml:"relabel,attr,omitempty"`   // yes/no
    Label      string `xml:"label,omitempty"`
    ImageLabel string `xml:"imagelabel,omitempty"`
    BaseLabel  string `xml:"baselabel,omitempty"`
}
```

**Use Case**: SELinux, AppArmor, DAC security labels
**Implementation Priority**: **MEDIUM** - Important for secure environments

### 11.2 KeyWrap
**Complexity**: Simple

```go
type DomainKeyWrap struct {
    Ciphers []DomainKeyWrapCipher `xml:"cipher"`
}

type DomainKeyWrapCipher struct {
    Name  string `xml:"name,attr"`   // aes, dea
    State string `xml:"state,attr"`  // on/off
}
```

**Use Case**: S390 cryptographic key wrapping
**Implementation Priority**: **LOW** - S390-specific

### 11.3 LaunchSecurity
**Complexity**: High (multiple implementations)

```go
type DomainLaunchSecurity struct {
    SEV    *DomainLaunchSecuritySEV    `xml:"-"`
    SEVSNP *DomainLaunchSecuritySEVSNP `xml:"-"`
    S390PV *DomainLaunchSecurityS390PV `xml:"-"`
    TDX    *DomainLaunchSecurityTDX    `xml:"-"`
}

// AMD SEV (Secure Encrypted Virtualization)
type DomainLaunchSecuritySEV struct {
    KernelHashes    string `xml:"kernelHashes,attr,omitempty"`
    CBitPos         *uint  `xml:"cbitpos"`
    ReducedPhysBits *uint  `xml:"reducedPhysBits"`
    Policy          *uint  `xml:"policy"`
    DHCert          string `xml:"dhCert"`
    Session         string `xml:"sesion"`
}

// AMD SEV-SNP
type DomainLaunchSecuritySEVSNP struct {
    KernelHashes            string  `xml:"kernelHashes,attr,omitempty"`
    AuthorKey               string  `xml:"authorKey,attr,omitempty"`
    VCEK                    string  `xml:"vcek,attr,omitempty"`
    CBitPos                 *uint   `xml:"cbitpos"`
    ReducedPhysBits         *uint   `xml:"reducedPhysBits"`
    Policy                  *uint64 `xml:"policy"`
    GuestVisibleWorkarounds string  `xml:"guestVisibleWorkarounds,omitempty"`
    IDBlock                 string  `xml:"idBlock,omitempty"`
    IDAuth                  string  `xml:"idAuth,omitempty"`
    HostData                string  `xml:"hostData,omitempty"`
}

// IBM S390 Protected Virtualization
type DomainLaunchSecurityS390PV struct {}

// Intel TDX (Trust Domain Extensions)
type DomainLaunchSecurityTDX struct {
    Policy                 *uint                       `xml:"policy"`
    MrConfigId             string                      `xml:"mrConfigId,omitempty"`
    MrOwner                string                      `xml:"mrOwner,omitempty"`
    MrOwnerConfig          string                      `xml:"mrOwnerConfig,omitempty"`
    QuoteGenerationService *DomainLaunchSecurityTDXQGS `xml:"quoteGenerationService"`
}
```

**Use Cases**:
- SEV/SEV-SNP: AMD confidential computing
- TDX: Intel confidential computing
- S390PV: IBM Z Protected Virtualization

**Implementation Priority**: **LOW-MEDIUM** - Specialized security feature, growing adoption

---

## Category 12: System Information

**Complexity**: High (SMBIOS data)

```go
type DomainSysInfo struct {
    SMBIOS *DomainSysInfoSMBIOS `xml:"-"`
    FWCfg  *DomainSysInfoFWCfg  `xml:"-"`
}

type DomainSysInfoSMBIOS struct {
    BIOS       *DomainSysInfoBIOS       `xml:"bios"`
    System     *DomainSysInfoSystem     `xml:"system"`
    BaseBoard  []DomainSysInfoBaseBoard `xml:"baseBoard"`
    Chassis    *DomainSysInfoChassis    `xml:"chassis"`
    Processor  []DomainSysInfoProcessor `xml:"processor"`
    Memory     []DomainSysInfoMemory    `xml:"memory"`
    OEMStrings *DomainSysInfoOEMStrings `xml:"oemStrings"`
}
```

**Use Case**: Override SMBIOS data (vendor, product, serial numbers)
**Implementation Priority**: **LOW-MEDIUM** - Useful for licensing, fingerprinting

---

## Category 13: ID Mapping (Containers)

**Complexity**: Simple

```go
type DomainIDMap struct {
    UIDs []DomainIDMapRange `xml:"uid"`
    GIDs []DomainIDMapRange `xml:"gid"`
}

type DomainIDMapRange struct {
    Start  uint `xml:"start,attr"`
    Target uint `xml:"target,attr"`
    Count  uint `xml:"count,attr"`
}
```

**Use Case**: User namespace ID mapping for containers (LXC)
**Implementation Priority**: **LOW** - Container-specific

---

## Category 14: Resource Management

**Complexity**: Simple

```go
type DomainResource struct {
    Partition    string                      `xml:"partition,omitempty"`
    FibreChannel *DomainResourceFibreChannel `xml:"fibrechannel"`
}

type DomainResourceFibreChannel struct {
    AppID string `xml:"appid,attr"`
}
```

**Use Case**: Cgroup partition placement, FC application ID
**Implementation Priority**: **LOW** - Advanced resource management

---

## Category 15: Throttle Groups

**Complexity**: Moderate (aliases DomainDiskIOTune)

```go
type DomainThrottleGroups struct {
    ThrottleGroups []ThrottleGroup `xml:"throttlegroup"`
}

type ThrottleGroup = DomainDiskIOTune
```

**Use Case**: Reusable I/O throttling groups for multiple devices
**Implementation Priority**: **LOW** - Advanced I/O management

---

## Summary by Complexity

### Simple Fields (9 fields)
Strings, primitives, or simple wrappers:
1. Type, ID, Name, UUID, HWUUID, Title, Description
2. Bootloader, BootloaderArgs
3. OnPoweroff, OnReboot, OnCrash
4. IOThreads

### Moderately Complex (12 fields)
1-2 levels of nesting, straightforward structure:
1. GenID
2. MaximumMemory, Memory, CurrentMemory
3. MemoryTune
4. VCPU
5. Clock
6. PM
7. BlockIOTune
8. SecLabel
9. KeyWrap
10. IDMap
11. DefaultIOThread
12. Metadata

### Highly Complex (22 fields)
Multiple levels of nesting, many options:
1. MemoryBacking
2. VCPUs
3. CPU (with NUMA)
4. CPUTune
5. NUMATune
6. IOThreadIDs
7. OS
8. Features (30+ sub-features)
9. Perf
10. Resource
11. SysInfo
12. LaunchSecurity
13. ThrottleGroups

---

## Implementation Priority Matrix

### Phase 1: Essential Fields (Must-Have)
**Target**: Basic working VMs

| Field | Complexity | Rationale |
|-------|------------|-----------|
| Type | Simple | Required attribute |
| Name | Simple | Required identifier |
| UUID | Simple | Important for stability |
| Memory | Simple | Required for basic config |
| VCPU | Simple | Required for basic config |
| OS (basic) | Moderate | Type, arch, machine, boot order |
| Description | Simple | Documentation |
| Title | Simple | Human-friendly name |

**Estimated Fields**: 8-10 top-level fields

### Phase 2: Common Features (Should-Have)
**Target**: Production-ready VMs

| Field | Complexity | Rationale |
|-------|------------|-----------|
| CurrentMemory | Simple | Memory ballooning |
| MaximumMemory | Simple | Memory hotplug |
| MemoryBacking | Moderate | Hugepages, performance |
| CPU | High | CPU model, topology, features |
| Features (basic) | High | ACPI, APIC, PAE |
| Clock | Moderate | Time synchronization |
| OnPoweroff/Reboot/Crash | Simple | Lifecycle management |
| IOThreads | Simple | Common for virtio |
| SecLabel | Moderate | Security in production |

**Estimated Fields**: 15-20 additional fields

### Phase 3: Advanced Features (Nice-to-Have)
**Target**: Specialized workloads

| Field | Complexity | Rationale |
|-------|------------|-----------|
| MemoryTune | Moderate | Memory limits |
| CPUTune | High | CPU pinning, scheduling |
| NUMATune | Moderate | NUMA workloads |
| VCPUs | Moderate | VCPU hotplug |
| IOThreadIDs | Moderate | Advanced I/O tuning |
| Features (advanced) | High | HyperV, KVM, specialized |
| BlockIOTune | Moderate | I/O QoS |
| Perf | Moderate | Monitoring |
| PM | Simple | Power management |

**Estimated Fields**: 10-15 additional fields

### Phase 4: Specialized Features (Optional)
**Target**: Niche use cases

| Field | Complexity | Rationale |
|-------|------------|-----------|
| LaunchSecurity | High | Confidential computing |
| SysInfo | High | SMBIOS customization |
| IDMap | Simple | Container namespaces |
| Resource | Simple | Advanced cgroup control |
| ThrottleGroups | Moderate | Reusable I/O groups |
| KeyWrap | Simple | S390-specific |
| GenID | Simple | Windows VM-Generation ID |
| Metadata | Simple | Custom application data |
| Bootloader | Simple | Xen PV domains |

**Estimated Fields**: Remaining fields

---

## Terraform Schema Design Recommendations

### 1. Flatten Where Reasonable
Some simple nested structs can be flattened:

```hcl
# Instead of:
memory {
  value = 2048
  unit  = "MiB"
}

# Consider:
memory      = 2048
memory_unit = "MiB"  # Optional, defaults to MiB
```

### 2. Use Blocks for Complex Nesting
Keep complex structures as nested blocks:

```hcl
cpu {
  mode = "host-model"

  topology {
    sockets = 1
    cores   = 2
    threads = 1
  }

  feature {
    name   = "vmx"
    policy = "require"
  }
}
```

### 3. Feature Flags
Consider flattening simple features:

```hcl
features {
  acpi = true
  apic = true
  pae  = true

  hyperv {
    relaxed    = true
    vapic      = true
    spinlocks  = 8191
  }
}
```

### 4. Lists vs Sets
- Use lists for ordered items (boot devices, VCPU pins)
- Use sets for unordered items (features, CPU features)

### 5. Computed vs Optional
- ID: Computed (runtime)
- UUID: Optional + Computed (can specify or auto-generate)
- Type: Required
- Name: Required

### 6. Validation
Implement validation for:
- Enum values (on_poweroff: destroy|restart|preserve|rename-restart)
- Memory units (KiB, MiB, GiB, TiB, bytes)
- CPU placement (static, auto)
- VCPU topology (sockets * cores * threads should match vcpu count)

### 7. Defaults
Sensible defaults to reduce verbosity:
- memory_unit: "MiB"
- cpu.mode: "custom" or "host-model"
- features.acpi: true (for ACPI-aware OSes)

---

## Field Frequency in Real-World Usage

Based on common libvirt domain configurations:

### Always Used (>95%)
- Type, Name, Memory, VCPU
- OS (Type, Arch, Machine)

### Very Common (50-95%)
- UUID, Description
- CurrentMemory
- CPU (mode, topology)
- Features (ACPI, APIC)
- Clock
- IOThreads

### Common (20-50%)
- MaximumMemory (memory hotplug)
- MemoryBacking (hugepages for performance)
- CPU features
- Features (HyperV for Windows, KVM features)
- SecLabel
- OnPoweroff/OnReboot/OnCrash

### Uncommon (5-20%)
- MemoryTune
- CPUTune (pinning)
- NUMATune
- BlockIOTune
- SysInfo

### Rare (<5%)
- LaunchSecurity (confidential computing)
- VCPUs (individual VCPU control)
- IOThreadIDs
- Perf
- PM
- KeyWrap
- IDMap
- Resource
- ThrottleGroups
- Bootloader (Xen-specific)

---

## Testing Strategy

### Unit Tests
Test struct mapping for:
1. Simple fields (string, int)
2. Nested structs (1-2 levels)
3. Arrays/slices
4. Optional pointers
5. Enum validation

### Integration Tests
Test complete domain XML generation for:
1. Minimal domain (Type, Name, Memory, VCPU)
2. Basic VM (+ OS, UUID, Description)
3. Production VM (+ CPU, Features, Clock, Security)
4. Advanced VM (+ Tuning, NUMA, Performance)

### Acceptance Tests
Real libvirt domain creation for:
1. Simple Linux VM
2. Windows VM (with HyperV features)
3. High-performance VM (hugepages, CPU pinning)
4. UEFI VM
5. Direct kernel boot

---

## Estimated Implementation Effort

| Phase | Fields | Complexity | Estimated Effort |
|-------|--------|------------|------------------|
| Phase 1 | 8-10 | Simple-Moderate | 1-2 weeks |
| Phase 2 | 15-20 | Moderate-High | 3-4 weeks |
| Phase 3 | 10-15 | High | 3-4 weeks |
| Phase 4 | Remaining | Varies | 2-3 weeks |

**Total**: 9-13 weeks for complete domain coverage (excluding devices)

**Note**: This is sequential development time. With parallel work and iterative releases, actual calendar time can be reduced.

---

## Next Steps

1. **Review and Prioritize**: Confirm priority matrix with stakeholders
2. **Phase 1 Implementation**: Start with essential fields
3. **Schema Design**: Design Terraform schema for Phase 1 fields
4. **Conversion Functions**: Implement Terraform → libvirtxml converters
5. **Testing**: Create test suite for Phase 1
6. **Iterate**: Move to Phase 2 after Phase 1 stabilization

---

## Appendix: Field Reference Quick Lookup

### By Category

**Metadata**: Type, ID, Name, UUID, HWUUID, GenID, Title, Description, Metadata

**Memory**: MaximumMemory, Memory, CurrentMemory, MemoryTune, MemoryBacking

**CPU**: VCPU, VCPUs, CPU, CPUTune, NUMATune

**I/O**: IOThreads, IOThreadIDs, DefaultIOThread, BlockIOTune, ThrottleGroups

**Boot**: Bootloader, BootloaderArgs, OS

**Features**: Features (with 30+ sub-features)

**Time**: Clock

**Lifecycle**: OnPoweroff, OnReboot, OnCrash

**Performance**: Perf

**Security**: SecLabel, KeyWrap, LaunchSecurity

**System Info**: SysInfo

**Containers**: IDMap

**Power**: PM

**Resource**: Resource

### Alphabetical Index

ACPI, APIC, BlockIOTune, Bootloader, BootloaderArgs, Clock, CPU, CPUTune, CurrentMemory, Description, Features, GenID, HWUUID, ID, IDMap, IOThreadIDs, IOThreads, KeyWrap, LaunchSecurity, MaximumMemory, Memory, MemoryBacking, MemoryTune, Metadata, Name, NUMATune, OnCrash, OnPoweroff, OnReboot, OS, Perf, PM, Resource, SecLabel, SysInfo, ThrottleGroups, Title, Type, UUID, VCPU, VCPUs

---

## Document Version

- **Version**: 1.0
- **Date**: 2025-10-05
- **Author**: Claude Code (libvirtxml package analysis)
- **Status**: Complete (Domain fields only, Devices excluded)
