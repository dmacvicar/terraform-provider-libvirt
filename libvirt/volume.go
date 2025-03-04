package libvirt

import (
	"context"
	"fmt"
	"log"
	"time"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	volumeStateConfNotExists = resourceStateConfNotExists
	volumeStateConfExists    = resourceStateConfExists
	volumeStateConfError     = resourceStateConfError
	volumeStateConfPending   = resourceStateConfPending
	volumeStateConfDone      = resourceStateConfDone
)

func volumeExistsStateRefreshFunc(virConn *libvirt.Libvirt, key string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		_, err := virConn.StorageVolLookupByKey(key)
		if err != nil {
			if isError(err, libvirt.ErrNoStorageVol) {
				log.Printf("Volume %s does not exist", key)
				return virConn, resourceStateConfNotExists, nil
			}
		}
		return virConn, resourceStateConfExists, err
	}
}

func volumeResizeDoneStateRefreshFunc(virConn *libvirt.Libvirt, volume libvirt.StorageVol, targetSize uint64) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		_, capacity, _, err := virConn.StorageVolGetInfo(volume)
		if err != nil {
			return virConn, resourceStateConfError, fmt.Errorf("failed to query volume '%s' info: %w", volume.Name, err)
		}

		if capacity != targetSize {
			return virConn, resourceStateConfPending, nil
		}
		return virConn, resourceStateConfDone, nil
	}
}

func waitForStateVolumeExists(ctx context.Context, virConn *libvirt.Libvirt, key string) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{volumeStateConfNotExists},
		Target:     []string{volumeStateConfExists},
		Refresh:    volumeExistsStateRefreshFunc(virConn, key),
		Timeout:    resourceStateTimeout,
		MinTimeout: resourceStateMinTimeout,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return err
	}
	return nil
}

func waitForStateVolumeResizeDone(ctx context.Context, virConn *libvirt.Libvirt, volume libvirt.StorageVol, targetSize uint64) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{volumeStateConfPending},
		Target:     []string{volumeStateConfDone},
		Refresh:    volumeResizeDoneStateRefreshFunc(virConn, volume, targetSize),
		Timeout:    resourceStateTimeout,
		MinTimeout: 1 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return err
	}
	return nil
}

// volumeDelete removes the volume identified by `key` from libvirt.
func volumeDelete(ctx context.Context, client *Client, key string) error {
	virConn := client.libvirt

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

	if err := retry.RetryContext(ctx, resourceStateTimeout, func() *retry.RetryError {
		if err := virConn.StoragePoolRefresh(volPool, 0); err != nil {
			return retry.RetryableError(fmt.Errorf("error refreshing pool for volume: %w", err))
		}
		return nil
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

	stateConf := &retry.StateChangeConf{
		Pending:    []string{resourceStateConfExists},
		Target:     []string{resourceStateConfNotExists},
		Refresh:    volumeExistsStateRefreshFunc(virConn, key),
		Timeout:    resourceStateTimeout,
		MinTimeout: resourceStateMinTimeout,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return err
	}
	return nil
}

// volumeResizeCheck checks whether it is possible to increase the size of the provided volume by the given amount.
func volumeResizeCheck(client *Client, volume libvirt.StorageVol, pool libvirt.StoragePool, sizeIncrease uint64) error {
	virConn := client.libvirt

	state, _, _, poolAvailable, err := virConn.StoragePoolGetInfo(pool)
	if err != nil {
		return fmt.Errorf("error retrieving info for storage pool '%s' : %w", pool.Name, err)
	}

	if state != poolStateRunning {
		return fmt.Errorf("the storage pool '%s' is in an invalid state (%d) for resizing", pool.Name, state)
	}

	_, volumeCapacity, volumeAllocated, err := virConn.StorageVolGetInfo(volume)
	if err != nil {
		return fmt.Errorf("error retrieving info for volume '%s': %w", volume.Name, err)
	}
	log.Printf(
		"[DEBUG] '%s' volume capacity=%d allocated=%d - %s pool available=%d - requested size increase=%d",
		volume.Name, volumeCapacity, volumeAllocated, pool.Name, poolAvailable, sizeIncrease,
	)

	if sizeIncrease > poolAvailable {
		return fmt.Errorf("not enough available space for storage pool '%s' to resize volume %s", pool.Name, volume.Name)
	}

	return nil
}

// volumeResize increases the size of the volume identified by `key' from the old to the new provided size
func volumeResize(ctx context.Context, client *Client, key string, oldSize, newSize uint64) error {
	virConn := client.libvirt

	volume, err := virConn.StorageVolLookupByKey(key)
	if err != nil {
		return fmt.Errorf("volumeResize: Can't retrieve volume with key %s: %w", key, err)
	}

	pool, err := virConn.StoragePoolLookupByName(volume.Pool)
	if err != nil {
		return fmt.Errorf("volumeResize: Failed to retrieve volume's storage pool %s: %w", volume.Pool, err)
	}

	client.poolMutexKV.Lock(pool.Name)
	defer client.poolMutexKV.Unlock(pool.Name)

	sizeDelta := newSize - oldSize
	if err := volumeResizeCheck(client, volume, pool, sizeDelta); err != nil {
		return fmt.Errorf("volumeResize: Failed while determining if the volume %s can be resized: %w", volume.Name, err)
	}

	if err := virConn.StorageVolResize(volume, sizeDelta, libvirt.StorageVolResizeDelta); err != nil {
		return fmt.Errorf("volumeResize: Failed to resize volume %s: %w", volume.Name, err)
	}

	if err := waitForStateVolumeResizeDone(ctx, virConn, volume, newSize); err != nil {
		return err
	}
	log.Printf("[INFO] The volume %s has been resized. Filesystem expansion might be necessary", volume.Name)

	return nil
}
