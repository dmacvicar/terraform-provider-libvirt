package libvirt

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	libvirt "github.com/libvirt/libvirt-go"
)

func resourceLibvirtVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceLibvirtVolumeCreate,
		Read:   resourceLibvirtVolumeRead,
		Delete: resourceLibvirtVolumeDelete,
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

func remoteImageSize(url string) (int, error) {
	response, err := http.Head(url)
	if err != nil {
		return 0, err
	}
	length, err := strconv.Atoi(response.Header.Get("Content-Length"))
	if err != nil {
		return 0, err
	}
	return length, nil
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

	volumeFormat := "qcow2"
	if _, ok := d.GetOk("format"); ok {
		volumeFormat = d.Get("format").(string)
	}
	volumeDef.Target.Format.Type = volumeFormat

	var (
		img    image
		volume *libvirt.StorageVol
	)

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

		// update the image in the description, even if the file has not changed
		size, err := img.Size()
		if err != nil {
			return err
		}
		log.Printf("Image %s image is: %d bytes", img, size)
		volumeDef.Capacity.Unit = "B"
		volumeDef.Capacity.Value = size
	} else {
		_, noSize := d.GetOk("size")
		_, noBaseVol := d.GetOk("base_volume_id")

		if noSize && noBaseVol {
			return fmt.Errorf("'size' needs to be specified if no 'source' or 'base_volume_id' is given")
		}
		volumeDef.Capacity.Value = uint64(d.Get("size").(int))
	}

	if baseVolumeID, ok := d.GetOk("base_volume_id"); ok {
		if _, ok := d.GetOk("size"); ok {
			return fmt.Errorf("'size' can't be specified when also 'base_volume_id' is given (the size will be set to the size of the backing image")
		}

		if _, ok := d.GetOk("base_volume_name"); ok {
			return fmt.Errorf("'base_volume_name' can't be specified when also 'base_volume_id' is given")
		}

		volume = nil
		baseVolume, err := client.libvirt.LookupStorageVolByKey(baseVolumeID.(string))
		if err != nil {
			return fmt.Errorf("Can't retrieve volume %s", baseVolumeID.(string))
		}
		backingStoreDef, err := newDefBackingStoreFromLibvirt(baseVolume)
		if err != nil {
			return fmt.Errorf("Could not retrieve backing store %s", baseVolumeID.(string))
		}
		volumeDef.BackingStore = &backingStoreDef
	}

	if baseVolumeName, ok := d.GetOk("base_volume_name"); ok {
		if _, ok := d.GetOk("size"); ok {
			return fmt.Errorf("'size' can't be specified when also 'base_volume_name' is given (the size will be set to the size of the backing image")
		}

		volume = nil
		baseVolumePool := pool
		if _, ok := d.GetOk("base_volume_pool"); ok {
			baseVolumePoolName := d.Get("base_volume_pool").(string)
			baseVolumePool, err = client.libvirt.LookupStoragePoolByName(baseVolumePoolName)
			if err != nil {
				return fmt.Errorf("can't find storage pool '%s'", baseVolumePoolName)
			}
			defer baseVolumePool.Free()
		}
		baseVolume, err := baseVolumePool.LookupStorageVolByName(baseVolumeName.(string))
		if err != nil {
			return fmt.Errorf("Can't retrieve volume %s", baseVolumeName.(string))
		}
		backingStoreDef, err := newDefBackingStoreFromLibvirt(baseVolume)
		if err != nil {
			return fmt.Errorf("Could not retrieve backing store %s", baseVolumeName.(string))
		}
		volumeDef.BackingStore = &backingStoreDef
	}

	if volume == nil {
		volumeDefXML, err := xml.Marshal(volumeDef)
		if err != nil {
			return fmt.Errorf("Error serializing libvirt volume: %s", err)
		}

		// create the volume
		v, err := pool.StorageVolCreateXML(string(volumeDefXML), 0)
		if err != nil {
			return fmt.Errorf("Error creating libvirt volume: %s", err)
		}
		volume = v
		defer volume.Free()
	}

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
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	volume, err := virConn.LookupStorageVolByKey(d.Id())
	if err != nil {
		virErr := err.(libvirt.Error)
		if virErr.Code != libvirt.ERR_NO_STORAGE_VOL {
			return fmt.Errorf("Can't retrieve volume %s", d.Id())
		}
		volID := d.Id()

		log.Printf("[INFO] Volume %s not found, attempting to start its pool", d.Id())

		volPoolName := d.Get("pool").(string)
		volPool, err := virConn.LookupStoragePoolByName(volPoolName)
		if err != nil {
			return fmt.Errorf("Error retrieving pool %s for volume %s: %s", volPoolName, volID, err)
		}
		defer volPool.Free()

		active, err := volPool.IsActive()
		if err != nil {
			return fmt.Errorf("error retrieving status of pool %s for volume %s: %s", volPoolName, volID, err)
		}
		if active {
			return fmt.Errorf("can't retrieve volume %s", d.Id())
		}

		err = volPool.Create(0)
		if err != nil {
			return fmt.Errorf("error starting pool %s: %s", volPoolName, err)
		}

		// attempt a new lookup
		volume, err = virConn.LookupStorageVolByKey(d.Id())
		if err != nil {
			return fmt.Errorf("second attempt: Can't retrieve volume %s", d.Id())
		}
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

	return nil
}

func resourceLibvirtVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	if client.libvirt == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	return removeVolume(client, d.Id())
}
