package libvirt

import (
	"encoding/xml"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	libvirtc "github.com/libvirt/libvirt-go"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

func resourceLibvirtPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceLibvirtPoolCreate,
		Read:   resourceLibvirtPoolRead,
		Delete: resourceLibvirtPoolDelete,
		Exists: resourceLibvirtPoolExists,
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
				Type:     schema.TypeString,
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
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceLibvirtPoolCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	virConn := client.libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	poolType := d.Get("type").(string)
	if poolType != "dir" {
		return fmt.Errorf("Only storage pools of type \"dir\" are supported")
	}

	poolName := d.Get("name").(string)

	client.poolMutexKV.Lock(poolName)
	defer client.poolMutexKV.Unlock(poolName)

	// Check whether the storage pool already exists. Its name needs to be
	// unique.
	if _, err := virConn.StoragePoolLookupByName(poolName); err == nil {
		return fmt.Errorf("storage pool '%s' already exists", poolName)
	}
	log.Printf("[DEBUG] Pool with name '%s' does not exist yet", poolName)

	poolPath := d.Get("path").(string)
	if poolPath == "" {
		return fmt.Errorf("\"path\" attribute is requires for storage pools of type \"dir\"")
	}

	poolDef := libvirtxml.StoragePool{
		Type: "dir",
		Name: poolName,
		Target: &libvirtxml.StoragePoolTarget{
			Path: poolPath,
		},
	}
	data, err := xmlMarshallIndented(poolDef)
	if err != nil {
		return fmt.Errorf("Error serializing libvirt storage pool: %s", err)
	}
	log.Printf("[DEBUG] Generated XML for libvirt storage pool:\n%s", data)

	data, err = transformResourceXML(data, d)
	if err != nil {
		return fmt.Errorf("Error applying XSLT stylesheet: %s", err)
	}

	// create the pool
	pool, err := virConn.StoragePoolDefineXML(data, 0)
	if err != nil {
		return fmt.Errorf("Error creating libvirt storage pool: %s", err)
	}

	err = virConn.StoragePoolBuild(pool, 0)
	if err != nil {
		return fmt.Errorf("Error building libvirt storage pool: %s", err)
	}

	err = virConn.StoragePoolSetAutostart(pool, 1)
	if err != nil {
		return fmt.Errorf("Error setting up libvirt storage pool: %s", err)
	}

	err = virConn.StoragePoolCreate(pool, 0)
	if err != nil {
		return fmt.Errorf("Error starting libvirt storage pool: %s", err)
	}

	err = virConn.StoragePoolRefresh(pool, 0)
	if err != nil {
		return fmt.Errorf("Error refreshing libvirt storage pool: %s", err)
	}

	id := uuidString(pool.UUID)
	if id == "" {
		return fmt.Errorf("Error retrieving libvirt pool id: %s", pool.Name)
	}
	d.SetId(id)

	// make sure we record the id even if the rest of this gets interrupted
	d.Partial(true)
	d.Set("id", id)
	d.SetPartial("id")
	d.Partial(false)

	log.Printf("[INFO] Pool ID: %s", d.Id())

	if err := poolWaitForExists(client.libvirt, id); err != nil {
		return err
	}

	return resourceLibvirtPoolRead(d, meta)
}

func resourceLibvirtPoolRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	virConn := client.libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	pool, err := virConn.StoragePoolLookupByUUID(parseUUID(d.Id()))
	// TODO: validate change in test from empty struct to explicit error
	if err != nil {
		log.Printf("storage pool '%s' may have been deleted outside Terraform", d.Id())
		d.SetId("")
		return nil
	}

	if pool.Name == "" {
		return fmt.Errorf("error retrieving pool name: %s", err)
	}
	d.Set("name", pool.Name)

	_, capacity, allocation, available, err := virConn.StoragePoolGetInfo(pool)
	if err != nil {
		return fmt.Errorf("error retrieving pool info: %s", err)
	}
	d.Set("capacity", capacity)
	d.Set("allocation", allocation)
	d.Set("available", available)

	poolDefXML, err := virConn.StoragePoolGetXMLDesc(pool, 0)
	if err != nil {
		return fmt.Errorf("could not get XML description for pool %s: %s", pool.Name, err)
	}

	var poolDef libvirtxml.StoragePool
	err = xml.Unmarshal([]byte(poolDefXML), &poolDef)
	if err != nil {
		return fmt.Errorf("could not get a pool definition from XML for %s: %s", poolDef.Name, err)
	}

	var poolPath string
	if poolDef.Target != nil && poolDef.Target.Path != "" {
		poolPath = poolDef.Target.Path
	}

	if poolPath == "" {
		log.Printf("Pool %s has no path specified", pool.Name)
	} else {
		log.Printf("[DEBUG] Pool %s path: %s", pool.Name, poolPath)
		d.Set("path", poolPath)
	}

	return nil
}

func resourceLibvirtPoolDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	if client.libvirt == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	return deletePool(client, d.Id())
}

func resourceLibvirtPoolExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	log.Printf("[DEBUG] Check if resource libvirt_pool exists")
	client := meta.(*Client)
	virConn := client.libvirt
	if virConn == nil {
		return false, fmt.Errorf(LibVirtConIsNil)
	}

	_, err := virConn.StoragePoolLookupByUUID(parseUUID(d.Id()))
	if err != nil {
		virErr := err.(libvirtc.Error)
		if virErr.Code != libvirtc.ERR_NO_STORAGE_POOL {
			return false, fmt.Errorf("Can't retrieve pool %s", d.Id())
		}
		// does not exist, but no error
		return false, nil
	}

	return true, nil
}
