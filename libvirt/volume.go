package libvirt

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/libvirt/libvirt-go"
)

const (
	volExistsID    = "EXISTS"
	volNotExistsID = "NOT-EXISTS"
)

// volumeExists returns "EXISTS" or "NOT-EXISTS" depending on the current volume existence
func volumeExists(virConn *libvirt.Connect, key string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vol, err := virConn.LookupStorageVolByKey(key)
		if err != nil {
			if err.(libvirt.Error).Code == libvirt.ERR_NO_STORAGE_VOL {
				log.Printf("Volume %s does not exist", key)
				return virConn, "NOT-EXISTS", nil
			}
			log.Printf("Volume %s: error: %s", key, err.(libvirt.Error).Message)
		}
		defer vol.Free()
		return virConn, volExistsID, err
	}
}

// volumeWaitForExists waits for a storage volume to be up and timeout after 5 minutes.
func volumeWaitForExists(virConn *libvirt.Connect, key string) error {
	log.Printf("Waiting for volume %s to be active...", key)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{volNotExistsID},
		Target:     []string{volExistsID},
		Refresh:    volumeExists(virConn, key),
		Timeout:    1 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for volume to reach EXISTS state: %s", err)
	}
	return nil
}

// volumeWaitDeleted waits for a storage volume to be removed
func volumeWaitDeleted(virConn *libvirt.Connect, key string) error {
	log.Printf("Waiting for volume %s to be deleted...", key)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{volExistsID},
		Target:     []string{volNotExistsID},
		Refresh:    volumeExists(virConn, key),
		Timeout:    1 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for volume to reach NOT-EXISTS state: %s", err)
	}
	return nil
}

// volumeDelete removes the volume identified by `key` from libvirt
func volumeDelete(client *Client, key string) error {
	volume, err := client.libvirt.LookupStorageVolByKey(key)
	if err != nil {
		virErr := err.(libvirt.Error)
		if virErr.Code != libvirt.ERR_NO_STORAGE_VOL {
			return fmt.Errorf("volumeDelete: Can't retrieve volume %s: %v", key, err)
		}
		// Volume already deleted.
		return nil
	}
	defer volume.Free()

	// Refresh the pool of the volume so that libvirt knows it is
	// not longer in use.
	volPool, err := volume.LookupPoolByVolume()
	if err != nil {
		return fmt.Errorf("Error retrieving pool for volume: %s", err)
	}
	defer volPool.Free()

	poolName, err := volPool.GetName()
	if err != nil {
		return fmt.Errorf("Error retrieving name of volume: %s", err)
	}

	client.poolMutexKV.Lock(poolName)
	defer client.poolMutexKV.Unlock(poolName)

	waitForSuccess("Error refreshing pool for volume", func() error {
		return volPool.Refresh(0)
	})

	// Workaround for redhat#1293804
	// https://bugzilla.redhat.com/show_bug.cgi?id=1293804#c12
	// Does not solve the problem but it makes it happen less often.
	_, err = volume.GetXMLDesc(0)
	if err != nil {
		virErr := err.(libvirt.Error)
		if virErr.Code != libvirt.ERR_NO_STORAGE_VOL {
			return fmt.Errorf("Can't retrieve volume %s XML desc: %s", key, err)
		}
		// Volume is probably gone already, getting its XML description is pointless
	}

	err = volume.Delete(0)
	if err != nil {
		virErr := err.(libvirt.Error)
		if virErr.Code != libvirt.ERR_NO_STORAGE_VOL {
			return fmt.Errorf("Can't delete volume %s: %s", key, err)
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

	volume, err := virConn.LookupStorageVolByKey(key)
	if err != nil {
		virErr := err.(libvirt.Error)
		if virErr.Code != libvirt.ERR_NO_STORAGE_VOL {
			return nil, fmt.Errorf("Can't retrieve volume %s", key)
		}
		log.Printf("[INFO] Volume %s not found, attempting to start its pool", key)

		volPool, err := virConn.LookupStoragePoolByName(volPoolName)
		if err != nil {
			return nil, fmt.Errorf("Error retrieving pool %s for volume %s: %s", volPoolName, key, err)
		}
		defer volPool.Free()

		active, err := volPool.IsActive()
		if err != nil {
			return nil, fmt.Errorf("error retrieving status of pool %s for volume %s: %s", volPoolName, key, err)
		}
		if active {
			log.Printf("Can't retrieve volume %s (and pool is active)", key)
			return nil, nil
		}

		err = volPool.Create(0)
		if err != nil {
			return nil, fmt.Errorf("error starting pool %s: %s", volPoolName, err)
		}

		// attempt a new lookup
		volume, err = virConn.LookupStorageVolByKey(key)
		if err != nil {
			virErr := err.(libvirt.Error)
			if virErr.Code != libvirt.ERR_NO_STORAGE_VOL {
				return nil, fmt.Errorf("Can't retrieve volume %s", key)
			}
			// does not exist, but no error
			return nil, nil
		}
	}
	return volume, nil
}
