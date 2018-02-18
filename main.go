package main

import (
	"github.com/dmacvicar/terraform-provider-libvirt/libvirt"
	"github.com/hashicorp/terraform/plugin"
	"log"
	"math/rand"
	"time"
)

func main() {
	defer func() {
		if libvirt.LibvirtClient != nil {
			alive, err := libvirt.LibvirtClient.IsAlive()
			if err != nil {
				log.Printf("[ERROR] cannot determine libvirt connection status: %v", err)
			}
			if alive {
				ret, err := libvirt.LibvirtClient.Close()
				if err != nil {
					log.Printf("[ERROR] cannot close libvirt connection %d - %v", ret, err)
				} else {
					libvirt.LibvirtClient = nil
				}
			}
		}
	}()

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: libvirt.Provider,
	})
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
