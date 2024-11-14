package libvirt

import (
	"context"
	"log"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	poolStateConfExists    = resourceStateConfExists
	poolStateConfNotExists = resourceStateConfNotExists
)

func poolExistsStateRefreshFunc(virConn *libvirt.Libvirt, uuid libvirt.UUID) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		_, err := virConn.StoragePoolLookupByUUID(uuid)
		if err != nil {
			if isError(err, libvirt.ErrNoStoragePool) {
				log.Printf("pool %v does not exist", uuidString(uuid))
				return virConn, poolStateConfNotExists, nil
			}
			return virConn, poolStateConfNotExists, err
		}
		return virConn, poolStateConfExists, err
	}
}

func waitForStatePoolExists(ctx context.Context, virConn *libvirt.Libvirt, uuid libvirt.UUID) error {
	log.Printf("Waiting for pool %v to appear...", uuidString(uuid))
	stateConf := &retry.StateChangeConf{
		Pending:    []string{poolStateConfNotExists},
		Target:     []string{poolStateConfExists},
		Refresh:    poolExistsStateRefreshFunc(virConn, uuid),
		Timeout:    resourceStateTimeout,
		Delay:      resourceStateDelay,
		MinTimeout: resourceStateMinTimeout,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return err
	}
	return nil
}

// waitForStatePoolDeleted waits for a storage pool to be removed.
func waitForStatePoolDeleted(ctx context.Context, virConn *libvirt.Libvirt, uuid libvirt.UUID) error {
	log.Printf("waiting for pool %v to be deleted...", uuidString(uuid))
	stateConf := &retry.StateChangeConf{
		Pending:    []string{poolStateConfExists},
		Target:     []string{poolStateConfNotExists},
		Refresh:    poolExistsStateRefreshFunc(virConn, uuid),
		Timeout:    resourceStateTimeout,
		Delay:      resourceStateDelay,
		MinTimeout: resourceStateMinTimeout,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return err
	}
	return nil
}
