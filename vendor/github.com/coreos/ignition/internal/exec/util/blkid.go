// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build linux

package util

// #cgo LDFLAGS: -lblkid
// #include <stdlib.h>
// #include "blkid.h"
import "C"

import (
	"bytes"
	"fmt"
	"unsafe"
)

const (
	field_name_type  = "TYPE"
	field_name_uuid  = "UUID"
	field_name_label = "LABEL"
)

func FilesystemType(device string) (string, error) {
	return filesystemLookup(device, field_name_type)
}

func FilesystemUUID(device string) (string, error) {
	return filesystemLookup(device, field_name_uuid)
}

func FilesystemLabel(device string) (string, error) {
	return filesystemLookup(device, field_name_label)
}

func filesystemLookup(device string, fieldName string) (string, error) {
	var buf [256]byte

	cDevice := C.CString(device)
	defer C.free(unsafe.Pointer(cDevice))
	cFieldName := C.CString(fieldName)
	defer C.free(unsafe.Pointer(cFieldName))

	switch C.blkid_lookup(cDevice, cFieldName, (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(len(buf))) {
	case C.RESULT_OK:
		// trim off tailing NULLs
		return string(buf[:bytes.IndexByte(buf[:], 0)]), nil
	case C.RESULT_OPEN_FAILED:
		return "", fmt.Errorf("failed to open %q", device)
	case C.RESULT_PROBE_FAILED:
		// If the probe failed, there's no filesystem created yet on this device
		return "", nil
	case C.RESULT_LOOKUP_FAILED:
		return "", fmt.Errorf("failed to lookup filesystem type of %q", device)
	default:
		return "", fmt.Errorf("unknown error")
	}
}
