package libvirt

import (
	"fmt"
	"log"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
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
			"allocation": {
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
			"preallocate_metadata": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	poolName := "default"
	if _, ok := d.GetOk("pool"); ok {
		poolName = d.Get("pool").(string)
	}

	client.poolMutexKV.Lock(poolName)
	defer client.poolMutexKV.Unlock(poolName)

	pool, err := virConn.StoragePoolLookupByName(poolName)
	if err != nil {
		return fmt.Errorf("can't find storage pool '%s'", poolName)
	}

	// Refresh the pool of the volume so that libvirt knows it is
	// not longer in use.
	waitForSuccess("error refreshing pool for volume", func() error {
		return virConn.StoragePoolRefresh(pool, 0)
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
			return fmt.Errorf("error while determining image type for %s: %s", img.String(), err)
		}
		if isQCOW2 {
			volumeDef.Target.Format.Type = "qcow2"
		}

		if isFormatGiven && isQCOW2 && givenFormat != "qcow2" {
			return fmt.Errorf("format other than QCOW2 explicitly specified for image detected as QCOW2 image: %s", img.String())
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
		if _, ok := d.GetOkExists("allocation"); ok { //nolint:golint,staticcheck
			volumeDef.Allocation = &libvirtxml.StorageVolumeSize{
				Unit:  "bytes",
				Value: uint64(d.Get("allocation").(int)),
			}
		}

		// first handle whether it has a backing image
		// backing images can be specified by either (id), or by (name, pool)

		var baseVolume libvirt.StorageVol
		if baseVolumeID, ok := d.GetOk("base_volume_id"); ok {
			if _, ok := d.GetOk("base_volume_name"); ok {
				return fmt.Errorf("'base_volume_name' can't be specified when also 'base_volume_id' is given")
			}
			baseVolume, err = virConn.StorageVolLookupByKey(baseVolumeID.(string))
			if err != nil {
				return fmt.Errorf("can't retrieve volume ID '%s': %v", baseVolumeID.(string), err)
			}
		} else if baseVolumeName, ok := d.GetOk("base_volume_name"); ok {
			baseVolumePool := pool
			if _, ok := d.GetOk("base_volume_pool"); ok {
				baseVolumePoolName := d.Get("base_volume_pool").(string)
				baseVolumePool, err = virConn.StoragePoolLookupByName(baseVolumePoolName)
				if err != nil {
					return fmt.Errorf("can't find storage pool '%s'", baseVolumePoolName)
				}
			}
			baseVolume, err = virConn.StorageVolLookupByName(baseVolumePool, baseVolumeName.(string))
			if err != nil {
				return fmt.Errorf("can't retrieve base volume with name '%s': %v", baseVolumeName.(string), err)
			}
		}

		// FIXME - confirm test behaviour accurate
		// if baseVolume != nil {
		if baseVolume.Name != "" {
			backingStoreFragmentDef, err := newDefBackingStoreFromLibvirt(virConn, baseVolume)
			if err != nil {
				return fmt.Errorf("could not retrieve backing store definition: %s", err.Error())
			}

			backingStoreVolumeDef, err := newDefVolumeFromLibvirt(virConn, baseVolume)
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
				return fmt.Errorf("when 'size' is specified, it shouldn't be smaller than the backing store specified with 'base_volume_id' or 'base_volume_name/base_volume_pool'")
			}
			volumeDef.BackingStore = &backingStoreFragmentDef
		}
	}

	data, err := xmlMarshallIndented(volumeDef)
	if err != nil {
		return fmt.Errorf("error serializing libvirt volume: %s", err)
	}
	log.Printf("[DEBUG] Generated XML for libvirt volume:\n%s", data)

	data, err = transformResourceXML(data, d)
	if err != nil {
		return fmt.Errorf("error applying XSLT stylesheet: %s", err)
	}

	var flags libvirt.StorageVolCreateFlags
	if d.Get("preallocate_metadata").(bool) {
		flags = libvirt.StorageVolCreatePreallocMetadata
	}

	// create the volume
	volume, err := virConn.StorageVolCreateXML(pool, data, flags)
	if err != nil {
		virErr := err.(libvirt.Error)
		if virErr.Code != uint32(libvirt.ErrStorageVolExist) {
			return fmt.Errorf("error creating libvirt volume: %s", err)
		}
		// oops, volume exists already, read it and move on
		volume, err = virConn.StorageVolLookupByName(pool, volumeDef.Name)
		if err != nil {
			return fmt.Errorf("error looking up libvirt volume: %s", err)
		}
		log.Printf("[INFO] Volume about to be created was found and left as-is: %s", volumeDef.Name)
	}

	// we use the key as the id
	d.SetId(volume.Key)

	// make sure we record the id even if the rest of this gets interrupted
	d.Partial(true)
	d.Set("id", volume.Key)
	d.SetPartial("id")
	d.Partial(false)

	log.Printf("[INFO] Volume ID: %s", d.Id())

	// upload source if present
	if _, ok := d.GetOk("source"); ok {
		err = img.Import(newCopier(virConn, &volume, volumeDef.Capacity.Value), volumeDef)
		if err != nil {
			//  don't save volume ID  in case of error. This will taint the volume after.
			// If we don't throw away the id, we will keep instead a broken volume.
			// see for reference: https://github.com/dmacvicar/terraform-provider-libvirt/issues/494
			d.Set("id", "")
			return fmt.Errorf("error while uploading source %s: %s", img.String(), err)
		}
	}

	if err := volumeWaitForExists(client.libvirt, volume.Key); err != nil {
		return err
	}

	return resourceLibvirtVolumeRead(d, meta)
}

// resourceLibvirtVolumeRead returns the current state for a volume resource
func resourceLibvirtVolumeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	if client.libvirt == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}
	virConn := meta.(*Client).libvirt
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

	if volume.Name == "" {
		return fmt.Errorf("error retrieving volume name for volume: %s", d.Id())
	}

	volPool, err := virConn.StoragePoolLookupByVolume(*volume)
	if err != nil {
		return fmt.Errorf("error retrieving pool for volume: %s", err)
	}

	if volPool.Name == "" {
		return fmt.Errorf("error retrieving pool name for volume: %s", volume.Name)
	}

	d.Set("pool", volPool.Name)
	d.Set("name", volume.Name)

	_, size, _, err := virConn.StorageVolGetInfo(*volume)
	if err != nil {
		virErr := err.(libvirt.Error)
		if virErr.Code != uint32(libvirt.ErrNoStorageVol) {
			return fmt.Errorf("error retrieving volume info: %s", err)
		}
		log.Printf("Volume '%s' may have been deleted outside Terraform", d.Id())
		d.SetId("")
		return nil
	}
	d.Set("size", size)

	volumeDef, err := newDefVolumeFromLibvirt(virConn, *volume)
	if err != nil {
		return err
	}

	if volumeDef.Target == nil || volumeDef.Target.Format == nil || volumeDef.Target.Format.Type == "" {
		log.Printf("Volume has no format specified: %s", volume.Name)
	} else {
		log.Printf("[DEBUG] Volume %s format: %s", volume.Name, volumeDef.Target.Format.Type)
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

	return true, nil
}
