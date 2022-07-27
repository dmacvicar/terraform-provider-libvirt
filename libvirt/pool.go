package libvirt

import (
	"fmt"
	"log"
	"time"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	poolExistsID    = "EXISTS"
	poolNotExistsID = "NOT-EXISTS"
)

// poolExists returns "EXISTS" or "NOT-EXISTS" depending on the current pool existence
func poolExists(virConn *libvirt.Libvirt, uuid string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		_, err := virConn.StoragePoolLookupByUUID(parseUUID(uuid))
		if err != nil {
			if err.(libvirt.Error).Code == uint32(libvirt.ErrNoStoragePool) {
				log.Printf("Pool %s does not exist", uuid)
				return virConn, "NOT-EXISTS", nil
			}
			log.Printf("Pool %s: error: %s", uuid, err.(libvirt.Error).Message)
		}
		return virConn, poolExistsID, err
	}
}

// poolWaitForExists waits for a storage pool to be up and timeout after 5 minutes.
func poolWaitForExists(virConn *libvirt.Libvirt, uuid string) error {
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
func poolWaitDeleted(virConn *libvirt.Libvirt, uuid string) error {
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

	pool, err := virConn.StoragePoolLookupByUUID(parseUUID(uuid))
	if err != nil {
		return fmt.Errorf("error retrieving storage pool info: %s", err)
	}

	if pool.Name == "" {
		return fmt.Errorf("error retrieving storage pool name for uuid: %s", uuid)
	}

	client.poolMutexKV.Lock(pool.Name)
	defer client.poolMutexKV.Unlock(pool.Name)

	state, _, _, _, err := virConn.StoragePoolGetInfo(pool)
	if err != nil {
		return fmt.Errorf("error retrieving storage pool info: %s", err)
	}

	if state != uint8(libvirt.StoragePoolInactive) {
		err := virConn.StoragePoolDestroy(pool)
		if err != nil {
			return fmt.Errorf("error deleting storage pool: %s", err)
		}
	}

	poolDef, err := newDefPoolFromLibvirt(virConn, pool)
	if err != nil {
		return err
	}

	// if the logical pool has no source device then the volume group existed before we created the pool, so we don't delete it
	if poolDef.Type == "dir" || (poolDef.Type == "logical" && poolDef.Source != nil && poolDef.Source.Device != nil) {
		err = virConn.StoragePoolDelete(pool, 0)
		if err != nil {
			return fmt.Errorf("error deleting storage pool: %w", err)
		}
	}

	err = virConn.StoragePoolUndefine(pool)
	if err != nil {
		return fmt.Errorf("error deleting storage pool: %s", err)
	}

	return poolWaitDeleted(client.libvirt, uuid)
}
