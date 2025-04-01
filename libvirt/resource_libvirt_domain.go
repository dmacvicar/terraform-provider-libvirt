package libvirt

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/dmacvicar/terraform-provider-libvirt/libvirt/helper/suppress"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"libvirt.org/go/libvirtxml"
)

type pendingMapping struct {
	mac         string
	hostname    string
	networkName string
}

func init() {
	spew.Config.Indent = "\t"
}

func resourceLibvirtDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLibvirtDomainCreate,
		ReadContext:   resourceLibvirtDomainRead,
		DeleteContext: resourceLibvirtDomainDelete,
		UpdateContext: resourceLibvirtDomainUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			//nolint:mnd
			Create: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"metadata": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"vcpu": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
				ForceNew: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  defaultDomainMemoryMiB,
				ForceNew: true,
			},
			"firmware": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "kvm",
			},
			"nvram": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"file": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
						"template": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
					},
				},
			},
			"running": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: false,
				Required: false,
			},
			"cloudinit": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"coreos_ignition": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "",
			},
			"fw_cfg_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "opt/com.coreos/config",
			},
			"filesystem": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"accessmode": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "mapped",
						},
						"source": {
							Type:     schema.TypeString,
							Required: true,
						},
						"target": {
							Type:     schema.TypeString,
							Required: true,
						},
						"readonly": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},
			"disk": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"volume_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"file": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"scsi": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
							Default:  false,
						},
						"wwn": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"block_device": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"network_interface": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"network_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"bridge": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"vepa": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"private": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"macvtap": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"passthrough": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"hostname": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"mac": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							DiffSuppressFunc: suppress.CaseDifference,
						},
						"wait_for_lease": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"addresses": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"graphics": {
				ForceNew: true,
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "spice",
							ForceNew: true,
						},
						"autoport": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						"listen_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "none",
							ForceNew: true,
						},
						"listen_address": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "127.0.0.1",
							ForceNew: true,
						},
						"websocket": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"video": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "cirrus",
						},
					},
				},
			},
			"console": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"source_path": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"source_host": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "127.0.0.1",
						},
						"source_service": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "0",
						},
						"target_port": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"target_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"cpu": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"autostart": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"machine": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"arch": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"boot_device": {
				Type:     schema.TypeList,
				Optional: true,
				Required: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dev": {
							Type:     schema.TypeList,
							Optional: true,
							Required: false,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"emulator": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"kernel": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
			"initrd": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
			"cmdline": {
				Type:     schema.TypeList,
				Optional: true,
				Required: false,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
			},
			"qemu_agent": {
				Type:     schema.TypeBool,
				Optional: true,
				Required: false,
				Default:  false,
				ForceNew: false,
			},
			"tpm": {
				Type:     schema.TypeList,
				Optional: true,
				// Error defining libvirt domain: unsupported configuration: only a single TPM non-proxy device is supported
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"model": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"backend_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "emulator",
						},
						"backend_device_path": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"backend_encryption_secret": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"backend_version": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"backend_persistent_state": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"xml": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"xslt": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"launch_security": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"sev", "sev-snp", "s390-pv",
							}, false),
							Description: "Launch security type (sev, sev-snp, or s390-pv)",
						},
						// SEV specific settings
						"cbitpos": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "C-bit position for SEV",
						},
						"reduced_phys_bits": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Reduced physical address bits for SEV",
						},
						"policy": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Policy value for SEV",
						},
						"dh_cert": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Diffie-Hellman certificate for SEV",
						},
						"session": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Session information for SEV",
						},
						// SEV-SNP specific settings
						"kernel_hashes": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Kernel hashes for SEV-SNP",
						},
						"author_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Author key for SEV-SNP",
						},
						"vcek": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "VCEK for SEV-SNP",
						},
						"guest_visible_workarounds": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Guest visible workarounds for SEV-SNP",
						},
						"id_block": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ID block for SEV-SNP",
						},
						"id_auth": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ID auth for SEV-SNP",
						},
						"host_data": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Host data for SEV-SNP",
						},
					},
				},
			},
			"memory_backing": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"file", "anonymous", "memfd",
							}, false),
							Description: "Memory backing source type (file, anonymous, or memfd)",
						},
						"access_mode": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"shared", "private",
							}, false),
							Description: "Memory access mode (shared or private)",
						},
						"allocation_mode": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"immediate", "ondemand",
							}, false),
							Description: "Memory allocation mode (immediate or ondemand)",
						},
						"discard": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Enable memory discard",
						},
						"nosharepages": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Disable memory sharing between guests",
						},
						"locked": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Lock memory to prevent swapping",
						},
						"hugepages": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Huge page size in KiB",
									},
									"nodeset": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "NUMA nodes to allocate huge pages from",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func setLaunchSecurity(d *schema.ResourceData, domainDef *libvirtxml.Domain) error {
	if launchSecurityList, ok := d.GetOk("launch_security"); ok {
		if len(launchSecurityList.([]interface{})) > 0 {
			launchSecurity := launchSecurityList.([]interface{})[0].(map[string]interface{})
			secType := launchSecurity["type"].(string)

			// Initialize the launch security element
			domainDef.LaunchSecurity = &libvirtxml.DomainLaunchSecurity{}

			switch secType {
			case "sev":
				sev := &libvirtxml.DomainLaunchSecuritySEV{}

				if v, ok := launchSecurity["cbitpos"]; ok && v.(int) > 0 {
					cbitpos := uint(v.(int))
					sev.CBitPos = &cbitpos
				}

				if v, ok := launchSecurity["reduced_phys_bits"]; ok && v.(int) > 0 {
					reducedPhysBits := uint(v.(int))
					sev.ReducedPhysBits = &reducedPhysBits
				}

				if v, ok := launchSecurity["policy"]; ok && v.(int) > 0 {
					policy := uint(v.(int))
					sev.Policy = &policy
				}

				if v, ok := launchSecurity["dh_cert"]; ok {
					sev.DHCert = v.(string)
				}

				if v, ok := launchSecurity["session"]; ok {
					sev.Session = v.(string)
				}

				domainDef.LaunchSecurity.SEV = sev

			case "sev-snp":
				sevSnp := &libvirtxml.DomainLaunchSecuritySEVSNP{}

				if v, ok := launchSecurity["kernel_hashes"]; ok {
					sevSnp.KernelHashes = v.(string)
				}

				if v, ok := launchSecurity["author_key"]; ok {
					sevSnp.AuthorKey = v.(string)
				}

				if v, ok := launchSecurity["vcek"]; ok {
					sevSnp.VCEK = v.(string)
				}

				if v, ok := launchSecurity["cbitpos"]; ok && v.(int) > 0 {
					cbitpos := uint(v.(int))
					sevSnp.CBitPos = &cbitpos
				}

				if v, ok := launchSecurity["reduced_phys_bits"]; ok && v.(int) > 0 {
					reducedPhysBits := uint(v.(int))
					sevSnp.ReducedPhysBits = &reducedPhysBits
				}

				if v, ok := launchSecurity["policy"]; ok && v.(int) > 0 {
					policy := uint64(v.(int))
					sevSnp.Policy = &policy
				}

				if v, ok := launchSecurity["guest_visible_workarounds"]; ok {
					sevSnp.GuestVisibleWorkarounds = v.(string)
				}

				if v, ok := launchSecurity["id_block"]; ok {
					sevSnp.IDBlock = v.(string)
				}

				if v, ok := launchSecurity["id_auth"]; ok {
					sevSnp.IDAuth = v.(string)
				}

				if v, ok := launchSecurity["host_data"]; ok {
					sevSnp.HostData = v.(string)
				}

				domainDef.LaunchSecurity.SEVSNP = sevSnp

			case "s390-pv":
				// S390 Protected Virtualization doesn't have any additional parameters
				domainDef.LaunchSecurity.S390PV = &libvirtxml.DomainLaunchSecurityS390PV{}
			}
		}
	}

	return nil
}

func resourceLibvirtDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Create resource libvirt_domain")

	virConn := meta.(*Client).libvirt

	domainDef, err := newDomainDefForConnection(virConn, d)
	if err != nil {
		return diag.FromErr(err)
	}

	if name, ok := d.GetOk("name"); ok {
		domainDef.Name = name.(string)
	}

	if cpuMode, ok := d.GetOk("cpu.0.mode"); ok {
		domainDef.CPU = &libvirtxml.DomainCPU{
			Mode: cpuMode.(string),
		}
	}

	domainDef.Memory = &libvirtxml.DomainMemory{
		Value: uint(d.Get("memory").(int)),
		Unit:  "MiB",
	}
	domainDef.VCPU = &libvirtxml.DomainVCPU{
		Value: uint(d.Get("vcpu").(int)),
	}
	domainDef.Description = d.Get("description").(string)

	domainDef.OS.Kernel = d.Get("kernel").(string)
	domainDef.OS.Initrd = d.Get("initrd").(string)
	domainDef.OS.Type.Arch = d.Get("arch").(string)

	domainDef.Devices.Emulator = d.Get("emulator").(string)

	if v := os.Getenv("TERRAFORM_LIBVIRT_TEST_DOMAIN_TYPE"); v != "" {
		domainDef.Type = v
	} else {
		domainDef.Type = d.Get("type").(string)
	}

	arch, err := getHostArchitecture(virConn)
	if err != nil {
		return diag.Errorf("error retrieving host architecture: %s", err)
	}

	if err := setGraphics(d, &domainDef, arch); err != nil {
		return diag.FromErr(err)
	}

	setVideo(d, &domainDef)

	if err := setConsoles(d, &domainDef); err != nil {
		return diag.FromErr(err)
	}

	setCmdlineArgs(d, &domainDef)
	setFirmware(d, &domainDef)
	setBootDevices(d, &domainDef)
	setTPMs(d, &domainDef)

	if err := setCoreOSIgnition(d, &domainDef, arch); err != nil {
		return diag.FromErr(err)
	}

	if err := setDisks(d, &domainDef, virConn); err != nil {
		return diag.FromErr(err)
	}

	if err := setFilesystems(d, &domainDef); err != nil {
		return diag.FromErr(err)
	}

	if err := setCloudinit(d, &domainDef, virConn); err != nil {
		return diag.FromErr(err)
	}

	if err := setMemoryBacking(d, &domainDef); err != nil {
		return diag.FromErr(err)
	}

	if err := setLaunchSecurity(d, &domainDef); err != nil {
		return diag.FromErr(err)
	}

	var waitForLeases []*libvirtxml.DomainInterface
	partialNetIfaces := make(map[string]*pendingMapping, d.Get("network_interface.#").(int))

	if err := setNetworkInterfaces(d, &domainDef, virConn, partialNetIfaces, &waitForLeases); err != nil {
		return diag.FromErr(err)
	}

	connectURI, err := virConn.ConnectGetUri()
	if err != nil {
		return diag.Errorf("error retrieving libvirt connection URI: %s", err)
	}
	log.Printf("[INFO] Creating libvirt domain at %s", connectURI)

	data, err := xmlMarshallIndented(domainDef)
	if err != nil {
		return diag.Errorf("error serializing libvirt domain: %s", err)
	}
	log.Printf("[DEBUG] Generated XML for libvirt domain:\n%s", data)

	data, err = transformResourceXML(data, d)
	if err != nil {
		return diag.Errorf("error applying XSLT stylesheet: %s", err)
	}

	domain, err := virConn.DomainDefineXML(data)
	if err != nil {
		return diag.Errorf("error defining libvirt domain: %s", err)
	}

	if autostart, ok := d.GetOk("autostart"); ok {
		var autostartInt int32
		if autostart.(bool) {
			autostartInt = 1
		}
		err = virConn.DomainSetAutostart(domain, autostartInt)
		if err != nil {
			return diag.Errorf("error setting autostart for domain: %s", err)
		}
	}

	err = virConn.DomainCreate(domain)
	if err != nil {
		return diag.Errorf("error creating libvirt domain: %s", err)
	}

	id := uuidString(domain.UUID)
	d.SetId(id)
	log.Printf("[INFO] Domain ID: %s", d.Id())

	if len(waitForLeases) > 0 {
		if err := waitForStateDomainLeaseDone(ctx, virConn, domain, waitForLeases, d); err != nil {
			return diag.FromErr(err)
		}
	}

	// We save runnig state to not mix what we have and what we want
	requiredStatus := d.Get("running")

	if diag := resourceLibvirtDomainRead(ctx, d, meta); diag.HasError() {
		return diag
	}

	d.Set("running", requiredStatus)

	// we must read devices again in order to set some missing ip/MAC/host mappings
	for i := 0; i < d.Get("network_interface.#").(int); i++ {
		prefix := fmt.Sprintf("network_interface.%d", i)
		mac := strings.ToUpper(d.Get(prefix + ".mac").(string))
		log.Printf("[DEBUG] Reading network_interface.%d with MAC: %s\n", i, mac)

		// if we were waiting for an IP address for this MAC, go ahead.
		if pending, ok := partialNetIfaces[mac]; ok {
			// we should have the address now
			addressesI, ok := d.GetOk(prefix + ".addresses")
			if !ok {
				log.Printf("Did not obtain the IP address for MAC=%s", mac)
				continue
			}

			network, err := virConn.NetworkLookupByName(pending.networkName)
			if err != nil {
				log.Printf("Can't retrieve network '%s'", pending.networkName)
				continue
			}

			for _, addressI := range addressesI.([]interface{}) {
				address := addressI.(string)
				log.Printf("[INFO] Finally adding IP/MAC/host=%s/%s/%s", address, mac, pending.hostname)

				err = updateOrAddHost(virConn, network, address, mac, pending.hostname)
				if err != nil {
					log.Printf("Could not add IP/MAC/host=%s/%s/%s: %s", address, mac, pending.hostname, err)
				}
			}
		}
	}

	if err := destroyDomainByUserRequest(virConn, d, domain); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func setMemoryBacking(d *schema.ResourceData, domainDef *libvirtxml.Domain) error {
	if memBackingList, ok := d.GetOk("memory_backing"); ok {
		if len(memBackingList.([]interface{})) > 0 {
			memBacking := memBackingList.([]interface{})[0].(map[string]interface{})

			// Initialize the memory backing element if not present
			if domainDef.MemoryBacking == nil {
				domainDef.MemoryBacking = &libvirtxml.DomainMemoryBacking{}
			}

			// Set source type
			if sourceType, ok := memBacking["source_type"].(string); ok && sourceType != "" {
				if domainDef.MemoryBacking.MemorySource == nil {
					domainDef.MemoryBacking.MemorySource = &libvirtxml.DomainMemorySource{}
				}
				domainDef.MemoryBacking.MemorySource.Type = sourceType
			}

			// Set access mode
			if accessMode, ok := memBacking["access_mode"].(string); ok && accessMode != "" {
				if domainDef.MemoryBacking.MemoryAccess == nil {
					domainDef.MemoryBacking.MemoryAccess = &libvirtxml.DomainMemoryAccess{}
				}
				domainDef.MemoryBacking.MemoryAccess.Mode = accessMode
			}

			// Set allocation mode
			if allocMode, ok := memBacking["allocation_mode"].(string); ok && allocMode != "" {
				if domainDef.MemoryBacking.MemoryAllocation == nil {
					domainDef.MemoryBacking.MemoryAllocation = &libvirtxml.DomainMemoryAllocation{}
				}
				domainDef.MemoryBacking.MemoryAllocation.Mode = allocMode
			}

			// Set discard
			if discard, ok := memBacking["discard"].(bool); ok && discard {
				domainDef.MemoryBacking.MemoryDiscard = &libvirtxml.DomainMemoryDiscard{}
			}

			// Set nosharepages
			if nosharepages, ok := memBacking["nosharepages"].(bool); ok && nosharepages {
				domainDef.MemoryBacking.MemoryNosharepages = &libvirtxml.DomainMemoryNosharepages{}
			}

			// Set locked
			if locked, ok := memBacking["locked"].(bool); ok && locked {
				domainDef.MemoryBacking.MemoryLocked = &libvirtxml.DomainMemoryLocked{}
			}

			// Set hugepages
			if hugepagesList, ok := memBacking["hugepages"].([]interface{}); ok && len(hugepagesList) > 0 {
				hugepages := &libvirtxml.DomainMemoryHugepages{
					Hugepages: []libvirtxml.DomainMemoryHugepage{},
				}

				for _, hpInterface := range hugepagesList {
					hp := hpInterface.(map[string]interface{})

					hugepage := libvirtxml.DomainMemoryHugepage{
						Size: uint(hp["size"].(int)),
					}

					if nodeset, ok := hp["nodeset"].(string); ok && nodeset != "" {
						hugepage.Nodeset = nodeset
					}

					hugepages.Hugepages = append(hugepages.Hugepages, hugepage)
				}

				domainDef.MemoryBacking.MemoryHugePages = hugepages
			}
		}
	}

	return nil
}

func resourceLibvirtDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Update resource libvirt_domain")

	virConn := meta.(*Client).libvirt

	uuid := parseUUID(d.Id())

	domain, err := virConn.DomainLookupByUUID(uuid)
	if err != nil {
		return diag.Errorf("error retrieving libvirt domain by update: %s", err)
	}

	domainRunningNow, err := domainIsRunning(virConn, domain)
	if err != nil {
		return diag.FromErr(err)
	}

	if !domainRunningNow {
		err = virConn.DomainCreate(domain)
		if err != nil {
			return diag.Errorf("error creating libvirt domain: %s", err)
		}
	}

	if d.HasChange("cloudinit") {
		cloudinitID, err := getCloudInitVolumeKeyFromTerraformID(d.Get("cloudinit").(string))
		if err != nil {
			return diag.FromErr(err)
		}

		disk, err := newDiskForCloudInit(virConn, cloudinitID)
		if err != nil {
			return diag.FromErr(err)
		}

		data, err := xml.Marshal(disk)
		if err != nil {
			return diag.Errorf("error serializing cloudinit disk: %s", err)
		}

		err = virConn.DomainUpdateDeviceFlags(domain,
			string(data),
			libvirt.DomainDeviceModifyConfig|libvirt.DomainDeviceModifyCurrent|libvirt.DomainDeviceModifyLive)
		if err != nil {
			return diag.Errorf("error while changing the cloudinit volume: %s", err)
		}
	}

	if d.HasChange("autostart") {
		var autoStart int32
		if d.Get("autostart").(bool) {
			autoStart = 1
		}

		err = virConn.DomainSetAutostart(domain, autoStart)
		if err != nil {
			return diag.Errorf("error setting autostart for domain: %s", err)
		}
	}

	netIfacesCount := d.Get("network_interface.#").(int)

	for i := 0; i < netIfacesCount; i++ {
		prefix := fmt.Sprintf("network_interface.%d", i)
		if d.HasChange(prefix+".hostname") || d.HasChange(prefix+".addresses") || d.HasChange(prefix+".mac") {
			networkUUID, ok := d.GetOk(prefix + ".network_id")
			log.Printf("[INFO] NetworkUUID: %s\n", networkUUID)
			if !ok {
				continue
			}

			uuid := parseUUID(networkUUID.(string))

			network, err := virConn.NetworkLookupByUUID(uuid)
			if err != nil {
				return diag.Errorf("can't retrieve network ID %s", networkUUID)
			}

			hostname := d.Get(prefix + ".hostname").(string)
			mac := d.Get(prefix + ".mac").(string)
			addresses := d.Get(prefix + ".addresses")
			for _, addressI := range addresses.([]interface{}) {
				address := addressI.(string)

				ip := net.ParseIP(address)
				if ip == nil {
					return diag.Errorf("could not parse addresses '%s'", address)
				}

				log.Printf("[INFO] Updating IP/MAC/host=%s/%s/%s in '%s' network", ip.String(), mac, hostname, network.Name)

				if err := updateOrAddHost(virConn, network, ip.String(), mac, hostname); err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	return nil
}

func resourceLibvirtDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Read resource libvirt_domain")

	virConn := meta.(*Client).libvirt

	uuid := parseUUID(d.Id())

	domain, err := virConn.DomainLookupByUUID(uuid)
	if err != nil {
		if isError(err, libvirt.ErrNoDomain) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving libvirt domain: %s", err)
	}

	xmlDesc, err := virConn.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return diag.Errorf("error retrieving libvirt domain XML description: %s", err)
	}

	log.Printf("[DEBUG] read: obtained XML desc for domain:\n%s", xmlDesc)

	domainDef, err := newDomainDefForConnection(virConn, d)
	if err != nil {
		return diag.FromErr(err)
	}

	err = xml.Unmarshal([]byte(xmlDesc), &domainDef)
	if err != nil {
		return diag.Errorf("error reading libvirt domain XML description: %s", err)
	}

	autostart, err := virConn.DomainGetAutostart(domain)
	if err != nil {
		return diag.Errorf("error reading domain autostart setting: %s", err)
	}
	_ = d.Set("autostart", autostart > 0)

	domainRunningNow, err := domainIsRunning(virConn, domain)
	if err != nil {
		return diag.Errorf("error reading domain running state : %s", err)
	}

	d.Set("name", domainDef.Name)
	d.Set("description", domainDef.Description)
	d.Set("vcpu", domainDef.VCPU.Value)

	switch domainDef.Memory.Unit {
	case "KiB":
		d.Set("memory", domainDef.Memory.Value/1024)
	case "MiB":
		d.Set("memory", domainDef.Memory.Value)
	default:
		return diag.Errorf("invalid memory unit : %s", domainDef.Memory.Unit)
	}

	if domainDef.OS.Loader != nil {
		d.Set("firmware", domainDef.OS.Loader.Path)
	}

	if domainDef.OS.NVRam != nil {
		nvram := map[string]interface{}{}
		if domainDef.OS.NVRam.NVRam != "" {
			nvram["file"] = domainDef.OS.NVRam.NVRam
		}

		if domainDef.OS.NVRam.Template != "" {
			nvram["template"] = domainDef.OS.NVRam.Template
		}

		d.Set("nvram", []map[string]interface{}{nvram})
	}

	if domainDef.CPU != nil {
		cpu := make(map[string]interface{})
		var cpus []map[string]interface{}
		if domainDef.CPU.Mode != "" {
			cpu["mode"] = domainDef.CPU.Mode
		}
		if len(cpu) > 0 {
			cpus = append(cpus, cpu)
			d.Set("cpu", cpus)
		}
	}

	d.Set("arch", domainDef.OS.Type.Arch)
	d.Set("running", domainRunningNow)

	cmdLines := splitKernelCmdLine(domainDef.OS.Cmdline)

	d.Set("cmdline", cmdLines)
	d.Set("kernel", domainDef.OS.Kernel)
	d.Set("initrd", domainDef.OS.Initrd)

	caps, err := getHostCapabilities(virConn)
	if err != nil {
		return diag.FromErr(err)
	}
	machine, err := getOriginalMachineName(caps, domainDef.OS.Type.Arch, domainDef.OS.Type.Type,
		domainDef.OS.Type.Machine)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("machine", machine)

	// Emulator is the same as the default don't set it in domainDef
	// or it will show as changed
	d.Set("emulator", domainDef.Devices.Emulator)
	var (
		disks []map[string]interface{}
		disk  map[string]interface{}
	)
	for _, diskDef := range domainDef.Devices.Disks {
		// network drives do not have a volume associated
		if diskDef.Source.Network != nil {
			if len(diskDef.Source.Network.Hosts) < 1 {
				return diag.Errorf("network disk does not contain any hosts")
			}
			url, err := url.Parse(fmt.Sprintf("%s://%s:%s%s",
				diskDef.Source.Network.Protocol,
				diskDef.Source.Network.Hosts[0].Name,
				diskDef.Source.Network.Hosts[0].Port,
				diskDef.Source.Network.Name))
			if err != nil {
				return diag.FromErr(err)
			}
			disk = map[string]interface{}{
				"url": url.String(),
			}
		} else if diskDef.Device == "cdrom" {
			// HACK we marked the disk as belonging to the cloudinit
			// resource so we can ignore it
			if diskDef.Serial == "cloudinit" {
				continue
			}

			disk = map[string]interface{}{
				"file": diskDef.Source.File.File,
			}
		} else if diskDef.Source.Block != nil {
			disk = map[string]interface{}{
				"block_device": diskDef.Source.Block.Dev,
			}
		} else if diskDef.Source.File != nil {
			// LEGACY way of handling volumes using "file", which we replaced
			// by the diskdef.Source.Volume once we realized it existed.
			// This code will be removed in future versions of the provider.
			virVol, err := virConn.StorageVolLookupByPath(diskDef.Source.File.File)
			if err != nil {
				return diag.Errorf("error retrieving volume for disk: %s", err)
			}

			disk = map[string]interface{}{
				"volume_id": virVol.Key,
			}
		} else {
			pool, err := virConn.StoragePoolLookupByName(diskDef.Source.Volume.Pool)
			if err != nil {
				return diag.Errorf("error retrieving pool for disk: %s", err)
			}

			virVol, err := virConn.StorageVolLookupByName(pool, diskDef.Source.Volume.Volume)
			if err != nil {
				return diag.Errorf("error retrieving volume for disk: %s", err)
			}

			disk = map[string]interface{}{
				"volume_id": virVol.Key,
			}
		}

		if diskDef.Target != nil && diskDef.Target.Bus == "scsi" {
			disk["scsi"] = true
			disk["wwn"] = diskDef.WWN
		} else {
			disk["scsi"] = false
		}

		disks = append(disks, disk)
	}

	if len(disks) > 0 {
		d.Set("disk", disks)
	}

	var filesystems []map[string]interface{}
	for _, fsDef := range domainDef.Devices.Filesystems {
		fs := map[string]interface{}{
			"accessmode": fsDef.AccessMode,
			"source":     fsDef.Source.Mount.Dir,
			"target":     fsDef.Target.Dir,
			"readonly":   fsDef.ReadOnly,
		}
		filesystems = append(filesystems, fs)
	}

	if len(filesystems) > 0 {
		d.Set("filesystem", filesystems)
	}

	// Read launch security configuration
	if domainDef.LaunchSecurity != nil {
		launchSecurity := make(map[string]interface{})

		if domainDef.LaunchSecurity.SEV != nil {
			launchSecurity["type"] = "sev"

			if domainDef.LaunchSecurity.SEV.CBitPos != nil {
				launchSecurity["cbitpos"] = int(*domainDef.LaunchSecurity.SEV.CBitPos)
			}

			if domainDef.LaunchSecurity.SEV.ReducedPhysBits != nil {
				launchSecurity["reduced_phys_bits"] = int(*domainDef.LaunchSecurity.SEV.ReducedPhysBits)
			}

			if domainDef.LaunchSecurity.SEV.Policy != nil {
				launchSecurity["policy"] = int(*domainDef.LaunchSecurity.SEV.Policy)
			}

			launchSecurity["dh_cert"] = domainDef.LaunchSecurity.SEV.DHCert
			launchSecurity["session"] = domainDef.LaunchSecurity.SEV.Session

		} else if domainDef.LaunchSecurity.SEVSNP != nil {
			launchSecurity["type"] = "sev-snp"

			launchSecurity["kernel_hashes"] = domainDef.LaunchSecurity.SEVSNP.KernelHashes
			launchSecurity["author_key"] = domainDef.LaunchSecurity.SEVSNP.AuthorKey
			launchSecurity["vcek"] = domainDef.LaunchSecurity.SEVSNP.VCEK

			if domainDef.LaunchSecurity.SEVSNP.CBitPos != nil {
				launchSecurity["cbitpos"] = int(*domainDef.LaunchSecurity.SEVSNP.CBitPos)
			}

			if domainDef.LaunchSecurity.SEVSNP.ReducedPhysBits != nil {
				launchSecurity["reduced_phys_bits"] = int(*domainDef.LaunchSecurity.SEVSNP.ReducedPhysBits)
			}

			if domainDef.LaunchSecurity.SEVSNP.Policy != nil {
				launchSecurity["policy"] = int(*domainDef.LaunchSecurity.SEVSNP.Policy)
			}

			launchSecurity["guest_visible_workarounds"] = domainDef.LaunchSecurity.SEVSNP.GuestVisibleWorkarounds
			launchSecurity["id_block"] = domainDef.LaunchSecurity.SEVSNP.IDBlock
			launchSecurity["id_auth"] = domainDef.LaunchSecurity.SEVSNP.IDAuth
			launchSecurity["host_data"] = domainDef.LaunchSecurity.SEVSNP.HostData

		} else if domainDef.LaunchSecurity.S390PV != nil {
			launchSecurity["type"] = "s390-pv"
		}

		d.Set("launch_security", []map[string]interface{}{launchSecurity})
	}

	// Read memory backing configuration
	if domainDef.MemoryBacking != nil {
		memBacking := make(map[string]interface{})

		if domainDef.MemoryBacking.MemorySource != nil {
			memBacking["source_type"] = domainDef.MemoryBacking.MemorySource.Type
		}

		if domainDef.MemoryBacking.MemoryAccess != nil {
			memBacking["access_mode"] = domainDef.MemoryBacking.MemoryAccess.Mode
		}

		if domainDef.MemoryBacking.MemoryAllocation != nil {
			memBacking["allocation_mode"] = domainDef.MemoryBacking.MemoryAllocation.Mode
		}

		if domainDef.MemoryBacking.MemoryDiscard != nil {
			memBacking["discard"] = true
		}

		if domainDef.MemoryBacking.MemoryNosharepages != nil {
			memBacking["nosharepages"] = true
		}

		if domainDef.MemoryBacking.MemoryLocked != nil {
			memBacking["locked"] = true
		}

		if domainDef.MemoryBacking.MemoryHugePages != nil && len(domainDef.MemoryBacking.MemoryHugePages.Hugepages) > 0 {
			hugepages := make([]map[string]interface{}, 0, len(domainDef.MemoryBacking.MemoryHugePages.Hugepages))

			for _, hp := range domainDef.MemoryBacking.MemoryHugePages.Hugepages {
				hugepage := make(map[string]interface{})
				hugepage["size"] = hp.Size

				if hp.Nodeset != "" {
					hugepage["nodeset"] = hp.Nodeset
				}

				hugepages = append(hugepages, hugepage)
			}

			memBacking["hugepages"] = hugepages
		}

		d.Set("memory_backing", []map[string]interface{}{memBacking})
	}

	// lookup interfaces with addresses
	ifacesWithAddr, err := domainGetIfacesInfo(virConn, domain, d)
	if err != nil {
		return diag.Errorf("error retrieving interface addresses: %s", err)
	}

	addressesForMac := func(mac string) []string {
		// look for an ip address and try to match it with the mac address
		// not sure if using the target device name is a better idea here
		var addrs []string
		for _, ifaceWithAddr := range ifacesWithAddr {
			if len(ifaceWithAddr.Hwaddr) > 0 && strings.ToUpper(ifaceWithAddr.Hwaddr[0]) == mac {
				for _, addr := range ifaceWithAddr.Addrs {
					addrs = append(addrs, addr.Addr)
				}
			}
		}
		return addrs
	}

	var netIfaces []map[string]interface{}
	for i, networkInterfaceDef := range domainDef.Devices.Interfaces {
		// we need it to read old values
		prefix := fmt.Sprintf("network_interface.%d", i)

		mac := strings.ToUpper(networkInterfaceDef.MAC.Address)
		netIface := map[string]interface{}{
			"network_id":     "",
			"network_name":   "",
			"bridge":         "",
			"vepa":           "",
			"private":        "",
			"macvtap":        "",
			"passthrough":    "",
			"mac":            mac,
			"hostname":       "",
			"wait_for_lease": false,
		}

		netIface["wait_for_lease"] = d.Get(prefix + ".wait_for_lease").(bool)
		netIface["hostname"] = d.Get(prefix + ".hostname").(string)
		netIface["addresses"] = addressesForMac(mac)
		log.Printf("[DEBUG] read: addresses for '%s': %+v", mac, netIface["addresses"])

		if networkInterfaceDef.Source.Network != nil {
			network, err := virConn.NetworkLookupByName(networkInterfaceDef.Source.Network.Network)
			if err != nil {
				return diag.Errorf("can't retrieve network ID for '%s'", networkInterfaceDef.Source.Network.Network)
			}

			netIface["network_id"] = uuidString(network.UUID)
			if err != nil {
				return diag.Errorf("can't retrieve network ID for '%s'", networkInterfaceDef.Source.Network.Network)
			}

			networkDef, err := getXMLNetworkDefFromLibvirt(virConn, network)
			if err != nil {
				return diag.FromErr(err)
			}

			netIface["network_name"] = networkInterfaceDef.Source.Network.Network

			// try to look for this MAC in the DHCP configuration for this VM
			if HasDHCP(networkDef) {
			hostnameSearch:
				for _, ip := range networkDef.IPs {
					if ip.DHCP != nil {
						for _, host := range ip.DHCP.Hosts {
							if strings.ToUpper(host.MAC) == netIface["mac"] {
								log.Printf("[DEBUG] read: hostname for '%s': '%s'", netIface["mac"], host.Name)
								netIface["hostname"] = host.Name
								break hostnameSearch
							}
						}
					}
				}
			}
		} else if networkInterfaceDef.Source.Bridge != nil {
			netIface["bridge"] = networkInterfaceDef.Source.Bridge.Bridge
		} else if networkInterfaceDef.Source.Direct != nil {
			switch networkInterfaceDef.Source.Direct.Mode {
			case "vepa":
				netIface["vepa"] = networkInterfaceDef.Source.Direct.Dev
			case "private":
				netIface["private"] = networkInterfaceDef.Source.Direct.Dev
			case "bridge":
				netIface["macvtap"] = networkInterfaceDef.Source.Direct.Dev
			case "passthrough":
				netIface["passthrough"] = networkInterfaceDef.Source.Direct.Dev
			}
		}
		netIfaces = append(netIfaces, netIface)
	}
	log.Printf("[DEBUG] read: ifaces for '%s':\n%s", domainDef.Name, spew.Sdump(netIfaces))

	if len(netIfaces) > 0 {
		d.Set("network_interface", netIfaces)
	}

	if len(ifacesWithAddr) > 0 {
		d.SetConnInfo(map[string]string{
			"type": "ssh",
			"host": ifacesWithAddr[0].Addrs[0].Addr,
		})
	}
	return nil
}

func resourceLibvirtDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Delete resource libvirt_domain")

	virConn := meta.(*Client).libvirt

	log.Printf("[DEBUG] Deleting domain %s", d.Id())

	uuid := parseUUID(d.Id())

	domain, err := virConn.DomainLookupByUUID(uuid)
	if err != nil {
		return diag.Errorf("error retrieving libvirt domain by delete: %s", err)
	}

	xmlDesc, err := virConn.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return diag.Errorf("error retrieving libvirt domain XML description: %s", err)
	}

	domainDef, err := newDomainDefForConnection(virConn, d)
	if err != nil {
		return diag.FromErr(err)
	}

	err = xml.Unmarshal([]byte(xmlDesc), &domainDef)
	if err != nil {
		return diag.Errorf("error reading libvirt domain XML description: %s", err)
	}

	state, _, err := virConn.DomainGetState(domain, 0)
	if err != nil {
		return diag.Errorf("couldn't get info about domain: %s", err)
	}

	if state == int32(libvirt.DomainRunning) || state == int32(libvirt.DomainPaused) {
		if err := virConn.DomainDestroy(domain); err != nil {
			return diag.Errorf("couldn't destroy libvirt domain: %s", err)
		}
	}

	if err := virConn.DomainUndefineFlags(domain, libvirt.DomainUndefineNvram|
		libvirt.DomainUndefineSnapshotsMetadata|libvirt.DomainUndefineManagedSave|
		libvirt.DomainUndefineCheckpointsMetadata); err != nil {
		if isError(err, libvirt.ErrNoSupport) || isError(err, libvirt.ErrInvalidArg) {
			log.Printf("libvirt does not support undefine flags: will try again without flags")
			if err := virConn.DomainUndefine(domain); err != nil {
				return diag.Errorf("couldn't undefine libvirt domain: %s", err)
			}
		} else {
			return diag.Errorf("couldn't undefine libvirt domain with flags: %s", err)
		}
	}

	return nil
}
