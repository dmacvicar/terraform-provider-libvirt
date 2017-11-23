package libvirt

import (
	"fmt"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
	"log"
	"strings"
)

func getGuestForArchType(caps libvirtxml.Caps, arch string, virttype string) (libvirtxml.CapsGuest, error) {
	for _, guest := range caps.Guests {
		log.Printf("[TRACE] Checking for %s/%s against %s/%s\n", arch, virttype, guest.Arch.Name, guest.OSType)
		if guest.Arch.Name == arch && guest.OSType == virttype {
			log.Printf("[DEBUG] Found %d machines in guest for %s/%s", len(guest.Arch.Machines), arch, virttype)
			return guest, nil
		}
	}
	return libvirtxml.CapsGuest{}, fmt.Errorf("[DEBUG] Could not find any guests for architecure type %s/%s", virttype, arch)
}

func getCanonicalMachineName(caps libvirtxml.Caps, arch string, virttype string, targetmachine string) (string, error) {
	guest, err := getGuestForArchType(caps, arch, virttype)
	if err != nil {
		return "", err
	}

	for _, machine := range guest.Arch.Machines {
		if machine.Name == targetmachine {
			if machine.Canonical != nil {
				return *machine.Canonical, nil
			}
			return machine.Name, nil
		}
	}
	return "", fmt.Errorf("[WARN] Cannot find machine type %s for %s/%s in %v", targetmachine, virttype, arch, caps)
}

func getOriginalMachineName(caps libvirtxml.Caps, arch string, virttype string, targetmachine string) (string, error) {
	guest, err := getGuestForArchType(caps, arch, virttype)
	if err != nil {
		return "", err
	}

	for _, machine := range guest.Arch.Machines {
		if machine.Canonical != nil && *machine.Canonical == targetmachine {
			return machine.Name, nil
		}
	}
	return targetmachine, nil // There wasn't a canonical mapping to this
}

// as kernal args allow duplicate keys, we use a list of maps
// we jump to a next map as soon as we find a duplicate
// key
func splitKernelCmdLine(cmdLine string) ([]map[string]string, error) {
	var cmdLines []map[string]string
	currCmdLine := make(map[string]string)
	argVals := strings.Split(cmdLine, " ")
	for _, argVal := range argVals {
		kv := strings.Split(argVal, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("Can't parse kernel command line: '%s'", cmdLine)
		}
		k, v := kv[0], kv[1]
		// if the key is duplicate, start a new map
		if _, ok := currCmdLine[k]; ok {
			cmdLines = append(cmdLines, currCmdLine)
			currCmdLine = make(map[string]string)
		}
		currCmdLine[k] = v
	}
	if len(currCmdLine) > 0 {
		cmdLines = append(cmdLines, currCmdLine)
	}
	return cmdLines, nil
}
