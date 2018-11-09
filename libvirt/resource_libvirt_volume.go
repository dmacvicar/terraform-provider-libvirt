package libvirt

import (
	"encoding/xml"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
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

	// Check whether the storage volume already exists. Its name needs to be
	// unique.
	if _, err := pool.LookupStorageVolByName(d.Get("name").(string)); err == nil {
		return fmt.Errorf("storage volume '%s' already exists", d.Get("name").(string))
	}

	volumeDef := newDefVolume()
	volumeDef.Name = d.Get("name").(string)

	var (
		img image
	)

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

		// the volume does not have a source image to upload, first handle
		// whether it has a backing image
		//
		// backing images can be specified by either (id), or by (name, pool)
		var baseVolume *libvirt.StorageVol

		if baseVolumeID, ok := d.GetOk("base_volume_id"); ok {
			if _, ok := d.GetOk("base_volume_name"); ok {
				return fmt.Errorf("'base_volume_name' can't be specified when also 'base_volume_id' is given")
			}
			baseVolume, err = client.libvirt.LookupStorageVolByKey(baseVolumeID.(string))
			if err != nil {
				return fmt.Errorf("Can't retrieve volume %s: %v", baseVolumeID.(string), err)
			}
		}

		if baseVolumeName, ok := d.GetOk("base_volume_name"); ok {
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
				return fmt.Errorf("Can't retrieve volume %s: %v", baseVolumeName.(string), err)
			}
		}

		if baseVolume != nil {
			backingStoreDef, err := newDefBackingStoreFromLibvirt(baseVolume)
			if err != nil {
				return fmt.Errorf("Could not retrieve backing store definition: %s", err.Error())
			}

			// does the backing store have some size information?, check at least that it is not smaller than the backing store
			volumeDef.Capacity.Value = uint64(d.Get("size").(int))
			if _, ok := d.GetOk("size"); ok {
				backingStoreVolumeDef, err := newDefVolumeFromLibvirt(baseVolume)
				if err != nil {
					return err
				}

				if backingStoreVolumeDef.Capacity != nil && volumeDef.Capacity.Value < backingStoreVolumeDef.Capacity.Value {
					return fmt.Errorf("When 'size' is specified, it shouldn't be smaller than the backing store specified with 'base_volume_id' or 'base_volume_name/base_volume_pool'")
				}
			}
			volumeDef.BackingStore = &backingStoreDef
		}
	}

	volumeDef.Capacity.Value = uint64(d.Get("size").(int))
	volumeDefXML, err := xml.Marshal(volumeDef)
	if err != nil {
		return fmt.Errorf("Error serializing libvirt volume: %s", err)
	}

	// create the volume
	volume, err := pool.StorageVolCreateXML(string(volumeDefXML), 0)
	if err != nil {
		return fmt.Errorf("Error creating libvirt volume: %s", err)
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
			return fmt.Errorf("Error while uploading source %s: %s", img.String(), err)
		}
	}

	return resourceLibvirtVolumeRead(d, meta)
}

func resourceLibvirtVolumeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	virConn := client.libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	volume, err := lookupVolumeReallyHard(client, d.Get("pool").(string), d.Id())
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
		return fmt.Errorf("error retrieving volume name: %s", err)
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

func resourceLibvirtVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	if client.libvirt == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	return removeVolume(client, d.Id())
}

func resourceLibvirtVolumeExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	log.Printf("[DEBUG] Check if resource libvirt_volume exists")
	client := meta.(*Client)

	volPoolName := d.Get("pool").(string)
	volume, err := lookupVolumeReallyHard(client, volPoolName, d.Id())
	if err != nil {
		return false, err
	}

	if volume == nil {
		return false, nil
	}
	defer volume.Free()

	return true, nil
}
