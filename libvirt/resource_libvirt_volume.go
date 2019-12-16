package libvirt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	libvirt "github.com/libvirt/libvirt-go"
)

func resourceLibvirtVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceLibvirtVolumeCreate,
		Read:   resourceLibvirtVolumeRead,
		Delete: resourceLibvirtVolumeDelete,
		Exists: resourceLibvirtVolumeExists,
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
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceLibvirtVolumeCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	if client.libvirt == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	poolName := "default"
	if _, ok := d.GetOk("pool"); ok {
		poolName = d.Get("pool").(string)
	}

	client.poolMutexKV.Lock(poolName)
	defer client.poolMutexKV.Unlock(poolName)

	pool, err := client.libvirt.LookupStoragePoolByName(poolName)
	if err != nil {
		return fmt.Errorf("can't find storage pool '%s'", poolName)
	}
	defer pool.Free()

	// Refresh the pool of the volume so that libvirt knows it is
	// not longer in use.
	waitForSuccess("error refreshing pool for volume", func() error {
		return pool.Refresh(0)
	})

	volumeDef := newDefVolume()
	if name, ok := d.GetOk("name"); ok {
		volumeDef.Name = name.(string)
	}

	var img image

	givenFormat, isFormatGiven := d.GetOk("format")
	if isFormatGiven {
		volumeDef.Target.Format.Type = givenFormat.(string)
	}

	// an source image was given, this mean we can't choose size
	if source, ok := d.GetOk("source"); ok {
		// source and size conflict
		if _, ok := d.GetOk("size"); ok {
			return fmt.Errorf("'size' can't be specified when also 'source' is given (the size will be set to the size of the source image")
		}
		if _, ok := d.GetOk("base_volume_id"); ok {
			return fmt.Errorf("'base_volume_id' can't be specified when also 'source' is given")
		}

		if _, ok := d.GetOk("base_volume_name"); ok {
			return fmt.Errorf("'base_volume_name' can't be specified when also 'source' is given")
		}

		if img, err = newImage(source.(string)); err != nil {
			return err
		}

		// figure out the format of the image
		isQCOW2, err := img.IsQCOW2()
		if err != nil {
			return fmt.Errorf("Error while determining image type for %s: %s", img.String(), err)
		}
		if isQCOW2 {
			volumeDef.Target.Format.Type = "qcow2"
		}

		if isFormatGiven && isQCOW2 && givenFormat != "qcow2" {
			return fmt.Errorf("Format other than QCOW2 explicitly specified for image detected as QCOW2 image: %s", img.String())
		}

		// update the image in the description, even if the file has not changed
		size, err := img.Size()
		if err != nil {
			return err
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

		//first handle whether it has a backing image
		// backing images can be specified by either (id), or by (name, pool)
		var baseVolume *libvirt.StorageVol
		if baseVolumeID, ok := d.GetOk("base_volume_id"); ok {
			if _, ok := d.GetOk("base_volume_name"); ok {
				return fmt.Errorf("'base_volume_name' can't be specified when also 'base_volume_id' is given")
			}
			baseVolume, err = client.libvirt.LookupStorageVolByKey(baseVolumeID.(string))
			if err != nil {
				return fmt.Errorf("Can't retrieve volume ID '%s': %v", baseVolumeID.(string), err)
			}
		} else if baseVolumeName, ok := d.GetOk("base_volume_name"); ok {
			baseVolumePool := pool
			if _, ok := d.GetOk("base_volume_pool"); ok {
				baseVolumePoolName := d.Get("base_volume_pool").(string)
				baseVolumePool, err = client.libvirt.LookupStoragePoolByName(baseVolumePoolName)
				if err != nil {
					return fmt.Errorf("can't find storage pool '%s'", baseVolumePoolName)
				}
				defer baseVolumePool.Free()
			}
			baseVolume, err = baseVolumePool.LookupStorageVolByName(baseVolumeName.(string))
			if err != nil {
				return fmt.Errorf("Can't retrieve base volume with name '%s': %v", baseVolumeName.(string), err)
			}
		}

		if baseVolume != nil {
			backingStoreFragmentDef, err := newDefBackingStoreFromLibvirt(baseVolume)
			if err != nil {
				return fmt.Errorf("Could not retrieve backing store definition: %s", err.Error())
			}

			backingStoreVolumeDef, err := newDefVolumeFromLibvirt(baseVolume)
			if err != nil {
				return err
			}

			// if the volume does not specify size, set it to the size of the backing store
			if _, ok := d.GetOk("size"); !ok {
				volumeDef.Capacity.Value = backingStoreVolumeDef.Capacity.Value
			}

			// Always check that the size, specified or taken from the backing store
			// is at least the size of the backing store itself
			if backingStoreVolumeDef.Capacity != nil && volumeDef.Capacity.Value < backingStoreVolumeDef.Capacity.Value {
				return fmt.Errorf("When 'size' is specified, it shouldn't be smaller than the backing store specified with 'base_volume_id' or 'base_volume_name/base_volume_pool'")
			}
			volumeDef.BackingStore = &backingStoreFragmentDef
		}
	}

	data, err := xmlMarshallIndented(volumeDef)
	if err != nil {
		return fmt.Errorf("Error serializing libvirt volume: %s", err)
	}
	log.Printf("[DEBUG] Generated XML for libvirt volume:\n%s", data)

	data, err = transformResourceXML(data, d)
	if err != nil {
		return fmt.Errorf("Error applying XSLT stylesheet: %s", err)
	}

	// create the volume
	volume, err := pool.StorageVolCreateXML(data, 0)
	if err != nil {
		virErr := err.(libvirt.Error)
		if virErr.Code != libvirt.ERR_STORAGE_VOL_EXIST {
			return fmt.Errorf("Error creating libvirt volume: %s", err)
		}
		// oops, volume exists already, read it and move on
		volume, err = pool.LookupStorageVolByName(volumeDef.Name)
		if err != nil {
			return fmt.Errorf("Error looking up libvirt volume: %s", err)
		}
		log.Printf("[INFO] Volume about to be created was found and left as-is: %s", volumeDef.Name)
	}
	defer volume.Free()

	// we use the key as the id
	key, err := volume.GetKey()
	if err != nil {
		return fmt.Errorf("Error retrieving volume key: %s", err)
	}
	d.SetId(key)

	// make sure we record the id even if the rest of this gets interrupted
	d.Partial(true)
	d.Set("id", key)
	d.SetPartial("id")
	d.Partial(false)

	log.Printf("[INFO] Volume ID: %s", d.Id())

	// upload source if present
	if _, ok := d.GetOk("source"); ok {
		err = img.Import(newCopier(client.libvirt, volume, volumeDef.Capacity.Value), volumeDef)
		if err != nil {
			//  don't save volume ID  in case of error. This will taint the volume after.
			// If we don't throw away the id, we will keep instead a broken volume.
			// see for reference: https://github.com/dmacvicar/terraform-provider-libvirt/issues/494
			d.Set("id", "")
			return fmt.Errorf("Error while uploading source %s: %s", img.String(), err)
		}
	}

	if err := volumeWaitForExists(client.libvirt, key); err != nil {
		return err
	}

	return resourceLibvirtVolumeRead(d, meta)
}

// resourceLibvirtVolumeRead returns the current state for a volume resource
func resourceLibvirtVolumeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	virConn := client.libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	volume, err := volumeLookupReallyHard(client, d.Get("pool").(string), d.Id())
	if err != nil {
		return err
	}

	if volume == nil {
		log.Printf("Volume '%s' may have been deleted outside Terraform", d.Id())
		d.SetId("")
		return nil
	}
	defer volume.Free()

	volName, err := volume.GetName()
	if err != nil {
		return fmt.Errorf("error retrieving volume name: %s", err)
	}

	volPool, err := volume.LookupPoolByVolume()
	if err != nil {
		return fmt.Errorf("error retrieving pool for volume: %s", err)
	}
	defer volPool.Free()

	volPoolName, err := volPool.GetName()
	if err != nil {
		return fmt.Errorf("error retrieving pool name: %s", err)
	}

	d.Set("pool", volPoolName)
	d.Set("name", volName)

	info, err := volume.GetInfo()
	if err != nil {
		virErr := err.(libvirt.Error)
		if virErr.Code != libvirt.ERR_NO_STORAGE_VOL {
			return fmt.Errorf("error retrieving volume info: %s", err)
		}
		log.Printf("Volume '%s' may have been deleted outside Terraform", d.Id())
		d.SetId("")
		return nil
	}
	d.Set("size", info.Capacity)

	volumeDef, err := newDefVolumeFromLibvirt(volume)
	if err != nil {
		return err
	}

	if volumeDef.Target == nil || volumeDef.Target.Format == nil || volumeDef.Target.Format.Type == "" {
		log.Printf("Volume has no format specified: %s", volName)
	} else {
		log.Printf("[DEBUG] Volume %s format: %s", volName, volumeDef.Target.Format.Type)
		d.Set("format", volumeDef.Target.Format.Type)
	}

	return nil
}

// resourceLibvirtVolumeDelete removed a volume resource
func resourceLibvirtVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	if client.libvirt == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	return volumeDelete(client, d.Id())
}

// resourceLibvirtVolumeExists returns True if the volume resource exists
func resourceLibvirtVolumeExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	log.Printf("[DEBUG] Check if resource libvirt_volume exists")
	client := meta.(*Client)

	volPoolName := d.Get("pool").(string)
	volume, err := volumeLookupReallyHard(client, volPoolName, d.Id())
	if err != nil {
		return false, err
	}

	if volume == nil {
		return false, nil
	}
	defer volume.Free()

	return true, nil
}
