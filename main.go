package main

import (
	"github.com/dmacvicar/terraform-provider-libvirt/libvirt"
	"github.com/hashicorp/terraform/plugin"
	"math/rand"
	"time"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: libvirt.Provider,
	})
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
