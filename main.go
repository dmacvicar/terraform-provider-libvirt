package main

import (
	"github.com/dmacvicar/terraform-provider-libvirt/libvirt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	defer libvirt.CleanupLibvirtConnections()

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: libvirt.Provider,
	})
}
