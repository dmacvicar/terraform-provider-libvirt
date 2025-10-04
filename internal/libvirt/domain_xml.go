package libvirt

import (
	"fmt"

	"libvirt.org/go/libvirtxml"
)

// MarshalDomainXML serializes a libvirtxml.Domain to XML string
func MarshalDomainXML(domain *libvirtxml.Domain) (string, error) {
	xmlBytes, err := domain.Marshal()
	if err != nil {
		return "", fmt.Errorf("failed to marshal domain XML: %w", err)
	}
	return string(xmlBytes), nil //nolint:unconvert // string() is needed for []byte to string
}

// UnmarshalDomainXML deserializes XML to libvirtxml.Domain
func UnmarshalDomainXML(data string) (*libvirtxml.Domain, error) {
	domain := &libvirtxml.Domain{}
	if err := domain.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal domain XML: %w", err)
	}
	return domain, nil
}
