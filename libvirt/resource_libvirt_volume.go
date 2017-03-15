package libvirt

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	libvirt "github.com/dmacvicar/libvirt-go"
	"github.com/hashicorp/terraform/helper/schema"
)

func volumeCommonSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"pool": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Default:  "default",
			ForceNew: true,
		},
		"source": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"size": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"base_volume_id": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"base_volume_pool": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"base_volume_name": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
	}
}

func resourceLibvirtVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceLibvirtVolumeCreate,
		Read:   resourceLibvirtVolumeRead,
		Delete: resourceLibvirtVolumeDelete,
		Schema: volumeCommonSchema(),
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
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	poolName := "default"
	if _, ok := d.GetOk("pool"); ok {
		poolName = d.Get("pool").(string)
	}

	PoolSync.AcquireLock(poolName)
	defer PoolSync.ReleaseLock(poolName)

	pool, err := virConn.LookupStoragePoolByName(poolName)
	if err != nil {
		return fmt.Errorf("can't find storage pool '%s'", poolName)
	}
	defer pool.Free()

	// Refresh the pool of the volume so that libvirt knows it is
	// not longer in use.
	WaitForSuccess("Error refreshing pool for volume", func() error {
		return pool.Refresh(0)
	})

	volumeDef := newDefVolume()

	if name, ok := d.GetOk("name"); ok {
		volumeDef.Name = name.(string)
	}

	var (
		img    image
		volume *libvirt.VirStorageVol = nil
	)

	// an source image was given, this mean we can't choose size
	if source, ok := d.GetOk("source"); ok {
		// source and size conflict
		if _, ok := d.GetOk("size"); ok {
			return fmt.Errorf("'size' can't be specified when also 'source' is given (the size will be set to the size of the source image.")
		}
		if _, ok := d.GetOk("base_volume_id"); ok {
			return fmt.Errorf("'base_volume_id' can't be specified when also 'source' is given.")
		}

		if _, ok := d.GetOk("base_volume_name"); ok {
			return fmt.Errorf("'base_volume_name' can't be specified when also 'source' is given.")
		}

		// Check if we already have this image in the pool
		if len(volumeDef.Name) > 0 {
			if v, err := pool.LookupStorageVolByName(volumeDef.Name); err != nil {
				log.Printf("Could not find image %s in pool %s", volumeDef.Name, poolName)
			} else {
				volume = &v
				volumeDef, err = newDefVolumeFromLibvirt(volume)
				if err != nil {
					return fmt.Errorf("could not get a volume definition from XML for %s: %s.", volumeDef.Name, err)
				}
			}
		}

		if img, err = newImage(source.(string)); err != nil {
			return err
		}

		// update the image in the description, even if the file has not changed
		if size, err := img.Size(); err != nil {
			return err
		} else {
			log.Printf("Image %s image is: %d bytes", img, size)
			volumeDef.Capacity.Unit = "B"
			volumeDef.Capacity.Amount = size
		}
	} else {
		_, noSize := d.GetOk("size")
		_, noBaseVol := d.GetOk("base_volume_id")

		if noSize && noBaseVol {
			return fmt.Errorf("'size' needs to be specified if no 'source' or 'base_volume_id' is given.")
		}
		volumeDef.Capacity.Amount = uint64(d.Get("size").(int))
	}

	if baseVolumeId, ok := d.GetOk("base_volume_id"); ok {
		if _, ok := d.GetOk("size"); ok {
			return fmt.Errorf("'size' can't be specified when also 'base_volume_id' is given (the size will be set to the size of the backing image.")
		}

		if _, ok := d.GetOk("base_volume_name"); ok {
			return fmt.Errorf("'base_volume_name' can't be specified when also 'base_volume_id' is given.")
		}

		volume = nil
		volumeDef.BackingStore = new(defBackingStore)
		volumeDef.BackingStore.Format.Type = "qcow2"
		baseVolume, err := virConn.LookupStorageVolByKey(baseVolumeId.(string))
		if err != nil {
			return fmt.Errorf("Can't retrieve volume %s", baseVolumeId.(string))
		}
		baseVolPath, err := baseVolume.GetPath()
		if err != nil {
			return fmt.Errorf("can't get name for base image '%s'", baseVolumeId)
		}
		volumeDef.BackingStore.Path = baseVolPath
	}

	if baseVolumeName, ok := d.GetOk("base_volume_name"); ok {
		if _, ok := d.GetOk("size"); ok {
			return fmt.Errorf("'size' can't be specified when also 'base_volume_name' is given (the size will be set to the size of the backing image.")
		}

		volume = nil
		baseVolumePool := pool
		if _, ok := d.GetOk("base_volume_pool"); ok {
			baseVolumePoolName := d.Get("base_volume_pool").(string)
			baseVolumePool, err = virConn.LookupStoragePoolByName(baseVolumePoolName)
			if err != nil {
				return fmt.Errorf("can't find storage pool '%s'", baseVolumePoolName)
			}
			defer baseVolumePool.Free()
		}

		volumeDef.BackingStore = new(defBackingStore)
		volumeDef.BackingStore.Format.Type = "qcow2"
		baseVolume, err := baseVolumePool.LookupStorageVolByName(baseVolumeName.(string))
		if err != nil {
			return fmt.Errorf("Can't retrieve volume %s", baseVolumeName.(string))
		}
		baseVolPath, err := baseVolume.GetPath()
		if err != nil {
			return fmt.Errorf("can't get name for base image '%s'", baseVolumeName)
		}
		volumeDef.BackingStore.Path = baseVolPath
	}

	if volume == nil {
		volumeDefXml, err := xml.Marshal(volumeDef)
		if err != nil {
			return fmt.Errorf("Error serializing libvirt volume: %s", err)
		}

		// create the volume
		v, err := pool.StorageVolCreateXML(string(volumeDefXml), 0)
		if err != nil {
			return fmt.Errorf("Error creating libvirt volume: %s", err)
		}
		volume = &v
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
		copier := func(src io.Reader) error {
			stream, err := libvirt.NewVirStream(virConn, 0)
			if err != nil {
				return err
			}
			defer stream.Close()

			volume.Upload(stream, 0, volumeDef.Capacity.Amount, 0)

			n, err := io.Copy(stream, src)
			if err != nil {
				return fmt.Errorf("Error while downloading %s: %s", img.String(), err)
			}
			log.Printf("%d bytes uploaded\n", n)
			return nil
		}

		err = img.Import(copier, volumeDef)
		if err != nil {
			return err
		}
	}

	return resourceLibvirtVolumeRead(d, meta)
}

func resourceLibvirtVolumeRead(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	volume, err := virConn.LookupStorageVolByKey(d.Id())
	if err != nil {
		virErr := err.(libvirt.VirError)
		if virErr.Code != libvirt.VIR_ERR_NO_STORAGE_VOL {
			return fmt.Errorf("Can't retrieve volume %s", d.Id())
		}

		log.Printf("[INFO] Volume %s not found, attempting to start its pool")

		volId := d.Id()
		volPoolName := d.Get("pool").(string)
		volPool, err := virConn.LookupStoragePoolByName(volPoolName)
		if err != nil {
			return fmt.Errorf("Error retrieving pool %s for volume %s: %s", volPoolName, volId, err)
		}
		defer volPool.Free()

		active, err := volPool.IsActive()
		if err != nil {
			return fmt.Errorf("Error retrieving status of pool %s for volume %s: %s", volPoolName, volId, err)
		}
		if active {
			return fmt.Errorf("Can't retrieve volume %s", d.Id())
		}

		err = volPool.Create(0)
		if err != nil {
			return fmt.Errorf("Error starting pool %s: %s", volPoolName, err)
		}

		// attempt a new lookup
		volume, err = virConn.LookupStorageVolByKey(d.Id())
		if err != nil {
			return fmt.Errorf("Second attempt: Can't retrieve volume %s", d.Id())
		}
	}
	defer volume.Free()

	volName, err := volume.GetName()
	if err != nil {
		return fmt.Errorf("Error retrieving volume name: %s", err)
	}

	volPool, err := volume.LookupPoolByVolume()
	if err != nil {
		return fmt.Errorf("Error retrieving pool for volume: %s", err)
	}
	defer volPool.Free()

	volPoolName, err := volPool.GetName()
	if err != nil {
		return fmt.Errorf("Error retrieving pool name: %s", err)
	}

	d.Set("pool", volPoolName)
	d.Set("name", volName)

	info, err := volume.GetInfo()
	if err != nil {
		return fmt.Errorf("Error retrieving volume name: %s", err)
	}
	d.Set("size", info.GetCapacityInBytes())

	return nil
}

func resourceLibvirtVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	return RemoveVolume(virConn, d.Id())
}
