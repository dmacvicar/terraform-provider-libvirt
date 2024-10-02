package libvirt

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"libvirt.org/go/libvirtxml"
)

func resourceLibvirtPool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLibvirtPoolCreate,
		ReadContext:   resourceLibvirtPoolRead,
		DeleteContext: resourceLibvirtPoolDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"capacity": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"allocation": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"available": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"xml": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
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

			// Dir-specific attributes
			"path": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			// logical-specific attributes
			"source_devices": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceLibvirtPoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	virConn := client.libvirt

	poolType := d.Get("type").(string)
	if poolType != "dir" && poolType != "logical" {
		return diag.Errorf("only storage pools of type \"dir\" and \"logical\" are supported")
	}

	poolName := d.Get("name").(string)

	client.poolMutexKV.Lock(poolName)
	defer client.poolMutexKV.Unlock(poolName)

	// Check whether the storage pool already exists. Its name needs to be
	// unique.
	if _, err := virConn.StoragePoolLookupByName(poolName); err == nil {
		return diag.Errorf("storage pool '%s' already exists", poolName)
	}
	log.Printf("[DEBUG] Pool with name '%s' does not exist yet", poolName)

	poolPath := d.Get("path").(string)
	var poolDef *libvirtxml.StoragePool
	poolCreateFlags := libvirt.StoragePoolBuildNew

	if poolType == "dir" {
		if poolPath == "" {
			return diag.Errorf("\"path\" attribute is requires for storage pools of type \"dir\"")
		}

		sourceDevices := d.Get("source_devices.#").(int)
		if sourceDevices != 0 {
			return diag.Errorf("\"source_devices\" attribute cannot be used for storage pool of type \"dir\"")
		}

		poolDef = &libvirtxml.StoragePool{
			Type: "dir",
			Name: poolName,
			Target: &libvirtxml.StoragePoolTarget{
				Path: poolPath,
			},
		}
	} else if poolType == "logical" {
		// path is auto-generated for lvm pools, so we don't set/read it
		if poolPath != "" {
			return diag.Errorf("\"path\" attribute cannot be used for storage pool of type \"logical\"")
		}

		poolDef = &libvirtxml.StoragePool{
			Type: "logical",
			Name: poolName,
		}

		var devices []libvirtxml.StoragePoolSourceDevice

		for i := 0; i < d.Get("source_devices.#").(int); i++ {
			device := d.Get(fmt.Sprintf("source_devices.%d", i)).(string)
			devices = append(devices, libvirtxml.StoragePoolSourceDevice{Path: device})
		}

		if devices != nil {
			poolDef.Source = &libvirtxml.StoragePoolSource{
				Device: devices,
			}
		} else {
			poolCreateFlags = libvirt.StoragePoolBuildNoOverwrite
		}
	}

	data, err := xmlMarshallIndented(poolDef)
	if err != nil {
		return diag.Errorf("error serializing libvirt storage pool: %s", err)
	}
	log.Printf("[DEBUG] Generated XML for libvirt storage pool:\n%s", data)

	data, err = transformResourceXML(data, d)
	if err != nil {
		return diag.Errorf("error applying XSLT stylesheet: %s", err)
	}

	// create the pool
	pool, err := virConn.StoragePoolDefineXML(data, 0)
	if err != nil {
		return diag.Errorf("error creating libvirt storage pool: %s", err)
	}

	err = virConn.StoragePoolBuild(pool, 0)
	if err != nil {
		return diag.Errorf("error building libvirt storage pool: %s", err)
	}

	err = virConn.StoragePoolSetAutostart(pool, 1)
	if err != nil {
		return diag.Errorf("error setting up libvirt storage pool: %s", err)
	}

	err = virConn.StoragePoolCreate(pool, libvirt.StoragePoolCreateFlags(poolCreateFlags))
	if err != nil {
		return diag.Errorf("error starting libvirt storage pool: %s", err)
	}

	err = virConn.StoragePoolRefresh(pool, 0)
	if err != nil {
		return diag.Errorf("error refreshing libvirt storage pool: %s", err)
	}

	id := uuidString(pool.UUID)
	if id == "" {
		return diag.Errorf("error retrieving libvirt pool id: %s", pool.Name)
	}
	d.SetId(id)

	log.Printf("[INFO] Pool ID: %s", d.Id())

	if err := waitForStatePoolExists(ctx, client.libvirt, pool.UUID); err != nil {
		return diag.FromErr(err)
	}

	return resourceLibvirtPoolRead(ctx, d, meta)
}

func resourceLibvirtPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	virConn := client.libvirt

	uuid := parseUUID(d.Id())

	pool, err := virConn.StoragePoolLookupByUUID(uuid)
	if err != nil {
		if isError(err, libvirt.ErrNoStoragePool) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving libvirt pool %s", err)
	}

	d.Set("name", pool.Name)

	_, capacity, allocation, available, err := virConn.StoragePoolGetInfo(pool)
	if err != nil {
		return diag.Errorf("error retrieving pool info: %s", err)
	}

	d.Set("capacity", capacity)
	d.Set("allocation", allocation)
	d.Set("available", available)

	poolDefXML, err := virConn.StoragePoolGetXMLDesc(pool, 0)
	if err != nil {
		return diag.Errorf("could not get XML description for pool %s: %s", pool.Name, err)
	}

	var poolDef libvirtxml.StoragePool
	err = xml.Unmarshal([]byte(poolDefXML), &poolDef)
	if err != nil {
		return diag.Errorf("could not get a pool definition from XML for %s: %s", poolDef.Name, err)
	}

	var poolPath string
	if poolDef.Target != nil && poolDef.Target.Path != "" {
		poolPath = poolDef.Target.Path
	}

	// for logical pool the path auto-generated, so we don't set/read it
	if poolDef.Type != "logical" {
		if poolPath == "" {
			log.Printf("Pool %s has no path specified", pool.Name)
		} else {
			log.Printf("[DEBUG] Pool %s path: %s", pool.Name, poolPath)
			d.Set("path", poolPath)
		}
	}

	if poolType := poolDef.Type; poolType == "" {
		log.Printf("Pool %s has no type specified", pool.Name)
	} else {
		log.Printf("[DEBUG] Pool %s type: %s", pool.Name, poolType)
		d.Set("type", poolType)
	}

	return nil
}

func resourceLibvirtPoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	virConn := client.libvirt

	uuid := parseUUID(d.Id())

	pool, err := virConn.StoragePoolLookupByUUID(uuid)
	if err != nil {
		if isError(err, libvirt.ErrNoStoragePool) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving storage pool info: %s", err)
	}

	client.poolMutexKV.Lock(pool.Name)
	defer client.poolMutexKV.Unlock(pool.Name)

	state, _, _, _, err := virConn.StoragePoolGetInfo(pool)
	if err != nil {
		return diag.Errorf("error retrieving storage pool info: %s", err)
	}

	if state != uint8(libvirt.StoragePoolInactive) {
		err := virConn.StoragePoolDestroy(pool)
		if err != nil {
			return diag.Errorf("error deleting storage pool: %s", err)
		}
	}

	poolDef, dia := newDefPoolFromLibvirt(virConn, pool)
	if dia != nil {
		return dia
	}

	// if the logical pool has no source device then the volume group existed before we created the pool, so we don't delete it
	if poolDef.Type == "dir" || (poolDef.Type == "logical" && poolDef.Source != nil && poolDef.Source.Device != nil) {
		err = virConn.StoragePoolDelete(pool, 0)
		if err != nil {
			return diag.Errorf("error deleting storage pool: %s", err)
		}
	}

	err = virConn.StoragePoolUndefine(pool)
	if err != nil {
		return diag.Errorf("error deleting storage pool: %s", err)
	}

	return diag.FromErr(waitForStatePoolDeleted(ctx, client.libvirt, uuid))
}
