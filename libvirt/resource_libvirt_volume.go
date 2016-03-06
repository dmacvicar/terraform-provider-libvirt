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
	libvirt "gopkg.in/alexzorin/libvirt-go.v2"
)

func resourceLibvirtVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceLibvirtVolumeCreate,
		Read:   resourceLibvirtVolumeRead,
		Update: resourceLibvirtVolumeUpdate,
		Delete: resourceLibvirtVolumeDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"pool": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
			},
			"source": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"size": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			    Default: -1,
			},
			"base_volume": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
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
	virConn := meta.(*Client).libvirt
	if virConn == nil {
		return fmt.Errorf("The libvirt connection was nil.")
	}

	poolName := d.Get("pool").(string)
	if poolName == "" {
		poolName = "default"
	}
	pool, err := virConn.LookupStoragePoolByName(poolName)
	if err != nil {
		return fmt.Errorf("can't find storage pool '%s'", poolName)
	}

	volumeDef := newDefVolume()
	volumeDef.Name = d.Get("name").(string)

	// an existing image was given, this mean we can't choose size
	url := d.Get("source").(string)
	if url != "" {

		// source and size conflict
		if d.Get("size").(int) != -1 {
			return fmt.Errorf("'size' can't be specified when also 'source' is given (the size will be set to the size of the source image.")
		}

		size, err := remoteImageSize(url)
		if err != nil {
			return err
		}
		log.Printf("Remote image is: %d bytes", size)
		volumeDef.Capacity.Unit = "B"
		volumeDef.Capacity.Amount = size

	} else {
		if d.Get("size").(int) == -1 {
			return fmt.Errorf("'size' needs to be specified if no 'source' is given.")
		}

		volumeDef.Capacity.Unit = "MB"
		volumeDef.Capacity.Amount = d.Get("size").(int)
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
	if url != "" {
		stream, err := libvirt.NewVirStream(virConn, 0)
		defer stream.Close()

		volume.Upload(stream, 0, uint64(volumeDef.Capacity.Amount), 0)
		response, err := http.Get(url)
		defer response.Body.Close()
		if err != nil {
			return fmt.Errorf("Error while downloading %s: %s", url, err)
		}

		n, err := io.Copy(stream, response.Body)
		if err != nil {
			return fmt.Errorf("Error while downloading %s: %s", url, err)
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

	err = volume.Delete(0)
	if err != nil {
		return fmt.Errorf("Can't delete volume %s", d.Id())
	}


	return nil
}
