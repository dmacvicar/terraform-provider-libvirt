package libvirt

import (
	"context"
	"fmt"
	"log"

	libvirt "github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	volumeStateConfNotExists = resourceStateConfNotExists
	volumeStateConfExists    = resourceStateConfExists
)

// UnitsMap is used for converting storage size units from xml representation into bytes
// https://pkg.go.dev/github.com/libvirt/libvirt-go-xml#StorageVolumeSize
// https://libvirt.org/formatstorage.html#storage-volume-general-metadata
//nolint: mnd
var UnitsMap map[string]uint64 = map[string]uint64{
	"":      1,
	"B":     1,
	"bytes": 1,
	"KB":    1000,
	"K":     1024,
	"KiB":   1024,
	"MB":    1_000_000,
	"M":     1_048_576,
	"MiB":   1_048_576,
	"GB":    1_000_000_000,
	"G":     1_073_741_824,
	"GiB":   1_073_741_824,
	"TB":    1_000_000_000_000,
	"T":     1_099_511_627_776,
	"TiB":   1_099_511_627_776,
	"PB":    1_000_000_000_000_000,
	"P":     1_125_899_906_842_624,
	"PiB":   1_125_899_906_842_624,
	"EB":    1_000_000_000_000_000_000,
	"E":     1_152_921_504_606_846_976,
	"EiB":   1_152_921_504_606_846_976,
}

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

// volumeRequiresResize checks whether a volume needs resizing after being created with StorageVolCreateXMLFrom.
// StorageVolCreateXMLFrom may ignore requested volume capacity in some cases. For example when qcow2 is involved,
// libvirt clones the volume using `qemu-img convert` which creates a new volume with the same capacity as the original.
func volumeRequiresResize(
	virConn *libvirt.Libvirt,
	d *schema.ResourceData,
	volume,
	baseVolume libvirt.StorageVol,
	volumePool libvirt.StoragePool,
) (bool, diag.Diagnostics) {
	if !d.Get("base_volume_copy").(bool) {
		return false, nil
	}

	size := d.Get("size")
	if size == nil {
		return false, nil
	}

	volumeXML, err := newDefVolumeFromLibvirt(virConn, volume)
	if err != nil {
		return false, diag.Errorf("could not get volume '%s' xml definition: %s", volume.Name, err)
	}

	baseVolumeXML, err := newDefVolumeFromLibvirt(virConn, baseVolume)
	if err != nil {
		return false, diag.Errorf("could not get volume '%s' xml definition: %s", baseVolume.Name, err)
	}

	// do not resize in case allocation > requested size. Happens when there is substantial metadata overhead
	if volumeXML.Allocation == nil || size.(int) <= int(volumeXML.Allocation.Value*UnitsMap[volumeXML.Allocation.Unit]) {
		return false, nil
	}

	if baseVolumeXML.Capacity == nil || size.(int) <= int(baseVolumeXML.Capacity.Value*UnitsMap[baseVolumeXML.Capacity.Unit]) {
		return false, nil
	}

	if volumePoolXML, err := newDefPoolFromLibvirt(virConn, volumePool); err != nil {
		return false, err
	} else if volumePoolXML.Type != "dir" {
		return false, nil
	}

	if volumeXML.Target != nil && volumeXML.Target.Format != nil && baseVolumeXML.Target != nil && baseVolumeXML.Target.Format != nil {
		if volumeXML.Target.Format.Type == "qcow2" || baseVolumeXML.Target.Format.Type == "qcow2" {
			return true, nil
		}
	}

	return false, nil
}
