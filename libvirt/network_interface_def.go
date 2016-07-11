package libvirt

import (
	"encoding/xml"
)

// An interface definition, as returned/understood by libvirt
// (see https://libvirt.org/formatdomain.html#elementsNICS)
//
// Something like:
//   <interface type='network'>
//       <source network='default'/>
//   </interface>
//
type defNetworkInterface struct {
	XMLName xml.Name `xml:"interface"`
	Type    string   `xml:"type,attr"`
	Mac     struct {
		Address string `xml:"address,attr"`
	} `xml:"mac"`
	Source struct {
		Network string `xml:"network,attr,omitempty"`
		Bridge  string `xml:"bridge,attr,omitempty"`
		Dev     string `xml:"dev,attr,omitempty"`
		Mode    string `xml:"mode,attr"`
	} `xml:"source"`
	Model struct {
		Type string `xml:"type,attr"`
	} `xml:"model"`
	waitForLease bool
}
