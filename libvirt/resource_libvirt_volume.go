package libvirt

import (
	"context"
	"fmt"
	"log"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceLibvirtVolume() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLibvirtVolumeCreate,
		ReadContext:   resourceLibvirtVolumeRead,
		DeleteContext: resourceLibvirtVolumeDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"pool": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
				ForceNew: true,
			},
			"source": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"format": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"base_volume_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"base_volume_pool": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"base_volume_name": {
				Type:     schema.TypeString,
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
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceLibvirtVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return diag.Errorf(LibVirtConIsNil)
	}

	poolName := "default"
	if _, ok := d.GetOk("pool"); ok {
		poolName = d.Get("pool").(string)
	}

	client.poolMutexKV.Lock(poolName)
	defer client.poolMutexKV.Unlock(poolName)

	pool, err := virConn.StoragePoolLookupByName(poolName)
	if err != nil {
		return diag.Errorf("can't find storage pool '%s'", poolName)
	}

	// Refresh the pool of the volume so that libvirt knows it is
	// not longer in use.
	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		if err := virConn.StoragePoolRefresh(pool, 0); err != nil {
			return resource.RetryableError(err)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	volumeDef := newDefVolume()
	if name, ok := d.GetOk("name"); ok {
		volumeDef.Name = name.(string)
	}

	givenFormat, isFormatGiven := d.GetOk("format")
	if isFormatGiven {
		volumeDef.Target.Format.Type = givenFormat.(string)
	}

	var img image
	// an source image was given, this mean we can't choose size
	if source, ok := d.GetOk("source"); ok {
		// source and size conflict
		if _, ok := d.GetOk("size"); ok {
			return diag.Errorf("'size' can't be specified when also 'source' is given (the size will be set to the size of the source image")
		}
		if _, ok := d.GetOk("base_volume_id"); ok {
			return diag.Errorf("'base_volume_id' can't be specified when also 'source' is given")
		}

		if _, ok := d.GetOk("base_volume_name"); ok {
			return diag.Errorf("'base_volume_name' can't be specified when also 'source' is given")
		}

		if img, err = newImage(source.(string)); err != nil {
			return diag.FromErr(err)
		}

		// figure out the format of the image
		isQCOW2, err := img.IsQCOW2()
		if err != nil {
			return diag.Errorf("error while determining image type for %s: %s", img.String(), err)
		}
		if isQCOW2 {
			volumeDef.Target.Format.Type = "qcow2"
		}

		if isFormatGiven && isQCOW2 && givenFormat != "qcow2" {
			return diag.Errorf("format other than QCOW2 explicitly specified for image detected as QCOW2 image: %s", img.String())
		}

		// update the image in the description, even if the file has not changed
		size, err := img.Size()
		if err != nil {
			return diag.FromErr(err)
		}

		log.Printf("Image %s image is: %d bytes", img, size)
		volumeDef.Capacity.Unit = "B"
		volumeDef.Capacity.Value = size
	} else {
		// the volume does not have a source image to upload

		// if size is given, set it to the specified value
		if _, ok := d.GetOk("size"); ok {
			volumeDef.Capacity.Value = uint64(d.Get("size").(int))
		}

		// first handle whether it has a backing image
		// backing images can be specified by either (id), or by (name, pool)

		var baseVolume libvirt.StorageVol
		if baseVolumeID, ok := d.GetOk("base_volume_id"); ok {
			if _, ok := d.GetOk("base_volume_name"); ok {
				return diag.Errorf("'base_volume_name' can't be specified when also 'base_volume_id' is given")
			}

			baseVolume, err = virConn.StorageVolLookupByKey(baseVolumeID.(string))
			if err != nil {
				return diag.Errorf("can't retrieve volume ID '%s': %v", baseVolumeID.(string), err)
			}
		} else if baseVolumeName, ok := d.GetOk("base_volume_name"); ok {
			baseVolumePool := pool
			if _, ok := d.GetOk("base_volume_pool"); ok {
				baseVolumePoolName := d.Get("base_volume_pool").(string)
				baseVolumePool, err = virConn.StoragePoolLookupByName(baseVolumePoolName)
				if err != nil {
					return diag.Errorf("can't find storage pool '%s'", baseVolumePoolName)
				}
			}
			baseVolume, err = virConn.StorageVolLookupByName(baseVolumePool, baseVolumeName.(string))
			if err != nil {
				return diag.Errorf("can't retrieve base volume with name '%s': %s", baseVolumeName.(string), err)
			}
		}

		// FIXME - confirm test behaviour accurate
		// if baseVolume != nil {
		if baseVolume.Name != "" {
			backingStoreFragmentDef, err := newDefBackingStoreFromLibvirt(virConn, baseVolume)
			if err != nil {
				return diag.Errorf("could not retrieve backing store definition: %s", err.Error())
			}

			backingStoreVolumeDef, err := newDefVolumeFromLibvirt(virConn, baseVolume)
			if err != nil {
				return diag.FromErr(err)
			}

			// if the volume does not specify size, set it to the size of the backing store
			if _, ok := d.GetOk("size"); !ok {
				volumeDef.Capacity.Value = backingStoreVolumeDef.Capacity.Value
			}

			// Always check that the size, specified or taken from the backing store
			// is at least the size of the backing store itself
			if backingStoreVolumeDef.Capacity != nil && volumeDef.Capacity.Value < backingStoreVolumeDef.Capacity.Value {
				return diag.Errorf(`when 'size' is specified, it shouldn't
be smaller than the backing store specified with
'base_volume_id' or 'base_volume_name/base_volume_pool'`)
			}
			volumeDef.BackingStore = &backingStoreFragmentDef
		}
	}

	data, err := xmlMarshallIndented(volumeDef)
	if err != nil {
		return diag.Errorf("error serializing libvirt volume: %s", err)
	}
	log.Printf("[DEBUG] Generated XML for libvirt volume:\n%s", data)

	data, err = transformResourceXML(data, d)
	if err != nil {
		return diag.Errorf("error applying XSLT stylesheet: %s", err)
	}

	volume, err := virConn.StorageVolCreateXML(pool, data, 0)
	if err != nil {
		if !isError(err, libvirt.ErrStorageVolExist) {
			return diag.Errorf("error creating libvirt volume: %s", err)
		}
		// oops, volume exists already, read it and move on
		volume, err = virConn.StorageVolLookupByName(pool, volumeDef.Name)
		if err != nil {
			return diag.Errorf("error looking up libvirt volume: %s", err)
		}
		log.Printf("[INFO] Volume about to be created was found and left as-is: %s", volumeDef.Name)
	}

	// we use the key as the id
	d.SetId(volume.Key)
	log.Printf("[INFO] Volume ID: %s", d.Id())

	// upload source if present
	if _, ok := d.GetOk("source"); ok {
		err = img.Import(newCopier(virConn, &volume, volumeDef.Capacity.Value), volumeDef)
		if err != nil {
			//  don't save volume ID  in case of error. This will taint the volume after.
			// If we don't throw away the id, we will keep instead a broken volume.
			// see for reference: https://github.com/dmacvicar/terraform-provider-libvirt/issues/494
			d.Set("id", "")
			return diag.Errorf("error while uploading source %s: %s", img.String(), err)
		}
	}

	if err := waitForStateVolumeExists(ctx, client.libvirt, volume.Key); err != nil {
		return diag.FromErr(err)
	}

	return resourceLibvirtVolumeRead(ctx, d, meta)
}

// resourceLibvirtVolumeRead returns the current state for a volume resource.
func resourceLibvirtVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	virConn := client.libvirt
	if virConn == nil {
		return diag.Errorf(LibVirtConIsNil)
	}

	poolName := d.Get("pool").(string)

	var volume libvirt.StorageVol
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		var lookupErr error
		volume, lookupErr = virConn.StorageVolLookupByKey(d.Id())
		if lookupErr == nil {
			return nil
		}

		if !isError(lookupErr, libvirt.ErrNoStorageVol) {
			return resource.NonRetryableError(lookupErr)
		}

		// volume not found, try to start the pool before retry
		volPool, err := virConn.StoragePoolLookupByName(poolName)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error retrieving pool %s for volume %s: %w", poolName, d.Id(), err))
		}

		active, err := virConn.StoragePoolIsActive(volPool)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error retrieving status of pool %s for volume %s: %w", poolName, d.Id(), err))
		}

		// pool was already started, nothing else to do
		if active == 1 {
			return resource.NonRetryableError(lookupErr)
		}

		if err := virConn.StoragePoolCreate(volPool, 0); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error starting pool %s: %w", poolName, err))
		}

		// pool started successfully, retry
		return resource.RetryableError(lookupErr)
	})

	if isError(err, libvirt.ErrNoStorageVol) {
		log.Printf("volume '%s' may have been deleted outside terraform", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(err)
	}

	volPool, err := virConn.StoragePoolLookupByVolume(volume)
	if err != nil {
		return diag.Errorf("error retrieving pool for volume: %s", err)
	}

	d.Set("pool", volPool.Name)
	d.Set("name", volume.Name)

	_, size, _, err := virConn.StorageVolGetInfo(volume)
	if err != nil {
		if isError(err, libvirt.ErrNoStorageVol) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving volume info: %s", err)
	}
	d.Set("size", size)

	volumeDef, err := newDefVolumeFromLibvirt(virConn, volume)
	if err != nil {
		return diag.FromErr(err)
	}

	if volumeDef.Target == nil || volumeDef.Target.Format == nil || volumeDef.Target.Format.Type == "" {
		log.Printf("Volume has no format specified: %s", volume.Name)
	} else {
		log.Printf("[DEBUG] volume %s format: %s", volume.Name, volumeDef.Target.Format.Type)
		d.Set("format", volumeDef.Target.Format.Type)
	}

	return nil
}

// resourceLibvirtVolumeDelete removed a volume resource.
func resourceLibvirtVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	if client.libvirt == nil {
		return diag.Errorf(LibVirtConIsNil)
	}

	return diag.FromErr(volumeDelete(ctx, client, d.Id()))
}
