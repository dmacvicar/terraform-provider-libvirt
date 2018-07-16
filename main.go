package main

import (
	"github.com/dmacvicar/terraform-provider-libvirt/libvirt"
	"github.com/hashicorp/terraform/plugin"
	"log"
	"math/rand"
	"time"
)

func main() {
	defer libvirt.CleanupLibvirtConnections()

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: libvirt.Provider,
	})
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
