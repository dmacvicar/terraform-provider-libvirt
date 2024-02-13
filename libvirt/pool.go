package libvirt

import (
	"context"
	"log"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	poolStateExists    = "EXISTS"
	poolStateNotExists = "NOT-EXISTS"
)

// poolExists returns "EXISTS" or "NOT-EXISTS" depending on the current pool existence.
func poolStateRefreshFunc(virConn *libvirt.Libvirt, uuid libvirt.UUID) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		_, err := virConn.StoragePoolLookupByUUID(uuid)
		if err != nil {
			if isError(err, libvirt.ErrNoStoragePool) {
				log.Printf("pool %s does not exist", uuid)
				return virConn, poolStateNotExists, nil
			}
			return virConn, poolStateNotExists, err
		}
		return virConn, poolStateExists, err
	}
}

// waitForStatePoolExists waits for a storage pool to be up and timeout after 5 minutes.
func waitForStatePoolExists(ctx context.Context, virConn *libvirt.Libvirt, uuid libvirt.UUID) error {
	log.Printf("Waiting for pool %s to appear...", uuid)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{poolStateNotExists},
		Target:     []string{poolStateExists},
		Refresh:    poolStateRefreshFunc(virConn, uuid),
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
	log.Printf("waiting for pool %s to be deleted...", uuid)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{poolStateExists},
		Target:     []string{poolStateNotExists},
		Refresh:    poolStateRefreshFunc(virConn, uuid),
		Timeout:    resourceStateTimeout,
		Delay:      resourceStateDelay,
		MinTimeout: resourceStateMinTimeout,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return err
	}
	return nil
}
