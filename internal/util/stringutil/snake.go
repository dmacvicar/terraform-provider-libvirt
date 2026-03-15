package stringutil

import (
	"strings"
	"unicode"
)

var snakeCaseAcronyms = map[string]string{
	"IPs":    "ips",
	"IPv6":   "ipv6",
	"IPv4":   "ipv4",
	"DNS":    "dns",
	"DHCP":   "dhcp",
	"MAC":    "mac",
	"UUID":   "uuid",
	"XML":    "xml",
	"HTTP":   "http",
	"HTTPS":  "https",
	"API":    "api",
	"URI":    "uri",
	"URL":    "url",
	"VLAN":   "vlan",
	"MTU":    "mtu",
	"TFTP":   "tftp",
	"NFS":    "nfs",
	"SCSI":   "scsi",
	"SATA":   "sata",
	"IDE":    "ide",
	"USB":    "usb",
	"PCI":    "pci",
	"VNC":    "vnc",
	"RDP":    "rdp",
	"VGA":    "vga",
	"CPU":    "cpu",
	"VCPU":   "vcpu",
	"RAM":    "ram",
	"ROM":    "rom",
	"BIOS":   "bios",
	"UEFI":   "uefi",
	"TPM":    "tpm",
	"RNG":    "rng",
	"WWN":    "wwn",
	"QEMU":   "qemu",
	"LXC":    "lxc",
	"BHyve":  "bhyve",
	"VMWare": "vmware",
}

// SnakeCase converts CamelCase to snake_case while keeping known acronyms intact.
func SnakeCase(s string) string {
	if s == "" {
		return s
	}

	if strings.HasSuffix(s, "s") {
		base := s[:len(s)-1]
		if snake, ok := snakeCaseAcronyms[base]; ok {
			return snake + "s"
		}
	}

	if snake, ok := snakeCaseAcronyms[s]; ok {
		return snake
	}

	for _, prefix := range []string{"VMWare", "BHyve", "QEMU", "LXC"} {
		if strings.HasPrefix(s, prefix) && len(s) > len(prefix) {
			rest := s[len(prefix):]
			runes := []rune(rest)
			if len(runes) > 0 && unicode.IsUpper(runes[0]) {
				return snakeCaseAcronyms[prefix] + "_" + SnakeCase(rest)
			}
		}
	}

	var b strings.Builder
	runes := []rune(s)

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if i > 0 && unicode.IsUpper(r) {
			prevLower := i > 0 && unicode.IsLower(runes[i-1])
			prevUpper := i > 0 && unicode.IsUpper(runes[i-1])
			nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])

			if prevLower {
				b.WriteRune('_')
			} else if prevUpper && nextLower && i > 1 {
				b.WriteRune('_')
			}
		}

		b.WriteRune(unicode.ToLower(r))
	}

	return b.String()
}
