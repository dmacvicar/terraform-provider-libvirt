package libvirt

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	//"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	//libvirt "gopkg.in/alexzorin/libvirt-go.v2"
	libvirt "github.com/dmacvicar/libvirt-go"
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
		"base_volume": &schema.Schema{
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

	// Refresh the pool of the volume so that libvirt knows it is
	// not longer in use.
	err = pool.Refresh(0)
	if err != nil {
		return fmt.Errorf("Error refreshing pool for volume: %s", err)
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

		size, err := remoteImageSize(url.(string))
		if err != nil {
			return err
		}
		log.Printf("Remote image is: %d bytes", size)
		volumeDef.Capacity.Unit = "B"
		volumeDef.Capacity.Amount = size
	} else {
		_, noSize := d.GetOk("size")
		_, noBaseVol := d.GetOk("base_volume")

		if noSize && noBaseVol {
			return fmt.Errorf("'size' needs to be specified if no 'source' or 'base_vol' is given.")
		}
		volumeDef.Capacity.Amount = d.Get("size").(int)
	}

	if baseVolumeId, ok := d.GetOk("base_volume"); ok {
		if _, ok := d.GetOk("size"); ok {
			return fmt.Errorf("'size' can't be specified when also 'base_volume' is given (the size will be set to the size of the backing image.")
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

	_, err := virConn.LookupStorageVolByKey(d.Id())
	if err != nil {
		return fmt.Errorf("Can't retrieve volume %s", d.Id())
	}

	return nil
}

func resourceLibvirtVolumeUpdate(d *schema.ResourceData, meta interface{}) error {
	return fmt.Errorf("Couldn't update libvirt domain")
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

	// Refresh the pool of the volume so that libvirt knows it is
	// not longer in use.
	volPool, err := volume.LookupPoolByVolume()
	if err != nil {
		return fmt.Errorf("Error retrieving pool for volume: %s", err)
	}

	err = volPool.Refresh(0)
	if err != nil {
		return fmt.Errorf("Error refreshing pool for volume: %s", err)
	}

	err = volume.Delete(0)
	if err != nil {
		return fmt.Errorf("Can't delete volume %s", d.Id())
	}

	return nil
}
