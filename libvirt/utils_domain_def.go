package libvirt

import (
	"encoding/xml"
	"fmt"
	"log"
	"strings"

	libvirt "github.com/libvirt/libvirt-go"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
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

func lookupMachine(machines []libvirtxml.CapsGuestMachine, targetmachine string) string {
	for _, machine := range machines {
		if machine.Name == targetmachine {
			if machine.Canonical != "" {
				return machine.Canonical
			}
			return machine.Name
		}
	}
	return ""
}

func getCanonicalMachineName(caps libvirtxml.Caps, arch string, virttype string, targetmachine string) (string, error) {
	guest, err := getGuestForArchType(caps, arch, virttype)
	if err != nil {
		return "", err
	}

	/* Machine entries can be in the guest.Arch.Machines level as well as
	   under each guest.Arch.Domains[].Machines */

	name := lookupMachine(guest.Arch.Machines, targetmachine)
	if name != "" {
		return name, nil
	}

	for _, domain := range guest.Arch.Domains {
		name := lookupMachine(domain.Machines, targetmachine)
		if name != "" {
			return name, nil
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
		if machine.Canonical != "" && machine.Canonical == targetmachine {
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
	if len(cmdLine) == 0 {
		return cmdLines, nil
	}

	currCmdLine := make(map[string]string)
	keylessCmdLineArgs := []string{}

	argVals := strings.Split(cmdLine, " ")
	for _, argVal := range argVals {
		if !strings.Contains(argVal, "=") {
			// keyless cmd line (eg: nosplash)
			keylessCmdLineArgs = append(keylessCmdLineArgs, argVal)
			continue
		}

		kv := strings.SplitN(argVal, "=", 2)
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
	if len(keylessCmdLineArgs) > 0 {
		cl := make(map[string]string)
		cl["_"] = strings.Join(keylessCmdLineArgs, " ")
		cmdLines = append(cmdLines, cl)
	}
	return cmdLines, nil
}

func getHostArchitecture(virConn *libvirt.Connect) (string, error) {
	type HostCapabilities struct {
		XMLName xml.Name `xml:"capabilities"`
		Host    struct {
			XMLName xml.Name `xml:"host"`
			CPU     struct {
				XMLName xml.Name `xml:"cpu"`
				Arch    string   `xml:"arch"`
			}
		}
	}

	info, err := virConn.GetCapabilities()
	if err != nil {
		return "", err
	}

	capabilities := HostCapabilities{}
	xml.Unmarshal([]byte(info), &capabilities)

	return capabilities.Host.CPU.Arch, nil
}

func getHostCapabilities(virConn *libvirt.Connect) (libvirtxml.Caps, error) {
	// We should perhaps think of storing this on the connect object
	// on first call to avoid the back and forth
	caps := libvirtxml.Caps{}
	capsXML, err := virConn.GetCapabilities()
	if err != nil {
		return caps, err
	}
	xml.Unmarshal([]byte(capsXML), &caps)
	log.Printf("[TRACE] Capabilities of host \n %+v", caps)
	return caps, nil
}
