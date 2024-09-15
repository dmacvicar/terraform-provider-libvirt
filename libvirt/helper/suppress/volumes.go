package suppress

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
)

const qcow2SectorSize = 512

// Suppress the diff if the specified volume size is qemu-img round up to sector size.
func Qcow2SizeDiffSuppressFunc(_, oldSizeStr, newSizeStr string, k *schema.ResourceData) bool {
	if format := k.Get("format"); format != "qcow2" {
		return false
	}

	// On first apply these are nil strings, so there's always a diff.
	if oldSizeStr == "" {
		return false
	}

	oldSize, _ := strconv.ParseUint(oldSizeStr, 10, 64)
	newSize, _ := strconv.ParseUint(newSizeStr, 10, 64)

	return oldSize-newSize < qcow2SectorSize
}
