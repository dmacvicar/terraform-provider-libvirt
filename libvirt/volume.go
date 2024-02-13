package libvirt

import (
	"context"
	"fmt"
	"log"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	volExistsID    = "EXISTS"
	volNotExistsID = "NOT-EXISTS"
)

// volumeExistsStateRefreshFunc returns "EXISTS" or "NOT-EXISTS" depending on the current volume existence.
func volumeExistsStateRefreshFunc(virConn *libvirt.Libvirt, key string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		_, err := virConn.StorageVolLookupByKey(key)
		if err != nil {
			if isError(err, libvirt.ErrNoStorageVol) {
				log.Printf("Volume %s does not exist", key)
				return virConn, volNotExistsID, nil
			}
		}
		return virConn, volExistsID, err
	}
}

// waitForStateVolumeExists waits for a storage volume to be up and timeout after 5 minutes.
func waitForStateVolumeExists(ctx context.Context, virConn *libvirt.Libvirt, key string) error {
	log.Printf("Waiting for volume %s to be active...", key)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{volNotExistsID},
		Target:     []string{volExistsID},
		Refresh:    volumeExistsStateRefreshFunc(virConn, key),
		Timeout:    resourceStateTimeout,
		MinTimeout: resourceStateMinTimeout,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return err
	}
	return nil
}

// volumeWaitDeleted waits for a storage volume to be removed.
func volumeWaitDeleted(ctx context.Context, virConn *libvirt.Libvirt, key string) error {
	log.Printf("Waiting for volume %s to be deleted...", key)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{volExistsID},
		Target:     []string{volNotExistsID},
		Refresh:    volumeExistsStateRefreshFunc(virConn, key),
		Timeout:    resourceStateTimeout,
		MinTimeout: resourceStateMinTimeout,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return err
	}
	return nil
}

// volumeDelete removes the volume identified by `key` from libvirt.
func volumeDelete(ctx context.Context, client *Client, key string) error {
	virConn := client.libvirt
	if virConn == nil {
		return fmt.Errorf(LibVirtConIsNil)
	}
	volume, err := virConn.StorageVolLookupByKey(key)
	if err != nil {
		if isError(err, libvirt.ErrNoStorageVol) {
			// Volume already deleted.
			return nil
		}
		return fmt.Errorf("volumeDelete: Can't retrieve volume %s: %w", key, err)
	}

	// Refresh the pool of the volume so that libvirt knows it is
	// no longer in use.
	volPool, err := virConn.StoragePoolLookupByVolume(volume)
	if err != nil {
		return fmt.Errorf("error retrieving pool for volume: %w", err)
	}

	if volPool.Name == "" {
		return fmt.Errorf("error retrieving name of pool for volume key: %s", volume.Key)
	}

	client.poolMutexKV.Lock(volPool.Name)
	defer client.poolMutexKV.Unlock(volPool.Name)

	if err := waitForSuccess("error refreshing pool for volume", func() error {
		return virConn.StoragePoolRefresh(volPool, 0)
	}); err != nil {
		return err
	}

	// Workaround for redhat#1293804
	// https://bugzilla.redhat.com/show_bug.cgi?id=1293804#c12
	// Does not solve the problem but it makes it happen less often.
	_, err = virConn.StorageVolGetXMLDesc(volume, 0)
	if err != nil {
		if !isError(err, libvirt.ErrNoStorageVol) {
			return fmt.Errorf("can't retrieve volume %s XML desc: %w", key, err)
		}
		// Volume is probably gone already, getting its XML description is pointless
	}

	err = virConn.StorageVolDelete(volume, 0)
	if err != nil {
		if !isError(err, libvirt.ErrNoStorageVol) {
			return fmt.Errorf("can't delete volume %s: %w", key, err)
		}
		// Volume is gone already
		return nil
	}

	return volumeWaitDeleted(ctx, client.libvirt, key)
}
