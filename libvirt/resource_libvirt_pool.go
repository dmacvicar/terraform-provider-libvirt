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

func resourceLibvirtPoolCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	// target path is computed for logical
	if target := diff.GetRawConfig().GetAttr("target"); target.IsKnown() && len(target.AsValueSlice()) > 0 {
		if path := target.AsValueSlice()[0].GetAttr("path"); path.IsKnown() {
			oldTargetPath, newTargetPath := diff.GetChange("target.0.path")
			if oldTargetPath != newTargetPath {
				if err := diff.ForceNew("target.0.path"); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

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
			"source": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"path": {
										Type:          schema.TypeString,
										Optional:      true,
										ForceNew:      true,
										ConflictsWith: []string{"path"},
									},
								},
							},
						},
					},
				},
			},
			"target": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
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
			// deprecated dir specific attribute
			"path": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Deprecated:    "use target.path instead",
				ConflictsWith: []string{"target.0.path"},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: resourceLibvirtPoolCustomizeDiff,
	}
}

func resourceLibvirtPoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	virConn := client.libvirt

	poolName := d.Get("name").(string)

	client.poolMutexKV.Lock(poolName)
	defer client.poolMutexKV.Unlock(poolName)

	// Check whether the storage pool already exists. Its name needs to be
	// unique.
	if _, err := virConn.StoragePoolLookupByName(poolName); err == nil {
		return diag.Errorf("storage pool '%s' already exists", poolName)
	}
	log.Printf("[DEBUG] Pool with name '%s' does not exist yet", poolName)

	var poolDef *libvirtxml.StoragePool
	poolCreateFlags := libvirt.StoragePoolBuildNew

	poolType := d.Get("type").(string)

	if poolType == "dir" {
		poolPath := d.Get("path").(string)
		if poolPath == "" {
			poolPath = d.Get("target.0.path").(string)
		}

		poolDef = &libvirtxml.StoragePool{
			Type: "dir",
			Name: poolName,
			Target: &libvirtxml.StoragePoolTarget{
				Path: poolPath,
			},
		}
	} else if poolType == "logical" {
		poolDef = &libvirtxml.StoragePool{
			Type: "logical",
			Name: poolName,
		}

		var devices []libvirtxml.StoragePoolSourceDevice

		for i := 0; i < d.Get("source.0.device.#").(int); i++ {
			devicePath := d.Get(fmt.Sprintf("source.0.device.%d.path", i)).(string)
			devices = append(devices, libvirtxml.StoragePoolSourceDevice{Path: devicePath})
		}

		if devices != nil {
			poolDef.Source = &libvirtxml.StoragePoolSource{
				Device: devices,
			}
		} else {
			poolCreateFlags = libvirt.StoragePoolBuildNoOverwrite
		}
	} else {
		return diag.Errorf("only storage pools of type \"dir\" and \"logical\" are supported")
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

	d.Set("type", poolDef.Type)

	if poolDef.Source != nil {
		source := map[string]interface{}{}

		if len(poolDef.Source.Device) > 0 {
			var devices []interface{}

			for _, device := range poolDef.Source.Device {
				deviceMap := make(map[string]interface{})
				deviceMap["path"] = device.Path
				devices = append(devices, deviceMap)
			}
			source["device"] = devices
		}

		if len(source) > 0 {
			if err := d.Set("source", []interface{}{source}); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if poolDef.Target != nil {
		if _, ok := d.GetOk("path"); ok {
			// old deprecated value is set
			d.Set("path", poolDef.Target.Path)
		} else {
			target := map[string]interface{}{}
			target["path"] = poolDef.Target.Path
			d.Set("target", []interface{}{target})
		}
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
