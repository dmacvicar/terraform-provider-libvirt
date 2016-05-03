package libvirt

import (
	"encoding/xml"
	"fmt"
	libvirt "github.com/dmacvicar/libvirt-go"
	"github.com/hashicorp/terraform/helper/schema"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
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

	pool, err := virConn.LookupStoragePoolByName(poolName)
	if err != nil {
		return fmt.Errorf("can't find storage pool '%s'", poolName)
	}
	defer pool.Free()

	// Refresh the pool of the volume so that libvirt knows it is
	// not longer in use.
       
        try            := 0
        maxTries       := 10
        sleepTimeInSec := 2
  
        for try < maxTries  {
		err = pool.Refresh(0)
		if err != nil {
			if try >= maxTries {
				return fmt.Errorf("Error refreshing pool for volume: %s , try: %i", err, try )
			} else {
				time.Sleep( time.Duration(sleepTimeInSec) * time.Second )
			}
		} else {
			break
		}
        }

	volumeDef := newDefVolume()

	if name, ok := d.GetOk("name"); ok {
		volumeDef.Name = name.(string)
	}

	// an existing image was given, this mean we can't choose size
	if url, ok := d.GetOk("source"); ok {
		// source and size conflict
		if _, ok := d.GetOk("size"); ok {
			return fmt.Errorf("'size' can't be specified when also 'source' is given (the size will be set to the size of the source image.")
		}
		if _, ok := d.GetOk("base_volume_id"); ok {
			return fmt.Errorf("'base_volume_id' can't be specified when also 'source' is given (the size will be set to the size of the base image.")
		}

		size, err := remoteImageSize(url.(string))
		if err != nil {
			return err
		}
		log.Printf("Remote image is: %d bytes", size)
		volumeDef.Capacity.Unit = "B"
		volumeDef.Capacity.Amount = size
	} else {
		_, noSize := d.GetOk("size")
		_, noBaseVol := d.GetOk("base_volume_id")

		if noSize && noBaseVol {
			return fmt.Errorf("'size' needs to be specified if no 'source' or 'base_volume_id' is given.")
		}
		volumeDef.Capacity.Amount = d.Get("size").(int)
	}

	if baseVolumeId, ok := d.GetOk("base_volume_id"); ok {
		if _, ok := d.GetOk("size"); ok {
			return fmt.Errorf("'size' can't be specified when also 'base_volume_id' is given (the size will be set to the size of the backing image.")
		}

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

	volumeDefXml, err := xml.Marshal(volumeDef)
	if err != nil {
		return fmt.Errorf("Error serializing libvirt volume: %s", err)
	}

	// create the volume
	volume, err := pool.StorageVolCreateXML(string(volumeDefXml), 0)
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
	log.Printf("[INFO] Volume ID: %s", d.Id())

	// upload source if present
	if url, ok := d.GetOk("source"); ok {
		stream, err := libvirt.NewVirStream(virConn, 0)
		defer stream.Close()

		volume.Upload(stream, 0, uint64(volumeDef.Capacity.Amount), 0)
		response, err := http.Get(url.(string))
		defer response.Body.Close()
		if err != nil {
			return fmt.Errorf("Error while downloading %s: %s", url.(string), err)
		}

		n, err := io.Copy(stream, response.Body)
		if err != nil {
			return fmt.Errorf("Error while downloading %s: %s", url.(string), err)
		}
		log.Printf("%d bytes uploaded\n", n)
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
		return fmt.Errorf("Can't retrieve volume %s", d.Id())
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

	volume, err := virConn.LookupStorageVolByKey(d.Id())
	if err != nil {
		return fmt.Errorf("Can't retrieve volume %s", d.Id())
	}
	defer volume.Free()

	// Refresh the pool of the volume so that libvirt knows it is
	// not longer in use.
	volPool, err := volume.LookupPoolByVolume()
	if err != nil {
		return fmt.Errorf("Error retrieving pool for volume: %s", err)
	}
	defer volPool.Free()

	err = volPool.Refresh(0)
	if err != nil {
		return fmt.Errorf("Error refreshing pool for volume: %s", err)
	}

	// Workaround for redhat#1293804
	// https://bugzilla.redhat.com/show_bug.cgi?id=1293804#c12
	// Does not solve the problem but it makes it happen less often.
	_, err = volume.GetXMLDesc(0)
	if err != nil {
		return fmt.Errorf("Can't retrieve volume %s XML desc: %s", d.Id(), err)
	}

	err = volume.Delete(0)
	if err != nil {
		return fmt.Errorf("Can't delete volume %s: %s", d.Id(), err)
	}

	return nil
}
