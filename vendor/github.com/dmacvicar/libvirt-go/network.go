package libvirt

/*
#cgo LDFLAGS: -lvirt
#include <libvirt/libvirt.h>
#include <libvirt/virterror.h>
#include <stdlib.h>
*/
import "C"

import (
	"unsafe"
)

const (
	VIR_NETWORK_UPDATE_COMMAND_NONE      = C.VIR_NETWORK_UPDATE_COMMAND_NONE
	VIR_NETWORK_UPDATE_COMMAND_MODIFY    = C.VIR_NETWORK_UPDATE_COMMAND_MODIFY
	VIR_NETWORK_UPDATE_COMMAND_DELETE    = C.VIR_NETWORK_UPDATE_COMMAND_DELETE
	VIR_NETWORK_UPDATE_COMMAND_ADD_LAST  = C.VIR_NETWORK_UPDATE_COMMAND_ADD_LAST
	VIR_NETWORK_UPDATE_COMMAND_ADD_FIRST = C.VIR_NETWORK_UPDATE_COMMAND_ADD_FIRST
)

const (
	VIR_NETWORK_UPDATE_AFFECT_CURRENT = C.VIR_NETWORK_UPDATE_AFFECT_CURRENT
	VIR_NETWORK_UPDATE_AFFECT_LIVE    = C.VIR_NETWORK_UPDATE_AFFECT_LIVE
	VIR_NETWORK_UPDATE_AFFECT_CONFIG  = C.VIR_NETWORK_UPDATE_AFFECT_CONFIG
)

const (
	VIR_NETWORK_SECTION_NONE              = C.VIR_NETWORK_SECTION_NONE
	VIR_NETWORK_SECTION_BRIDGE            = C.VIR_NETWORK_SECTION_BRIDGE
	VIR_NETWORK_SECTION_DOMAIN            = C.VIR_NETWORK_SECTION_DOMAIN
	VIR_NETWORK_SECTION_IP                = C.VIR_NETWORK_SECTION_IP
	VIR_NETWORK_SECTION_IP_DHCP_HOST      = C.VIR_NETWORK_SECTION_IP_DHCP_HOST
	VIR_NETWORK_SECTION_IP_DHCP_RANGE     = C.VIR_NETWORK_SECTION_IP_DHCP_RANGE
	VIR_NETWORK_SECTION_FORWARD           = C.VIR_NETWORK_SECTION_FORWARD
	VIR_NETWORK_SECTION_FORWARD_INTERFACE = C.VIR_NETWORK_SECTION_FORWARD_INTERFACE
	VIR_NETWORK_SECTION_FORWARD_PF        = C.VIR_NETWORK_SECTION_FORWARD_PF
	VIR_NETWORK_SECTION_PORTGROUP         = C.VIR_NETWORK_SECTION_PORTGROUP
	VIR_NETWORK_SECTION_DNS_HOST          = C.VIR_NETWORK_SECTION_DNS_HOST
	VIR_NETWORK_SECTION_DNS_TXT           = C.VIR_NETWORK_SECTION_DNS_TXT
	VIR_NETWORK_SECTION_DNS_SRV           = C.VIR_NETWORK_SECTION_DNS_SRV
)

type VirNetwork struct {
	ptr C.virNetworkPtr
}

func (n *VirNetwork) Free() error {
	if result := C.virNetworkFree(n.ptr); result != 0 {
		return GetLastError()
	}
	return nil
}

func (n *VirNetwork) Create() error {
	result := C.virNetworkCreate(n.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (n *VirNetwork) Destroy() error {
	result := C.virNetworkDestroy(n.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (n *VirNetwork) IsActive() (bool, error) {
	result := C.virNetworkIsActive(n.ptr)
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (n *VirNetwork) IsPersistent() (bool, error) {
	result := C.virNetworkIsPersistent(n.ptr)
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (n *VirNetwork) GetAutostart() (bool, error) {
	var out C.int
	result := C.virNetworkGetAutostart(n.ptr, (*C.int)(unsafe.Pointer(&out)))
	if result == -1 {
		return false, GetLastError()
	}
	switch out {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}

func (n *VirNetwork) SetAutostart(autostart bool) error {
	var cAutostart C.int
	switch autostart {
	case true:
		cAutostart = 1
	default:
		cAutostart = 0
	}
	result := C.virNetworkSetAutostart(n.ptr, cAutostart)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (n *VirNetwork) GetName() (string, error) {
	name := C.virNetworkGetName(n.ptr)
	if name == nil {
		return "", GetLastError()
	}
	return C.GoString(name), nil
}

func (n *VirNetwork) GetUUID() ([]byte, error) {
	var cUuid [C.VIR_UUID_BUFLEN](byte)
	cuidPtr := unsafe.Pointer(&cUuid)
	result := C.virNetworkGetUUID(n.ptr, (*C.uchar)(cuidPtr))
	if result != 0 {
		return []byte{}, GetLastError()
	}
	return C.GoBytes(cuidPtr, C.VIR_UUID_BUFLEN), nil
}

func (n *VirNetwork) GetUUIDString() (string, error) {
	var cUuid [C.VIR_UUID_STRING_BUFLEN](C.char)
	cuidPtr := unsafe.Pointer(&cUuid)
	result := C.virNetworkGetUUIDString(n.ptr, (*C.char)(cuidPtr))
	if result != 0 {
		return "", GetLastError()
	}
	return C.GoString((*C.char)(cuidPtr)), nil
}

func (n *VirNetwork) GetBridgeName() (string, error) {
	result := C.virNetworkGetBridgeName(n.ptr)
	if result == nil {
		return "", GetLastError()
	}
	bridge := C.GoString(result)
	C.free(unsafe.Pointer(result))
	return bridge, nil
}

func (n *VirNetwork) GetXMLDesc(flags uint32) (string, error) {
	result := C.virNetworkGetXMLDesc(n.ptr, C.uint(flags))
	if result == nil {
		return "", GetLastError()
	}
	xml := C.GoString(result)
	C.free(unsafe.Pointer(result))
	return xml, nil
}

func (n *VirNetwork) UpdateXMLDesc(xmldesc string, command, section int) error {
	xmldescC := C.CString(xmldesc)
	result := C.virNetworkUpdate(n.ptr, C.uint(command), C.uint(section), C.int(-1), xmldescC, C.uint(C.VIR_NETWORK_UPDATE_AFFECT_CURRENT))
	C.free(unsafe.Pointer(xmldescC))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (n *VirNetwork) Undefine() error {
	result := C.virNetworkUndefine(n.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}
