package libvirt

import (
	"context"
	"log"
	"time"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	poolExistsID    = "EXISTS"
	poolNotExistsID = "NOT-EXISTS"
)

// poolExists returns "EXISTS" or "NOT-EXISTS" depending on the current pool existence.
func poolStateRefreshFunc(virConn *libvirt.Libvirt, uuid libvirt.UUID) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		_, err := virConn.StoragePoolLookupByUUID(uuid)
		if err != nil {
			if isError(err, libvirt.ErrNoStoragePool) {
				log.Printf("pool %s does not exist", uuid)
				return virConn, poolNotExistsID, nil
			}
			return virConn, poolNotExistsID, err
		}
		return virConn, poolExistsID, err
	}
}

// waitForStatePoolExists waits for a storage pool to be up and timeout after 5 minutes.
func waitForStatePoolExists(ctx context.Context, virConn *libvirt.Libvirt, uuid libvirt.UUID) error {
	log.Printf("Waiting for pool %s to appear...", uuid)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{poolNotExistsID},
		Target:     []string{poolExistsID},
		Refresh:    poolStateRefreshFunc(virConn, uuid),
		Timeout:    1 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return err
	}
	return nil
}

// waitForStatePoolDeleted waits for a storage pool to be removed.
func waitForStatePoolDeleted(ctx context.Context, virConn *libvirt.Libvirt, uuid libvirt.UUID) error {
	log.Printf("waiting for pool %s to be deleted...", uuid)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{poolExistsID},
		Target:     []string{poolNotExistsID},
		Refresh:    poolStateRefreshFunc(virConn, uuid),
		Timeout:    1 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return err
	}
	return nil
}
