package libvirt

import (
	"encoding/xml"
	"fmt"
	libvirt "github.com/dmacvicar/libvirt-go"
	"github.com/hashicorp/terraform/helper/schema"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// network transparent image
type image interface {
	Size() (int64, error)
	WriteToStream(*libvirt.VirStream) error
	String() string
}

type localImage struct {
	path string
}

func (i *localImage) String() string {
	return i.path
}

func (i *localImage) Size() (int64, error) {
	file, err := os.Open(i.path)
	if err != nil {
		return 0, err
	}

	fi, err := file.Stat()
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

func (i *localImage) WriteToStream(stream *libvirt.VirStream) error {
	file, err := os.Open(i.path)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("Error while opening %s: %s", i.path, err)
	}

	n, err := io.Copy(stream, file)
	if err != nil {
		return fmt.Errorf("Error while reading %s: %s", i.path, err)
	}
	log.Printf("%d bytes uploaded\n", n)
	return nil
}

type httpImage struct {
	url *url.URL
}

func (i *httpImage) String() string {
	return i.url.String()
}

func (i *httpImage) Size() (int64, error) {
	response, err := http.Head(i.url.String())
	if err != nil {
		return 0, err
	}
	length, err := strconv.Atoi(response.Header.Get("Content-Length"))
	if err != nil {
		return 0, err
	}
	return int64(length), nil
}

func (i *httpImage) WriteToStream(stream *libvirt.VirStream) error {
	response, err := http.Get(i.url.String())
	defer response.Body.Close()
	if err != nil {
		return fmt.Errorf("Error while downloading %s: %s", i.url.String(), err)
	}

	n, err := io.Copy(stream, response.Body)
	if err != nil {
		return fmt.Errorf("Error while downloading %s: %s", i.url.String(), err)
	}
	log.Printf("%d bytes uploaded\n", n)
	return nil
}

func newImage(source string) (image, error) {
	url, err := url.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("Can't parse source '%s' as url: %s", source, err)
	}

	if strings.HasPrefix(url.Scheme, "http") {
		return &httpImage{url: url}, nil
	} else if url.Scheme == "file" || url.Scheme == "" {
		return &localImage{path: url.Path}, nil
	} else {
		return nil, fmt.Errorf("Don't know how to read from '%s': %s", url.String(), err)
	}
}

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

	// an existing image was given, this mean we can't choose size
	if url, ok := d.GetOk("source"); ok {
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

		img, err := newImage(url.(string))
		if err != nil {
			return err
		}

		size, err := img.Size()
		if err != nil {
			return err
		}
		log.Printf("Image %s image is: %d bytes", img, size)

		volumeDef.Capacity.Unit = "B"
		volumeDef.Capacity.Amount = size
	} else {
		_, noSize := d.GetOk("size")
		_, noBaseVol := d.GetOk("base_volume_id")

		if noSize && noBaseVol {
			return fmt.Errorf("'size' needs to be specified if no 'source' or 'base_volume_id' is given.")
		}
		volumeDef.Capacity.Amount = int64(d.Get("size").(int))
	}

	if baseVolumeId, ok := d.GetOk("base_volume_id"); ok {
		if _, ok := d.GetOk("size"); ok {
			return fmt.Errorf("'size' can't be specified when also 'base_volume_id' is given (the size will be set to the size of the backing image.")
		}

		if _, ok := d.GetOk("base_volume_name"); ok {
			return fmt.Errorf("'base_volume_name' can't be specified when also 'base_volume_id' is given.")
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

	if baseVolumeName, ok := d.GetOk("base_volume_name"); ok {
		if _, ok := d.GetOk("size"); ok {
			return fmt.Errorf("'size' can't be specified when also 'base_volume_name' is given (the size will be set to the size of the backing image.")
		}

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
	if source, ok := d.GetOk("source"); ok {
		stream, err := libvirt.NewVirStream(virConn, 0)
		if err != nil {
			return err
		}
		defer stream.Close()

		img, err := newImage(source.(string))
		if err != nil {
			return err
		}

		volume.Upload(stream, 0, uint64(volumeDef.Capacity.Amount), 0)
		err = img.WriteToStream(stream)
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

	return RemoveVolume(virConn, d.Id())
}
