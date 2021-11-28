package libvirt

import (
	"fmt"
	"log"
	"time"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	volExistsID    = "EXISTS"
	volNotExistsID = "NOT-EXISTS"
)

// volumeExists returns "EXISTS" or "NOT-EXISTS" depending on the current volume existence
func volumeExists(virConn *libvirt.Libvirt, key string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		_, err := virConn.StorageVolLookupByKey(key)
		if err != nil {
			if err.(libvirt.Error).Code == uint32(libvirt.ErrNoStorageVol) {
				log.Printf("Volume %s does not exist", key)
				return virConn, "NOT-EXISTS", nil
			}
			log.Printf("Volume %s: error: %s", key, err.(libvirt.Error).Message)
		}
		return virConn, volExistsID, err
	}
}

// volumeWaitForExists waits for a storage volume to be up and timeout after 5 minutes.
func volumeWaitForExists(virConn *libvirt.Libvirt, key string) error {
	log.Printf("Waiting for volume %s to be active...", key)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{volNotExistsID},
		Target:     []string{volExistsID},
		Refresh:    volumeExists(virConn, key),
		Timeout:    1 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error waiting for volume to reach EXISTS state: %s", err)
	}
	return nil
}

// volumeWaitDeleted waits for a storage volume to be removed
func volumeWaitDeleted(virConn *libvirt.Libvirt, key string) error {
	log.Printf("Waiting for volume %s to be deleted...", key)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{volExistsID},
		Target:     []string{volNotExistsID},
		Refresh:    volumeExists(virConn, key),
		Timeout:    1 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error waiting for volume to reach NOT-EXISTS state: %s", err)
	}
	return nil
}

// volumeDelete removes the volume identified by `key` from libvirt
func volumeDelete(client *Client, key string) error {
	virConn := client.libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}
	volume, err := virConn.StorageVolLookupByKey(key)
	if err != nil {
		virErr := err.(libvirt.Error)
		if virErr.Code != uint32(libvirt.ErrNoStorageVol) {
			return fmt.Errorf("volumeDelete: Can't retrieve volume %s: %v", key, err)
		}
		// Volume already deleted.
		return nil
	}

	// Refresh the pool of the volume so that libvirt knows it is
	// no longer in use.
	volPool, err := virConn.StoragePoolLookupByVolume(volume)
	if err != nil {
		return fmt.Errorf("error retrieving pool for volume: %s", err)
	}

	if volPool.Name == "" {
		return fmt.Errorf("error retrieving name of pool for volume key: %s", volume.Key)
	}

	client.poolMutexKV.Lock(volPool.Name)
	defer client.poolMutexKV.Unlock(volPool.Name)

	waitForSuccess("error refreshing pool for volume", func() error {
		return virConn.StoragePoolRefresh(volPool, 0)
	})

	// Workaround for redhat#1293804
	// https://bugzilla.redhat.com/show_bug.cgi?id=1293804#c12
	// Does not solve the problem but it makes it happen less often.
	_, err = virConn.StorageVolGetXMLDesc(volume, 0)
	if err != nil {
		virErr := err.(libvirt.Error)
		if virErr.Code != uint32(libvirt.ErrNoStorageVol) {
			return fmt.Errorf("can't retrieve volume %s XML desc: %s", key, err)
		}
		// Volume is probably gone already, getting its XML description is pointless
	}

	err = virConn.StorageVolDelete(volume, 0)
	if err != nil {
		virErr := err.(libvirt.Error)
		if virErr.Code != uint32(libvirt.ErrNoStorageVol) {
			return fmt.Errorf("can't delete volume %s: %s", key, err)
		}
		// Volume is gone already
		return nil
	}

	return volumeWaitDeleted(client.libvirt, key)
}

// tries really hard to find volume with `key`
// it will try to start the pool if it does not find it
//
// You have to call volume.Free() on the returned volume
func volumeLookupReallyHard(client *Client, volPoolName string, key string) (*libvirt.StorageVol, error) {
	virConn := client.libvirt
	if virConn == nil {
		return nil, fmt.Errorf(LibVirtConIsNil)
	}

	volume, err := virConn.StorageVolLookupByKey(key)
	if err != nil {
		virErr := err.(libvirt.Error)
		if virErr.Code != uint32(libvirt.ErrNoStorageVol) {
			return nil, fmt.Errorf("can't retrieve volume %s", key)
		}
		log.Printf("[INFO] Volume %s not found, attempting to start its pool", key)

		volPool, err := virConn.StoragePoolLookupByName(volPoolName)
		if err != nil {
			return nil, fmt.Errorf("error retrieving pool %s for volume %s: %s", volPoolName, key, err)
		}

		active, err := virConn.StoragePoolIsActive(volPool)
		if err != nil {
			return nil, fmt.Errorf("error retrieving status of pool %s for volume %s: %s", volPoolName, key, err)
		}
		if active == 1 {
			log.Printf("can't retrieve volume %s (and pool is active)", key)
			return nil, nil
		}

		err = virConn.StoragePoolCreate(volPool, 0)
		if err != nil {
			return nil, fmt.Errorf("error starting pool %s: %s", volPoolName, err)
		}

		// attempt a new lookup
		volume, err = virConn.StorageVolLookupByKey(key)
		if err != nil {
			virErr := err.(libvirt.Error)
			if virErr.Code != uint32(libvirt.ErrNoStorageVol) {
				return nil, fmt.Errorf("can't retrieve volume %s", key)
			}
			// does not exist, but no error
			return nil, nil
		}
	}
	return &volume, nil
}
