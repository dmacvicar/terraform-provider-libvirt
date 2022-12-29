package main

import (
	"math/rand"
	"time"

	"github.com/dmacvicar/terraform-provider-libvirt/libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	defer libvirt.CleanupLibvirtConnections()

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: libvirt.Provider,
	})
}
