package libvirt

import (
	"fmt"
	"log"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/libvirt/libvirt-go-xml"
)

// getRoutesFromResource gets the libvirt network routes from a network definition
func getRoutesFromResource(d *schema.ResourceData) ([]libvirtxml.NetworkRoute, error) {
	routesCount, ok := d.GetOk("routes.#")
	if !ok {
		log.Printf("[INFO] No routes defined")
		return []libvirtxml.NetworkRoute{}, nil
	}

	routes := []libvirtxml.NetworkRoute{}
	log.Printf("[INFO] %d routes defined", routesCount)
	for i := 0; i < routesCount.(int); i++ {
		route := libvirtxml.NetworkRoute{}
		routePrefix := fmt.Sprintf("routes.%d", i)

		if cidr, ok := d.GetOk(routePrefix + ".cidr"); ok {
			addr, net, err := net.ParseCIDR(cidr.(string))
			if err != nil {
				return nil, fmt.Errorf("Error parsing static route in network: %s", err)
			}

			if addr.To4() == nil {
				route.Family = "ipv6"
			}

			route.Address = addr.String()

			ones, _ := net.Mask.Size()
			route.Prefix = (uint)(ones)
		} else {
			return nil, fmt.Errorf("no address defined for static route")
		}

		if gw, ok := d.GetOk(routePrefix + ".gateway"); ok {
			ip := net.ParseIP(gw.(string))
			if ip == nil {
				return nil, fmt.Errorf("Error parsing IP '%s' in static route in network", gw)
			}
			route.Gateway = ip.String()
		} else {
			return nil, fmt.Errorf("no gateway defined for static route")
		}

		routes = append(routes, route)
	}

	return routes, nil
}
