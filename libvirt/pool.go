package libvirt

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	libvirt "github.com/libvirt/libvirt-go"
)

const (
	poolExistsID    = "EXISTS"
	poolNotExistsID = "NOT-EXISTS"
)

// poolExists returns "EXISTS" or "NOT-EXISTS" depending on the current pool existence
func poolExists(virConn *libvirt.Connect, uuid string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pool, err := virConn.LookupStoragePoolByUUIDString(uuid)
		if err != nil {
			if err.(libvirt.Error).Code == libvirt.ERR_NO_STORAGE_POOL {
				log.Printf("Pool %s does not exist", uuid)
				return virConn, "NOT-EXISTS", nil
			}
			log.Printf("Pool %s: error: %s", uuid, err.(libvirt.Error).Message)
		}
		if pool != nil {
			defer pool.Free()
		}
		return virConn, poolExistsID, err
	}
}

// poolWaitForExists waits for a storage pool to be up and timeout after 5 minutes.
func poolWaitForExists(virConn *libvirt.Connect, uuid string) error {
	log.Printf("Waiting for pool %s to be active...", uuid)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{poolNotExistsID},
		Target:     []string{poolExistsID},
		Refresh:    poolExists(virConn, uuid),
		Timeout:    1 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		log.Printf("%s", err)
		return fmt.Errorf("unexpected error during pool creation operation. The operation did not complete successfully")
	}
	return nil
}

// poolWaitDeleted waits for a storage pool to be removed
func poolWaitDeleted(virConn *libvirt.Connect, uuid string) error {
	log.Printf("Waiting for pool %s to be deleted...", uuid)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{poolExistsID},
		Target:     []string{poolNotExistsID},
		Refresh:    poolExists(virConn, uuid),
		Timeout:    1 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		log.Printf("%s", err)
		return fmt.Errorf("unexpected error during pool destroy operation. The pool was not deleted")
	}
	return nil
}

// deletePool deletes the pool identified by `uuid` from libvirt
func deletePool(client *Client, uuid string) error {
	virConn := client.libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}

	pool, err := virConn.LookupStoragePoolByUUIDString(uuid)
	if err != nil {
		return fmt.Errorf("error retrieving storage pool info: %s", err)
	}

	poolName, err := pool.GetName()
	if err != nil {
		return fmt.Errorf("error retrieving storage pool name: %s", err)
	}
	client.poolMutexKV.Lock(poolName)
	defer client.poolMutexKV.Unlock(poolName)

	info, err := pool.GetInfo()
	if err != nil {
		return fmt.Errorf("error retrieving storage pool info: %s", err)
	}

	if info.State != libvirt.STORAGE_POOL_INACTIVE {
		err := pool.Destroy()
		if err != nil {
			return fmt.Errorf("error deleting storage pool: %s", err)
		}
	}

	err = pool.Delete(0)
	if err != nil {
		return fmt.Errorf("error deleting storage pool: %s", err)
	}

	err = pool.Undefine()
	if err != nil {
		return fmt.Errorf("error deleting storage pool: %s", err)
	}

	return poolWaitDeleted(client.libvirt, uuid)
}
