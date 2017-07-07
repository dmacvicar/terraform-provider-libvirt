package libvirt

import "github.com/libvirt/libvirt-go-xml"

// An interface definition, as returned/understood by libvirt
// (see https://libvirt.org/formatdomain.html#elementsNICS)
//
// Something like:
//   <interface type='network'>
//       <source network='default'/>
//   </interface>
//

var waitForLeases map[libvirtxml.DomainInterface]bool
