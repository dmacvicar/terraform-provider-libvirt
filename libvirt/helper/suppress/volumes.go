package suppress

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"strconv"
)

// Suppress the diff if the specified volume size is not an even multiple of 1024 bytes
func Qcow2RoundOffset(_, reportedSize, newSize string, k *schema.ResourceData) bool {
	format := k.Get("format")

	// On first apply these are nil strings, so there's always a diff
	if reportedSize == "" || format == "" {
		return false
	}

	nSize, _ := strconv.ParseUint(newSize, 10, 64)
	rSize, _ := strconv.ParseUint(reportedSize, 10, 64)
	newSizeRounded := roundUp(nSize)

	// rounded size and reported size are the same
	// image is qcow2 format, so suppress the diff
	if newSizeRounded == rSize && format == "qcow2" {
		return true
	}

	return false
}

// roundUpNewSize will make sure an existing/old image size
// matches what qemu-img calculates by rounding up to the
// nearest multiple of 1024
func roundUp(i uint64) uint64 {
	if i%1024 != 0 {
		i = (i + 1024) - (i % 1024)
	}
	return i
}
